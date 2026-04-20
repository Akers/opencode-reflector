# 三、功能需求

> 本文档是 [需求与设计方案主文档](../requirements-and-design.md) 的组成部分。

---

### 3.1 触发机制

#### FR-001：定时触发（TIME）

- **描述**：每天在用户配置的固定时间点自动触发反思
- **默认时间**：00:00（零点）
- **配置项**：
  - `trigger.time`：24 小时制时间字符串，如 `"00:00"`、`"08:30"`
- **行为**：反思引擎对自上次反思完成以来到当前时间的所有新会话执行分析

#### FR-002：事件触发（EVENTS）

- **描述**：在特定事件发生时自动触发反思
- **支持的事件类型**：

| 事件类型 | 标识 | 触发条件 | 优先级 |
|---------|------|---------|--------|
| 任务完成 | `TASK_FINISHED` | 检测到任务结束信号（详见 FR-007） | 高 |
| N 轮对话 | `N_MESSAGES` | 自上次反思以来累计完成 N 轮对话（默认 N=10） | 中 |

- **配置项**：
  - `trigger.events`：事件类型数组，支持多选，如 `["TASK_FINISHED", "10_MESSAGES"]`
  - `trigger.messageInterval`：对话轮次间隔，默认 `10`

#### FR-003：手动触发（MANUAL）

- **描述**：用户主动发送指令触发反思
- **触发指令**：`/or:reflect_now`
- **行为**：立即对当前会话及自上次反思以来的所有未分析会话执行反思

#### FR-004：触发防抖与并发控制

- **描述**：防止短时间内重复触发导致的资源浪费和数据冲突
- **规则**：
  1. 同一触发源在 5 分钟内不重复触发（防抖）
  2. 若反思引擎正在运行，新的触发请求排队等待，不并发执行
  3. 手动触发优先级最高，可中断排队中的定时/事件触发

### 3.2 会话数据获取

#### FR-005：Session 数据采集

- **描述**：从智能体工具获取自上次反思完成以来的所有新增/更新会话数据
- **Watermark 机制**：
  - 使用"上次反思完成时间戳"作为 watermark
  - Watermark 持久化存储，重启后不丢失
  - 支持手动重置 watermark（用于调试或补分析历史数据）
- **数据范围**：获取 `watermark < session.lastMessageTime <= now` 的所有会话

#### FR-006：多工具适配（Adapter 模式）

- **描述**：通过 Adapter 层屏蔽不同智能体工具的 Session 数据格式差异
- **Adapter 接口契约**：

```typescript
interface SessionAdapter {
  /** 返回适配的智能体工具类型 */
  getToolType(): AgentToolType;

  /** 获取指定时间之后的所有会话 */
  getSessions(since: Date): Promise<CanonicalSession[]>;

  /** 注册触发事件监听 */
  onTrigger(event: TriggerEvent, callback: () => void): void;

  /** 获取当前工具的能力矩阵（哪些指标可获取） */
  getCapabilityMatrix(): CapabilityMatrix;
}
```

- **能力矩阵（Capability Matrix）**：
  - 每个 Adapter 声明自己能提供哪些数据指标
  - 不可获取的指标在日报中标注为 `N/A`
  - 避免"强行实现"导致不可靠的数据

#### FR-007：任务完成识别

- **描述**：自动判断会话中的任务是否成功完成
- **多级判定策略**（按优先级从高到低）：

| 级别 | 判定方式 | 可靠度 | 适用场景 |
|------|---------|--------|---------|
| L1 | OpenSpec `/opsx:archive` 指令完成 | ★★★★★ | 用户使用 OpenSpec 工作流时 |
| L2 | LLM 意图分类 | ★★★★ | 通用场景 |
| L3 | 关键词/模式匹配（兜底） | ★★★ | 模型不可用时的降级方案 |

