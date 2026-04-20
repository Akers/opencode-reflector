# 情感分析功能重设计规格

> 基于 2026-04-20 头脑风暴确认的设计决策

## 设计决策摘要

| 决策项 | 选择 |
|--------|------|
| 执行位置 | Adapter 侧（Node.js），通过 opencode SDK `session.prompt()` |
| 影子 Session | 用完即删 — 每次分析创建临时 session，完成后删除 |
| 模型获取 | 配置文件优先 + SDK `client.config.get()` 回退 |
| 执行时机 | Adapter 预处理 — 在调用 Core Engine HTTP API 之前完成 |
| L2 分类器 | 统一迁移到 Adapter — 所有 LLM 调用集中在 Adapter 侧 |
| 降级策略 | 规则引擎 — SDK 不可用时用关键词匹配 + 统计规则 |
| 分析粒度 | 单 session 单次调用 — 每个 session 独立分析 |

---

## 1. 新架构

### 1.1 整体架构变更

```
                         opencode 进程
                    ┌───────────────────────────┐
                    │     opencode Runtime       │
                    │  ┌─────────────────────┐  │
                    │  │    Main Agent        │  │
                    │  └─────────────────────┘  │
                    │                            │
                    │  ┌─────────────────────┐  │
                    │  │  Reflector Plugin    │  │
                    │  │    (Adapter)         │  │
                    │  │                      │  │
                    │  │  ┌────────────────┐ │  │
                    │  │  │ LLM Agent      │ │  │
                    │  │  │                │ │  │
                    │  │  │ • 情感分析     │ │  │ ← SDK session.prompt()
                    │  │  │ • L2 任务分类  │ │  │    + 自定义 system prompt
                    │  │  │                │ │  │    + 宿主默认模型
                    │  │  │ 降级: 规则引擎 │ │  │
                    │  │  └────────────────┘ │  │
                    │  └─────────────────────┘  │
                    └────────────┬───────────────┘
                                 │ HTTP POST /api/v1/reflect
                                 │ (session 已附带 sentiment + task_status)
                                 ▼
                    ┌───────────────────────────┐
                    │    Core Engine (Go)        │
                    │                            │
                    │  • 从 metadata 读取结果    │
                    │  • 不再调用 LLM            │
                    │  • 负责存储/指标/报告      │
                    └───────────────────────────┘
```

### 1.2 核心原则

**让宿主工具做它最擅长的事（调 LLM），让 Core Engine 做它最擅长的事（存储、指标计算、报告生成）。**

- Adapter 负责：所有 LLM 交互（情感分析 + 任务分类）
- Core Engine 负责：指标提取、数据持久化、日报生成

---

## 2. 数据流

### 2.1 完整反思流程（新）

