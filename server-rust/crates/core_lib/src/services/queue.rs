//! 异步图片处理队列
//!
//! 基于 `tokio::sync::mpsc` 的内存队列，支持后台 worker 并发处理
//! 缩略图生成、水印添加等 CPU 密集型任务。
//!
//! 架构：
//! - `ImageQueue`：发送端句柄，用于投递任务
//! - `ImageTask`：任务枚举（缩略图、水印）
//! - worker 循环：消费任务并在 `spawn_blocking` 中执行 CPU 密集操作

use std::path::PathBuf;
use std::sync::Arc;

use bytes::Bytes;
use tokio::sync::mpsc;
use tokio::sync::Mutex;

use crate::config::WatermarkConfig;
use crate::services::storage_driver::StorageDriver;

/// 图片处理任务
///
/// `data` 携带原图字节（避免依赖本地临时文件）；
/// `driver` 为远端存储驱动（local 时为 None，产物直接写本地磁盘）。
#[derive(Clone)]
pub enum ImageTask {
    /// 生成缩略图
    Thumbnail {
        photo_id: i64,
        data: Vec<u8>,
        date_dir: String,
        stored_filename: String,
        upload_root: String,
        orig_width: u32,
        orig_height: u32,
        public_url: String,
        driver: Option<Arc<dyn StorageDriver>>,
    },
    /// 添加水印
    Watermark {
        photo_id: i64,
        data: Vec<u8>,
        date_dir: String,
        stored_filename: String,
        upload_root: String,
        public_url: String,
        watermark_text: String,
        config: WatermarkConfig,
        driver: Option<Arc<dyn StorageDriver>>,
    },
}

impl std::fmt::Debug for ImageTask {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        // 不打印驱动实例与图片数据，仅输出任务概要
        match self {
            ImageTask::Thumbnail { photo_id, .. } => f
                .debug_struct("ImageTask::Thumbnail")
                .field("photo_id", photo_id)
                .finish(),
            ImageTask::Watermark { photo_id, .. } => f
                .debug_struct("ImageTask::Watermark")
                .field("photo_id", photo_id)
                .finish(),
        }
    }
}

/// 图片处理队列（发送端）
#[derive(Clone)]
pub struct ImageQueue {
    sender: mpsc::Sender<ImageTask>,
    #[allow(dead_code)]
    pool: sqlx::SqlitePool,
}

impl ImageQueue {
    /// 创建队列并启动后台 workers
    ///
    /// `worker_count` 指定并发 worker 数量，建议等于 CPU 核心数。
    pub fn new(pool: sqlx::SqlitePool, worker_count: usize) -> Self {
        let (sender, receiver) = mpsc::channel::<ImageTask>(1024);
        let receiver = Arc::new(Mutex::new(receiver));

        for i in 0..worker_count.max(1) {
            let rx = receiver.clone();
            let pool_clone = pool.clone();
            tokio::spawn(async move {
                worker_loop(i, rx, pool_clone).await;
            });
        }

        Self { sender, pool }
    }

    /// 创建未初始化的队列（用于测试/禁用场景）
    pub fn none(pool: sqlx::SqlitePool) -> Option<Self> {
        // 仍然创建一个单 worker 队列，但不对外暴露
        Some(Self::new(pool, 1))
    }

    /// 投递一个任务到队列
    pub async fn enqueue(&self, task: ImageTask) -> crate::AppResult<()> {
        self.sender
            .send(task)
            .await
            .map_err(|_| crate::error::AppError::Internal("图片队列已关闭".to_string()))?;
        Ok(())
    }
}

/// Worker 循环：持续接收并处理任务
async fn worker_loop(
    id: usize,
    receiver: Arc<Mutex<mpsc::Receiver<ImageTask>>>,
    pool: sqlx::SqlitePool,
) {
    tracing::info!("🖼️ 图片处理 worker #{} 已启动", id);
    loop {
        let task = {
            let mut rx = receiver.lock().await;
            rx.recv().await
        };

        match task {
            Some(t) => {
                if let Err(e) = process_task(&t, &pool).await {
                    tracing::error!(worker = id, error = %e, "图片处理任务失败");
                }
            }
            None => {
                tracing::info!("🖼️ 图片处理 worker #{} 退出（队列关闭）", id);
                break;
            }
        }
    }
}

