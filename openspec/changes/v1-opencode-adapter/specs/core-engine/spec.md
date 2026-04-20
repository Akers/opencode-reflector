## ADDED Requirements

### Requirement: Core Engine HTTP API 服务
Core Engine SHALL 作为独立 Go 进程运行，通过 HTTP localhost 暴露 RESTful API，监听端口可通过配置文件修改（默认 19870）。

#### Scenario: Adapter 触发反思分析
- **WHEN** Adapter 向 `POST /api/v1/reflect` 发送包含 CanonicalSession 数组和配置的 JSON 请求
- **THEN** Core Engine 执行完整的反思流程并返回分析结果（状态、分析会话数、日报路径、Token 消耗、耗时）

#### Scenario: 健康检查
- **WHEN** 任何客户端向 `GET /api/v1/health` 发送请求
- **THEN** 返回 HTTP 200 和服务状态 JSON

#### Scenario: 运行统计
- **WHEN** 任何客户端向 `GET /api/v1/stats` 发送请求
- **THEN** 返回反思工具自身的累计运行统计（总反思次数、总 Token 消耗、总分析会话数）

### Requirement: 触发管理
Core Engine SHALL 支持定时触发（TIME）和手动触发（MANUAL）两种模式。定时触发在用户配置的固定时间执行；手动触发由 Adapter 转发用户的 `/or:reflect_now` 指令。

#### Scenario: 定时触发防抖
- **WHEN** 定时触发在 5 分钟内被重复调用
- **THEN** 系统忽略重复触发，仅执行一次反思

#### Scenario: 并发触发排队
- **WHEN** 反思引擎正在运行时收到新的触发请求
- **THEN** 新请求排队等待，不并发执行

### Requirement: Session 解析与指标提取
Core Engine SHALL 接收 CanonicalSession 格式的数据，提取 7 类共 25 项监控指标（时间 M-001~004、Token M-005~008、工具/MCP/Skill M-009~014、智能体参与 M-015~016、消息统计 M-017~019、人类参与度 M-020~022、情感 M-023~025）。不可获取的指标 SHALL 标注为 N/A。

#### Scenario: 完整指标提取
- **WHEN** 收到一个包含 100 条消息的完整 CanonicalSession
- **THEN** 系统在 2 秒内提取所有可用指标并写入 SQLite

#### Scenario: Token 数据不可用
- **WHEN** CanonicalSession 中消息的 `promptTokens` 和 `completionTokens` 字段为空
- **THEN** 对应的 M-005~M-008 指标值记录为 N/A（-1）

### Requirement: 任务状态多级判定
Core Engine SHALL 实现三级任务状态判定策略（L1 > L2 > L3），判定结果为 COMPLETED / INTERRUPTED / UNCERTAIN / ABANDONED 之一。

#### Scenario: L1 判定 - OpenSpec 指令
- **WHEN** 会话中检测到用户发送 `/opsx:archive` 指令且智能体成功执行
- **THEN** 任务状态判定为 COMPLETED

#### Scenario: L2 判定 - LLM 意图分类
- **WHEN** L1 未命中且反思模型可用
- **THEN** 调用配置的模型分析最后一条 Agent 消息，返回结构化 JSON 判定结果

#### Scenario: L3 判定 - 关键词匹配兜底
- **WHEN** L1 和 L2 均未命中（或模型不可用）
- **THEN** 使用中英文关键词匹配进行兜底判定

### Requirement: 情感分析模块
Core Engine SHALL 对会话中所有人类消息执行情感分析，计算负面情绪占比（M-023）、态度评分（M-024）、认可度（M-025）。情感分析可通过配置完全禁用。

#### Scenario: 情感分析正常执行
- **WHEN** 情感分析已启用且模型可用，收到包含 20 条人类消息的会话
- **THEN** 在 5 秒内返回情感分析结果，Token 消耗记入 `reflector_tokens`

#### Scenario: 情感分析被禁用
- **WHEN** 配置中 `sentiment.enabled` 为 false
- **THEN** 跳过情感分析，M-023~M-025 标记为 N/A