```
Step 0: 触发
  ├─ hooks.tool["reflect"]     (手动)
  ├─ hooks["chat.message"]     (N轮消息)
  └─ scheduled timer           (定时)

Step 1: Adapter 预处理
  │
  ├─ 1.1 获取模型配置
  │     client.config.get() → Config.model
  │     如果为空 → 使用 reflector.yaml 中的 model.override
  │     → modelInfo = { providerID, modelID }
  │
  ├─ 1.2 获取会话数据
  │     fetchSessions(client) → CanonicalSession[]
  │
  ├─ 1.3 情感分析 (对每个有 human messages 的 session)
  │     for each session:
  │       ├─ 脱敏: sanitizeMessages(session.humanMessages)
  │       ├─ 创建影子 session: client.session.create({ title: "__reflector_sa_" + id })
  │       ├─ 调用 LLM: client.session.prompt({
  │       │     id: shadowId,
  │       │     body: {
  │       │       system: sentimentPrompt,    // 情感分析 system prompt
  │       │       model: modelInfo,           // 宿主默认模型
  │       │       tools: {},                  // 禁用所有工具
  │       │       parts: [{ type: "text", text: sanitizedMessages }]
  │       │     }
  │       │   })
  │       ├─ 解析 JSON 响应 → SentimentResult
  │       ├─ 删除影子 session: client.session.delete({ id: shadowId })
  │       └─ 附加到 session.metadata.sentiment = result
  │
  ├─ 1.4 L2 任务分类 (对每个需要分类的 session)
  │     for each session:
  │       ├─ 加载 classify prompt
  │       ├─ 创建影子 session (或复用)
  │       ├─ 调用 LLM: session.prompt({ system: classifyPrompt, ... })
  │       ├─ 解析响应 → TaskClassification { status, confidence }
  │       ├─ 删除影子 session
  │       └─ 附加到 session.metadata.task_classification = result
  │
  └─ 1.5 调用 Core Engine
        engineManager.reflect({
          trigger_type, trigger_detail,
          sessions,  // 已附带 sentiment + task_classification
          config: { sentiment_enabled, ... }
        })

Step 2: Core Engine 处理（不变）
  │
  ├─ 遍历 sessions
  ├─ ExtractAll(metrics) — 时间/Token/工具/消息指标
  ├─ 从 metadata 提取 sentiment_result → 填充 metrics
  ├─ 从 metadata 提取 task_classification → 确定 TaskStatus
  │     (L1 和 L3 仍在 Core Engine，L2 从 metadata 读取)
  ├─ Save to SQLite
  └─ Generate report
```

### 2.2 降级流程

```
Adapter 调用 SDK prompt
  │
  ├─ 成功 → 使用 LLM 结果（优先级最高）
  │
  ├─ 失败原因: SDK 不可用 / opencode 不活跃 / 网络错误
  │   │
  │   └─ 降级到规则引擎
  │       │
  │       ├─ 情感分析降级:
  │       │   ├─ 正面关键词计数: 谢谢、好的、不错、可以、完美、great、thanks...
  │       │   ├─ 负面关键词计数: 不对、错误、不行、失败、糟糕、wrong、bad、fail...
  │       │   ├─ 中性关键词计数: 继续、下一个、然后、ok、next...
  │       │   ├─ negative_ratio = negative / total
  │       │   ├─ attitude_score = 5 + (positive - negative) / total * 5
  │       │   └─ approval_ratio = positive / (positive + negative) (如果有)
  │       │
  │       └─ 任务分类降级:
  │           └─ 跳过 L2，仅使用 L1 + L3（现有逻辑不变）
  │
  └─ 规则引擎也失败 → 返回 N/A (-1)
```

---

## 3. 接口变更

### 3.1 CanonicalSession 新增 Metadata 约定

```typescript
// session.metadata 新增字段 (Adapter 注入, Core Engine 读取):
interface SessionMetadataSentiment {
  negative_ratio: number;   // 0.0~1.0
  attitude_score: number;   // 1~10
  approval_ratio: number;   // 0.0~1.0
  source: "llm" | "builtin" | "na";  // 结果来源
  tokens?: {                 // LLM 调用消耗 (仅 source=llm 时)
    prompt: number;
    completion: number;
  };
}

interface SessionMetadataTaskClassification {
  status: string;            // COMPLETED | ABANDONED | UNCERTAIN
  confidence: number;        // 0.0~1.0
  source: "llm" | "l1" | "l3" | "rule";
}
```

### 3.2 ReflectRequest 变更

```go
// ReflectConfig 变更:
type ReflectConfig struct {
    ModelID          string `json:"model_id"`           // 移除 (不再需要)
    SentimentEnabled bool   `json:"sentiment_enabled"`  // 保留
    // 新增:
    SentimentSource  string `json:"sentiment_source"`   // "llm" | "builtin" | "na"
}
```

### 3.3 Core Engine server.go 变更

