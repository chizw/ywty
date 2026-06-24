# 贡献指南

感谢你愿意为 **ywty** 做出贡献！🎉

本文档说明如何参与项目开发、提 Issue、提 Pull Request。开始前请先阅读 [README.md](./README.md) 和 [REFACTOR_PLAN.md](./REFACTOR_PLAN.md) 以了解项目结构与路线图。

## 📜 行为准则

请阅读并遵守 [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)，我们期望所有参与者保持友好与专业。

## 🐛 报告 Bug

1. 先在 [Issues](https://github.com/scfcn/ywty/issues) 搜索是否已有相关报告
2. 使用 [Bug Report 模板](./.github/ISSUE_TEMPLATE/bug_report.yml) 提交
3. 描述应包含：
   - 复现步骤（尽量具体）
   - 期望行为 vs 实际行为
   - 截图 / 视频 / 日志
   - 环境信息：OS、Go 版本、Node 版本、数据库、部署方式

## 💡 提议新功能

1. 先在 [Discussions](https://github.com/scfcn/ywty/discussions) 发起讨论
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
   cd server && go build ./... && go vet ./...
   cd web-nuxt && npm run build
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

### 后端（Go）

- Go ≥ 1.25，使用 `gofmt` + `goimports`
- 提交前跑 `go vet ./...`
- 业务错误请使用 [`internal/errors`](./server/internal/errors) 的业务错误码
- 模型变更必须同步更新 GORM 迁移
- 优先使用 [接口+注册表](https://github.com/scfcn/ywty/blob/main/REFACTOR_PLAN.md) 模式扩展驱动
- 敏感配置（密钥、AK/SK）必须通过环境变量传入，不允许硬编码

### 前端（Nuxt 3 + TypeScript）

- TypeScript 严格模式，避免 `any`
- 组件命名采用 PascalCase（`<AppButton>`、`<AppUploader>`）
- 页面放在 `pages/`，跨页组件放在 `components/`
- 全局状态使用 Pinia Store
- 提交前跑 `npm run build`

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
- [ ] `go build ./...` 通过
- [ ] `npm run build` 通过
- [ ] PR 描述清晰，对应 Issue 已链接

## 💬 联系方式

- 一般问题：[Discussions](https://github.com/scfcn/ywty/discussions)
- 安全问题：见 [SECURITY.md](./SECURITY.md)（**不要**在公开 Issue 中讨论）

---

再次感谢你的贡献！❤️
