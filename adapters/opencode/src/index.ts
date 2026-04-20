/**
 * opencode-reflector Server Plugin.
 * Handles event listening, scheduled triggers, and tool registration.
 *
 * LLM interactions (sentiment analysis + task classification) are performed
 * here in the Adapter via opencode SDK, before sending data to Core Engine.
 */
import type { Plugin, PluginModule, PluginInput, Hooks } from "@opencode-ai/plugin";
import { tool } from "@opencode-ai/plugin";

import { EngineManager } from "./engine-manager.js";
import { loadConfig, type ReflectorConfig } from "./config.js";
import { fetchSessions, type CanonicalSession } from "./session-adapter.js";
import { OpencodeLLMAgent } from "./llm-agent.js";
import { SentimentAnalyzer } from "./sentiment-analyzer.js";
import { TaskClassifier } from "./task-classifier.js";

let engineManager: EngineManager | null = null;
let config: ReflectorConfig | null = null;
let scheduledTimer: ReturnType<typeof setTimeout> | null = null;
let messageCounter = 0;
let llmAgent: OpencodeLLMAgent | null = null;
let sentimentAnalyzer: SentimentAnalyzer | null = null;
let taskClassifier: TaskClassifier | null = null;

const reflectorPlugin: Plugin = async (input: PluginInput) => {
  const { client, directory, worktree } = input;

  // Load configuration
  config = loadConfig(directory);

  // Initialize engine manager
  engineManager = new EngineManager(config.port, directory);

  // Initialize LLM agent and analyzers
  llmAgent = new OpencodeLLMAgent(client as any);
  const promptDir = `${directory}/.reflector/prompts`;
  sentimentAnalyzer = new SentimentAnalyzer(llmAgent, promptDir, config.model.override);
  taskClassifier = new TaskClassifier(
    llmAgent, promptDir, config.model.override, config.classification.confidenceThreshold,
  );

  // Start Core Engine (non-blocking)
  engineManager.ensureRunning().then((ok) => {
    if (ok && config!.trigger.time.enabled) {
      setupSchedule(config!.trigger.time.schedule);
    }
  });

  const hooks: Hooks = {};

  // Register reflect tool
  hooks.tool = {
    reflect: tool({
      description:
        "触发一次反思分析。分析当前会话和未分析的历史会话，生成日报。",
      args: {
        detail: tool.schema
          .string()
          .optional()
          .describe("触发详情，如 'MANUAL' 或 'TASK_FINISHED'"),
      },
      async execute({ detail }, ctx) {
        if (!engineManager?.isRunning()) {
          return {
            output: "⚠️ Core Engine 未运行，无法执行反思分析。请检查 reflector 日志。",
          };
        }

        try {
          const sessions = await fetchSessions(client);

          ctx.metadata({
            title: `🔍 反思分析中... (${sessions.length} 个会话)`,
          });

          // Pre-process: sentiment analysis + task classification
          const processedSessions = await preprocessSessions(sessions, "MANUAL");

          const result = await engineManager.reflect({
            trigger_type: "MANUAL",
            trigger_detail: detail || "TOOL_TRIGGER",
            sessions: processedSessions,
            config: {
              sentiment_enabled: config!.sentiment.enabled,
              sentiment_source: config!.sentiment.mode === "off" ? "na" : config!.sentiment.mode,
            },
          });

          if (!result) {
            return { output: "❌ 反思分析失败，请检查 Core Engine 日志。" };
          }

          return {
            output:
              `✅ 反思分析完成！\n` +
              `- 分析会话数: ${result.sessions_analyzed}\n` +
              `- 状态: ${result.status}\n` +
              `- 日报路径: ${result.report_path}\n` +
              `- 耗时: ${result.duration_ms}ms\n` +
              `- Token 消耗: ${result.tokens_consumed}`,
          };
        } catch (err) {
          return {
            output: `❌ 反思分析异常: ${err}`,
          };
        }
      },
    }),

    // Cleanup tool
    cleanup: tool({
      description: "清理旧的反思数据。默认清理 90 天前的数据。",
      args: {
        days: tool.schema
          .number()
          .optional()
          .describe("保留最近多少天的数据，默认使用配置中的 retention.days"),
      },
      async execute({ days }, ctx) {
        if (!engineManager?.isRunning()) {
          return { output: "⚠️ Core Engine 未运行。" };
        }

        const retainDays = days || config!.retention.days;
        ctx.metadata({ title: `🧹 清理 ${retainDays} 天前的数据...` });

        const result = await engineManager.cleanup(retainDays);

        if (!result) {
          return { output: "❌ 清理失败。" };
        }

        return { output: `✅ 已清理 ${result.days} 天前的数据。` };
      },
    }),
  };

  // Listen for chat message events (for N_MESSAGES trigger)
  hooks["chat.message"] = async (input) => {
    if (!config?.trigger.events.enabled) return;
    if (!config.trigger.events.types.includes("N_MESSAGES")) return;
    if (!engineManager?.isRunning()) return;

    messageCounter++;
    if (messageCounter >= config.trigger.events.messageInterval) {
      messageCounter = 0;

      try {
        const sessions = await fetchSessions(client);

        // Skip sentiment on message trigger if configured
        const sentimentMode = config.sentiment.skipOnMessageTrigger
          ? "off" as const
          : config.sentiment.mode;

        const processedSessions = await preprocessSessions(sessions, "EVENTS", sentimentMode);

        await engineManager.reflect({
          trigger_type: "EVENTS",
          trigger_detail: "N_MESSAGES",
          sessions: processedSessions,
          config: {
            sentiment_enabled: config.sentiment.enabled && !config.sentiment.skipOnMessageTrigger,
            sentiment_source: config.sentiment.mode === "off" ? "na" : (sentimentMode === "off" ? "na" : sentimentMode),
          },
        });
      } catch (err) {
        console.error("[reflector] Event-triggered reflect failed:", err);
      }
    }
  };

  // Event handler for session-level events
  hooks.event = async ({ event }) => {
    // Could handle session.status events for TASK_FINISHED detection
  };

  return hooks;
};

