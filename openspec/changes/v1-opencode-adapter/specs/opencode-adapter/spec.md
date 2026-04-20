## ADDED Requirements

### Requirement: opencode 插件注册
opencode Adapter SHALL 作为 opencode 插件注册，使用 `@opencode-ai/plugin` 提供的插件 API 实现初始化、触发监听和指令注册。

#### Scenario: 插件加载初始化
- **WHEN** opencode 启动并加载本插件
- **THEN** Adapter 初始化，按需启动 Core Engine 进程，注册 `/or:reflect_now` 和 `/or:cleanup` 指令

#### Scenario: 插件卸载清理
- **WHEN** opencode 关闭或卸载本插件
- **THEN** Adapter 向 Core Engine 发送优雅关闭信号，等待进程退出

### Requirement: Core Engine 生命周期管理
opencode Adapter SHALL 负责 Core Engine 进程的启动、健康检查和关闭。启动时使用 `go build` 编译的单一二进制文件。

#### Scenario: 按需启动 Core Engine
- **WHEN** Adapter 初始化且 Core Engine 进程未运行
- **THEN** 启动 Core Engine 二进制，等待 `/health` 端点返回 200 后标记为就绪

#### Scenario: Core Engine 启动失败
- **WHEN** Core Engine 进程启动失败或健康检查超时（10秒）
- **THEN** 重试 3 次（指数退避），仍失败则记录错误日志并静默降级（插件不阻断 opencode 运行）

#### Scenario: Core Engine 已在运行
- **WHEN** Adapter 初始化且检测到 Core Engine 进程已在运行（健康检查通过）
- **THEN** 直接复用现有进程，不重复启动

### Requirement: Session 数据获取
opencode Adapter SHALL 通过 `@opencode-ai/sdk` 获取自上次反思以来的所有新增会话数据，并转换为 CanonicalSession 格式。

#### Scenario: 增量获取会话
- **WHEN** 触发反思时
- **THEN** Adapter 读取持久化的 watermark 时间戳，获取该时间之后的所有会话，转换为 CanonicalSession 数组

#### Scenario: SDK 不提供 Token 数据
- **WHEN** opencode SDK 的消息对象不包含 Token 计数字段
- **THEN** CanonicalMessage 的 `promptTokens` 和 `completionTokens` 设为 undefined，CapabilityMatrix 的 `tokenMetrics` 设为 false

### Requirement: 定时触发实现
opencode Adapter SHALL 使用 opencode 插件的定时能力（或内部 setInterval）在配置的时间点自动触发反思。

#### Scenario: 零点定时触发
- **WHEN** 配置的触发时间为 `"00:00"` 且到达该时间点
- **THEN** Adapter 获取自上次反思以来的会话数据，调用 Core Engine 的 `/api/v1/reflect` 接口

### Requirement: 手动触发指令
opencode Adapter SHALL 注册 `/or:reflect_now` 指令，用户发送该指令时立即触发反思。

#### Scenario: 用户手动触发反思
- **WHEN** 用户在 opencode 中发送 `/or:reflect_now` 指令
- **THEN** Adapter 立即获取当前会话及自上次反思以来的所有未分析会话，调用 Core Engine

### Requirement: 清理指令
opencode Adapter SHALL 注册 `/or:cleanup` 指令，支持 `--days N` 参数，委托 Core Engine 执行数据清理。

#### Scenario: 用户触发数据清理
- **WHEN** 用户发送 `/or:cleanup --days 30`
- **THEN** Adapter 调用 Core Engine 的清理 API，删除 30 天前的数据

### Requirement: Adapter 配置读取
opencode Adapter SHALL 从 `.reflector/reflector.yaml` 读取配置（复用 Core Engine 的配置文件），获取触发时间、模型配置等信息。

#### Scenario: 配置文件共享
- **WHEN** Adapter 启动时
- **THEN** 读取 `.reflector/reflector.yaml` 获取 Core Engine 端口、触发时间等配置，与 Core Engine 共享同一配置文件
