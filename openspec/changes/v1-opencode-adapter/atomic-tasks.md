# TDD 原子任务清单

> 将 `tasks.md` 中的 46 个任务拆分为 **2~5 分钟** 可完成的 TDD 原子任务。
> 每个任务遵循 **Red → Green → Verify** 循环：先写测试，再实现，最后验证通过。

---

## 原则

1. **一个原子任务 = 一个测试用例 + 一个最小实现**，完成后 `go test` 必须通过
2. 严格依赖顺序：上层任务依赖下层的测试已全部通过
3. 每个任务标注依赖的原子任务 ID（如 `→ AT-003`）
4. 估算时间均为单人操作（含编写测试 + 实现 + 运行验证）

---

## Phase 1: 项目骨架（AT-001 ~ AT-006）

### AT-001 — Go 模块初始化
- **原任务**: 1.1
- **Red**: 创建 `cmd/reflector/main.go`，写 `TestMainExists` 验证编译通过
- **Green**: `package main` + `func main() { fmt.Println("reflector") }`
- **验证**: `go build ./cmd/reflector/`
- **耗时**: ~3 min

### AT-002 — 目录结构创建
- **原任务**: 1.4
- **Red**: 写 `TestDirectoryStructure` 验证所有必要目录存在
- **Green**: 创建目录：`internal/{api,engine,model,store,config}`、`adapters/{shared,opencode/src}`、`prompts/`、`scripts/`
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-003 — 核心依赖引入
- **原任务**: 1.2
- **Red**: 写 `TestDependencies` 验证 chi、sqlite、yaml 包可 import
- **Green**: `go get` 添加依赖，`go mod tidy`
- **验证**: `go build ./...` 通过
- **耗时**: ~3 min

### AT-004 — .reflector 目录模板与 .gitignore
- **原任务**: 1.4
- **Red**: 写 `TestDefaultDirLayout` 验证 `EnsureDefaultLayout()` 创建 data/、reports/、logs/、prompts/、hooks/ 子目录
- **Green**: 实现 `internal/config/layout.go` → `EnsureDefaultLayout(basePath string) error`
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-005 — .gitignore 规则
- **原任务**: 1.4
- **Red**: 写 `TestGitignoreContent` 验证 `.reflector/` 条目存在
- **Green**: 在 `.gitignore` 中添加 `.reflector/` 规则
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-006 — 默认配置文件模板
- **原任务**: 9.3
- **Red**: 写 `TestDefaultConfigYAML` 验证默认 YAML 能被解析为 Config 结构
- **Green**: 创建 `reflector.yaml` 模板（含所有默认值）
- **验证**: 测试通过
- **耗时**: ~3 min

---

## Phase 2: 数据模型（AT-007 ~ AT-021）

> 依赖: Phase 1 完成

### AT-007 — AgentToolType 枚举与 TaskStatus 枚举
- **原任务**: 2.1
- **Red**: `TestAgentToolTypeValues` 验证 `"opencode"/"openclaw"/"claudecode"` 常量；`TestTaskStatusValues` 验证 `COMPLETED/INTERRUPTED/UNCERTAIN/ABANDONED`
- **Green**: `internal/model/enums.go` 定义 `type AgentToolType string` 和 `type TaskStatus string` 常量
- **验证**: `go test ./internal/model/...`
- **耗时**: ~2 min

### AT-008 — CanonicalMessage 结构体与 JSON 序列化
- **原任务**: 2.1
- **Red**: `TestCanonicalMessageJSON` — 构造完整 CanonicalMessage，`json.Marshal` 后 `json.Unmarshal`，验证字段一致
- **Green**: `internal/model/session.go` 定义 `CanonicalMessage` struct（含 JSON tags 与 TypeScript 定义对齐）
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-009 — CanonicalMessage 可选字段 nil 测试
- **原任务**: 2.1
- **Red**: `TestCanonicalMessageOptionalFields` — `AgentName`、`PromptTokens`、`CompletionTokens` 为 nil 时序列化/反序列化正确
- **Green**: 使用 `*string`/`*int64` 指针类型或 `omitempty`
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-010 — CanonicalToolCall 结构体与 JSON 序列化
- **原任务**: 2.1
- **Red**: `TestCanonicalToolCallJSON` — 构造 TOOL/MCP/SKILL 三种类型，验证序列化正确
- **Green**: `internal/model/session.go` 定义 `CanonicalToolCall` struct
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-011 — CanonicalSession 结构体与 JSON 序列化
- **原任务**: 2.1
- **Red**: `TestCanonicalSessionJSON` — 完整 session 含 messages 和 toolCalls，round-trip 序列化验证
- **Green**: `internal/model/session.go` 定义 `CanonicalSession` struct
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-012 — CanonicalSession 边界 case：空 messages/toolCalls
- **原任务**: 2.1
- **Red**: `TestCanonicalSessionEmptyFields` — messages=[] 和 toolCalls=[] 时序列化正确
- **Green**: 确保切片字段使用 `omitempty` 或默认空切片
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-013 — CapabilityMatrix 结构体
- **原任务**: 2.1
- **Red**: `TestCapabilityMatrixJSON` — 序列化/反序列化验证
- **Green**: `internal/model/session.go` 定义 `CapabilityMatrix` struct（5 个 bool 字段）
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-014 — SessionMetrics 结构体（时间指标 M-001~M-004）
- **原任务**: 2.2
- **Red**: `TestSessionMetricsTimeFields` — 验证 StartTime、EndTime、DurationSec、HumanDurationSec 字段
- **Green**: `internal/model/metrics.go` 定义 `SessionMetrics` 中时间相关字段
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-015 — SessionMetrics Token 指标（M-005~M-008）
- **原任务**: 2.2
- **Red**: `TestSessionMetricsTokenFields` — PromptTokens、CompletionTokens、TotalTokens、LLMRequestCount；-1 表示 N/A
- **Green**: 扩展 `SessionMetrics` Token 字段
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-016 — SessionMetrics 工具/消息/参与度/情感字段（M-009~M-025）
- **原任务**: 2.2
- **Red**: `TestSessionMetricsAllFields` — 验证所有剩余字段（ToolCalls map、AgentParticipations、Messages 等）序列化正确
- **Green**: 扩展 `SessionMetrics` 所有字段
- **验证**: 测试通过
- **耗时**: ~4 min