```go
// 原来 (sentiment.go 中的逻辑):
if cfg.SentimentEnabled {
    negativeRatio, attitudeScore, approvalRatio := engine.AnalyzeSentimentOrNA(ctx, ...)
    metrics.HumanNegativeRatio = negativeRatio
    ...
}

// 改为:
if cfg.SentimentEnabled {
    // 从 session metadata 提取
    if sr, ok := session.Metadata["sentiment"]; ok {
        if m, ok := sr.(map[string]interface{}); ok {
            metrics.HumanNegativeRatio = getFloat(m, "negative_ratio", -1)
            metrics.HumanAttitudeScore = getFloat(m, "attitude_score", -1)
            metrics.HumanApprovalRatio = getFloat(m, "approval_ratio", -1)
        }
    }
}
```

### 3.4 Core Engine classifier.go 变更

```go
// 原来:
func ClassifyTask(session, client LLMClient) {
    L1 → L2(client.ClassifyIntent) → L3
}

// 改为:
func ClassifyTask(session, metadata map[string]interface{}) {
    L1(检查 /opsx:archive)
    → 从 metadata["task_classification"] 读取 L2 结果
    → L3(关键词匹配)
}
```

---

## 4. Adapter 新增文件

### 4.1 `adapters/opencode/src/llm-agent.ts`

核心 LLM 调用封装，负责通过 SDK 执行 LLM 交互：

```typescript
export interface LLMResult<T> {
  data: T;
  tokens: { prompt: number; completion: number };
  source: "llm";
}

export class OpencodeLLMAgent {
  private client: OpencodeClient;
  private modelInfo: ModelInfo | null;

  // 通过 SDK session.prompt() 调用 LLM
  async callLLM(systemPrompt: string, userContent: string): Promise<LLMResult<string>>

  // 获取默认模型信息
  async resolveModel(): Promise<ModelInfo>

  // 创建和清理影子 session
  private async createShadowSession(prefix: string): Promise<string>
  private async deleteShadowSession(id: string): Promise<void>
}
```

### 4.2 `adapters/opencode/src/sentiment-analyzer.ts`

情感分析智能体：

```typescript
export class SentimentAnalyzer {
  private llmAgent: OpencodeLLMAgent;
  private builtinFallback: BuiltinSentimentEngine;

  // 分析单个 session 的情感
  async analyzeSession(session: CanonicalSession): Promise<SentimentResult>

  // 降级到规则引擎
  private builtinAnalyze(messages: string[]): SentimentResult
}
```

### 4.3 `adapters/opencode/src/task-classifier.ts`

任务分类智能体（L2 层）：

```typescript
export class TaskClassifier {
  private llmAgent: OpencodeLLMAgent;

  // L2 分类
  async classify(messages: CanonicalMessage[]): Promise<TaskClassification | null>
}
```

### 4.4 `adapters/opencode/src/builtin-sentiment.ts`

规则引擎降级：

```typescript
export class BuiltinSentimentEngine {
  // 基于关键词的简易情感分析
  analyze(messages: string[]): SentimentResult
}
```

### 4.5 `adapters/opencode/src/message-sanitizer.ts`

从 Go 移植的脱敏逻辑（API key、Bearer token、密码等正则替换）。

---

## 5. 配置变更

### 5.1 reflector.yaml

```yaml
# 模型配置（仅在 SDK 获取不到默认模型时使用）
model:
  override: ""                    # 强制覆盖宿主默认模型，为空则自动获取

# 情感分析配置
sentiment:
  enabled: true
  skipOnMessageTrigger: true      # N轮对话触发时跳过（节省 Token）
  mode: "agent"                   # agent(推荐) | builtin(规则引擎) | off(关闭)

# 任务分类配置（L2 层）
classification:
  enabled: true                   # 是否启用 L2 LLM 分类
  confidence_threshold: 0.7       # L2 置信度阈值
```

### 5.2 Adapter 配置接口变更

```typescript
export interface ReflectorConfig {
  // ... 保留现有字段 ...
  model: {
    id: string;          // 兼容旧配置
    override: string;    // 新增: 强制覆盖模型
  };
  sentiment: {
    enabled: boolean;
    skipOnMessageTrigger: boolean;
    mode: "agent" | "builtin" | "off";  // 新增
  };
  classification: {       // 新增
    enabled: boolean;
    confidenceThreshold: number;
  };
}
```

---

