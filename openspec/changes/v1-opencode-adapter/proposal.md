## Why

opencode-reflector 项目的需求与设计方案已完成，现在需要实现首个可运行的版本。当前没有任何代码，需要从零构建 Core Engine 和 opencode Adapter，验证 Core + Adapter 架构的可行性，并实现基本的反思监控能力。

## What Changes

- **新建 Go Core Engine**：实现 HTTP API 服务，接收标准化 Session 数据，执行指标提取、任务状态判定、情感分析、数据持久化和日报生成
- **新建 opencode Adapter（Node.js 插件）**：对接 opencode 的 `@opencode-ai/plugin` 插件机制，监听触发事件，获取 Session 数据，转换为 CanonicalSession 格式后调用 Core Engine
- **新建 Adapter 基础设施**：定义 `SessionAdapter` 接口契约、`CanonicalSession` 统一数据模型、`CapabilityMatrix` 能力矩阵——为 openclaw/claudecode 的未来适配提供扩展点
- **SQLite 持久化**：按设计文档 DDL 建表，实现会话、工具调用、反思日志的 CRUD
- **配置管理**：实现 `reflector.yaml` 加载与热更新
- **日报生成**：按模板生成 Markdown 格式的反思日报
- **提示词管理**：内置默认提示词，支持 `.reflector/prompts/` 目录热覆盖
- **Hook 机制**：实现脚本式 Hook 执行器，支持 10 个生命周期节点

## Capabilities

### New Capabilities

- `core-engine`: Go HTTP 服务，包含触发管理、Session 解析、指标提取、任务状态判定（L1/L2/L3）、情感分析、日报生成、数据持久化（SQLite）、Hook 执行等核心模块
- `adapter-infrastructure`: Adapter 接口契约定义（SessionAdapter、CanonicalSession、CapabilityMatrix 等类型），作为多工具适配的公共基础层
- `opencode-adapter`: 基于 `@opencode-ai/plugin` 的 opencode 适配器实现，负责触发监听、Session 获取与格式转换、Core Engine 调用

### Modified Capabilities

（无现有能力需要修改）

## Impact

- **新增代码**：Go 项目（`cmd/`、`internal/`）+ Node.js 项目（`adapters/opencode/`），总计约 15-20 个新文件
- **新增依赖**：Go 侧依赖 SQLite 驱动（如 `modernc.org/sqlite`）、HTTP 路由库；Node.js 侧依赖 `@opencode-ai/plugin`、`@opencode-ai/sdk`
- **运行时影响**：Core Engine 作为独立进程通过 HTTP localhost 暴露 API，Adapter 作为 opencode 插件加载并管理 Core Engine 生命周期
- **配置与数据**：引入 `.reflector/` 目录（配置、数据、日志、报告、提示词、Hooks）