### AT-017 — AgentParticipation 和 ToolCallRecord 结构体
- **原任务**: 2.2
- **Red**: `TestAgentParticipation` 和 `TestToolCallRecord` 两个测试
- **Green**: `internal/model/metrics.go` 定义两个 struct
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-018 — ReflectionLog 和 Watermark 结构体
- **原任务**: 2.2
- **Red**: `TestReflectionLogJSON` 和 `TestWatermarkJSON` 两个测试
- **Green**: `internal/model/metrics.go` 定义两个 struct
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-019 — ReflectRequest 结构体
- **原任务**: 2.1
- **Red**: `TestReflectRequestJSON` — 验证请求体（trigger_type、sessions 数组、config）序列化/反序列化
- **Green**: `internal/model/api.go` 定义 `ReflectRequest` struct
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-020 — ReflectResponse 结构体
- **原任务**: 2.1
- **Red**: `TestReflectResponseJSON` — 验证 status、sessions_analyzed、report_path 等字段
- **Green**: `internal/model/api.go` 定义 `ReflectResponse` struct
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-021 — JSON round-trip 端到端：完整请求→响应
- **原任务**: 2.1
- **Red**: `TestFullRoundTrip` — 构造含 2 条 message + 1 条 toolCall 的完整请求，序列化后反序列化为 ReflectRequest，验证嵌套结构完整
- **Green**: 确保所有嵌套结构体 JSON tag 正确（此步通常只是验证，无需新代码）
- **验证**: 测试通过
- **耗时**: ~4 min

---

## Phase 3: 配置管理（AT-022 ~ AT-027）

> 依赖: Phase 2 完成

### AT-022 — Config 结构体与 YAML tags
- **原任务**: 2.3
- **Red**: `TestConfigYAMLTags` — 构造完整 Config，`yaml.Marshal` 后验证输出包含正确的 YAML key
- **Green**: `internal/config/config.go` 定义 `Config` struct（嵌套 Trigger、Model、Sentiment、Retention、Report、LogLevel）
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-023 — DefaultConfig() 默认值
- **原任务**: 2.3
- **Red**: `TestDefaultConfig` — 验证默认端口=19870、时间="00:00"、保留=90天、日志=info
- **Green**: 实现 `DefaultConfig() *Config` 返回硬编码默认值
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-024 — Load() 无配置文件时返回默认值
- **原任务**: 2.3
- **Red**: `TestLoadNoFile` — 传入不存在的路径，验证返回 DefaultConfig 且无 error
- **Green**: 实现 `Load(path string) (*Config, error)`，文件不存在时返回 DefaultConfig()
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-025 — Load() 从 YAML 文件加载
- **原任务**: 2.3
- **Red**: `TestLoadFromYAML` — 创建临时 YAML 文件，自定义端口和模型，验证解析正确
- **Green**: `Load()` 中添加 `os.ReadFile` + `yaml.Unmarshal` 逻辑
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-026 — Load() 部分覆盖（YAML 只设部分字段，其余用默认值）
- **原任务**: 2.3
- **Red**: `TestLoadPartialOverride` — YAML 只改端口，其余字段保留默认值
- **Green**: 先 `DefaultConfig()` 再 `yaml.Unmarshal` 覆盖
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-027 — Watermark 持久化
- **原任务**: 2.4
- **Red**: `TestWatermarkPersistence` — 写入 watermark 时间戳到文件，重新加载验证一致
- **Green**: `internal/config/watermark.go` 实现 `LoadWatermark()` / `SaveWatermark()`
- **验证**: 测试通过
- **耗时**: ~3 min

---

## Phase 4: SQLite 数据持久化（AT-028 ~ AT-042）

> 依赖: Phase 2 + Phase 3 完成