/// 处理单个任务
async fn process_task(task: &ImageTask, pool: &sqlx::SqlitePool) -> crate::AppResult<()> {
    match task {
        ImageTask::Thumbnail {
            photo_id,
            data,
            date_dir,
            stored_filename,
            upload_root,
            orig_width,
            orig_height,
            public_url,
            driver,
        } => {
            let start = std::time::Instant::now();

            // 在 spawn_blocking 中执行 CPU 密集型缩略图生成（内存中编码，不落盘）
            let thumb_bytes = tokio::task::spawn_blocking({
                let data = data.clone();
                let filename = stored_filename.clone();
                let w = *orig_width;
                let h = *orig_height;
                move || generate_thumbnail_blocking(&data, &filename, w, h)
            })
            .await
            .map_err(|e| crate::error::AppError::Internal(format!("缩略图任务 panic: {}", e)))?;

            if let Some(thumb_bytes) = thumb_bytes {
                let thumb_filename = format!("thumb_{}", stored_filename);
                let thumb_key = format!("{}/{}", date_dir, thumb_filename);
                let content_type = thumbnail_content_type(stored_filename);

                // 缩略图与主图走同一驱动：远端上传，本地写盘
                let thumb_url = match driver {
                    Some(d) => {
                        let result = d
                            .upload(&thumb_key, Bytes::from(thumb_bytes), content_type)
                            .await?;
                        result.url
                    }
                    None => {
                        let dir = PathBuf::from(upload_root).join(date_dir);
                        tokio::fs::create_dir_all(&dir).await.map_err(|e| {
                            crate::error::AppError::Internal(format!("创建目录失败: {}", e))
                        })?;
                        tokio::fs::write(dir.join(&thumb_filename), &thumb_bytes)
                            .await
                            .map_err(|e| {
                                crate::error::AppError::Internal(format!("保存缩略图失败: {}", e))
                            })?;
                        format!(
                            "{}/{}",
                            public_url.trim_end_matches('/'),
                            encode_relative_path(&thumb_key)
                        )
                    }
                };

                // 更新数据库记录
                sqlx::query("UPDATE photos SET thumbnail_url = ?, updated_at = datetime('now') WHERE id = ?")
                    .bind(&thumb_url)
                    .bind(photo_id)
                    .execute(pool)
                    .await
                    .map_err(|e| crate::error::AppError::Internal(format!("更新缩略图 URL 失败: {}", e)))?;

                tracing::debug!(
                    photo_id = photo_id,
                    elapsed = ?start.elapsed(),
                    "缩略图生成完成"
                );
            }

            Ok(())
        }
        ImageTask::Watermark {
            photo_id,
            data,
            date_dir,
            stored_filename,
            upload_root,
            public_url,
            watermark_text,
            config,
            driver,
        } => {
            let start = std::time::Instant::now();

            // 在 spawn_blocking 中执行 CPU 密集型水印添加（需克隆为所有权数据跨线程）
            let config = config.clone();
            let wm_text = watermark_text.clone();
            let data = data.clone();
            let wm_data = tokio::task::spawn_blocking(move || {
                apply_watermark_blocking(&data, &wm_text, &config)
            })
            .await
            .map_err(|e| crate::error::AppError::Internal(format!("水印任务 panic: {}", e)))?;

            if let Some(wm_bytes) = wm_data {
                // 生成水印文件名：wm_{original_filename}
                let wm_filename = format!("wm_{}", stored_filename);
                let wm_key = format!("{}/{}", date_dir, wm_filename);

                // 水印产物与主图走同一驱动
                let wm_url = match driver {
                    Some(d) => {
                        let result = d
                            .upload(&wm_key, Bytes::from(wm_bytes), "image/png")
                            .await?;
                        result.url
                    }
                    None => {
                        let dir = PathBuf::from(upload_root).join(date_dir);
                        tokio::fs::create_dir_all(&dir).await.map_err(|e| {
                            crate::error::AppError::Internal(format!("创建目录失败: {}", e))
                        })?;
                        tokio::fs::write(dir.join(&wm_filename), &wm_bytes)
                            .await
                            .map_err(|e| {
                                crate::error::AppError::Internal(format!("保存水印图片失败: {}", e))
                            })?;
                        format!(
                            "{}/{}",
                            public_url.trim_end_matches('/'),
                            encode_relative_path(&wm_key)
                        )
                    }
                };

                // 更新数据库记录
                sqlx::query("UPDATE photos SET watermark_url = ?, updated_at = datetime('now') WHERE id = ?")
                    .bind(&wm_url)
                    .bind(photo_id)
                    .execute(pool)
                    .await
                    .map_err(|e| crate::error::AppError::Internal(format!("更新水印 URL 失败: {}", e)))?;

                tracing::debug!(
                    photo_id = photo_id,
                    elapsed = ?start.elapsed(),
                    "水印添加完成"
                );
            } else {
                tracing::warn!(photo_id, "水印生成返回空（可能字体未配置）");
            }

            Ok(())
        }
    }
}

