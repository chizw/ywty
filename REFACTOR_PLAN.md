# ywty 重构工程任务文档

> 项目代号：**ywty-refactor**
> 文档版本：v1.0
> 目标产物：Go 后端 + Nuxt 3 前端 + MySQL/SQLite/PostgreSQL 三驱动

---

## 1. 背景与目标

### 1.1 现有项目
- **ywty** 自托管图床 / 云相册
- 后端：Laravel 11 (PHP 8.2) + Filament 3 + Fortify + Sanctum + Octane
- 前端：Vue 3 + Vite + Pinia + Naive UI + Tailwind
- 数据库：MySQL / SQLite / PostgreSQL（共存）

### 1.2 重构目标
1. **统一技术栈**：前后端 TypeScript + Go，单一语言链路
2. **提升性能**：Go 高并发 + Nuxt SSR 优化首屏
3. **保持兼容**：REST 契约、数据库结构、驱动能力完全对齐老版本
4. **平滑迁移**：可与老版本并行运行，逐模块替换
5. **保留生态**：存储 / 支付 / 短信 / 邮件 / 社交 / 扫描 / 处理 7 大驱动体系

### 1.3 锁定技术栈

| 维度 | 选型 |
|---|---|
| 前端框架 | **Nuxt 3** + Vue 3 + TypeScript |
| 状态管理 | Pinia |
| UI 库 | Naive UI + Tailwind CSS |
| 工具 | VueUse、@nuxt/image、@vueuse/head |
| 国际化 | @nuxtjs/i18n |
| 后端语言 | **Go 1.22+** |
| Web 框架 | **Gin** |
| ORM | **GORM v2** |
| 队列 | **Asynq**（基于 Redis） |
| 配置 | Viper |
| 日志 | Zap |
| 鉴权 | JWT（前台）+ Session（后台） |
| RBAC | Casbin |
| 校验 | go-playground/validator |
| 存储 | S3 SDK / 阿里云 / 腾讯云 / 七牛 / 又拍 / FTP / SFTP / WebDAV |
| 图片处理 | bimg (libvips) + imaging（备选） |
| 数据库 | **MySQL 8.0（生产）+ SQLite（开发/单机）** |
| 缓存 | Redis + BigCache |
| 对象存储 | MinIO（开发） / S3 兼容 |
| 部署 | Docker + Docker Compose |
| 反向代理 | Nginx |

---

## 2. 顶层目录结构