/**
 * Pre-process sessions: run sentiment analysis and task classification
 * via SDK before sending to Core Engine.
 *
 * Results are injected into session.metadata for Core Engine to read.
 */
async function preprocessSessions(
  sessions: CanonicalSession[],
  triggerType: string,
  sentimentModeOverride?: "agent" | "builtin" | "off",
): Promise<CanonicalSession[]> {
  const sentimentMode = sentimentModeOverride ?? config!.sentiment.mode;

  for (const session of sessions) {
    // --- Sentiment Analysis ---
    if (config!.sentiment.enabled && sentimentMode !== "off") {
      const humanMessages = session.messages
        .filter((m) => m.role === "human")
        .map((m) => m.content);

      if (humanMessages.length > 0) {
        try {
          const sentimentResult = await sentimentAnalyzer!.analyzeSession(humanMessages, sentimentMode);
          session.metadata.sentiment = {
            negative_ratio: sentimentResult.negative_ratio,
            attitude_score: sentimentResult.attitude_score,
            approval_ratio: sentimentResult.approval_ratio,
            source: sentimentResult.source,
          };
          if (sentimentResult.tokens) {
            session.metadata.sentiment_tokens = sentimentResult.tokens;
          }
        } catch (err) {
          console.warn(`[reflector] Sentiment analysis failed for session ${session.id}:`, err);
        }
      }
    }

    // --- L2 Task Classification ---
    if (config!.classification.enabled) {
      try {
        const classification = await taskClassifier!.classify(session.messages);
        if (classification) {
          session.metadata.task_classification = {
            status: classification.status,
            confidence: classification.confidence,
            source: classification.source,
          };
        }
      } catch (err) {
        console.warn(`[reflector] Task classification failed for session ${session.id}:`, err);
      }
    }
  }

  return sessions;
}

/**
 * Setup a daily schedule for time-based reflection.
 */
function setupSchedule(scheduleStr: string): void {
  const [hours, minutes] = scheduleStr.split(":").map(Number);
  if (isNaN(hours) || isNaN(minutes)) {
    console.warn(`[reflector] Invalid schedule format: ${scheduleStr}`);
    return;
  }

  const scheduleNext = () => {
    const now = new Date();
    const next = new Date(now);
    next.setHours(hours, minutes, 0, 0);

    if (next <= now) {
      next.setDate(next.getDate() + 1);
    }

    const delayMs = next.getTime() - now.getTime();
    console.log(`[reflector] Next scheduled reflection at ${next.toISOString()} (in ${Math.round(delayMs / 60000)} minutes)`);

    scheduledTimer = setTimeout(async () => {
      if (!engineManager?.isRunning() || !config) return;

      try {
        console.log("[reflector] Executing scheduled reflection...");
        const sessions = await fetchSessions(undefined as any);

        // Time trigger always uses config sentiment mode
        const processedSessions = await preprocessSessions(sessions, "TIME");

        await engineManager.reflect({
          trigger_type: "TIME",
          trigger_detail: scheduleStr,
          sessions: processedSessions,
          config: {
            sentiment_enabled: config.sentiment.enabled,
            sentiment_source: config.sentiment.mode === "off" ? "na" : config.sentiment.mode,
          },
        });
      } catch (err) {
        console.error("[reflector] Scheduled reflect failed:", err);
      }

      scheduleNext();
    }, delayMs);
  };

  scheduleNext();
}

// Export as PluginModule
export default {
  server: reflectorPlugin,
} satisfies PluginModule;

// Cleanup on process exit
process.on("exit", () => {
  if (scheduledTimer) clearTimeout(scheduledTimer);
});

export { EngineManager, loadConfig };
