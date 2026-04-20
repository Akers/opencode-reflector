# 六、接口设计

> 本文档是 [需求与设计方案主文档](../requirements-and-design.md) 的组成部分。

---

### 6.1 Core Engine HTTP API

#### `POST /api/v1/reflect`

触发一次反思分析。

**Request**：
```json
{
  "trigger_type": "EVENTS",
  "trigger_detail": "TASK_FINISHED",
  "sessions": [
    {
      "id": "sess_abc123",
      "tool_type": "opencode",
      "title": "实现用户登录功能",
      "messages": [
        {
          "role": "human",
          "content": "帮我实现一个用户登录功能",
          "timestamp": "2026-04-19T09:30:00Z",
          "metadata": {}
        },
        {
          "role": "agent",
          "agent_name": "main",
          "content": "好的，我来帮你实现...",
          "timestamp": "2026-04-19T09:30:05Z",
          "metadata": {
            "prompt_tokens": 1500,
            "completion_tokens": 800
          }
        }
      ],
      "tool_calls": [
        {
          "type": "TOOL",
          "name": "read",
          "duration_ms": 120,
          "success": true,
          "called_at": "2026-04-19T09:30:10Z"
        }
      ],
      "metadata": {}
    }
  ],
  "config": {
    "model_id": "minimax-cn-coding-plan/MiniMax-M2.7-highspeed",
    "sentiment_enabled": true
  }
}
```

**Response**：
```json
{
  "status": "SUCCESS",
  "sessions_analyzed": 1,
  "report_path": ".reflector/reports/dayreport-2026-04-19.md",
  "tokens_consumed": 2500,
  "duration_ms": 3200
}
```

#### `GET /api/v1/health`

健康检查。

#### `GET /api/v1/stats`

获取反思工具自身的运行统计。

### 6.2 Adapter 公共接口

```typescript
// 所有 Adapter 必须实现的接口
interface SessionAdapter {
  getToolType(): AgentToolType;
  getSessions(since: Date): Promise<CanonicalSession[]>;
  onTrigger(event: TriggerEvent, callback: () => void): void;
  getCapabilityMatrix(): CapabilityMatrix;
}

// 统一的会话数据模型
interface CanonicalSession {
  id: string;
  toolType: AgentToolType;
  title?: string;
  messages: CanonicalMessage[];
  toolCalls: CanonicalToolCall[];
  metadata: Record<string, unknown>;
}

interface CanonicalMessage {
  role: "human" | "agent" | "system";
  content: string;
  timestamp: string;         // ISO 8601
  agentName?: string;        // 智能体名称（role=agent时）
  promptTokens?: number;
  completionTokens?: number;
  metadata: Record<string, unknown>;
}

interface CanonicalToolCall {
  type: "TOOL" | "MCP" | "SKILL";
  name: string;
  durationMs?: number;
  success: boolean;
  calledAt: string;          // ISO 8601
}

// 能力矩阵
interface CapabilityMatrix {
  tokenMetrics: boolean;      // 是否能获取Token消耗
  toolCallDetails: boolean;   // 是否能获取工具调用详情
  mcpCallDetails: boolean;    // 是否能获取MCP调用详情
  skillCallDetails: boolean;  // 是否能获取Skill调用详情
  agentNames: boolean;        // 是否能区分不同智能体/子代理
}
```
