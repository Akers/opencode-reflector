import { OpencodeLLMAgent } from "./llm-agent.js";
import { builtinAnalyze } from "./builtin-sentiment.js";
import { sanitizeMessages } from "./message-sanitizer.js";
import { readFileSync, existsSync } from "node:fs";
import { join } from "node:path";

export interface SentimentAnalysisResult {
    negative_ratio: number;
    attitude_score: number;
    approval_ratio: number;
    source: "llm" | "builtin" | "na";
    tokens?: { prompt: number; completion: number };
}

export class SentimentAnalyzer {
    private llmAgent: OpencodeLLMAgent;
    private sentimentPrompt: string;
    private modelOverride: string;

    constructor(llmAgent: OpencodeLLMAgent, promptDir: string, modelOverride: string = "") {
        this.llmAgent = llmAgent;
        this.modelOverride = modelOverride;
        this.sentimentPrompt = this.loadPrompt(promptDir);
    }

    async analyzeSession(
        humanMessages: string[],
        mode: "agent" | "builtin" | "off" = "agent",
    ): Promise<SentimentAnalysisResult> {
        if (humanMessages.length === 0 || mode === "off") {
            return { negative_ratio: -1, attitude_score: -1, approval_ratio: -1, source: "na" };
        }

        if (mode === "builtin") {
            return builtinAnalyze(humanMessages);
        }

        try {
            const sanitized = sanitizeMessages(humanMessages);
            const userContent = sanitized.join("\n---\n");

            const result = await this.llmAgent.callLLM(this.sentimentPrompt, userContent, this.modelOverride);

            const parsed = this.extractJSON(result.text);
            if (parsed && typeof parsed.negative_ratio === "number") {
                return {
                    negative_ratio: parsed.negative_ratio,
                    attitude_score: parsed.attitude_score ?? -1,
                    approval_ratio: parsed.approval_ratio ?? -1,
                    source: "llm",
                    tokens: result.tokens,
                };
            }
        } catch (err) {
            console.warn("[reflector] LLM sentiment analysis failed, falling back to builtin:", err);
        }

        return builtinAnalyze(humanMessages);
    }

    private loadPrompt(promptDir: string): string {
        const promptPath = join(promptDir, "sentiment.md");
        if (existsSync(promptPath)) {
            return readFileSync(promptPath, "utf-8");
        }
        return `Analyze the sentiment of the following human messages from a coding session. 
Return ONLY a JSON object with these fields:
- "negative_ratio": float 0.0-1.0, ratio of negative sentiment
- "attitude_score": float 1-10, overall attitude score (1=very negative, 10=very positive)
- "approval_ratio": float 0.0-1.0, ratio of approval/positive feedback

Messages:`;
    }

    private extractJSON(text: string): any | null {
        const jsonMatch = text.match(/\{[\s\S]*\}/);
        if (!jsonMatch) return null;
        try {
            return JSON.parse(jsonMatch[0]);
        } catch {
            return null;
        }
    }
}