## 6. Core Engine 变更

### 6.1 删除的代码

| 文件 | 删除内容 |
|------|---------|
| `engine/sentiment.go` | `SentimentAnalyzer` 接口、`AnalyzeSentiment`、`AnalyzeSentimentOrNA` 函数 |
| `engine/sentiment.go` | 保留: `SentimentResult`、`SanitizeMessages`（Go 端可保留用于其他场景或删除） |
| `engine/classifier.go` | `LLMClient` 接口、`checkL2` 函数 |
| `server/server.go` | `NewServer` 的 `llmClient` 和 `analyzer` 参数 |
| `server/server.go` | `Server.llmClient` 和 `Server.analyzer` 字段 |
| `server/server_test.go` | mock LLM/analyzer 相关代码 |
| `server/integration_test.go` | 相关调整 |

### 6.2 修改的代码

| 文件 | 修改内容 |
|------|---------|
| `engine/classifier.go` | `ClassifyTask` 签名改为接受 `metadata map[string]interface{}` |
| `server/server.go` | 情感分析从 metadata 读取，不再调用 analyzer |
| `server/server.go` | `DetermineTaskStatus` 传入 metadata |
| `model/api.go` | `ReflectConfig` 移除 `ModelID`，新增 `SentimentSource` |

### 6.3 新增的代码

| 文件 | 新增内容 |
|------|---------|
| `model/api.go` 或新文件 | metadata 解析 helper 函数 |

---

## 7. 开放问题

### 7.1 待 PoC 验证

1. **`session.prompt()` 的 `system` 参数是否生效？** SDK 类型定义中有此字段，但需验证运行时行为。
2. **`tools: {}` 是否能禁用所有工具？** 需确认 LLM 不会在情感分析时调用工具。
3. **影子 session 的 Token 是否影响用户统计？** 删除 session 后 token 是否也从统计中移除。
4. **`session.prompt()` 响应的 parts 结构？** 需确认如何正确提取纯文本 JSON。

### 7.2 多工具兼容性

| 工具 | SDK prompt 支持 | 状态 |
|------|----------------|------|
| opencode | ✅ `client.session.prompt()` | 原生支持 |
| claudecode | ❓ 待调研 | 可能通过 MCP tool |
| openclaw | ❓ 待调研 | 可能类似 opencode |

降级方案：不支持 SDK prompt 的工具 → 规则引擎降级。

### 7.3 Token 消耗追踪

情感分析消耗的 Token 需要单独统计：
- 影子 session 的 AssistantMessage.tokens 可获取 prompt/completion token
- 日报中应将"反思工具自身 Token 消耗"与"任务 Token 消耗"分开显示

---

## 8. 实施计划

| Phase | 内容 | 预估 |
|-------|------|------|
| Phase 0 | SDK PoC 验证（system prompt / tools:{} / 影子 session） | 0.5 天 |
| Phase 1 | Adapter 侧实现（llm-agent + sentiment-analyzer + task-classifier + builtin-sentiment + sanitizer） | 2 天 |
| Phase 2 | Core Engine 适配（修改 server.go / classifier.go，删除旧接口） | 1 天 |
| Phase 3 | 测试更新（单元测试 + 集成测试 + 配置更新） | 1 天 |
| Phase 4 | 文档更新 + E2E 验证 | 0.5 天 |

---

## 9. 风险与缓解

| 风险 | 等级 | 缓解措施 |
|------|------|---------|
| SDK `system` 参数不生效 | 🔴 高 | PoC 先验证，不生效则用 user message 前缀替代 |
| 影子 session 创建/删除失败 | 🟡 中 | try/catch + 最终清理保证 + 超时控制 |
| 定时触发时 opencode 不活跃 | 🟡 中 | 规则引擎降级 |
| 情感分析 Token 消耗过高 | 🟢 低 | skipOnMessageTrigger + mode=builtin 选项 |
| 多工具 Adapter 差异 | 🟡 中 | LLM Agent 抽象接口，每个工具实现自己的 Adapter |