- **L1 规则**：检测到用户发送 `/opsx:archive` 指令且智能体成功完成该指令
- **L2 规则**：调用配置的反思模型，以结构化方式判断最后一条 Agent 消息是否表示任务完成（输出 JSON：`{ status: "completed" | "interrupted" | "uncertain", confidence: number }`）
- **L3 规则**：中英文关键词匹配
  - 中文：包含"任务已完成"、"完成工作"、"工作完成"、"已完成"等
  - 英文：包含 "task completed"、"done"、"finished"、"all done" 等
  - 排除：包含否定语境的表述（如"无法完成"、"未能完成"）

#### FR-008：任务状态枚举

| 状态 | 标识 | 判定条件 |
|------|------|---------|
| 已完成 | `COMPLETED` | L1/L2/L3 判定为完成 |
| 已中断 | `INTERRUPTED` | 最后一条为 Agent 消息且未判定为完成 |
| 不确定 | `UNCERTAIN` | L2 置信度过低且 L3 无匹配 |
| 用户放弃 | `ABANDONED` | 最后一条为 Human 消息且之后无 Agent 回复 |

### 3.3 监控指标提取

#### FR-009：指标体系定义

所有指标以 Session 为采集主体，分为以下几类：

##### A. 时间指标

| 指标 ID | 指标名称 | 数据类型 | 说明 |
|---------|---------|---------|------|
| M-001 | 任务开始时间 | `datetime` | 会话第一条消息的产生时间 |
| M-002 | 任务结束时间 | `datetime` | 会话最后一条消息的产生时间 |
| M-003 | 任务耗时 | `duration` | M-002 - M-001，单位秒 |
| M-004 | 人类参与时长 | `duration` | 估算值：人类消息发送的时间跨度（首条人类消息到末条人类消息） |

##### B. Token 与请求指标

| 指标 ID | 指标名称 | 数据类型 | 说明 |
|---------|---------|---------|------|
| M-005 | Prompt Token 消耗 | `integer` | 会话中所有 LLM 请求的 prompt token 总和 |
| M-006 | Completion Token 消耗 | `integer` | 会话中所有 LLM 请求的 completion token 总和 |
| M-007 | 总 Token 消耗 | `integer` | M-005 + M-006 |
| M-008 | LLM 请求次数 | `integer` | 一次上行+一次下行计为一次 |

##### C. 工具/MCP/Skills 调用指标

| 指标 ID | 指标名称 | 数据类型 | 说明 |
|---------|---------|---------|------|
| M-009 | 工具调用次数 | `map<string, int>` | 按工具名分组统计，如 `{read: 7, write: 3, bash: 12}` |
| M-010 | 工具调用清单 | `array<ToolCall>` | 每次调用的详细信息（工具名、耗时、是否成功） |
| M-011 | MCP 调用次数 | `map<string, int>` | 按 MCP 服务名分组统计 |
| M-012 | MCP 调用清单 | `array<MCPCall>` | 每次调用的详细信息（MCP 名、耗时、是否成功） |
| M-013 | Skill 调用次数 | `map<string, int>` | 按 Skill 名分组统计 |
| M-014 | Skill 调用清单 | `array<SkillCall>` | 每次调用的详细信息（Skill 名） |

##### D. 智能体参与指标

| 指标 ID | 指标名称 | 数据类型 | 说明 |
|---------|---------|---------|------|
| M-015 | 智能体/子代理数量 | `integer` | 参与当前会话的不同智能体/子代理数量 |
| M-016 | 智能体/子代理参与度 | `map<string, float>` | 每个智能体消息数 / 总智能体消息数 × 100% |

##### E. 消息统计指标

| 指标 ID | 指标名称 | 数据类型 | 说明 |
|---------|---------|---------|------|
| M-017 | 会话总消息数 | `integer` | 包含 human/agent/tool/mcp 所有消息 |
| M-018 | 智能体消息数 | `integer` | 不含人类发送消息的消息总数 |
| M-019 | 人类消息数 | `integer` | 人类发送的指令/消息数量 |

##### F. 人类参与度指标

