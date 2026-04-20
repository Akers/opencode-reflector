## ADDED Requirements

### Requirement: SessionAdapter 接口契约
系统 SHALL 定义 `SessionAdapter` TypeScript 接口，作为所有智能体工具适配器的公共契约。接口包含 `getToolType()`、`getSessions(since)`、`onTrigger(event, callback)`、`getCapabilityMatrix()` 四个方法。

#### Scenario: 新工具接入实现接口
- **WHEN** 开发者为新智能体工具创建 Adapter
- **THEN** 只需实现 `SessionAdapter` 接口的四个方法即可完成基础接入

### Requirement: CanonicalSession 统一数据模型
系统 SHALL 定义 `CanonicalSession` TypeScript 接口，作为跨工具的统一会话数据模型。包含 `id`、`toolType`、`title`、`messages`（CanonicalMessage 数组）、`toolCalls`（CanonicalToolCall 数组）、`metadata` 字段。

#### Scenario: Adapter 转换 Session 数据
- **WHEN** opencode Adapter 获取到原生 Session 数据
- **THEN** 将其转换为 CanonicalSession 格式后发送给 Core Engine

#### Scenario: CanonicalMessage 包含可选 Token 字段
- **WHEN** 原生消息数据包含 Token 计数
- **THEN** 映射到 CanonicalMessage 的 `promptTokens` 和 `completionTokens` 字段

#### Scenario: 原生消息无 Token 数据
- **WHEN** 原生消息数据不包含 Token 计数
- **THEN** `promptTokens` 和 `completionTokens` 为 undefined，Core Engine 将对应指标标记为 N/A

### Requirement: CapabilityMatrix 能力矩阵
系统 SHALL 定义 `CapabilityMatrix` TypeScript 接口，声明每个 Adapter 能提供哪些数据指标。包含 `tokenMetrics`、`toolCallDetails`、`mcpCallDetails`、`skillCallDetails`、`agentNames` 五个布尔字段。

#### Scenario: opencode 能力矩阵声明
- **WHEN** opencode Adapter 初始化时
- **THEN** 根据实际 SDK 能力填写 CapabilityMatrix（Token 指标取决于 PoC 验证结果）

#### Scenario: 未来工具能力差异
- **WHEN** openclaw Adapter 接入且其 SDK 不支持 MCP 调用详情
- **THEN** `mcpCallDetails` 为 false，Core Engine 将 MCP 相关指标标记为 N/A

### Requirement: Adapter 公共类型定义包
系统 SHALL 将 SessionAdapter、CanonicalSession、CanonicalMessage、CanonicalToolCall、CapabilityMatrix 等类型定义发布为共享的 TypeScript 包（`adapters/shared/`），供所有 Adapter 复用。

#### Scenario: 多 Adapter 复用类型
- **WHEN** openclaw Adapter 项目创建时
- **THEN** 可直接 import 共享包中的类型定义，无需重复定义

### Requirement: Adapter 与 Core Engine 的通信契约
所有 Adapter SHALL 通过 HTTP POST `/api/v1/reflect` 将 CanonicalSession 数组和配置发送给 Core Engine，接收标准化的 JSON 响应。

#### Scenario: 请求格式标准化
- **WHEN** 任何 Adapter 向 Core Engine 发送反思请求
- **THEN** 请求体 SHALL 包含 `trigger_type`、`trigger_detail`、`sessions`（CanonicalSession 数组）、`config` 字段

#### Scenario: 响应格式标准化
- **WHEN** Core Engine 完成反思分析
- **THEN** 返回包含 `status`、`sessions_analyzed`、`report_path`、`tokens_consumed`、`duration_ms` 的 JSON 响应
