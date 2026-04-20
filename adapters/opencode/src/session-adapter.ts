/**
 * Session data adapter.
 * Fetches sessions from opencode SDK and converts to CanonicalSession format.
 */
import type { Session, Message, Part } from "@opencode-ai/sdk";

// Re-export the SDK client type for convenience
type OpencodeSessionClient = {
  list(options?: { query?: { directory?: string } }): Promise<
    | { data: Session[]; error: undefined; request: Request; response: Response }
    | { data: undefined; error: unknown; request: Request; response: Response }
    | undefined
  >;
  messages(options: {
    path: { id: string };
    query?: { directory?: string; limit?: number };
  }): Promise<
    | {
        data: { info: Message; parts: Part[] }[];
        error: undefined;
        request: Request;
        response: Response;
      }
    | { data: undefined; error: unknown; request: Request; response: Response }
    | undefined
  >;
};

// Canonical types matching the Go Core Engine model
export interface CanonicalMessage {
  role: "human" | "agent" | "system";
  content: string;
  timestamp: string; // ISO 8601
  agent_name?: string;
  prompt_tokens?: number;
  completion_tokens?: number;
  metadata: Record<string, unknown>;
}

export interface CanonicalToolCall {
  type: "TOOL" | "MCP" | "SKILL";
  name: string;
  duration_ms?: number;
  success: boolean;
  called_at: string; // ISO 8601
}

export interface CanonicalSession {
  id: string;
  tool_type: "opencode";
  title?: string;
  messages: CanonicalMessage[];
  tool_calls: CanonicalToolCall[];
  metadata: Record<string, unknown>;
}

export interface CapabilityMatrix {
  tokenMetrics: boolean;
  toolCallDetails: boolean;
  mcpCallDetails: boolean;
  skillCallDetails: boolean;
  agentNames: boolean;
}

/**
 * Fetch sessions created/updated since the given timestamp
 * and convert them to CanonicalSession format.
 */
export async function fetchSessions(
  client: { session: OpencodeSessionClient },
  since?: Date,
): Promise<CanonicalSession[]> {
  const result = await client.session.list();
  if (!result || result.error !== undefined) {
    console.warn("[reflector] Failed to list sessions:", result?.error);
    return [];
  }

  const sessions = result.data ?? [];
  const canonicalSessions: CanonicalSession[] = [];

  for (const session of sessions) {
    // Filter by time if watermark provided
    if (since && session.time.updated < since.getTime() / 1000) {
      continue;
    }

    const canonical = await convertSession(client.session, session.id, session.title);
    if (canonical) {
      canonicalSessions.push(canonical);
    }
  }

  return canonicalSessions;
}

/**
 * Convert a single opencode session to CanonicalSession format.
 */
async function convertSession(
  sessionClient: OpencodeSessionClient,
  sessionID: string,
  title?: string,
): Promise<CanonicalSession | null> {
  try {
    const result = await sessionClient.messages({ path: { id: sessionID } });
    if (!result || result.error !== undefined) {
      console.warn(`[reflector] Failed to get messages for session ${sessionID}:`, result?.error);
      return null;
    }

    const messagesWithParts = result.data ?? [];
    const canonicalMessages: CanonicalMessage[] = [];
    const toolCalls: CanonicalToolCall[] = [];

    for (const { info: msg, parts } of messagesWithParts) {
      if (msg.role === "user") {
        const textContent = extractTextFromParts(parts);
        canonicalMessages.push({
          role: "human",
          content: textContent,
          timestamp: new Date(msg.time.created).toISOString(),
          metadata: {
            id: msg.id,
            agent: (msg as Record<string, unknown>).agent,
            model: (msg as Record<string, unknown>).model,
          },
        });
      } else if (msg.role === "assistant") {
        const textContent = extractTextFromParts(parts);
        const assistantMsg = msg as Record<string, unknown>;
        const tokens = assistantMsg.tokens as {
          input: number;
          output: number;
          reasoning: number;
          cache: { read: number; write: number };
        } | undefined;

        canonicalMessages.push({
          role: "agent",
          content: textContent,
          timestamp: new Date(msg.time.created).toISOString(),
          prompt_tokens: tokens?.input,
          completion_tokens: tokens?.output,
          metadata: {
            id: msg.id,
            modelID: (assistantMsg as { modelID?: string }).modelID,
            providerID: (assistantMsg as { providerID?: string }).providerID,
            cost: (assistantMsg as { cost?: number }).cost,
            finish: (assistantMsg as { finish?: string }).finish,
            tokens: tokens
              ? {
                  input: tokens.input,
                  output: tokens.output,
                  reasoning: tokens.reasoning,
                  cache_read: tokens.cache?.read,
                  cache_write: tokens.cache?.write,
                }
              : undefined,
          },
        });

        // Extract tool calls from parts
        extractToolCallsFromParts(parts, toolCalls);
      }
    }

    return {
      id: sessionID,
      tool_type: "opencode",
      title: title || undefined,
      messages: canonicalMessages,
      tool_calls: toolCalls,
      metadata: {},
    };
  } catch (err) {
    console.error(`[reflector] Failed to convert session ${sessionID}:`, err);
    return null;
  }
}

/**
 * Extract text content from message parts.
 */
function extractTextFromParts(parts: Part[]): string {
  const texts: string[] = [];
  for (const part of parts) {
    if (part.type === "text") {
      const textPart = part as { type: "text"; text: string };
      texts.push(textPart.text);
    }
  }
  return texts.join("\n");
}

/**
 * Extract tool calls from message parts.
 */
function extractToolCallsFromParts(
  parts: Part[],
  toolCalls: CanonicalToolCall[],
): void {
  for (const part of parts) {
    if (part.type === "tool") {
      const toolPart = part as unknown as {
        type: string;
        tool: string;
        callID?: string;
        state?: string;
        id?: string;
      };
      const toolName = toolPart.tool || "unknown";
      const state = toolPart.state || "";

      toolCalls.push({
        type: "TOOL",
        name: toolName,
        success: state === "completed" || state === "success",
        called_at: new Date().toISOString(),
      });
    }
  }
}

/**
 * Get the capability matrix for opencode.
 */
export function getCapabilityMatrix(): CapabilityMatrix {
  return {
    tokenMetrics: true,
    toolCallDetails: true,
    mcpCallDetails: true,
    skillCallDetails: true,
    agentNames: true,
  };
}