| 指标 ID | 指标名称 | 数据类型 | 说明 |
|---------|---------|---------|------|
| M-020 | 人类参与率 | `float` | M-019 / M-017 × 100% |
| M-021 | 人类介入次数 | `integer` | 人类在智能体工作过程中主动介入（非首条消息）的次数 |
| M-022 | 会话中断次数 | `integer` | 包含中断/恢复关键词的消息对数量 |

##### G. 情感分析指标

| 指标 ID | 指标名称 | 数据类型 | 说明 |
|---------|---------|---------|------|
| M-023 | 人类负面情绪占比 | `float` | 负面情绪消息数 / 人类消息总数 × 100% |
| M-024 | 人类态度评分 | `integer(1-10)` | 1=极端暴躁/质疑/反感，10=非常满意 |
| M-025 | 人类认可度 | `integer(1-10)` | 对 Agent 工作成果的认可程度 |

#### FR-010：情感分析实现规范

- **分析方式**：调用用户配置的反思模型（LLM），批量送入会话中所有人类消息
- **Prompt 设计**：
  - 使用 `.md` 文件存储，支持热修改（详见 FR-015）
  - 要求模型输出结构化 JSON：`{ sentiment_score: number, attitude_score: number, recognition_score: number, negative_ratio: number, analysis_summary: string }`
- **成本控制**：
  - 情感分析可通过配置 `sentiment.enabled: false` 完全禁用
  - 仅在任务完成触发和定时触发时执行，10 轮对话触发默认跳过情感分析
  - 在日报中单独统计反思工具自身的 Token 消耗，与任务 Token 消耗区分
- **脱敏处理**：送入模型前，对消息内容做脱敏（移除 API Key、密码、Token 等 pattern）

### 3.4 数据持久化

#### FR-011：本地持久化存储

- **描述**：将所有采集的指标数据持久化存储到本地
- **存储方案**：SQLite 数据库（单文件，轻量，与 Go 天然亲和）
- **存储路径**：`<project-root>/.reflector/data/reflector.db`
- **核心数据表设计**：

```sql
-- 会话基础信息
CREATE TABLE sessions (
    id              TEXT PRIMARY KEY,       -- 会话ID
    agent_tool      TEXT NOT NULL,          -- 智能体工具标识(opencode/openclaw/claudecode)
    title           TEXT,                   -- 会话标题
    start_time      DATETIME NOT NULL,      -- 任务开始时间
    end_time        DATETIME NOT NULL,      -- 任务结束时间
    status          TEXT NOT NULL,          -- COMPLETED/INTERRUPTED/UNCERTAIN/ABANDONED
    prompt_tokens   INTEGER DEFAULT 0,      -- Prompt Token 消耗
    completion_tokens INTEGER DEFAULT 0,    -- Completion Token 消耗
    llm_request_count INTEGER DEFAULT 0,    -- LLM 请求次数
    total_messages  INTEGER DEFAULT 0,      -- 总消息数
    agent_messages  INTEGER DEFAULT 0,      -- 智能体消息数
    human_messages  INTEGER DEFAULT 0,      -- 人类消息数
    human_participation_rate REAL DEFAULT 0,-- 人类参与率
    human_intervention_count INTEGER DEFAULT 0, -- 人类介入次数
    interrupt_count INTEGER DEFAULT 0,      -- 中断次数
    agent_count     INTEGER DEFAULT 0,      -- 智能体/子代理数量
    negative_emotion_ratio REAL DEFAULT 0,  -- 负面情绪占比
    attitude_score  INTEGER,                -- 态度评分(1-10)
    recognition_score INTEGER,              -- 认可度(1-10)
    reflector_tokens INTEGER DEFAULT 0,     -- 反思工具自身消耗的Token
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 智能体参与详情
CREATE TABLE agent_participations (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id      TEXT NOT NULL REFERENCES sessions(id),
    agent_name      TEXT NOT NULL,          -- 智能体/子代理名称
    message_count   INTEGER NOT NULL,       -- 该智能体的消息数
    participation_rate REAL NOT NULL,       -- 参与度百分比
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 工具/MCP/Skill 调用记录
CREATE TABLE tool_calls (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id      TEXT NOT NULL REFERENCES sessions(id),
    call_type       TEXT NOT NULL,          -- TOOL/MCP/SKILL
    tool_name       TEXT NOT NULL,          -- 工具/MCP/Skill名称
    duration_ms     INTEGER,               -- 调用耗时(毫秒)
    success         BOOLEAN DEFAULT TRUE,   -- 是否成功
    called_at       DATETIME               -- 调用时间
);

-- 反思执行日志
CREATE TABLE reflection_logs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    trigger_type    TEXT NOT NULL,          -- TIME/EVENTS/MANUAL
    trigger_detail  TEXT,                   -- 触发详情(如事件类型)
    sessions_count  INTEGER DEFAULT 0,      -- 本次分析的会话数
    tokens_consumed INTEGER DEFAULT 0,      -- 本次反思消耗的Token
    status          TEXT NOT NULL,          -- SUCCESS/PARTIAL/FAILED
    error_message   TEXT,                   -- 错误信息(如有)
    started_at      DATETIME NOT NULL,
    completed_at    DATETIME
);

-- Watermark（水位标记）
CREATE TABLE watermarks (
    id              INTEGER PRIMARY KEY CHECK (id = 1), -- 单行表
    last_reflection_time DATETIME NOT NULL,   -- 上次反思完成时间
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_sessions_start_time ON sessions(start_time);
CREATE INDEX idx_sessions_end_time ON sessions(end_time);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_tool_calls_session ON tool_calls(session_id);
CREATE INDEX idx_tool_calls_type_name ON tool_calls(call_type, tool_name);
```