### AT-028 — Store 接口定义
- **原任务**: 3.1
- **Red**: `TestStoreInterface` — 定义 `Store` interface 并验证 SQLiteStore 实现该接口（编译时检查）
- **Green**: `internal/store/store.go` 定义 interface
- **验证**: 编译通过
- **耗时**: ~2 min

### AT-029 — NewStore() 数据库初始化 + DDL 建表
- **原任务**: 3.1
- **Red**: `TestNewStoreCreatesTables` — NewStore() 后查询 sqlite_master 验证 5 张表存在
- **Green**: `internal/store/sqlite.go` 实现 `NewStore()` + DDL 执行
- **验证**: 测试通过
- **耗时**: ~4 min

### AT-030 — NewStore() 幂等性（重复调用不报错）
- **原任务**: 3.1
- **Red**: `TestNewStoreIdempotent` — 连续两次 NewStore() 同一路径，第二次不报错且表结构完整
- **Green**: 使用 `CREATE TABLE IF NOT EXISTS`
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-031 — InsertSession() 基本插入
- **原任务**: 3.2
- **Red**: `TestInsertSession` — 插入一条完整 SessionMetrics，查询验证所有字段正确
- **Green**: 实现 `InsertSession(ctx, metrics *model.SessionMetrics) error`
- **验证**: 测试通过
- **耗时**: ~4 min

### AT-032 — InsertSession() N/A 指标处理（-1 值）
- **原任务**: 3.2
- **Red**: `TestInsertSessionWithNA` — Token 字段为 -1 时正确存入 NULL
- **Green**: 插入时 -1 转为 nil
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-033 — GetSessionsByDateRange() 按日期查询
- **原任务**: 3.2
- **Red**: `TestGetSessionsByDateRange` — 插入 3 条不同日期 session，查其中 2 条
- **Green**: 实现 `GetSessionsByDateRange(ctx, start, end time.Time) ([]model.SessionMetrics, error)`
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-034 — GetSessionsByStatus() 按状态查询
- **原任务**: 3.2
- **Red**: `TestGetSessionsByStatus` — 插入 COMPLETED + INTERRUPTED session，查 COMPLETED
- **Green**: 实现 `GetSessionsByStatus(ctx, status model.TaskStatus) ([]model.SessionMetrics, error)`
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-035 — InsertAgentParticipations() 批量插入
- **原任务**: 3.3
- **Red**: `TestInsertAgentParticipations` — 插入 3 条参与记录，验证数据正确
- **Green**: 实现 `InsertAgentParticipations(ctx, sessionID string, participations []model.AgentParticipation) error`
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-036 — InsertToolCalls() 批量插入
- **原任务**: 3.4
- **Red**: `TestInsertToolCalls` — 插入 TOOL/MCP/SKILL 各一条，验证 call_type 正确
- **Green**: 实现 `InsertToolCalls(ctx, sessionID string, calls []model.ToolCallRecord) error`
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-037 — InsertReflectionLog() 插入反思日志
- **原任务**: 3.5
- **Red**: `TestInsertReflectionLog` — 插入一条 SUCCESS 日志，查询验证
- **Green**: 实现 `InsertReflectionLog(ctx, log *model.ReflectionLog) error`
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-038 — GetWatermark() / UpdateWatermark()
- **原任务**: 3.6
- **Red**: `TestWatermarkCRUD` — 初始为空 → Update → Get 验证 → 再 Update → Get 验证更新
- **Green**: 实现 `GetWatermark(ctx) (time.Time, error)` 和 `UpdateWatermark(ctx, t time.Time) error`
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-039 — Cleanup() 按天数删除过期数据
- **原任务**: 3.7
- **Red**: `TestCleanup` — 插入 30 天前 + 1 天前各一条 session，Cleanup(days=7) 后验证只剩 1 条
- **Green**: 实现 `Cleanup(ctx, days int) error`，DELETE WHERE end_time < ?
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-040 — Cleanup() 级联删除关联记录
- **原任务**: 3.7
- **Red**: `TestCleanupCascade` — 插入 session + agent_participations + tool_calls，Cleanup 后验证关联记录也被删除
- **Green**: Cleanup 中增加对 agent_participations 和 tool_calls 的 DELETE
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-041 — FallbackWriter：写入 JSON 文件
- **原任务**: 3.8
- **Red**: `TestFallbackWrite` — 模拟 SQLite 失败，验证数据写入 fallback JSON 文件
- **Green**: `internal/store/fallback.go` 实现 `WriteFallback(data interface{}) error`
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-042 — 写入重试机制
- **原任务**: 3.8
- **Red**: `TestInsertWithRetry` — mock Store 让前 2 次失败第 3 次成功，验证重试逻辑
- **Green**: `internal/store/retry.go` 实现 `WithRetry(fn func() error, maxRetries int) error`
- **验证**: 测试通过
- **耗时**: ~3 min

---

## Phase 5: 核心引擎 — Session 解析与指标提取（AT-043 ~ AT-056）

> 依赖: Phase 2 + Phase 4 完成

