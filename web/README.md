# ywty 用户端前端

Vue 3 + TypeScript + Pinia + Naive UI + TailwindCSS，基于 Vite 构建。

## 说明

- 构建产物输出到 `dist/`（由 Go 服务托管）；`npm run build-only` 跳过类型检查。
- 可以使用 `npm run openapi-ts` 命令通过 `openapi.json` 文件生成接口服务代码和枚举类。

### 构建

```sh
npm install
npm run build
```

### 热重载开发

```sh
npm run dev
```