```
d:\projects\ywty\
├── README.md
├── REFACTOR_PLAN.md                # 本文档
├── server/                          # Go 后端工程
│   ├── cmd/
│   │   ├── api/main.go              # API 服务入口
│   │   ├── worker/main.go           # 队列 Worker 入口
│   │   └── migrate/main.go          # 数据迁移 CLI
│   ├── internal/
│   │   ├── config/                  # 配置加载（Viper）
│   │   ├── database/                # DB 工厂（MySQL/PostgreSQL/SQLite）
│   │   ├── cache/                   # 缓存抽象
│   │   ├── queue/                   # Asynq 客户端
│   │   ├── logger/                  # Zap 初始化
│   │   ├── middleware/              # Gin 中间件
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── ratelimit.go
│   │   │   └── upload_verify.go
│   │   ├── modules/                 # 业务模块（按域划分）
│   │   │   ├── user/
│   │   │   ├── group/
│   │   │   ├── album/
│   │   │   ├── photo/
│   │   │   ├── tag/
│   │   │   ├── share/
│   │   │   ├── explore/
│   │   │   ├── order/
│   │   │   ├── plan/
│   │   │   ├── coupon/
│   │   │   ├── payment/
│   │   │   ├── capacity/
│   │   │   ├── ticket/
│   │   │   ├── report/
│   │   │   ├── violation/
│   │   │   ├── notice/
│   │   │   ├── page/
│   │   │   ├── feedback/
│   │   │   ├── stat/
│   │   │   ├── auth/
│   │   │   └── admin/
│   │   ├── drivers/                 # 驱动层
│   │   │   ├── storage/             # 9 个存储驱动
│   │   │   ├── payment/             # 6 个支付驱动
│   │   │   ├── sms/                 # 多短信驱动
│   │   │   ├── mail/                # 多邮件驱动
│   │   │   ├── social/              # 多社交登录
│   │   │   ├── scan/                # 多图片扫描
│   │   │   └── process/             # 多图片处理
│   │   ├── pkg/                     # 通用工具
│   │   ├── response/                # 统一响应
│   │   ├── errors/                  # 业务错误码
│   │   └── router/                  # 路由注册
│   ├── migrations/                  # GORM 迁移文件
│   ├── configs/                     # 配置文件
│   │   ├── config.yaml
│   │   ├── config.dev.yaml
│   │   └── config.prod.yaml
│   ├── scripts/                     # 构建/部署脚本
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
├── web/                             # 现有 Vue 3 工程（保留至迁移完成）
└── web-nuxt/                        # 新 Nuxt 3 工程
    ├── app/
    │   ├── app.vue
    │   ├── error.vue
    │   ├── layouts/
    │   │   ├── default.vue
    │   │   ├── user.vue
    │   │   └── admin.vue
    │   ├── pages/
    │   │   ├── index.vue
    │   │   ├── explore/
    │   │   ├── share/[token].vue
    │   │   ├── page/[slug].vue
    │   │   ├── user/
    │   │   │   ├── dashboard/
    │   │   │   ├── album/
    │   │   │   ├── photo/
    │   │   │   ├── share/
    │   │   │   ├── order/
    │   │   │   ├── plan/
    │   │   │   ├── ticket/
    │   │   │   ├── profile/
    │   │   │   └── setting/
    │   │   ├── admin/               # 管理后台
    │   │   │   ├── login.vue
    │   │   │   ├── dashboard.vue
    │   │   │   ├── users/
    │   │   │   ├── groups/
    │   │   │   ├── albums/
    │   │   │   ├── photos/
    │   │   │   ├── plans/
    │   │   │   ├── orders/
    │   │   │   ├── coupons/
    │   │   │   ├── shares/
    │   │   │   ├── reports/
    │   │   │   ├── tickets/
    │   │   │   ├── notices/
    │   │   │   ├── pages/
    │   │   │   ├── violations/
    │   │   │   ├── storages/
    │   │   │   ├── payments/
    │   │   │   ├── smss/
    │   │   │   ├── mails/
    │   │   │   ├── socials/
    │   │   │   ├── scans/
    │   │   │   ├── processes/
    │   │   │   └── settings/
    │   │   ├── auth/
    │   │   │   ├── login.vue
    │   │   │   ├── register.vue
    │   │   │   └── forgot-password.vue
    │   │   └── [...notFound].vue
    │   ├── components/
    │   │   ├── common/
    │   │   ├── photo/
    │   │   ├── album/
    │   │   ├── upload/
    │   │   ├── payment/
    │   │   └── admin/                # 后台专用组件
    │   ├── composables/
    │   │   ├── useApi.ts
    │   │   ├── useAuth.ts
    │   │   ├── useUpload.ts
    │   │   └── usePermission.ts
    │   ├── stores/                   # Pinia
    │   │   ├── auth.ts
    │   │   ├── user.ts
    │   │   ├── album.ts
    │   │   ├── photo.ts
    │   │   ├── upload.ts
    │   │   ├── order.ts
    │   │   ├── plan.ts
    │   │   └── admin.ts
    │   ├── middleware/
    │   │   ├── auth.ts
    │   │   ├── guest.ts
    │   │   └── admin.ts
    │   ├── plugins/
    │   ├── utils/
    │   └── locales/                  # i18n
    │       ├── zh-CN.json
    │       └── en-US.json
    ├── server/
    │   ├── api/                      # Nitro BFF（可选聚合）
    │   └── middleware/
    ├── nuxt.config.ts
    ├── package.json
    ├── tsconfig.json
    └── Dockerfile
├── deploy/
│   ├── docker-compose.yml            # 一键拉起 MySQL/PostgreSQL/Redis/Server/Nuxt
│   ├── nginx.conf
│   ├── mysql/init.sql
│   └── postgres/init.sql
└── .env.example
```

---

## 3. 阶段化任务总览

| 阶段 | 名称 | 周期 | 关键交付 |
|---|---|---|---|
| **P0** | 工程脚手架 | - | Go + Nuxt 工程跑通、双 DB 连通、Hello World |
| **P1** | 数据库与基础 | - | 19 张表 1:1 迁移、模型、迁移 CLI |
| **P2** | 鉴权体系 | - | 注册/登录/Token/RBAC/Casbin |
| **P3** | 核心域 | - | 用户/群组/相册/照片/标签/分享 |
| **P4** | 存储驱动 | - | 9 个存储驱动 + 工厂 + 配额 |
| **P5** | 商业域 | - | 套餐/订单/优惠券/支付/容量 |
| **P6** | 扩展驱动 | - | 短信/邮件/社交/扫描/处理 |
| **P7** | 运营域 | - | 工单/举报/违规/通知/公告/单页 |
| **P8** | Nuxt 公共端 | - | Explore/Share/Page/Auth(SSR) |
| **P9** | Nuxt 用户中心 | - | 相册/照片/订单/工单/设置 |
| **P10** | Nuxt 后台 | - | `/admin/*` 全模块 |
| **P11** | 测试与部署 | - | 单元/集成测试、Docker、CI、迁移工具 |

