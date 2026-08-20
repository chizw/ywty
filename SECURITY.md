# 安全策略

## ⚠️ 请勿在公开 Issue 中披露安全漏洞

如果你发现了潜在的安全问题（包括但不限于越权访问、注入漏洞、敏感信息泄露等），**请不要在 GitHub Issue / Discussion / PR 中公开**。请按下面的方式私下报告。

## 📮 报告方式

- **邮件**：`security@chizw.dev`（首选）
- **GitHub Private Vulnerability Reporting**：
  <https://github.com/chizw/ywty/security/advisories/new>

请在报告中尽量提供：

1. 问题简述
2. 复现步骤 / PoC（截图、curl 命令、Payload）
3. 受影响的版本 / Commit
4. 你的环境信息（部署方式、配置）
5. 联系方式（以便我们反馈修复进展）

我们会在 **48 小时内**确认收到你的报告，并在 **7 天内**给出初步评估和修复时间表。

## 🔐 支持的版本

| 版本 | 是否支持安全更新 |
|---|---|
| v3.x（重构版） | ✅ |
| v2.x（Lsky Pro+） | ❌ 请升级到 v3 |
| v1.x（Lsky Pro+） | ❌ |

## 🛡️ 安全实践建议

部署本项目时建议：

- **使用 HTTPS**：生产环境必须通过 Nginx / Caddy 终止 TLS
- **修改默认密钥**：
  - `YWTY_AUTH_JWT_SECRET` 必须改为强随机值
  - 数据库密码、Redis 密码、各云服务 AK/SK 不允许使用默认值
- **限制管理后台访问**：通过 Nginx 限制 `/admin` 来源 IP
- **启用速率限制**：`YWTY_RATELIMIT_ENABLE=true`（默认开启）
- **定期备份数据库**：`sqlite` 部署务必做文件系统级备份
- **最小化驱动权限**：S3 / OSS 等驱动使用子账号 + 最小权限策略
- **保持依赖更新**：`go get -u` 与 `npm update` 定期执行

## 🏆 致谢

我们感谢以下安全研究者的负责任披露（按时间顺序）：

_（暂无记录，欢迎成为第一个贡献者）_

## 📜 历史漏洞

无

---

再次感谢你帮助让 ywty 更安全！🙏