### AT-043 — ParseSession() 基本解析
- **原任务**: 4.1
- **Red**: `TestParseSession` — 输入 CanonicalSession JSON，验证解析为内部 SessionMetrics
- **Green**: `internal/engine/parser.go` 实现 `ParseSession(raw json.RawMessage) (*model.CanonicalSession, error)`
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-044 — ParseSession() 无效 JSON 处理
- **原任务**: 4.1
- **Red**: `TestParseSessionInvalidJSON` — 输入非法 JSON，验证返回明确错误
- **Green**: 错误处理逻辑
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-045 — ExtractTimeMetrics() — M-001~M-004
- **原任务**: 4.2
- **Red**: `TestExtractTimeMetrics` — 构造 3 条不同时间 message，验证 start/end/duration/human_duration
- **Green**: `internal/engine/extractor.go` 实现 `ExtractTimeMetrics(session *model.CanonicalSession) (start, end time.Time, duration, humanDur int64)`
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-046 — ExtractTimeMetrics() — 空消息处理
- **原任务**: 4.2
- **Red**: `TestExtractTimeMetricsEmpty` — messages=[] 时返回零值不 panic
- **Green**: 添加空消息边界检查
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-047 — ExtractTokenMetrics() — M-005~M-008（有 Token 数据）
- **原任务**: 4.2
- **Red**: `TestExtractTokenMetrics` — messages 含 promptTokens/completionTokens，验证累计计算正确
- **Green**: `ExtractTokenMetrics(session) (prompt, completion, total, requests int64)`
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-048 — ExtractTokenMetrics() — Token 不可用（nil 字段）
- **原任务**: 4.2
- **Red**: `TestExtractTokenMetricsNA` — messages 的 token 字段全为 nil，验证返回 -1
- **Green**: nil 检测 → 返回 -1
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-049 — ExtractToolMetrics() — M-009~M-014
- **原任务**: 4.2
- **Red**: `TestExtractToolMetrics` — toolCalls 含 3 种类型（TOOL/MCP/SKILL），验证按类型和名称分组统计正确
- **Green**: `ExtractToolMetrics(session) (counts map[string]int, details []model.ToolCallRecord)`
- **验证**: 测试通过
- **耗时**: ~4 min

### AT-050 — ExtractAgentMetrics() — M-015~M-016
- **原任务**: 4.2
- **Red**: `TestExtractAgentMetrics` — 3 条 agent message 含 2 个不同 agentName，验证计数和参与度
- **Green**: `ExtractAgentMetrics(session) (count int, participations map[string]float64)`
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-051 — ExtractMessageMetrics() — M-017~M-019
- **原任务**: 4.2
- **Red**: `TestExtractMessageMetrics` — 5 human + 8 agent + 2 system 消息，验证总数和分类
- **Green**: `ExtractMessageMetrics(session) (total, agent, human int)`
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-052 — ExtractHumanParticipation() — M-020~M-022
- **原任务**: 4.2
- **Red**: `TestExtractHumanParticipation` — 验证参与率、介入次数（非首条 human 消息）、中断次数（关键词匹配）
- **Green**: `ExtractHumanParticipation(session) (rate float64, interventions, interrupts int)`
- **验证**: 测试通过
- **耗时**: ~4 min

### AT-053 — ExtractAll() 整合所有指标
- **原任务**: 4.2
- **Red**: `TestExtractAll` — 构造完整 session，调用 ExtractAll()，验证返回的 SessionMetrics 所有字段已填充
- **Green**: `ExtractAll(session) *model.SessionMetrics` 编排调用所有子函数
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-054 — ExtractAll() — 含空 session
- **原任务**: 4.2
- **Red**: `TestExtractAllEmpty` — session 无消息无 toolCall，验证不 panic 且返回合理零值
- **Green**: 各子函数添加边界保护
- **验证**: 测试通过
- **耗时**: ~2 min

---

## Phase 6: 核心引擎 — 任务状态判定（AT-055 ~ AT-061）

> 依赖: Phase 5 完成

### AT-055 — L1: OpenSpec /opsx:archive 检测（正向）
- **原任务**: 4.3
- **Red**: `TestL1ArchiveDetected` — 最后一条 human 消息包含 `/opsx:archive`，验证返回 COMPLETED
- **Green**: `internal/engine/classifier.go` 实现 `checkL1(messages []model.CanonicalMessage) (model.TaskStatus, bool)`
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-056 — L1: 无 OpenSpec 指令时返回 false
- **原任务**: 4.3
- **Red**: `TestL1NoArchive` — 消息中无 `/opsx:archive`，验证返回 false
- **Green**: 已实现，确保无匹配时返回 `(UNCERTAIN, false)`
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-057 — L3: 中文关键词匹配
- **原任务**: 4.3
- **Red**: `TestL3ChineseKeywords` — "任务已完成"、"完成工作"、"工作完成" → COMPLETED；"无法完成" → 不是 COMPLETED
- **Green**: `checkL3Chinese(content string) (model.TaskStatus, bool)`
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-058 — L3: 英文关键词匹配
- **原任务**: 4.3
- **Red**: `TestL3EnglishKeywords` — "task completed"、"all done"、"finished" → COMPLETED
- **Green**: `checkL3English(content string) (model.TaskStatus, bool)`
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-059 — L3: 否定语境排除
- **原任务**: 4.3
- **Red**: `TestL3NegativeContext` — "无法完成"、"未能完成"、"did not complete" → 不是 COMPLETED
- **Green**: 添加否定词过滤逻辑
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-060 — L2: LLM 意图分类（mock client）
- **原任务**: 4.3
- **Red**: `TestL2WithMockClient` — 注入 mock LLM client 返回 `{status: "completed", confidence: 0.9}`，验证判定结果
- **Green**: `checkL2(messages, llmClient) (model.TaskStatus, float64, error)` 定义 LLMClient interface
- **验证**: 测试通过
- **耗时**: ~4 min

