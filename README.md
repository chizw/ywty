<div align="center">

# 云雾图驿

**自托管图床 / 云相册 · Rust + Astro**

[![GitHub release](https://img.shields.io/github/v/release/chizw/ywty?label=release&style=flat-square)](https://github.com/chizw/ywty/releases)
[![License](https://img.shields.io/github/license/chizw/ywty?style=flat-square)](./LICENSE.md)
[![Rust](https://img.shields.io/badge/Rust-1.85+-000000?style=flat-square&logo=rust&logoColor=white)](https://www.rust-lang.org)
[![Astro](https://img.shields.io/badge/Astro-5-BC52EE?style=flat-square&logo=astro&logoColor=white)](https://astro.build)
[![shadcn--vue](https://img.shields.io/badge/shadcn--vue%20%2F%20shadcn--ui-000000?style=flat-square&logo=shadcnui&logoColor=white)](https://shadcn-vue.com)
[![Build Status](https://img.shields.io/github/actions/workflow/status/chizw/ywty/ci.yml?label=CI&style=flat-square)](https://github.com/chizw/ywty/actions)
[![Stars](https://img.shields.io/github/stars/chizw/ywty?style=flat-square)](https://github.com/chizw/ywty/stargazers)
[![Forks](https://img.shields.io/github/forks/chizw/ywty?style=flat-square)](https://github.com/chizw/ywty/network/members)

[English](./README.en.md) · 简体中文

</div>

---

## ✨ 项目简介

**ywty** 是一款**自托管图床 / 云相册**系统：从经典 Lsky Pro+ 完全重写而来，后端采用 **Rust + axum + sqlx**，前端为 **Astro 5 SSR + Vue/React Islands**（公开页 Vue，用户中心/后台 React），UI 组件库统一 Tailwind 设计 token。

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
- 两套验证码：图形 / 邮箱
- **RBAC**（基于 Role 枚举）+ 角色组 + 用户组
- **API Token**（用户自管理 + 能力授权）
- **OAuth 社交登录**：GitHub / Google / 微信 / QQ / 钉钉 / Gitee / 微博

### 📦 存储驱动（9 种可插拔）
Local · **S3** · 阿里云 **OSS** · 腾讯云 **COS** · 七牛云 · 又拍云 · **FTP** · **SFTP** · **WebDAV**
- 浏览器直传签名（OSS/S3 跳过服务器中转）
- 用户配额 / 容量统计

### 💳 商业域
- 套餐 / 订阅 / 订单全链路
- 6 种支付驱动：**支付宝** · 微信 · **PayPal** · **Stripe** · **EPay 彩虹** · Mock
- 异步通知（签名校验 + 防重放）
- 优惠券（满减 / 折扣 / 固定额）

### 🔌 扩展驱动
| 类别 | 已支持 |
|---|---|
| 邮件 | SMTP · 阿里云 DirectMail · Log |
| 社交登录 | GitHub · Google · 微信 · QQ · 钉钉 · Gitee · 微博 |
| 图片处理 | Local（imaging，缩略图/水印；自定义 HTTP 未实现） |

### 🗃️ 数据库
- sqlx + 手写 SQL 覆盖 **30+ 张业务表**
- 默认 **SQLite** 轻量化，深度支持 **MySQL**（推荐 MariaDB 10.6.20+）
- 迁移 CLI：`migrate up / down / status / seed`

## 🏗️ 技术栈

### 后端
| 技术 | 用途 |
|---|---|
| **Rust 1.85+** | 主语言 |
| **axum** | HTTP 框架 |
| **sqlx** |异步 SQL（SQLite/MySQL/PostgreSQL） |
| **serde** | 序列化（YAML + 环境变量配置） |
| **tokio** | 异步运行时 |
| **tokio::sync::mpsc** | 异步图片处理队列 |
| **Redis** | 可选缓存层 |
| **jsonwebtoken** | JWT |
| **argon2** | 密码哈希 |
| **utoipa** | OpenAPI/Swagger 文档 |

### 前端
| 技术 | 用途 |
|---|---|
| **Astro 5** | SSR 框架（node standalone） |
| **Vue 3.5 Islands** | 公开页（首页/探索/分享/登录） |
| **React 19 Islands** | 用户中心 + 管理后台 |
| **shadcn-vue / shadcn/ui** | UI 组件库（Radix 原语） |
| **Tailwind CSS 3.4** | 原子化样式 + CSS 变量主题（双框架共享 token） |
| **class-variance-authority** | 组件变体管理 |
| **@lucide/vue / lucide-react** | 图标库 |
| **Pinia / Zustand** | 状态管理（Vue / React） |
| **PhotoSwipe** | 图片灯箱（框架无关） |

### DevOps
| 技术 | 用途 |
|---|---|
| **Docker** + **Docker Compose** | 容器化 |
| **GitHub Actions** | CI（fmt / clippy / test / build / Docker 多架构） |
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

# 数据库迁移（sqlx migrate）
cd server-rust && sqlx migrate run
```

**默认管理员**：`admin` / `admin123456`（请尽快修改）

### 方式 B：本地开发

```bash
# ---- 后端 ----
cd server-rust
cargo build --bin api      # 首次编译
cargo run --bin api        # 启动 API（默认 :3000）

# ---- 前端（独立开发模式）----
cd web-astro
npm install
npm run dev                # 启动 Web（默认 :4321，/api 反代到 :3000）
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
├── server-rust/               # Rust 后端
│   ├── crates/
│   │   ├── api/              # 薄二进制壳（main.rs）
│   │   └── core_lib/         # 业务核心库
│   │       ├── src/
│   │       │   ├── handlers/ # HTTP 处理器（22 个域）
│   │       │   ├── services/ # 业务服务层
│   │       │   ├── models/   # 领域模型
│   │       │   ├── dto/      # 请求/响应 DTO
│   │       │   ├── middleware/# auth/cors/rate_limit
│   │       │   ├── auth/     # JWT + RBAC + 密码哈希
│   │       │   └── utils/    # 响应信封/分页/验证码
│   │       └── tests/        # 集成测试
│   ├── config.yaml            # 统一配置
│   └── Dockerfile
├── web-astro/                   # Astro 5 前端（Vue/React Islands）
│   ├── src/pages/               # 公共端 + 用户中心 + 管理后台
│   ├── src/layouts/             # Public / Dashboard / Admin 布局
│   ├── src/components/
│   │   ├── vue/                 # Vue Islands 组件（公开页 + ui/）
│   │   └── react/               # React Islands 组件（用户中心/后台 + ui/）
│   ├── src/lib/                 # 共享 API 客户端 / 认证 / 工具
│   ├── src/stores/              # Pinia / Zustand 状态
│   ├── src/middleware.ts        # 路由守卫 + 生产环境 API 反代
│   └── astro.config.mjs
├── deploy/                    # 部署脚本
│   ├── deploy.sh             # 一键编译（cargo）+ 构建 + 推送
│   ├── dev-up.sh / dev-down.sh
│   ├── backup-db.sh          # 多数据库备份
│   ├── docker-compose.yml    # 开发环境
│   ├── nginx.conf            # 反向代理
│   └── supervisor.conf       # 进程守护
├── .github/workflows/         # CI / Release
├── docker-compose.yml
├── .env.example
├── Dockerfile                # 合并镜像（API + Web）
└── README.md
```

### 开发规范

- 后端遵循 **Rust API Guidelines**，提交前跑：
  ```bash
  cargo fmt --all -- --check   # 格式检查
  cargo clippy -- -D warnings  # 静态检查
  cargo test --workspace       # 测试
  ```
- 前端遵循 **Astro + TypeScript 规范**，提交前跑：
  ```bash
  npx astro check
  npm run build
  ```
- 提交信息遵循 **Conventional Commits**（`feat:` / `fix:` / `docs:` / `refactor:` / `test:` / `chore:`）

### 调试技巧

```bash
# 查看 API 健康
curl http://localhost:3000/api/v1/healthz

# 实时日志
docker compose logs -f ywty

# 进入容器调试
docker compose exec ywty sh
```

## 🧪 测试

```bash
# 后端
cd server-rust
cargo test --workspace --verbose
cargo tarpaulin --out Html   # 查看覆盖率（需安装 cargo-tarpaulin）

# 前端 E2E（待补）
# 暂无前端测试脚本
```

当前覆盖：认证、专辑、用户、图片等核心域集成测试（13 项）+ 单元测试（9 项）。

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
| P8 | 公共端（Vue Islands） | ✅ |
| P9 | 用户中心（React Islands） | ✅ |
| P10 | 管理后台（React Islands） | ✅ |
| P11 | 测试与部署 | ✅ |
| P12 | 前端迁移 Astro Islands | ✅ |
| v1.0 | 正式版 | 🎯 2026 Q4 |
| 未来 | 移动端 App / 多租户 / 联邦 | 💭 |

## 📜 许可证

本项目采用 [MIT License](./LICENSE.md) 开源。

## 🙏 致谢

- [Lsky Pro+](https://www.lsky.pro) — 原始项目作者及其团队
- 所有贡献者与社区用户
- ⭐ **如果觉得不错，给我们一个 Star 吧！** ⭐