#### FR-012：数据保留与清理

- **描述**：提供数据生命周期管理能力
- **默认保留期**：90 天
- **清理方式**：提供 `/or:cleanup` 命令，支持 `--days N` 参数指定清理 N 天前的数据
- **行为**：清理时同时删除数据库记录和对应的日报文件

### 3.5 日报生成

#### FR-013：日报生成规则

- **生成路径**：`<project-root>/.reflector/reports/dayreport-<report-day>.md`
- **日期归属规则**：
  - **定时触发（默认 00:00）**：`report-day` = 前一天日期（记录前一个自然日的工作）
  - **事件/手动触发**：`report-day` = 当前日期（记录当日截至触发时刻的工作）
  - **简化原则**：以自然日（00:00-24:00）为边界，不做跨日切分
- **幂等性**：同一 `report-day` 的日报，如已存在则追加更新（不覆盖已有内容，仅追加新分析的会话数据）

#### FR-014：日报内容结构

```markdown
# 📊 反思日报 - {report-day}

## 一、任务清单

| # | 会话ID | 会话标题 | 开始时间 | 结束时间 | Token消耗 | 人类参与度 | 负面情绪 | 任务状态 |
|---|--------|---------|---------|---------|----------|----------|---------|---------|
| 1 | abc123 | xxx     | 09:30   | 11:45   | 15,230   | 23.5%    | 5.2%    | ✅ 完成  |
| 2 | def456 | xxx     | 14:00   | 14:30   | 3,120    | 45.0%    | 0%      | ⚠️ 中断  |

## 二、汇总统计

### 2.1 整体指标
- **总会话数**：N
- **任务完成率**：X%
- **总 Token 消耗**：XXX（Prompt: XXX, Completion: XXX）
- **总 LLM 请求次数**：XXX
- **反思工具自身 Token 消耗**：XXX
- **总人类参与时长**：Xh Xm
- **平均人类参与率**：X%
- **平均人类态度评分**：X/10

### 2.2 态度分布
- 😊 满意（8-10分）：X%
- 😐 中立（4-7分）：X%
- 😠 不满（1-3分）：X%

## 三、工具使用统计

### 3.1 内置工具
| 工具名 | 调用次数 |
|--------|---------|
| read   | 42      |
| write  | 18      |
| bash   | 35      |

### 3.2 MCP 服务
| MCP名称 | 调用次数 |
|---------|---------|
| contex7 | 7       |

### 3.3 Skills
| Skill名称 | 调用次数 |
|----------|---------|
| pptx     | 1       |
| pdf      | 10      |

## 四、反思工具自身开销
- 本次反思 Token 消耗：XXX
- 反思触发方式：TIME / EVENTS / MANUAL
```