/// 将相对路径的各段做 URL 编码（保留 `/`），用于拼接公开 URL
fn encode_relative_path(rel: &str) -> String {
    rel.split('/')
        .map(urlencoding::encode)
        .collect::<Vec<_>>()
        .join("/")
}

/// 根据原文件扩展名推断缩略图的 Content-Type / 编码格式
fn thumbnail_content_type(stored_filename: &str) -> &'static str {
    let ext = std::path::Path::new(stored_filename)
        .extension()
        .and_then(|e| e.to_str())
        .unwrap_or("")
        .to_lowercase();
    match ext.as_str() {
        "jpg" | "jpeg" => "image/jpeg",
        _ => "image/png",
    }
}

/// 同步生成缩略图（在 spawn_blocking 中调用）
///
/// 在内存中完成缩放与编码，返回缩略图字节；图片小于阈值时返回 None。
fn generate_thumbnail_blocking(
    image_data: &[u8],
    original_filename: &str,
    orig_width: u32,
    orig_height: u32,
) -> Option<Vec<u8>> {
    use image::imageops::FilterType;

    const THUMB_MAX: u32 = 300;

    if orig_width <= THUMB_MAX && orig_height <= THUMB_MAX {
        return None;
    }

    let (thumb_w, thumb_h) = if orig_width > orig_height {
        let w = THUMB_MAX;
        let h = (orig_height * THUMB_MAX) / orig_width;
        (w, h)
    } else {
        let h = THUMB_MAX;
        let w = (orig_width * THUMB_MAX) / orig_height;
        (w, h)
    };

    let img = image::load_from_memory(image_data).ok()?;
    let thumbnail = img.resize(thumb_w, thumb_h, FilterType::Lanczos3);

    // 按原扩展名选择编码格式（JPEG 原图编码为 JPEG，其余统一 PNG）
    let ext = std::path::Path::new(original_filename)
        .extension()
        .and_then(|e| e.to_str())
        .unwrap_or("")
        .to_lowercase();
    let mut output: Vec<u8> = Vec::new();
    if ext == "jpg" || ext == "jpeg" {
        image::DynamicImage::write_to(
            &thumbnail,
            &mut std::io::Cursor::new(&mut output),
            image::ImageFormat::Jpeg,
        )
        .ok()?;
    } else {
        image::DynamicImage::write_to(
            &thumbnail,
            &mut std::io::Cursor::new(&mut output),
            image::ImageFormat::Png,
        )
        .ok()?;
    }
    Some(output)
}

/// 同步添加水印（在 spawn_blocking 中调用）
///
/// 在图片上绘制文字水印，支持 9 宫格位置和透明度。
/// 返回水印图片的 PNG 字节数据，若字体未配置则返回 None。
fn apply_watermark_blocking(
    image_data: &[u8],
    watermark_text: &str,
    config: &WatermarkConfig,
) -> Option<Vec<u8>> {
    use image::GenericImageView;
    use image::Rgba;
    use imageproc::drawing::draw_text_mut;

    // 加载原始图片
    let mut img = image::load_from_memory(image_data).ok()?;
    let (img_width, img_height) = img.dimensions();

    // 加载字体
    let font = load_watermark_font(config)?;
    let font_scale = ab_glyph::PxScale {
        x: config.font_size,
        y: config.font_size,
    };

    // 估算文字尺寸（简单估算：字符数 * 字体大小 * 0.6）
    let text_width = (watermark_text.len() as f32 * config.font_size * 0.6) as u32;
    let text_height = config.font_size as u32;

    // 计算边距（图片尺寸的 2%，最小 10px）
    let margin_x = std::cmp::max(10, img_width / 50);
    let margin_y = std::cmp::max(10, img_height / 50);

    // 根据 position 计算文字左上角坐标（支持 9 宫格）
    let (x, y) = compute_watermark_position(
        config.position.as_str(),
        img_width,
        img_height,
        text_width,
        text_height,
        margin_x,
        margin_y,
    );

    // 文字颜色：白色 + 配置的透明度
    let text_color = Rgba([255u8, 255u8, 255u8, config.opacity]);

    // 在图片上直接绘制文字
    draw_text_mut(
        &mut img,
        text_color,
        x as i32,
        y as i32,
        font_scale,
        &font,
        watermark_text,
    );

    // 编码为 PNG 字节
    let mut output_data: Vec<u8> = Vec::new();
    let mut cursor = std::io::Cursor::new(&mut output_data);
    img.write_to(&mut cursor, image::ImageFormat::Png).ok()?;

    Some(output_data)
}