### AT-061 — ClassifyTask() L1→L2→L3 级联判定
- **原任务**: 4.3
- **Red**: `TestClassifyTaskCascade` — 三个 case 分别命中 L1/L2/L3，以及全部未命中→UNCERTAIN
- **Green**: `ClassifyTask(session, llmClient) model.TaskStatus` 实现级联逻辑
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-062 — DetermineTaskStatus() ABANDONED 判定
- **原任务**: 4.3
- **Red**: `TestAbandonedStatus` — 最后一条消息是 human → ABANDONED；最后一条是 agent → 由 ClassifyTask 决定
- **Green**: `DetermineTaskStatus(session, llmClient) model.TaskStatus`
- **验证**: 测试通过
- **耗时**: ~2 min

---

## Phase 7: 核心引擎 — 情感分析（AT-063 ~ AT-068）

> 依赖: Phase 5 完成

### AT-063 — SentimentResult 结构体
- **原任务**: 4.4
- **Red**: `TestSentimentResultJSON` — 验证结构化 JSON 输出可解析为 SentimentResult
- **Green**: `internal/engine/sentiment.go` 定义 `SentimentResult` struct
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-064 — SanitizeMessages() 脱敏处理
- **原任务**: 4.4
- **Red**: `TestSanitizeMessages` — 输入含 API Key（`sk-xxxx`）、密码（`password=xxx`）、token（`Bearer xxx`）的文本，验证输出已脱敏
- **Green**: `SanitizeMessages(messages []string) []string`，正则替换敏感 pattern
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-065 — AnalyzeSentiment() 正常调用（mock client）
- **原任务**: 4.4
- **Red**: `TestAnalyzeSentimentSuccess` — mock LLM client 返回有效 JSON，验证返回 SentimentResult
- **Green**: `AnalyzeSentiment(ctx, messages []string, client LLMClient) (*SentimentResult, error)`
- **验证**: 测试通过
- **耗时**: ~4 min

### AT-066 — AnalyzeSentiment() 模型返回非 JSON
- **原任务**: 4.4
- **Red**: `TestAnalyzeSentimentInvalidResponse` — mock 返回非 JSON 文本，验证返回 error
- **Green**: JSON 解析错误处理
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-067 — 情感分析禁用时跳过
- **原任务**: 4.4
- **Red**: `TestSentimentDisabled` — config.sentiment.enabled=false 时，AnalyzeSentiment 返回 nil 不报错
- **Green**: 入口处检查 config
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-068 — 情感分析模型不可用时降级
- **原任务**: 4.4
- **Red**: `TestSentimentModelUnavailable` — mock client 返回 error，验证返回 nil + M-023~025 标为 N/A
- **Green**: error 时返回 nil 指标，由调用方处理 N/A
- **验证**: 测试通过
- **耗时**: ~2 min

---

## Phase 8: 核心引擎 — 触发管理与 Hook（AT-069 ~ AT-076）

> 依赖: Phase 2 完成

### AT-069 — TriggerType 枚举与 TriggerManager 结构体
- **原任务**: 4.5
- **Red**: `TestTriggerTypeValues` — TIME/EVENTS/MANUAL 常量存在
- **Green**: `internal/engine/trigger.go` 定义类型
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-070 — 防抖：5 分钟内重复触发忽略
- **原任务**: 4.5
- **Red**: `TestDebounce` — 快速调用两次 Trigger()，验证只执行一次回调
- **Green**: `TriggerManager` 添加 lastTrigger 字段 + 5min 防抖逻辑
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-071 — 并发排队：反思运行中时新请求排队
- **原任务**: 4.5
- **Red**: `TestQueueDuringExecution` — 第一次 Trigger 正在执行时第二次 Trigger 不并发，排队等待
- **Green**: 使用 `sync.Mutex` 或 channel 控制并发
- **验证**: 测试通过
- **耗时**: ~4 min