#### Scenario: 情感分析模型不可用
- **WHEN** 情感分析模型调用超时或失败
- **THEN** 跳过情感分析，记录错误日志，M-023~M-025 标记为 N/A

### Requirement: SQLite 数据持久化
Core Engine SHALL 使用 SQLite 数据库（`.reflector/data/reflector.db`）持久化所有指标数据，按设计文档 DDL 建表（sessions、agent_participations、tool_calls、reflection_logs、watermarks），并创建必要索引。

#### Scenario: 数据库首次初始化
- **WHEN** Core Engine 首次启动且数据库文件不存在
- **THEN** 自动创建 `.reflector/data/` 目录和数据库文件，执行建表 DDL

#### Scenario: 数据库写入失败降级
- **WHEN** SQLite 写入失败且重试 3 次仍不成功
- **THEN** 将数据写入 fallback JSON 文件（`.reflector/data/fallback/`），记录错误日志

### Requirement: 日报生成
Core Engine SHALL 按自然日生成 Markdown 格式的反思日报，存储在 `.reflector/reports/dayreport-<date>.md`。同一日期的日报追加更新，不覆盖已有内容。

#### Scenario: 定时触发生成日报
- **WHEN** 零点定时触发反思，分析 10 个会话
- **THEN** 生成前一天日期的日报文件，包含任务清单、汇总统计、工具使用统计、反思工具自身开销

#### Scenario: 日报幂等追加
- **WHEN** 同一日期已有日报文件，再次触发反思
- **THEN** 追加新分析的会话数据到已有日报，不覆盖已有内容

### Requirement: 配置管理
Core Engine SHALL 从 `.reflector/reflector.yaml` 加载 YAML 配置，支持热更新（下次触发时自动重新加载）。配置文件不存在时使用内置默认值。

#### Scenario: 配置文件不存在
- **WHEN** `.reflector/reflector.yaml` 文件不存在
- **THEN** 使用内置默认配置启动，并在日志中输出 INFO 级别提示

#### Scenario: 配置热更新
- **WHEN** 运行期间修改了 `reflector.yaml` 文件
- **THEN** 下次反思触发时自动加载新配置，无需重启 Core Engine

### Requirement: 提示词管理
Core Engine SHALL 支持从 `.reflector/prompts/` 目录加载 LLM 提示词（`.md` 文件），覆盖内置默认提示词。文件不存在或为空时使用内置默认值并输出 WARNING。

#### Scenario: 自定义提示词加载
- **WHEN** `.reflector/prompts/sentiment-analysis.md` 文件存在且内容非空
- **THEN** 使用该文件内容作为情感分析提示词

#### Scenario: 提示词文件缺失降级
- **WHEN** `.reflector/prompts/sentiment-analysis.md` 文件不存在
- **THEN** 使用编译时嵌入的默认提示词，并在日志中输出 WARNING

### Requirement: Hook 执行器
Core Engine SHALL 在 10 个生命周期节点执行 `.reflector/hooks/` 目录下的脚本式 Hook。通过 stdin 传入 JSON，stdout 接收 JSON，exit code 标识成功/失败。Hook 失败不阻断主流程。

#### Scenario: Hook 正常执行
- **WHEN** `before-save-metrics` 节点存在对应的 Hook 脚本且执行成功
- **THEN** Hook 返回的 JSON 数据用于修改后续流程的输入

#### Scenario: Hook 执行失败
- **WHEN** Hook 脚本 exit code 非零
- **THEN** 记录 WARNING 日志，继续执行主流程，不阻断

### Requirement: 数据保留与清理
Core Engine SHALL 提供数据清理能力，默认保留 90 天，通过 `/or:cleanup --days N` 命令触发清理。

#### Scenario: 执行数据清理
- **WHEN** 用户发送 `/or:cleanup --days 30` 指令
- **THEN** 删除 30 天前的数据库记录和对应的日报文件