---

## 4. 阶段详细任务

### P0 - 工程脚手架

**目标**：双工程可独立运行，双数据库可切换

| # | 任务 | 验收 |
|---|---|---|
| 0.1 | 初始化 `server/`（`go mod init`、Gin 框架接入） | `go run cmd/api/main.go` 返回 200 |
| 0.2 | 初始化 `web-nuxt/`（`npx nuxi init`） | `npm run dev` 渲染首页 |
| 0.3 | 配置 Viper（YAML + 环境变量覆盖） | 修改 yaml 生效 |
| 0.4 | 实现数据库工厂 `internal/database/factory.go` | 切换配置项同时支持 MySQL / PostgreSQL / SQLite |
| 0.5 | GORM 初始化 + 连接池配置 | `db.Raw("select 1").Scan(&n)` 成功 |
| 0.6 | 统一响应封装 `internal/response` | `code/msg/data` 三段式 |
| 0.7 | 统一错误码 `internal/errors` | 错误码常量与文档一致 |
| 0.8 | Zap 日志初始化（按 level、按文件切割） | 控制台与文件均输出 |
| 0.9 | 跨域中间件、健康检查 `/healthz` | curl 通过 |
| 0.10 | Docker Compose（MySQL + Redis + Server + Nuxt） | `docker compose up` 拉起全部 |

**目录产出**：`server/`、`web-nuxt/`、`deploy/`

---

### P1 - 数据库与基础

**目标**：19 张表 1:1 迁移、模型、迁移 CLI

| # | 任务 | 对应老迁移文件 |
|---|---|---|
| 1.1 | GORM 模型：`User` | `2024_04_24_172040_create_users_table` |
| 1.2 | GORM 模型：`Group` | `2024_04_24_161046_create_groups_table` |
| 1.3 | GORM 模型：`Driver` | `2024_04_24_161047_create_drivers_table` |
| 1.4 | GORM 模型：`Storage` | `2024_04_24_161050_create_storages_table` |
| 1.5 | GORM 模型：`Album` + `AlbumPhoto` | `2024_04_24_172200_create_albums_table` |
| 1.6 | GORM 模型：`Photo` | `2024_04_24_172219_create_photos_table` |
| 1.7 | GORM 模型：`Tag` + `Taggable`（多态） | `2024_04_24_172221_create_tags_table` |
| 1.8 | GORM 模型：`Share` + `Shareable`（多态） | `2024_04_24_172223_create_shares_table` |
| 1.9 | GORM 模型：`Notice` | `2024_04_24_172555_create_notices_table` |
| 1.10 | GORM 模型：`Page` | `2024_04_24_172714_create_pages_table` |
| 1.11 | GORM 模型：`Ticket` + `TicketReply` | `2024_04_24_172823_create_tickets_table` |
| 1.12 | GORM 模型：`Report` | `2024_04_24_173007_create_reports_table` |
| 1.13 | GORM 模型：`Like` | `2024_04_24_205207_create_likes_table` |
| 1.14 | GORM 模型：`Plan` + `PlanCapacity` + `PlanPrice` | `2024_04_24_225049_create_plans_table` |
| 1.15 | GORM 模型：`Coupon` | `2024_04_24_225418_create_coupons_table` |
| 1.16 | GORM 模型：`Order` | `2024_04_24_225450_create_orders_table` |
| 1.17 | GORM 模型：`OAuth` | 老版本已有 |
| 1.18 | GORM 模型：`Feedback` | 老版本已有 |
| 1.19 | GORM 模型：`Violation` | 老版本已有 |
| 1.20 | Spatie settings 表适配 | `settings/*` 迁移 |
| 1.21 | Jobs / JobBatch / FailedJob 队列表 | `0001_01_01_000002_create_jobs_table` |
| 1.22 | `cmd/migrate/main.go` 命令：`migrate up/down/status` | 命令行可用 |
| 1.23 | `cmd/migrate/main.go` 命令：`seed` 初始化数据 | 跑通 InitializeSeeder |
| 1.24 | **老数据导入工具**（mysqldump → 新库） | 字段一一对应导入成功 |

**验收**：新建空库跑 `migrate up` 后表结构与老库 `mysqldump --no-data` 完全一致；老数据导入无丢失。

---

### P2 - 鉴权体系