### AT-072 — 手动触发优先级
- **原任务**: 4.5
- **Red**: `TestManualPriority` — 排队中存在 TIME 触发时 MANUAL 可抢占
- **Green**: 使用优先级 channel 或条件变量
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-073 — HookPoint 枚举（10 个节点）
- **原任务**: 4.6
- **Red**: `TestHookPointConstants` — 验证 10 个 hook 节点常量存在
- **Green**: `internal/engine/hook.go` 定义 `HookPoint` type 和 10 个常量
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-074 — HookRunner 扫描 hooks 目录
- **原任务**: 4.6
- **Red**: `TestScanHooksDir` — 创建 `.reflector/hooks/before-save-metrics.sh`，验证 ScanHooks() 返回正确映射
- **Green**: `ScanHooks(hooksDir string) (map[HookPoint]string, error)`
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-075 — ExecuteHook() stdin/stdout JSON 通信
- **原任务**: 4.6
- **Red**: `TestExecuteHook` — 创建一个测试脚本（读 stdin JSON → 修改 → 写 stdout），验证 ExecuteHook 正确传递和接收 JSON
- **Green**: `ExecuteHook(ctx, hookPath string, input interface{}, output interface{}) error`
- **验证**: 测试通过
- **耗时**: ~4 min

### AT-076 — ExecuteHook() 失败不阻断
- **原任务**: 4.6
- **Red**: `TestExecuteHookFailure` — 脚本 exit code 非零，验证返回 error 但主流程继续（由调用方忽略）
- **Green**: 非零 exit code 返回 error（不 panic）
- **验证**: 测试通过
- **耗时**: ~2 min

---

## Phase 9: 提示词管理（AT-077 ~ AT-080）

> 依赖: Phase 1 完成

### AT-077 — embed 默认提示词文件
- **原任务**: 5.3
- **Red**: `TestEmbeddedPrompts` — 验证 `prompts/` 下三个 .md 文件可通过 embed.FS 读取
- **Green**: `prompts/` 目录下创建三个默认 .md 文件 + `//go:embed prompts/*.md`
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-078 — PromptLoader：优先从文件系统读取
- **原任务**: 5.2
- **Red**: `TestLoadPromptFromFile` — 创建临时 .md 文件，LoadPrompt 返回文件内容
- **Green**: `LoadPrompt(name string, embeddedFS embed.FS) (string, error)` 先读文件系统
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-079 — PromptLoader：文件不存在时 fallback 到嵌入默认值
- **原任务**: 5.2
- **Red**: `TestLoadPromptFallback` — 文件不存在时返回嵌入默认值，不报错
- **Green**: 文件读取失败时 fallback 到 embed.FS
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-080 — PromptLoader：空文件 fallback
- **原任务**: 5.2
- **Red**: `TestLoadPromptEmptyFile` — 文件存在但内容为空，返回嵌入默认值 + warning
- **Green**: 检查内容为空 → fallback
- **验证**: 测试通过
- **耗时**: ~2 min

---

## Phase 10: 日报生成（AT-081 ~ AT-086）

> 依赖: Phase 4 + Phase 5 完成

### AT-081 — ReportGenerator 基本模板渲染
- **原任务**: 6.1
- **Red**: `TestGenerateReport` — 传入 2 个 SessionMetrics，验证生成的 Markdown 包含：标题、任务清单表、汇总统计
- **Green**: `internal/engine/report.go` 实现 `GenerateReport(date string, metrics []model.SessionMetrics, reflectorTokens int64) (string, error)`
- **验证**: 测试通过
- **耗时**: ~4 min

### AT-082 — 日报内容：工具使用统计（3.1/3.2/3.3 节）
- **原任务**: 6.1
- **Red**: `TestReportToolStats` — SessionMetrics 含 toolCalls 数据，验证日报包含工具/MCP/Skill 分组统计表
- **Green**: 在 GenerateReport 中添加工具统计渲染
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-083 — 日期归属逻辑
- **原任务**: 6.2
- **Red**: `TestReportDateAttribution` — TIME 触发 → 前一天；MANUAL 触发 → 今天
- **Green**: `GetReportDate(triggerType model.TriggerType, now time.Time) string`
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-084 — 日报幂等追加：文件已存在时追加
- **原任务**: 6.3
- **Red**: `TestAppendReport` — 先写一个日报文件，再追加新数据，验证追加后内容完整
- **Green**: `SaveReport(path string, content string) error`，文件存在时追加
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-085 — 日报文件名格式验证
- **原任务**: 6.1
- **Red**: `TestReportFilename` — 验证 `dayreport-2026-04-19.md` 格式正确
- **Green**: `ReportFilePath(basePath, date string) string`
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-086 — 日报含反思工具自身 Token 开销
- **原任务**: 6.1
- **Red**: `TestReportReflectorCost` — 传入 reflectorTokens=500，验证日报"四、反思工具自身开销"节包含该值
- **Green**: 模板中添加反思工具开销节
- **验证**: 测试通过
- **耗时**: ~2 min

---

## Phase 11: HTTP API 服务（AT-087 ~ AT-097）

> 依赖: Phase 5~10 全部完成

