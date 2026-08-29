# ywty（云雾图驿）

基于 **ywty - **（PHP/Laravel）的 **Go 重写版本**。
PHP 原版源码保留在本仓库根目录作为移植参照；活跃开发在 [`server-go/`](server-go/)。

```
Docker 镜像：node 构建 web + admin → golang 构建 server-go → alpine(+vips) 运行
            （CGO_ENABLED=0 纯静态单二进制，SQLite/MySQL 双方言）
```

## 架构

| 目录 | 说明 |
| --- | --- |
| `server-go/` | Go 1.26 单二进制（HTTP 服务 + 数据库队列 worker + 调度） |
| `web/` | 用户端 Vue 3 SPA（与 PHP 版共用，构建产物由 Go 托管） |
| `admin/` | 管理后台 Vue 3 SPA（Vue + Naive UI，挂 `/admin`） |
| `app/` 等根目录 PHP | PHP 版参照实现（基线，不再演进） |

## Go 版里程碑状态

- ✅ M0：骨架、双方言迁移（37 表对齐 Laravel schema）、安装流程、静态托管
- ✅ M1：Sanctum 令牌认证、会话、API 权限映射、验证码、configs/group、用户资料/令牌
- ✅ M2：上传管线（命名/容量/频率限制）、纯 Go 图像处理 + vips 外挂、图片直出、相册、legacy v1
- ✅ M3：分享（密码/过期）、广场、公告、页面、举报、点赞
- ✅ M4：套餐/订单/优惠码、支付 SPI（EPay：RSA-SHA256）、支付回调、延时取消
- ✅ M5(部分)：工单完整闭环
- ✅ M6：`/api/admin` + admin SPA（用户/图片/公告/页面/订单/优惠码/工单/设置/储存/驱动）
- ⏳ 后续：S3/OSS/COS/WebDAV 储存驱动、支付宝/微信支付、短信与内容审核、OAuth 登录

## 与 PHP 版的兼容承诺

- **数据库**：表结构与 Laravel 迁移逐列对齐；PHP 版数据库可直接被 Go 版使用（启动时自动检测，
  旧库自动修补 `users.email` 可空约束）
- **API**：`openapi.json` 的 90 个路径 + `/api/v1` 旧客户端契约；响应 envelope
  `{status, message, data, time}` 逐键一致
- **认证**：Sanctum 令牌（sha256 存储、`{id}|{plain}` 明文）与 bcrypt 密码直接互认
- **加密设置**：`settings` 加密字段使用 Laravel Crypt（AES-256-CBC）格式，APP_KEY 可沿用
- 已知差异：未匹配 API 路径返回 HTTP 200 + `{"status":"error"}`（CI 冒烟约定；PHP 为 404）

## 快速开始（Docker）

```bash
docker compose up -d --build
# 首次启动在 environment 中提供：
#   APP_URL / APP_LICENSE_KEY / ADMIN_USERNAME / ADMIN_EMAIL / ADMIN_PASSWORD
# 即自动完成安装；或调用 POST /api/v2/install
```

- 服务端口：`PORT`（默认 3000），健康检查 `/healthz`
- 用户端：`/`（web 主题），管理后台：`/admin`
- 数据卷：`./data`（SQLite + app_key + installed.lock）、`./uploads`（本地储存）

### 环境变量

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `HOST` / `PORT` | 127.0.0.1 / 3000 | 监听地址 |
| `APP_URL` | http://localhost | 站点 URL（生成图片链接等） |
| `DB_DRIVER` | sqlite | `sqlite` / `mysql`（MariaDB ≥ 10.6） |
| `DB_PATH` | `{DATA_DIR}/ywty.db` | SQLite 文件路径 |
| `DB_HOST/PORT/USER/PASSWORD/NAME` | — | MySQL 连接 |
| `DATA_DIR` | data | 数据目录 |
| `UPLOADS_DIR` | uploads | 本地储存兜底根目录 |
| `STATIC_DIR` | public | 前端主题目录 |
| `ADMIN_STATIC_DIR` | admin/dist | 管理后台目录 |
| `APP_KEY` | 自动生成 | Laravel APP_KEY（`base64:...`，迁移旧库必填） |
| `APP_LICENSE_KEY` + `ADMIN_*` | — | 环境变量一键安装 |
| `REDIS_ADDR/PASSWORD` | — | 预留（当前使用 DB cache/queue） |

## 本地开发

```bash
# Go 后端
cd server-go && go test ./... && go run ./cmd/ywty

# 前端主题（产物输出到 public/）
cd web && npm install && npm run build

# 管理后台
cd admin && npm install && npm run build
```

## 原版