**目标**：注册 / 登录 / Token / RBAC / Casbin 跑通

| # | 任务 | 对应老接口 |
|---|---|---|
| 2.1 | `POST /api/v2/auth/register` | `V2/AuthController::register` |
| 2.2 | `POST /api/v2/auth/login`（账号密码） | `V2/AuthController::login` |
| 2.3 | `POST /api/v2/auth/logout` | `V2/AuthController::logout` |
| 2.4 | `POST /api/v2/auth/reset-password`（邮箱） | `V2/AuthController::resetPassword` |
| 2.5 | `POST /api/v2/auth/reset-password`（短信） | `V2/AuthController::resetPassword` |
| 2.6 | `POST /api/v2/auth/email/code`（发送验证码） | `V2/AuthController::sendEmailCode` |
| 2.7 | `POST /api/v2/auth/sms/code`（发送验证码） | `V2/AuthController::sendSmsCode` |
| 2.8 | `POST /api/v2/oauth/{provider}`（社交登录） | `V2/OAuthController` |
| 2.9 | `POST /api/v2/tokens`（生成 API Token） | `V2/UserTokenController::store` |
| 2.10 | `GET /api/v2/tokens` `DELETE /api/v2/tokens/{id}` | Token 列表与吊销 |
| 2.11 | JWT 中间件（前台） | `Middleware/CheckTokenPermission` |
| 2.12 | Session 中间件（后台） | Laravel 默认 session |
| 2.13 | Casbin 模型文件（`model.conf` + `policy.csv`） | RBAC 策略 |
| 2.14 | Casbin 适配器（DB 持久化） | 策略可热更新 |
| 2.15 | 管理员账号体系（`is_admin` / `group_id` 关联） | 与老版本字段一致 |
| 2.16 | 后台登录 `POST /api/admin/auth/login` | `Fortify/CreateNewUser` |
| 2.17 | 验证码（图形 / 邮箱 / 短信）三套 | `Mews\Captcha` / `MailService` / `SmsService` |
| 2.18 | 接口频率限制中间件 | `Jobs/Middleware/RateLimited` |

**验收**：用老版本账号在新版本登录成功；老 Token（Sanctum 格式）可平滑迁移或共存。

---

### P3 - 核心域（用户/相册/照片/标签/分享）

**目标**：核心上传/浏览链路完整

| # | 任务 | 对应老接口 |
|---|---|---|
| 3.1 | `GET/POST /api/v2/user/profile` | `V2/UserController` |
| 3.2 | `POST /api/v2/user/change-password` | `UpdateUserPassword` |
| 3.3 | `POST /api/v2/user/change-email` `change-phone` | `UserController` |
| 3.4 | `GET/POST /api/v2/albums` | `V2/UserAlbumController` |
| 3.5 | `GET/PUT/DELETE /api/v2/albums/{id}` | `UserAlbumController` |
| 3.6 | `POST /api/v2/photos`（上传单/多） | `V1/ImageController::upload` |
| 3.7 | `GET /api/v2/photos`（列表 + 过滤 + 排序） | `V2/UserPhotoController::index` |
| 3.8 | `GET/PUT/DELETE /api/v2/photos/{id}` | `UserPhotoController` |
| 3.9 | `POST /api/v2/photos/{id}/move-to-album` | 移入相册 |
| 3.10 | `POST /api/v2/photos/{id}/copy` | 复制照片 |
| 3.11 | `POST /api/v2/photos/batch-delete` | 批量删除 |
| 3.12 | `GET /api/v2/images/{id}`（对外图片直链） | `V1/ImageController::image` |
| 3.13 | `POST /api/v2/tags` `GET /api/v2/tags` | 标签管理 |
| 3.14 | `POST /api/v2/shares`（生成分享） | `V2/UserShareController::store` |
| 3.15 | `GET /api/v2/shares` `DELETE /api/v2/shares/{id}` | 分享管理 |
| 3.16 | `GET /s/{token}` 分享公开页（不需登录） | `V2/ShareController` |
| 3.17 | `POST /api/v2/likes` 点赞 | `Like` 模型 |
| 3.18 | `POST /api/v2/reports` 举报 | `V2/ReportController` |
| 3.19 | 上传频率限制中间件 | `UploadFrequencyLimit` |
| 3.20 | 缩略图生成 Job（Asynq） | `GeneratePhotoThumbnailJob` |
| 3.21 | 图片处理 Job（Resize / Watermark） | `HandlePhotoJob` |
| 3.22 | 自动删除 Job | `AutoDeletePhotoJob` |
| 3.23 | 单图/多图/粘贴/拖拽上传前端实现 | `web/src/components/Upload/ImageUploader.vue` |