/// 计算水印文字位置（9 宫格）
///
/// 支持的位置：
/// - top-left, top-center, top-right
/// - middle-left, center, middle-right
/// - bottom-left, bottom-center, bottom-right
#[allow(clippy::too_many_arguments)]
fn compute_watermark_position(
    position: &str,
    img_width: u32,
    img_height: u32,
    text_width: u32,
    text_height: u32,
    margin_x: u32,
    margin_y: u32,
) -> (u32, u32) {
    let x = match position {
        "top-left" | "middle-left" | "bottom-left" => margin_x,
        "top-center" | "center" | "bottom-center" => img_width.saturating_sub(text_width) / 2,
        // top-right, middle-right, bottom-right, 其他默认
        _ => img_width.saturating_sub(text_width + margin_x),
    };

    let y = match position {
        "top-left" | "top-center" | "top-right" => margin_y,
        "middle-left" | "center" | "middle-right" => img_height.saturating_sub(text_height) / 2,
        // bottom-left, bottom-center, bottom-right, 其他默认
        _ => img_height.saturating_sub(text_height + margin_y),
    };

    (x, y)
}

/// 加载水印字体
///
/// 优先使用 config.font_path 指定的字体文件，
/// 否则尝试常见的系统字体路径。
/// 如果都失败，返回 None（水印功能降级为不生效）。
///
/// 返回 `FontRef<'static>`，字体数据通过 `Box::leak` 转为静态生命周期
/// （字体在程序运行期间始终存在，无需释放）。
fn load_watermark_font(config: &WatermarkConfig) -> Option<ab_glyph::FontRef<'static>> {
    // 收集所有候选字体路径
    let mut candidates: Vec<String> = Vec::new();

    // 1. 用户配置的字体路径优先级最高
    if let Some(font_path) = &config.font_path {
        candidates.push(font_path.clone());
    }

    // 2. 常见系统字体路径
    if cfg!(windows) {
        candidates.extend([
            "C:/Windows/Fonts/msyh.ttc".to_string(),   // 微软雅黑
            "C:/Windows/Fonts/msyhbd.ttc".to_string(), // 微软雅黑 Bold
            "C:/Windows/Fonts/simhei.ttf".to_string(), // 黑体
            "C:/Windows/Fonts/simsun.ttc".to_string(), // 宋体
            "C:/Windows/Fonts/arial.ttf".to_string(),  // Arial (fallback)
        ]);
    } else if cfg!(target_os = "macos") {
        candidates.extend([
            "/System/Library/Fonts/PingFang.ttc".to_string(), // 苹方
            "/System/Library/Fonts/STHeiti Light.ttc".to_string(), // 黑体
            "/Library/Fonts/Arial Unicode.ttf".to_string(),   // Arial Unicode
        ]);
    } else {
        // Linux
        candidates.extend([
            "/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc".to_string(),
            "/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc".to_string(),
            "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf".to_string(),
            "/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf".to_string(),
        ]);
    }

    // 尝试加载每个候选字体
    for font_path in &candidates {
        if let Ok(font_data) = std::fs::read(font_path) {
            // 将字体数据 leak 为静态生命周期（字体在整个程序生命周期内有效）
            let static_data: &'static [u8] = Box::leak(font_data.into_boxed_slice());
            match ab_glyph::FontRef::try_from_slice(static_data) {
                Ok(font) => {
                    tracing::info!("使用字体: {}", font_path);
                    return Some(font);
                }
                Err(_) => continue,
            }
        }
    }

    // 无可用字体，降级
    tracing::warn!("未找到可用字体，水印功能不可用。请配置 watermark.font_path 或安装系统字体。");
    None
}