### AT-087 — chi 路由初始化 + 中间件注册
- **原任务**: 7.1
- **Red**: `TestNewServer` — NewServer() 后验证路由注册（可用 httptest）
- **Green**: `internal/api/server.go` 实现 `NewServer(store, config, engine) *Server`，注册 chi 中间件
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-088 — GET /api/v1/health 正常响应
- **原任务**: 7.3
- **Red**: `TestHealthEndpoint` — httptest GET /api/v1/health，验证 200 + `{"status": "ok"}`
- **Green**: 实现 health handler
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-089 — GET /api/v1/stats 正常响应
- **原任务**: 7.4
- **Red**: `TestStatsEndpoint` — 插入 2 条 reflection_log 后 GET /api/v1/stats，验证 total_reflections=2
- **Green**: 实现 stats handler，从 store 聚合统计
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-090 — POST /api/v1/reflect 请求验证（空 body）
- **原任务**: 7.2
- **Red**: `TestReflectEmptyBody` — POST 空 body，验证返回 400
- **Green**: 请求体验证中间件/handler
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-091 — POST /api/v1/reflect 请求验证（无效 trigger_type）
- **原任务**: 7.2
- **Red**: `TestReflectInvalidTriggerType` — trigger_type="INVALID"，验证返回 400
- **Green**: trigger_type 校验逻辑
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-092 — POST /api/v1/reflect 请求验证（无 sessions）
- **原任务**: 7.2
- **Red**: `TestReflectNoSessions` — sessions=[] 空，验证返回 400
- **Green**: sessions 非空校验
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-093 — POST /api/v1/reflect 完整流程（mock store）
- **原任务**: 7.2
- **Red**: `TestReflectFullFlow` — mock store + mock LLM client，POST 完整请求，验证 200 + ReflectResponse 各字段
- **Green**: 反思 handler 编排：解析 → 提取 → 判定 → 情感 → 持久化 → 日报 → 返回
- **验证**: 测试通过
- **耗时**: ~5 min

### AT-094 — POST /api/v1/reflect 返回正确 duration_ms
- **原任务**: 7.2
- **Red**: `TestReflectDuration` — 验证 ReflectResponse 中 duration_ms > 0
- **Green**: handler 中添加耗时计算
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-095 — POST /api/v1/reflect 错误处理（store 失败）
- **原任务**: 7.2
- **Red**: `TestReflectStoreFailure` — mock store InsertSession 返回 error，验证返回 500
- **Green**: 错误处理链
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-096 — POST /api/v1/cleanup 正常调用
- **原任务**: 7.5
- **Red**: `TestCleanupEndpoint` — POST `{"days": 30}`，验证 200 + 调用了 store.Cleanup(30)
- **Green**: cleanup handler 实现
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-097 — main.go 启动流程
- **原任务**: 7.6
- **Red**: `TestMainStartup` — 验证 main() 启动后 health endpoint 可达（集成测试，临时端口）
- **Green**: `cmd/reflector/main.go` 实现：LoadConfig → NewStore → NewServer → ListenAndServe
- **验证**: 测试通过
- **耗时**: ~4 min

---

## Phase 12: opencode Adapter（Node.js）（AT-098 ~ AT-108）

> 依赖: Phase 11 完成（需要 Core Engine 可编译运行）

### AT-098 — Adapter 项目初始化
- **原任务**: 8.1
- **Red**: `npm run build` 验证 TypeScript 编译成功
- **Green**: `adapters/opencode/` 下创建 package.json、tsconfig.json、安装依赖
- **验证**: `npm run build` 通过
- **耗时**: ~3 min

### AT-099 — Adapter 共享类型包
- **原任务**: 1.3
- **Red**: import shared types 并验证 TypeScript 编译通过
- **Green**: `adapters/shared/types.ts` 定义所有 interface
- **验证**: 编译通过
- **耗时**: ~3 min

### AT-100 — ProcessManager: 启动 Core Engine 进程
- **原任务**: 8.3
- **Red**: `TestProcessManagerStart` — mock child_process.spawn，验证启动参数正确
- **Green**: `process-manager.ts` 实现 `start(binaryPath: string, port: number): Promise<void>`
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-101 — ProcessManager: 健康检查等待
- **原任务**: 8.3
- **Red**: `TestWaitForHealthy` — mock HTTP GET /health 返回 200，验证解析成功
- **Green**: 轮询 `/api/v1/health` 直到 200，超时 10 秒
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-102 — ProcessManager: 启动失败重试
- **原任务**: 8.3
- **Red**: `TestStartRetry` — mock 前两次失败第三次成功，验证重试 3 次
- **Green**: 指数退避重试逻辑
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-103 — ProcessManager: 优雅关闭
- **原任务**: 8.3
- **Red**: `TestGracefulShutdown` — 调用 shutdown()，验证 SIGTERM 发送 + 等待退出
- **Green**: `shutdown(): Promise<void>` 发送 SIGTERM + 等待
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-104 — SessionAdapter: getSessions() 数据获取
- **原任务**: 8.4
- **Red**: `TestGetSessions` — mock SDK 返回数据，验证转换为 CanonicalSession[]
- **Green**: `adapter.ts` 实现 `getSessions(since: Date): Promise<CanonicalSession[]>`
- **验证**: 测试通过
- **耗时**: ~4 min

### AT-105 — SessionAdapter: getCapabilityMatrix()
- **原任务**: 8.6
- **Red**: `TestGetCapabilityMatrix` — 验证返回值包含正确的 boolean 字段
- **Green**: 根据 SDK 实际能力返回 CapabilityMatrix
- **验证**: 测试通过
- **耗时**: ~2 min