**验收**：用户从注册到上传、相册管理、分享全链路打通。

---

### P4 - 存储驱动（9 个）

**目标**：所有存储驱动在 Go 端可插拔

| # | 驱动 | 库 |
|---|---|---|
| 4.1 | `StorageAbstract` 接口 | - |
| 4.2 | `StorageDriverFactory` 工厂 | - |
| 4.3 | Local Storage | `os` |
| 4.4 | S3 Storage | `aws-sdk-go-v2/service/s3` |
| 4.5 | Aliyun OSS | `aliyun-oss-go-sdk/oss` |
| 4.6 | Tencent COS | `cos-go-sdk-v5` |
| 4.7 | Qiniu | `qiniu/go-sdk/v7` |
| 4.8 | Upyun | `upyun-go-sdk` |
| 4.9 | FTP | `jlaffaye/ftp` |
| 4.10 | SFTP | `pkg/sftp` |
| 4.11 | WebDAV | `studio-b12/gowebdav` |
| 4.12 | 配额统计 Job（容量计算） | `UserCapacityService` |
| 4.13 | 跨存储复制 | 内部接口 |
| 4.14 | 直传签名（前端直传 OSS/S3 跳过服务器） | `POST /api/v2/uploads/sign` |
| 4.15 | X-Sendfile / X-Accel 响应 | `XSendfileResponseFactory` |

**验收**：9 个驱动均通过单元测试，前端能选择并上传成功。

---

### P5 - 商业域

**目标**：套餐、订单、支付、容量闭环

| # | 任务 | 对应老模块 |
|---|---|---|
| 5.1 | `Plan` CRUD（API + Admin） | `PlanResource` |
| 5.2 | `Plan` 价格表管理 | `PlanPrice` |
| 5.3 | `Plan` 容量策略 | `PlanCapacity` |
| 5.4 | `Plan` 群组策略 | `PlanGroup` |
| 5.5 | `GET /api/v2/plans`（公开套餐） | `V2/PlanController::index` |
| 5.6 | `POST /api/v2/orders`（下单） | `V2/UserOrderController::store` |
| 5.7 | `GET /api/v2/orders` `GET /api/v2/orders/{id}` | 订单列表与详情 |
| 5.8 | `POST /api/v2/orders/{id}/pay`（发起支付） | 调支付驱动 |
| 5.9 | `POST /api/v2/orders/{id}/cancel` | `CancelOrderJob` |
| 5.10 | 优惠券 `Coupon` CRUD + 校验 | `CouponResource` |
| 5.11 | 容量服务（用户当前容量/已用/剩余） | `UserCapacityService` |
| 5.12 | **支付驱动 6 个** | - |
| 5.12.1 | Alipay（`yansongda/pay` Go 替代） | `AlipayPayment` |
| 5.12.2 | WeChat Pay | `WeChatPayment` |
| 5.12.3 | PayPal | `plutov/paypal` |
| 5.12.4 | EPay（聚合） | `EPayPayment` |
| 5.12.5 | UniPay（银联） | `UniPayPayment` |
| 5.12.6 | PayPay（日） | `PayPayPayment` |
| 5.13 | 支付回调路由 | `web.php` pay |
| 5.14 | 退款流程 | 各驱动适配 |
| 5.15 | 订单统计报表 | `StatService` |

**验收**：能完成"购买套餐 → 支付 → 容量增加"全链路。

---

### P6 - 扩展驱动

**目标**：5 大扩展驱动全部可插拔

#### 6.1 短信（SMS）
| 驱动 | 库 |
|---|---|
| Aliyun | `aliyun/alibaba-cloud-sdk-go` |
| Tencent | `tencentcloud/tencentcloud-sdk-go` |
| Huawei | 自封装 |
| Twilio | `twilio/twilio-go` |
| EJoin | 自封装 |
| Qiniu | `qiniu/go-sdk` |

#### 6.2 邮件（Mail）
| 驱动 |
|---|
| SMTP |
| Aliyun DirectMail |
| Tencent SES |
| Mailgun |
| Amazon SES |
| Postmark |

#### 6.3 社交登录（Social）
| 驱动 |
|---|
| GitHub |
| Google |
| Facebook |
| Twitter |
| 微信 |
| QQ |
| 钉钉 |
| 飞书 |
| Gitee |
| 微博 |

#### 6.4 图片扫描（Scan）
| 驱动 |
|---|
| Aliyun Green |
| Tencent IMS |
| 自定义 HTTP |

