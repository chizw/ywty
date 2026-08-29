# ywty（云雾图驿）

自托管图床 / 云上相册系统，Go 1.26 实现的单二进制服务。

```
Docker 镜像：node 构建 web + admin → golang 构建 server-go → alpine(+vips) 运行
            （CGO_ENABLED=0 纯静态单二进制，SQLite/MySQL 双方言）
```

## 架构

| 目录 | 说明 |
| --- | --- |
| `server-go/` | Go 单二进制：HTTP 服务 + 数据库队列 worker + 调度 |
| `web/` | 用户端 Vue 3 SPA（构建产物由 Go 托管） |
| `admin/` | 管理后台 Vue 3 SPA（Vue + Naive UI，挂 `/admin`） |
| `server-go/internal/db/migrations/` | SQLite/MySQL 双方言 schema（37 张表） |

## 功能

- 上传管线：格式/大小/容量/频率校验、`{Ymd}/{md5}` 等命名规则、内容指纹去重
- 图像处理：缩略图、缩放/裁剪/旋转、文字与图片水印（九宫格/平铺/透明度）、可选 vips 外挂 heic/avif/webp
- 相册 / 标签 / 分享（密码、过期、slug）/ 图片广场
- 用户体系：注册登录、会话与访问令牌（细粒度接口权限）、邮箱/手机绑定、工单
- 商业化：套餐 / 订单 / 优惠码 / 容量与角色组发放、EPay 支付（RSA-SHA256）与回调
- 管理后台：仪表盘、用户、图片审核、公告/页面、套餐/订单/优惠码、工单、设置、储存与驱动管理
- 储存驱动：本地、S3 兼容（AWS/MinIO/OSS/COS/R2）、WebDAV
- 旧版 1.x 接口兼容（PicGo 等客户端）

## 快速开始（Docker）

```bash
docker compose up -d --build
# 首次启动在 environment 中提供：
#   APP_URL / APP_LICENSE_KEY / ADMIN_USERNAME / ADMIN_EMAIL / ADMIN_PASSWORD
# 即自动完成安装；或调用 POST /api/v2/install
```

- 服务端口：`PORT`（默认 3000），健康检查 `/healthz`
- 用户端：`/`，管理后台：`/admin`
- 数据卷：`./data`（SQLite + installed.lock）、`./uploads`（本地储存）

### 环境变量

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `HOST` / `PORT` | 127.0.0.1 / 3000 | 监听地址 |
| `APP_URL` | http://localhost | 站点 URL（生成图片链接等） |
| `APP_NAME` | ywty | 站点名称 |
| `DB_DRIVER` | sqlite | `sqlite` / `mysql`（MariaDB ≥ 10.6） |
| `DB_PATH` | `{DATA_DIR}/ywty.db` | SQLite 文件路径 |
| `DB_HOST/PORT/USER/PASSWORD/NAME` | — | MySQL 连接 |
| `DATA_DIR` | data | 数据目录 |
| `UPLOADS_DIR` | uploads | 本地储存兜底根目录 |
| `STATIC_DIR` | web/dist | 前端主题目录 |
| `ADMIN_STATIC_DIR` | admin/dist | 管理后台目录 |
| `APP_LICENSE_KEY` + `ADMIN_*` | — | 环境变量一键安装 |
| `REDIS_ADDR/PASSWORD` | — | 预留（当前使用 DB cache/queue） |

## 响应契约

统一 envelope：`{"status": "success"|"error", "message": "...", "data": ..., "time": unix秒}`。
业务错误 HTTP 状态码默认 200，校验错误 422，未认证 401，限流 429。

## 本地开发

```bash
# Go 后端
cd server-go && go test ./... && go run ./cmd/ywty

# 用户端主题（产物 web/dist）
cd web && npm install && npm run build

# 管理后台（产物 admin/dist）
cd admin && npm install && npm run build
```

## 致谢

- [Lsky Pro](https://www.lsky.pro) —— 本项目的数据库结构、接口契约与产品形态参考自 Lsky Pro 社区，特此致谢。