### AT-106 — Plugin 入口: 注册指令
- **原任务**: 8.2
- **Red**: `TestPluginRegistration` — 验证 registerCommand 被正确调用
- **Green**: `index.ts` 实现 `registerPlugin()` 注册 `/or:reflect_now` 和 `/or:cleanup`
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-107 — 定时触发: setInterval
- **原任务**: 8.5
- **Red**: `TestScheduleTrigger` — mock Date 到 00:00，验证 reflect 被调用
- **Green**: `triggers.ts` 实现定时逻辑
- **验证**: 测试通过
- **耗时**: ~3 min

### AT-108 — Adapter 配置读取
- **原任务**: 8.2
- **Red**: `TestLoadAdapterConfig` — 创建临时 reflector.yaml，验证解析出端口和触发时间
- **Green**: 使用 js-yaml 读取配置
- **验证**: 测试通过
- **耗时**: ~2 min

---

## Phase 13: 构建与安装脚本（AT-109 ~ AT-111）

> 依赖: Phase 11 + Phase 12 完成

### AT-109 — scripts/build.sh
- **原任务**: 9.1
- **Red**: `TestBuildScript` — 运行 build.sh，验证产出 Go 二进制 + Adapter .js 文件
- **Green**: 编写 build.sh
- **验证**: 脚本执行成功
- **耗时**: ~3 min

### AT-110 — scripts/install.sh
- **原任务**: 9.2
- **Red**: `TestInstallScript` — 运行 install.sh，验证 .reflector/ 目录创建 + 二进制复制到正确位置
- **Green**: 编写 install.sh
- **验证**: 脚本执行成功
- **耗时**: ~3 min

### AT-111 — 默认配置文件验证
- **原任务**: 9.3
- **Red**: `TestDefaultConfigValid` — 加载默认 reflector.yaml，验证 Config 结构完整
- **Green**: 已在 AT-006 完成，此步做端到端验证
- **验证**: 测试通过
- **耗时**: ~2 min

---

## Phase 14: 集成验证（AT-112 ~ AT-114）

> 依赖: 全部完成

### AT-112 — Go 集成测试: 完整反思流程
- **原任务**: 10.2
- **Red**: `TestIntegrationReflectFlow` — 启动真实 Core Engine → POST reflect → 验证 SQLite 数据 + 日报文件
- **Green**: 使用真实 SQLite（temp dir），mock LLM client
- **验证**: 测试通过
- **耗时**: ~5 min

### AT-113 — 端到端验证: curl 触发
- **原任务**: 10.3
- **Red**: 手动验证脚本 — 启动 Core Engine → curl POST → 检查数据和文件
- **Green**: 编写 `scripts/e2e-test.sh` 验证脚本
- **验证**: 脚本通过
- **耗时**: ~4 min

### AT-114 — Go 单元测试覆盖率检查
- **原任务**: 10.1
- **Red**: `go test -cover ./...` 验证覆盖率 > 70%
- **Green**: 补充缺失测试
- **验证**: 覆盖率达标
- **耗时**: ~5 min

---

## 汇总统计

| Phase | 描述 | 原子任务数 | 预估总耗时 |
|-------|------|-----------|-----------|
| 1 | 项目骨架 | 6 | 16 min |
| 2 | 数据模型 | 15 | 41 min |
| 3 | 配置管理 | 6 | 16 min |
| 4 | SQLite 持久化 | 15 | 44 min |
| 5 | 指标提取 | 14 | 37 min |
| 6 | 任务判定 | 8 | 20 min |
| 7 | 情感分析 | 6 | 15 min |
| 8 | 触发/Hook | 8 | 23 min |
| 9 | 提示词管理 | 4 | 9 min |
| 10 | 日报生成 | 6 | 16 min |
| 11 | HTTP API | 11 | 28 min |
| 12 | opencode Adapter | 11 | 31 min |
| 13 | 构建脚本 | 3 | 8 min |
| 14 | 集成验证 | 3 | 14 min |
| **合计** | | **116** | **~318 min (~5.3h)** |

### 依赖关系图

```
Phase 1 ──→ Phase 2 ──→ Phase 3 ──→ Phase 4
                                      │
                                      ├──→ Phase 5 ──→ Phase 6 (任务判定)
                                      │                 └──→ Phase 7 (情感分析)
                                      │
                                      └──→ Phase 8 (触发/Hook)
                                          
Phase 5 + 6 + 7 + 8 ──→ Phase 9 (提示词)
                      ──→ Phase 10 (日报)
                      ──→ Phase 11 (HTTP API)
                                         │
Phase 11 ──→ Phase 12 (Adapter) ──→ Phase 13 (构建)
                                     └──→ Phase 14 (集成)
```

### 可并行执行的 Phase

以下 Phase 之间无依赖，可并行开发：
- **Phase 5 + Phase 8**（指标提取 与 触发/Hook 互相独立）
- **Phase 6 + Phase 7**（任务判定 与 情感分析 互相独立）
- **Phase 9 + Phase 10**（提示词 与 日报互相独立）
