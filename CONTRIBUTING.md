# 贡献指南

感谢你愿意为 **ywty** 做出贡献！🎉

本文档说明如何参与项目开发、提 Issue、提 Pull Request。开始前请先阅读 [README.md](./README.md) 和 [AGENTS.md](./AGENTS.md) 以了解项目结构与路线图。

## 📜 行为准则

请阅读并遵守 [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)，我们期望所有参与者保持友好与专业。

## 🐛 报告 Bug

1. 先在 [Issues](https://github.com/chizw/ywty/issues) 搜索是否已有相关报告
2. 使用 [Bug Report 模板](./.github/ISSUE_TEMPLATE/bug_report.yml) 提交
3. 描述应包含：
   - 复现步骤（尽量具体）
   - 期望行为 vs 实际行为
   - 截图 / 视频 / 日志
   - 环境信息：OS、Rust 版本、Node 版本、数据库、部署方式

## 💡 提议新功能

1. 先在 [Discussions](https://github.com/chizw/ywty/discussions) 发起讨论
2. 确认方向后使用 [Feature Request 模板](./.github/ISSUE_TEMPLATE/feature_request.yml) 提交 Issue
3. 等待维护者反馈后再开始编码

## 🔧 提交 Pull Request

1. **Fork** 本仓库并创建分支
   ```bash
   git checkout -b feat/your-feature
   # 或
   git checkout -b fix/your-bug
   ```
2. **遵循代码规范**（见下文）
3. **补充测试**（如果是 bug 修复或新功能）
4. **本地跑通**
   ```bash
   cd server-rust
cargo fmt --all -- --check && cargo clippy --workspace --all-targets -- -D warnings && cargo test --workspace
   cd web-astro && npm run build
   ```
5. **提交** — 建议遵循 [Conventional Commits](https://www.conventionalcommits.org/zh-hans/) 规范
6. **推送并创建 PR** — 使用 [PR 模板](./.github/PULL_REQUEST_TEMPLATE.md)

### Commit 消息格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

常用 type：
- `feat`：新功能
- `fix`：bug 修复
- `docs`：文档变更
- `style`：代码格式（不影响功能）
- `refactor`：重构
- `perf`：性能优化
- `test`：测试
- `chore`：构建 / 工具 / 依赖

示例：
```
feat(photo): 支持 HEIC 格式上传
fix(auth): 修复刷新令牌过期后未自动登出的问题
docs(readme): 更新 Docker Compose 启动说明
```

## 🧑‍💻 代码规范

### 后端（Rust）

- Rust ≥ 1.85，使用 `cargo fmt` + `clippy`
- 提交前跑 `cargo clippy --workspace -- -D warnings`
- 业务错误请使用 `crates/core_lib/src/error.rs` 中的整数错误码
- 模型变更需同步更新 SQL 迁移（`db/migrations/`）
- 优先使用 trait + 注册表模式扩展驱动（存储驱动、支付驱动等）
- 敏感配置（密钥、AK/SK）必须通过环境变量传入，不允许硬编码

### 前端（Astro 7 + Vue/React Islands + TypeScript）

- TypeScript 严格模式，避免 `any`
- 公开页用 Vue Islands（`src/components/vue/`），用户中心/后台用 React Islands（`src/components/react/`）
- 页面放在 `src/pages/`，组件按框架放在 `src/components/{vue,react}/`
- 状态管理：Vue 用 Pinia，React 用 Zustand（`src/stores/`）
- 提交前跑 `npx astro check` 和 `npm run build`

### 数据库

- 修改模型时同步在迁移文件里写出"before/after"列
- 不要直接修改已经合并的迁移文件，新增一个迁移来变更

## 🧪 测试

- 后端单元测试覆盖率目标 ≥ 70%
- 涉及 API 变更请补充 HTTP 集成测试
- 前端组件可使用 Vitest

## 📦 提交 PR 前的清单

- [ ] 代码遵循项目规范
- [ ] 已添加 / 更新单元测试
- [ ] 已更新相关文档（README / CHANGELOG）
- [ ] `cargo build --workspace` 通过
- [ ] `npm run build` 通过
- [ ] PR 描述清晰，对应 Issue 已链接

## 💬 联系方式

- 一般问题：[Discussions](https://github.com/chizw/ywty/discussions)
- 安全问题：见 [SECURITY.md](./SECURITY.md)（**不要**在公开 Issue 中讨论）

---

再次感谢你的贡献！❤️
