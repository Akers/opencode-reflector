## Context

opencode-reflector 是一个 AI 编程智能体反思监控插件，设计文档已定义 Core Engine(Go) + Adapter(Node.js) 的双进程架构。当前项目处于从零起步阶段，没有任何代码实现。

本变更实现首个可用版本（v1），聚焦于 opencode 工具的完整适配，同时构建好 Adapter 基础设施以便后续快速接入 openclaw 和 claudecode。

**关键约束**：
- opencode 的插件机制为 Node.js（`@opencode-ai/plugin`），核心逻辑用 Go 实现
- 不修改 opencode/openclaw/claudecode 的源码
- 反思工具自身的 Token 消耗需单独统计
- 所有数据存储在本地 `.reflector/` 目录

## Goals / Non-Goals

**Goals:**

- 跑通 opencode 会话的完整反思流程：触发 → 获取 Session → 指标提取 → 任务判定 → 情感分析 → 持久化 → 日报生成
- 实现 Core Engine 作为独立 Go HTTP 服务运行
- 实现 opencode Adapter 作为 opencode 插件运行
- 定义清晰的 Adapter 接口契约，使新增一个工具的适配工作量最小化
- 支持定时触发（TIME）和手动触发（MANUAL）两种模式
- 实现配置文件（YAML）加载与热更新
- 实现脚本式 Hook 机制

**Non-Goals:**

- 不适配 openclaw 和 claudecode（仅预留接口）
- 不实现事件触发中的 `N_MESSAGES` 模式（仅实现 `TASK_FINISHED` 的框架）
- 不实现周报/月报/趋势分析（Phase 4）
- 不实现多项目支持（单 `.reflector` 目录）
- 不做 Web UI 或可视化仪表盘
- 不实现跨工具对比报告

## Decisions

### D1：Core Engine 端口与生命周期管理

**决策**：Core Engine 监听 `localhost:19870`（可通过配置修改），由 opencode Adapter 管理其生命周期——Adapter 插件初始化时按需启动 Core Engine 进程，插件卸载时优雅关闭。

**理由**：
- 使用固定端口简化配置，用户无需手动管理
- Adapter 作为入口，天然适合管理 Core Engine 的启停
- 按需启动避免后台常驻占用资源

**替代方案**：
- Unix Socket：性能更好但跨平台兼容性差，调试不便
- Core Engine 独立常驻服务：增加用户运维负担

### D2：SQLite 驱动选择

**决策**：使用 `modernc.org/sqlite`（纯 Go 实现，无 CGO 依赖）。

**理由**：
- 跨平台编译无需安装 C 工具链
- 单一二进制分发，用户零配置
- 性能对本项目数据量（日均 10 会话）完全足够

**替代方案**：
- `mattn/go-sqlite3`：需 CGO，编译和交叉发布复杂
- JSON 文件存储：查询能力弱，不利于未来趋势分析

### D3：Go HTTP 框架选择

**决策**：使用标准库 `net/http` + 轻量路由（`go-chi/chi`）。

**理由**：
- chi 是最轻量的 Go 路由库之一，兼容标准库 `http.Handler`
- 无额外抽象，便于维护
- 本项目 API 端点少（约 5 个），无需重框架

### D4：情感分析模型调用方式

**决策**：Core Engine 通过 OpenAI 兼容 API 调用用户配置的模型（如 MiniMax-M2.7），使用 OpenAI Go SDK。

**理由**：
- OpenAI 兼容 API 是业界标准，覆盖绝大多数模型供应商
- 用户通过 `reflector.yaml` 配置模型 ID 和 API endpoint
- 支持本地模型（如 Ollama）和云端模型

### D5：Adapter 与 Core Engine 的数据序列化

**决策**：使用 JSON over HTTP，CanonicalSession 定义为 JSON Schema。

**理由**：
- Go 和 Node.js 对 JSON 的支持都是原生的
- 可读性好，便于调试
- 无需引入 protobuf 等额外依赖

### D6：opencode Session 数据获取方式

**决策**：通过 `@opencode-ai/sdk` 提供的 API 获取 Session 数据。如果 SDK 不直接暴露 Token 统计，则在能力矩阵中将 Token 指标标记为 `N/A`。

**理由**：
- 遵循"不修改源码"约束
- 诚实标注不可获取的数据，优于猜测或 hack

**风险**：需 PoC 验证 SDK 能提供哪些数据。如果 SDK 能力严重不足，可能需要降级为读取 opencode 本地存储文件（如 SQLite DB）。

### D7：任务状态判定的默认策略

**决策**：v1 实现完整的 L1/L2/L3 三级策略。L2 调用配置的反思模型，L3 使用中英文关键词匹配。

**理由**：
- L1（OpenSpec 指令）本项目自身就使用 OpenSpec，可立即验证
- L2 是主要判定方式，准确度高
- L3 作为兜底，确保模型不可用时系统仍能运行

## Risks / Trade-offs

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| opencode SDK 不暴露 Token 统计 | M-005~M-008 指标不可获取 | 能力矩阵标注 N/A；降级为读取本地存储文件 |
| Core Engine 启动失败 | 整个反思流程不可用 | Adapter 实现重试机制（3次，指数退避），失败后记录日志并静默降级 |
| 情感分析模型 API 不可用 | 情感指标缺失 | 跳过情感分析，标记 N/A，使用 L3 关键词匹配降级 |
| 固定端口冲突 | Core Engine 无法启动 | 端口可配置，启动时检测冲突并报错 |
| Go + Node.js 双运行时增加用户负担 | 安装和部署复杂度 | 提供一键安装脚本；Go 编译为单一二进制，用户无需 Go 环境 |
| 情感分析自身的 Token 消耗 | 监控工具本身成为 Token 消耗者 | 日报中单独统计反思工具消耗；情感分析可配置禁用 |
