import { OpencodeLLMAgent } from "./llm-agent.js";
import { readFileSync, existsSync } from "node:fs";
import { join } from "node:path";

export interface TaskClassification {
    status: string; // COMPLETED | ABANDONED | UNCERTAIN
    confidence: number; // 0.0-1.0
    source: "llm";
}

export class TaskClassifier {
    private llmAgent: OpencodeLLMAgent;
    private classifyPrompt: string;
    private modelOverride: string;
    private confidenceThreshold: number;

    constructor(
        llmAgent: OpencodeLLMAgent,
        promptDir: string,
        modelOverride: string = "",
        confidenceThreshold: number = 0.7,
    ) {
        this.llmAgent = llmAgent;
        this.modelOverride = modelOverride;
        this.confidenceThreshold = confidenceThreshold;
        this.classifyPrompt = this.loadPrompt(promptDir);
    }

    async classify(
        messages: Array<{ role: string; content: string }>,
    ): Promise<TaskClassification | null> {
        if (messages.length === 0) return null;

        try {
            const messagesStr = messages.map((m) => `[${m.role}]: ${m.content}`).join("\n");

            const result = await this.llmAgent.callLLM(
                this.classifyPrompt,
                messagesStr,
                this.modelOverride,
            );

            const parsed = this.extractJSON(result.text);
            if (parsed && typeof parsed.status === "string" && typeof parsed.confidence === "number") {
                if (parsed.confidence >= this.confidenceThreshold) {
                    return {
                        status: parsed.status,
                        confidence: parsed.confidence,
                        source: "llm",
                    };
                }
            }
        } catch (err) {
            console.warn("[reflector] LLM task classification failed:", err);
        }

        return null;
    }

    private loadPrompt(promptDir: string): string {
        const promptPath = join(promptDir, "classify.md");
        if (existsSync(promptPath)) {
            return readFileSync(promptPath, "utf-8");
        }
        return `Classify whether the task in this coding session was completed.
Return ONLY a JSON object:
- "status": "COMPLETED" | "ABANDONED" | "UNCERTAIN"
- "confidence": float 0.0-1.0

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