#### 6.5 图片处理（Process / Handle）
| 驱动 |
|---|
| Local（bimg / imaging） |
| 第三方 HTTP 处理器 |

**验收**：每个驱动独立测试通过；可热插拔；新驱动注册即可用。

---

### P7 - 运营域

| # | 任务 | 对应老模块 |
|---|---|---|
| 7.1 | 通知 `Notice` CRUD + 用户读取 | `NoticeResource` |
| 7.2 | 单页 `Page` CRUD（CMS） | `PageResource` |
| 7.3 | 工单 `Ticket` + `TicketReply` 完整流程 | `TicketResource` |
| 7.4 | 工单创建/回复邮件通知 Job | `SendTicketCreateNotificationMailJob` |
| 7.5 | 举报 `Report` 处理 | `ReportResource` |
| 7.6 | 违规 `Violation` 处理 | `ViolationResource` |
| 7.7 | 反馈 `Feedback` | `FeedbackResource` |
| 7.8 | 统计 `Stat`（用户/上传/存储/订单） | `StatService` |
| 7.9 | 后台图表数据接口 | `Widgets/*Chart` |
| 7.10 | 群组 `Group` 管理 + 默认群组策略 | `GroupResource` |
| 7.11 | 升级（License）流程 | `UpgradeJob` + `UpgradeController` |

---

### P8 - Nuxt 公共端（SSR 优先）

| # | 任务 | 对应老页面 |
|---|---|---|
| 8.1 | 首页 `index.vue` | `themes/default/index.blade.php` |
| 8.2 | Explore `pages/explore/index.vue` | `IndexView.vue` |
| 8.3 | Explore 专辑 `pages/explore/album.vue` | `AlbumView.vue` |
| 8.4 | Explore 用户中心 `pages/explore/user.vue` | `UserCenterView.vue` |
| 8.5 | 分享公开页 `pages/share/[token].vue` | `ShareView.vue` |
| 8.6 | 单页 `pages/page/[slug].vue` | `PageView.vue` |
| 8.7 | 登录 `pages/auth/login.vue` | `LoginView.vue` |
| 8.8 | 注册 `pages/auth/register.vue` | `RegisterView.vue` |
| 8.9 | 找回密码 `pages/auth/forgot-password.vue` | `ForgetPassword.vue` |
| 8.10 | 404 `pages/[...notFound].vue` | `NotFoundView.vue` |
| 8.11 | `@nuxt/image` 接入 + 缩略图参数 | PhotoSwipe 替代 |
| 8.12 | PhotoSwipe 替换（Lightbox） | `usePhotoSwipe.ts` |
| 8.13 | Masonry 瀑布流布局 | `vue-next-masonry` |
| 8.14 | i18n `zh-CN / en-US` | 完整对齐 |
| 8.15 | SEO meta（OG / Twitter Card） | 动态 |

**验收**：首页 TTFB < 200ms；Lighthouse Performance > 90。

---

### P9 - Nuxt 用户中心

| # | 任务 |
|---|---|
| 9.1 | Layout `user.vue`（侧边栏 + 顶栏） |
| 9.2 | Dashboard `pages/user/dashboard/index.vue` |
| 9.3 | 相册列表/详情 `pages/user/album/*` |
| 9.4 | 照片列表/操作 `pages/user/photo/*` |
| 9.5 | 分享管理 `pages/user/share/index.vue` |
| 9.6 | 套餐列表/详情 `pages/user/plan/*` |
| 9.7 | 订单列表/详情/支付 `pages/user/order/*` |
| 9.8 | 工单创建/列表/详情 `pages/user/ticket/*` |
| 9.9 | 个人资料（基本信息/邮箱/手机/密码/社交） `pages/user/profile/*` |
| 9.10 | 用户设置 `pages/user/setting/index.vue` |
| 9.11 | 上传组件（拖拽/粘贴/批量/进度） |
| 9.12 | 验证码输入组件（图形/邮箱/短信） |
| 9.13 | 支付选择组件（图标 + 渠道） |
| 9.14 | 照片批量操作（删除/移入/分享） |

---

### P10 - Nuxt 管理后台 `/admin/*`