### 3.6 配置管理

#### FR-015：配置文件规范

- **配置文件路径**：`<project-root>/.reflector/reflector.yaml`
- **配置文件格式**：YAML
- **配置热更新**：修改配置文件后无需重启，下次触发时自动加载新配置

```yaml
# opencode-reflector 配置文件

# 触发配置
trigger:
  # 定时触发
  time:
    enabled: true
    schedule: "00:00"    # 24小时制

  # 事件触发
  events:
    enabled: true
    types:
      - TASK_FINISHED
      - N_MESSAGES
    messageInterval: 10  # N轮对话触发

# 模型配置
model:
  id: "minimax-cn-coding-plan/MiniMax-M2.7-highspeed"

# 情感分析配置
sentiment:
  enabled: true                    # 是否启用情感分析
  skipOnMessageTrigger: true       # N轮对话触发时跳过情感分析

# 数据保留
retention:
  days: 90                         # 数据保留天数

# 日报配置
report:
  template: "default"              # 日报模板名称（未来支持自定义模板）

# 日志级别
logLevel: "info"                   # debug/info/warn/error
```

#### FR-016：提示词热修改

- **描述**：所有 LLM 提示词以 `.md` 文件存储，支持运行时热修改
- **提示词存储路径**：`<project-root>/.reflector/prompts/`

| 文件名 | 用途 |
|--------|------|
| `task-status-classification.md` | 任务状态分类提示词（L2 判定） |
| `sentiment-analysis.md` | 情感分析提示词 |
| `report-summary.md` | 日报汇总生成提示词（预留） |

- **Fallback 机制**：若 `.md` 文件不存在或内容为空，使用编译时嵌入的默认提示词，并在日志中输出 WARNING

### 3.7 扩展机制

#### FR-017：Hook 生命周期

- **描述**：通过 Hook 机制提供便捷的自定义扩展能力
- **Hook 实现形式**：脚本式 Hook（语言无关）

```
.reflector/hooks/
  before-save-metrics.sh      # 保存指标前
  after-save-metrics.sh       # 保存指标后
  before-sentiment.py         # 情感分析前
  after-sentiment.py          # 情感分析后
  before-report.go            # 日报生成前（编译后的二进制）
  after-report.sh             # 日报生成后
```

- **通信协议**：
  - Core Engine 通过 `stdin` 传入 JSON 数据
  - Hook 脚本通过 `stdout` 返回 JSON 数据
  - 通过 `exit code` 标识成功（0）或失败（非 0）
  - Hook 抛异常不阻断主流程，仅记录警告日志
- **完整 Hook 节点**：

| Hook 节点 | 触发时机 | 输入数据 | 可修改 |
|----------|---------|---------|--------|
| `before-fetch-sessions` | 获取会话前 | watermark 时间 | ✅ 可修改时间范围 |
| `after-fetch-sessions` | 获取会话后 | 会话列表 | ✅ 可过滤/增强 |
| `before-task-classify` | 任务状态判定前 | 最后一条 Agent 消息 | ✅ 可预处理 |
| `after-task-classify` | 任务状态判定后 | 判定结果 | ✅ 可覆盖判定 |
| `before-sentiment` | 情感分析前 | 人类消息列表 | ✅ 可脱敏/过滤 |
| `after-sentiment` | 情感分析后 | 情感分析结果 | ✅ 可修正分数 |
| `before-save-metrics` | 保存指标前 | 完整指标数据 | ✅ 可增删改指标 |
| `after-save-metrics` | 保存指标后 | 已保存的指标数据 | ❌ 只读 |
| `before-report` | 日报生成前 | 当日指标汇总 | ✅ 可自定义数据 |
| `after-report` | 日报生成后 | 日报文件路径 | ❌ 只读 |
