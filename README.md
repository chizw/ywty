<div align="center">

# 云雾图驿

**自托管图床 / 云相册 · Go + Nuxt 重构版**

[![GitHub release](https://img.shields.io/github/v/release/chizw/ywty?label=release&style=flat-square)](https://github.com/chizw/ywty/releases)
[![License](https://img.shields.io/github/license/chizw/ywty?style=flat-square)](./LICENSE.md)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![Nuxt](https://img.shields.io/badge/Nuxt-3-00DC82?style=flat-square&logo=nuxt.js&logoColor=white)](https://nuxt.com)
[![shadcn--vue](https://img.shields.io/badge/shadcn--vue-000000?style=flat-square&logo=shadcnui&logoColor=white)](https://shadcn-vue.com)
[![Build Status](https://img.shields.io/github/actions/workflow/status/chizw/ywty/ci.yml?label=CI&style=flat-square)](https://github.com/chizw/ywty/actions)
[![Stars](https://img.shields.io/github/stars/chizw/ywty?style=flat-square)](https://github.com/chizw/ywty/stargazers)
[![Forks](https://img.shields.io/github/forks/chizw/ywty?style=flat-square)](https://github.com/chizw/ywty/network/members)

[English](./README.en.md) · 简体中文

</div>

---

## ✨ 项目简介

**ywty** 是一款**自托管图床 / 云相册**系统：从经典 ywty 完全重写而来，后端从 PHP/Laravel 迁移到 **Go + Gin + GORM**，前端从 Vue 3 + Vite 升级到 **Nuxt 3 SSR**，UI 组件库采用 **shadcn-vue**（Radix Vue + Tailwind CSS）。

我们保留了原有的 REST 契约、30+ 张业务表结构和「多驱动」生态，目标是给个人/团队提供**零成本、可扩展、长期可维护**的私有云相册方案。

> **当前状态**：P0–P11 全部完成 ✅，**生产可用 alpha 版**。欢迎 Star / Watch 关注更新。

## 🎯 核心特性

### 🖼️ 核心域
- 📁 **相册** - 公开/私有、批量管理、容量统计
- 🏷️ **标签** - 多对多绑定图片
- 🔗 **分享** - 密码保护、过期时间、QR Code
- ❤️ **点赞** - 去重计数
- 🚨 **举报** - 多类型、自动归集违规
- 🔍 **探索** - 公开内容瀑布流、用户/专辑页

### 🔐 鉴权 & 权限
- JWT（前台用户）+ Session（后台管理员）
- 三套验证码：图形 / 邮箱 / 短信
- **RBAC**（Casbin）+ 角色组 + 用户组
- **API Token**（用户自管理 + 能力授权）
- **OAuth 社交登录**：GitHub / Google / 微信 / QQ / 钉钉 / Gitee / 微博

### 📦 存储驱动（9 种可插拔）
Local · **S3** · 阿里云 **OSS** · 腾讯云 **COS** · 七牛云 · 又拍云 · **FTP** · **SFTP** · **WebDAV**
- 浏览器直传签名（OSS/S3 跳过服务器中转）
- 跨存储复制
- 用户配额 / 容量统计

### 💳 商业域
- 套餐 / 订阅 / 订单全链路
- 6 种支付驱动：**支付宝** · 微信 · **PayPal** · **Stripe** · **EPay 彩虹** · Mock
- 异步通知（签名校验 + 防重放）
- 优惠券（满减 / 折扣 / 固定额）

### 🔌 扩展驱动
| 类别 | 已支持 |
|---|---|
| 短信 | 阿里云 · 腾讯云 · Twilio · 七牛云 · Log |
| 邮件 | SMTP · 阿里云 DirectMail · Log |
| 社交登录 | GitHub · Google · 微信 · QQ · 钉钉 · Gitee · 微博 |
| 图片扫描 | 阿里云内容安全 · 腾讯云 IMS · 自定义 HTTP |
| 图片处理 | Local（imaging）· 自定义 HTTP |

### 🗃️ 数据库
- 一套 GORM 模型覆盖 **30+ 张业务表**
- 默认 **SQLite** 轻量化，深度支持 **MySQL**（推荐 MariaDB 10.6.20+）
- 迁移 CLI：`migrate up / down / status / seed`

## 🏗️ 技术栈

### 后端
| 技术 | 用途 |
|---|---|
| **Go 1.25+** | 主语言 |
| **Gin** | HTTP 框架 |
| **GORM v2** | ORM |
| **Viper** | 配置（YAML + 环境变量） |
| **Zap** | 日志（按等级 / 文件切割） |
| **Asynq** | 异步队列（基于 Redis） |
| **BigCache + Redis** | 多级缓存 |
| **Casbin** | RBAC |
| **golang-jwt** | JWT |
| **disintegration/imaging** | 图像处理 |
| **asaskevich/govalidator** | 验证 |

### 前端
| 技术 | 用途 |
|---|---|
| **Nuxt 3** | SSR / SSG 框架 |
| **Vue 3** + **TypeScript** | 视图层 |
| **shadcn-vue** | UI 组件库（42 个组件，基于 Radix Vue） |
| **Radix Vue** | 无样式原语（Accessible Primitives） |
| **Tailwind CSS** | 原子化样式 + CSS 变量主题 |
| **class-variance-authority** | 组件变体管理 |
| **@lucide/vue** | 图标库 |
| **Pinia** | 状态管理 |
| **Vue I18n** | 国际化（zh-CN / en-US） |
| **@nuxt/image** | 图片优化 |

### DevOps
| 技术 | 用途 |
|---|---|
| **Docker** + **Docker Compose** | 容器化 |
| **GitHub Actions** | CI（gofmt / vet / test / build / Docker 多架构） |
| **Nginx** | 反向代理 + HTTPS |
| **Supervisor** | 进程守护 |

## 🚀 部署

### 方式 A：Docker Compose（推荐）

```bash
# 1. 准备环境变量
cp .env.example .env
# 编辑 .env：至少改 JWT_SECRET 和 APP_URL

# 2. 一键启动（内置 MySQL + Redis）
docker compose up -d

# 3. 访问
open http://localhost:3000
```

**其他场景**：
```bash
# 连接宿主机已有 MySQL + Redis
DB_HOST=host.docker.internal REDIS_HOST=host.docker.internal \
  docker compose up -d --scale mysql=0 --scale redis=0

# 本地开发用 SQLite（无需任何数据库容器）
DB_DRIVER=sqlite docker compose up -d --scale mysql=0 --scale redis=0

# 数据库迁移
docker compose run --rm migrate
```

**默认管理员**：`admin` / `admin123456`（请尽快修改）

### 方式 B：本地开发

```bash
# ---- 后端 ----
cd server
go mod download
cp .env.example .env
go run ./cmd/migrate       # 初始化数据库 + 种子数据
go run ./cmd/api           # 启动 API（默认 :8080）
# 另起一个终端：
go run ./cmd/worker        # 启动 Worker（队列消费）

# ---- 前端 ----
cd web-nuxt
npm install
npm run dev                # 启动 Web（默认 :3000）
```

### 方式 C：生产部署

```bash
# 1. 编译产物 + 构建多架构镜像并推送
./deploy/deploy.sh prod v1.0.0

# 2. 拷贝 Nginx / Supervisor 模板
cp deploy/nginx.conf.example /etc/nginx/sites-available/ywty.conf
cp deploy/supervisor.conf /etc/supervisor/conf.d/ywty.conf

# 3. 数据库备份（自动按数据库驱动适配）
./deploy/backup-db.sh mysql
```

详细的部署配置和参数说明见 [deploy/README.md](./deploy/README.md)。

## 🛠️ 开发

### 目录结构

```
ywty/
├── server/                    # Go 后端
│   ├── cmd/
│   │   ├── api/              # API 服务入口
│   │   ├── worker/           # 队列 Worker
│   │   ├── migrate/          # 数据库迁移 CLI
│   │   └── import/           # 数据导入工具
│   ├── configs/              # Viper 配置 + Casbin 模型
│   ├── internal/
│   │   ├── auth/             # JWT
│   │   ├── handler/          # HTTP handlers
│   │   ├── service/          # 业务服务
│   │   ├── model/            # GORM 模型（30+ 张表）
│   │   ├── drivers/          # 支付/存储驱动
│   │   ├── notify/           # 短信 / 邮件
│   │   ├── social/           # 社交登录
│   │   ├── scan/             # 图片扫描
│   │   ├── process/          # 图片处理
│   │   ├── jobs/             # 异步任务
│   │   ├── queue/            # Asynq 封装
│   │   ├── rbac/             # Casbin 鉴权
│   │   ├── middleware/       # 限流 / CORS
│   │   ├── database/         # GORM 工厂
│   │   ├── errors/           # 业务错误码
│   │   ├── response/         # 统一响应
│   │   ├── router/           # 路由注册
│   │   ├── seed/             # 种子数据
│   │   ├── license/          # License 管理
│   │   └── logger/           # Zap
│   ├── Dockerfile
│   └── go.mod
├── web-nuxt/                  # Nuxt 3 前端
│   ├── pages/                # 公共端 + 用户中心 + 管理后台
│   ├── layouts/              # 4 套布局
│   ├── components/           # 通用组件
│   │   └── ui/               # shadcn-vue UI 组件（42 个）
│   ├── composables/          # useApi / useAuth / useMessage
│   ├── stores/               # Pinia 状态
│   ├── middleware/           # 路由守卫
│   ├── i18n/                 # 国际化
│   ├── types/                # TypeScript 类型
│   ├── server/api/           # Nitro BFF
│   └── Dockerfile
├── deploy/                    # 部署脚本
│   ├── deploy.sh             # 一键编译 + 构建 + 推送
│   ├── dev-up.sh / dev-down.sh
│   ├── backup-db.sh          # 多数据库备份
│   ├── nginx.conf.example    # 反向代理
│   └── supervisor.conf       # 进程守护
├── .github/workflows/         # CI / Release
├── docker-compose.yml
├── .env.example
├── Dockerfile                # 合并镜像（API + Web）
└── README.md
```

### 开发规范

- 后端遵循 **Uber Go Style Guide**，提交前跑：
  ```bash
  gofmt -l .          # 格式检查
  go vet ./...        # 静态检查
  go test -race ./...
  ```
- 前端遵循 **Vue 3 + TypeScript 规范**，提交前跑：
  ```bash
  npm run lint
  npm run typecheck
  npm run build
  ```
- 提交信息遵循 **Conventional Commits**（`feat:` / `fix:` / `docs:` / `refactor:` / `test:` / `chore:`）

### 调试技巧

```bash
# 查看 API 健康
curl http://localhost:8080/healthz

# 查看队列状态（需装 asynqmon）
# docker run -p 8081:8080 hibiken/asynqmon --redis-addr=host.docker.internal:6379

# 实时日志
docker compose logs -f ywty

# 进入 API 容器调试
docker compose exec ywty sh
```

## 🧪 测试

```bash
# 后端
cd server
go test -race -shuffle=on -coverprofile=coverage.out ./...
go tool cover -html=coverage.out   # 查看覆盖率

# 前端（待补 E2E）
cd web-nuxt
npm run test
```

当前覆盖：`errors` · `storage` · `drivers/payment`（含 EPay 签名校验、Mock 驱动单测）。

## 🤝 贡献

欢迎贡献代码 / 提 Issue / 提 PR。请阅读 [CONTRIBUTING.md](./CONTRIBUTING.md) 与 [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)。

代码改动请尽量带上：
- ✅ 单元测试
- ✅ CHANGELOG 条目
- ✅ 类型定义 / API 文档更新

## ⭐ Star 增长趋势

如果这个项目对你有帮助，请给我们一个 ⭐ 鼓励！你的支持是项目持续迭代的最大动力。

[![Star History Chart](https://api.star-history.com/svg?repos=chizw/ywty&type=Date)](https://star-history.com/#chizw/ywty&Date)

## 📐 API 约定

- 基础路径：`/api/v1`、`/api/v2`（双版本兼容）
- 鉴权：`Authorization: Bearer <access_token>` 或 `X-Token: <token>`
- 响应结构：
  ```json
  { "code": 0, "message": "ok", "data": {...} }
  ```
- 分页响应：
  ```json
  { "data": [...], "meta": { "current_page": 1, "total": 100, "per_page": 20 } }
  ```
- 错误码：业务错误码见 `internal/errors`，HTTP 状态码保持一致

## 🛣️ 路线图

| 阶段 | 名称 | 状态 |
|---|---|---|
| P0 | 工程脚手架 | ✅ |
| P1 | 数据库与基础 | ✅ |
| P2 | 鉴权体系 | ✅ |
| P3 | 核心域 | ✅ |
| P4 | 存储驱动（9 种） | ✅ |
| P5 | 商业域 | ✅ |
| P6 | 扩展驱动 | ✅ |
| P7 | 运营域 | ✅ |
| P8 | Nuxt 公共端 | ✅ |
| P9 | Nuxt 用户中心 | ✅ |
| P10 | Nuxt 管理后台 | ✅ |
| P11 | 测试与部署 | ✅ |
| P12 | shadcn-vue UI 重构 | ✅ |
| v1.0 | 正式版 | 🎯 2026 Q4 |
| 未来 | 移动端 App / 多租户 / 联邦 | 💭 |

## 📜 许可证

本项目采用 [MIT License](./LICENSE.md) 开源。

## 🙏 致谢

- 所有贡献者与社区用户
- ⭐ **如果觉得不错，给我们一个 Star 吧！** ⭐