| # | 任务 | 对应老 Filament Resource |
|---|---|---|
| 10.1 | Admin Layout `admin.vue` | - |
| 10.2 | Admin Login | `Filament/Pages/Auth/Login` |
| 10.3 | Dashboard 概览 + 图表 | `Dashboard.php` + `Widgets/*` |
| 10.4 | 用户管理 | `UserResource` |
| 10.5 | 群组管理 | `GroupResource` |
| 10.6 | 相册管理 | `AlbumResource` |
| 10.7 | 照片管理（查看/审核） | `PhotoResource` |
| 10.8 | 套餐管理 | `PlanResource` |
| 10.9 | 订单管理 | `OrderResource` |
| 10.10 | 优惠券管理 | `CouponResource` |
| 10.11 | 分享管理 | `ShareResource` |
| 10.12 | 举报管理 | `ReportResource` |
| 10.13 | 工单管理 | `TicketResource` |
| 10.14 | 通知管理 | `NoticeResource` |
| 10.15 | 单页管理 | `PageResource` |
| 10.16 | 违规管理 | `ViolationResource` |
| 10.17 | 反馈管理 | `FeedbackResource` |
| 10.18 | 存储驱动管理 | `StorageResource` |
| 10.19 | 支付驱动管理 | `PaymentDriverResource` |
| 10.20 | 短信驱动管理 | `SmsDriverResource` |
| 10.21 | 邮件驱动管理 | `MailDriverResource` |
| 10.22 | 社交驱动管理 | `SocialiteDriverResource` |
| 10.23 | 扫描驱动管理 | `ScanDriverResource` |
| 10.24 | 处理驱动管理 | `ProcessDriverResource` |
| 10.25 | 处理驱动管理 | `HandleDriverResource` |
| 10.26 | 系统设置（基础/站点/用户/队列/升级） | `Settings/*` |
| 10.27 | License 授权 | `Filament/Pages/License` |

**验收**：所有老后台功能 1:1 还原，支持权限隔离（普通管理员仅看分配范围）。

---

### P11 - 测试与部署

| # | 任务 |
|---|---|
| 11.1 | Go 单元测试（覆盖率 > 70%） |
| 11.2 | Go 集成测试（API 级） |
| 11.3 | Nuxt 组件单测（Vitest） |
| 11.4 | Nuxt E2E（Playwright） |
| 11.5 | 老数据迁移工具（一键 mysqldump → 导入） |
| 11.6 | 老 Laravel 与新 Go 双写期（影子模式） |
| 11.7 | 切流方案（按用户/按接口灰度） |
| 11.8 | Dockerfile（多阶段构建、镜像 < 100MB） |
| 11.9 | docker-compose.yml（生产级） |
| 11.10 | Nginx 配置（静态资源、限流、HTTPS） |
| 11.11 | GitHub Actions CI（lint + test + build） |
| 11.12 | 监控（Prometheus + Grafana） |
| 11.13 | 链路追踪（OpenTelemetry） |
| 11.14 | 文档站（部署 / 升级 / 驱动开发） |
| 11.15 | CHANGELOG 与版本号策略 |

---

## 5. 关键策略

### 5.1 数据库多驱动实现

```go
// internal/database/factory.go
func New(cfg *config.Config) (*gorm.DB, error) {
    var dialector gorm.Dialector
    switch cfg.Database.Driver {
    case "mysql":
        dialector = mysql.Open(cfg.Database.DSN)
    case "postgres":
        dialector = postgres.Open(cfg.Database.DSN)
    case "sqlite":
        dialector = sqlite.Open(cfg.Database.Path)
    default:
        return nil, errors.New("unsupported driver: " + cfg.Database.Driver)
    }
    return gorm.Open(dialector, &gorm.Config{...})
}
```

**驱动差异处理**：
- **MySQL**：`bigint` ↔ `int64`，`varchar(n)` ↔ `string`，`json` ↔ `json.RawMessage`
- **PostgreSQL**：`bigserial` ↔ `int64`，`varchar(n)` ↔ `string`，`jsonb` ↔ `json.RawMessage`（注意使用 `jsonb` 而非 `json`）
- **SQLite**：`INTEGER` ↔ `int64`，`TEXT` ↔ `string`，无原生 JSON 用 `TEXT` 存储 + 应用层序列化
- 软删除字段 `deleted_at` 一致（GORM 自动处理）
- 时间字段 `created_at` / `updated_at` 自动维护
- 多态字段 `taggable_type/id` / `shareable_type/id` 保持原结构
- 自增策略差异：MySQL `AUTO_INCREMENT`、PostgreSQL `SERIAL`/`BIGSERIAL`、SQLite `AUTOINCREMENT`（由 GORM tag 控制）
- 索引命名差异：长度限制 MySQL 64 字符、PostgreSQL 63 字符，迁移脚本需统一前缀避免冲突
- 分页语法差异：MySQL 用 `LIMIT ? OFFSET ?`、PostgreSQL 同 SQLite 也兼容、GORM 已封装

