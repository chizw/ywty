//! ywty Rust API 二进制入口
//!
//! 这是一个"薄二进制壳"：所有业务逻辑（配置加载、状态构建、迁移、路由、
//! 服务启动）均位于 `core_lib` 库的 [`core_lib::app::run`] 中。
//!
//! 保持本 crate 代码量极小的原因：Rust 1.97/LLVM 22 在大体积二进制 crate
//! 的最终代码生成阶段会触发 STATUS_ACCESS_VIOLATION；将代码下沉到库 crate
//! 可规避该问题（库 crate 走不同的代码生成路径，不受影响）。

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    core_lib::app::run().await
}