**GORM tag 适配示例**：
```go
type User struct {
    ID        uint           `gorm:"primaryKey;autoIncrement"`
    Name      string         `gorm:"type:varchar(100);not null;index"`
    Meta      datatypes.JSON `gorm:"type:json"`        // MySQL
    // Meta      datatypes.JSON `gorm:"type:jsonb"`     // PostgreSQL 需切换
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

**适配层抽象**（`internal/database/dialect.go`）：
- 根据驱动返回对应列类型映射
- 统一 JSON 字段读写
- 提供驱动专属的 Migration 钩子

### 5.2 REST 契约保持

- **路径不变**：`/api/v1/*` `/api/v2/*` 完全对齐
- **请求/响应 JSON 字段顺序、命名、类型不变**
- **错误码体系** 重新映射但 HTTP 状态码一致
- **分页结构** `{ data, meta: { current_page, total, per_page } }`

### 5.3 驱动注册表

```go
// internal/drivers/storage/factory.go
type Factory interface {
    Register(name string, driver StorageAbstract)
    Get(name string) (StorageAbstract, error)
}
```

- 启动时从 `drivers` 表加载配置并注册
- 业务调用统一 `factory.Get(group.DefaultStorageID)`
- 新增驱动只需实现 `StorageAbstract` 接口

### 5.4 后台 RBAC

- Casbin 策略文件按"角色 → 资源 → 操作"定义
- 后台中间件 `AdminAuth()` 验证角色
- 路由级 meta 标签控制可见按钮
- 数据级权限通过 `group_id` / `created_by` 过滤

### 5.5 切流策略

| 阶段 | 模式 | 比例 |
|---|---|---|
| 阶段 1 | 老 Laravel 主、新 Go 影子 | 0% |
| 阶段 2 | 新 Go 灰度 | 5% → 20% → 50% |
| 阶段 3 | 新 Go 为主、老 Laravel 备 | 80% → 100% |
| 阶段 4 | 完全切换 | 100% |

---

## 6. 风险与应对

| 风险 | 影响 | 应对 |
|---|---|---|
| 老 Token 格式不兼容 | 用户需重新登录 | 同时支持 Sanctum Token 与 JWT 解析 |
| 支付回调 IP 白名单 | 切流期回调混乱 | 新老各保留独立回调地址，灰度切换 |
| 大表迁移数据丢失 | 不可逆 | 全量备份 + 影子模式比对 + 灰度切流 |
| 图片直链 URL 不一致 | 外部引用 404 | 保留 `ImageController` 旧路径转发 |
| 主题/插件系统 | 第三方无法升级 | 主题系统保留为静态资源，文档告知 |
| i18n 词条对齐 | 用户体验下降 | 老 `.json` 文件直接复用 |
| Go SDK 缺失（部分支付/短信） | 无法实现 | 自封装 HTTP 客户端 |
| 性能回退 | 体验下降 | 切流后持续观察 2 周 |

---

## 7. 验收清单（按阶段）

- [ ] P0：双工程可独立 `docker compose up` 跑通（MySQL/PostgreSQL/SQLite 均可切换）
- [ ] P1：老库 `mysqldump --no-data` / `pg_dump --schema-only` 与新库 `migrate up` 字段一致（三数据库）
- [ ] P2：用老版本账号 + 密码可在新版本登录
- [ ] P3：上传/相册/分享全链路跑通
- [ ] P4：9 个存储驱动逐个测试通过
- [ ] P5：可购买套餐并完成支付
- [ ] P6：5 大扩展驱动均注册成功
- [ ] P7：所有运营模块 API 文档与老版本一致
- [ ] P8：公共端 Lighthouse 性能分 > 90
- [ ] P9：用户中心功能 1:1 还原
- [ ] P10：后台 27 个模块全部可用
- [ ] P11：CI 通过，Docker 镜像 < 100MB，监控就绪

---

## 8. 排期建议

| 阶段 | 建议周期 |
|---|---|
| P0 | 1 周 |
| P1 | 1 周 |
| P2 | 1.5 周 |
| P3 | 2 周 |
| P4 | 1.5 周 |
| P5 | 2 周 |
| P6 | 2 周 |
| P7 | 1 周 |
| P8 | 1.5 周 |
| P9 | 2 周 |
| P10 | 2.5 周 |
| P11 | 1.5 周 |

---

## 9. 后续动作

1. ✅ 确认本文档
2. ⏭️ 启动 P0（Go + Nuxt 工程脚手架）
3. ⏭️ 启动 P1（数据库迁移）
4. ⏭️ 逐阶段推进，每阶段结束产出可演示版本

---

> 文档维护人：项目组
> 最近更新：v1.0 初版
