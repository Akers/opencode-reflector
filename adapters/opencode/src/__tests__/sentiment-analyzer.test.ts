import * as assert from "node:assert";
import { SentimentAnalyzer } from "../sentiment-analyzer.js";
import { OpencodeLLMAgent } from "../llm-agent.js";
import { describe, test, getResults, reset } from "./runner.js";

describe("sentiment-analyzer", () => {
    test("agent mode success", async () => {
        const mockAgent = {
            callLLM: async () => ({
                text: '{"negative_ratio":0.2,"attitude_score":7.5,"approval_ratio":0.8}',
                tokens: { prompt: 5, completion: 10 },
            }),
        } as unknown as OpencodeLLMAgent;
        const analyzer = new SentimentAnalyzer(mockAgent, "/tmp/prompts", "");
        const result = await analyzer.analyzeSession(["thanks"], "agent");
        assert.strictEqual(result.source, "llm");
        assert.strictEqual(result.negative_ratio, 0.2);
        assert.strictEqual(result.attitude_score, 7.5);
        assert.strictEqual(result.approval_ratio, 0.8);
        assert.ok(result.tokens);
        assert.strictEqual(result.tokens!.prompt, 5);
    });

    test("agent mode fallback to builtin on error", async () => {
        const mockAgent = {
            callLLM: async () => { throw new Error("LLM error"); },
        } as unknown as OpencodeLLMAgent;
        const analyzer = new SentimentAnalyzer(mockAgent, "/tmp/prompts", "");
        const result = await analyzer.analyzeSession(["thanks"], "agent");
        assert.strictEqual(result.source, "builtin");
    });

    test("builtin mode", async () => {
        const mockAgent = {} as unknown as OpencodeLLMAgent;
        const analyzer = new SentimentAnalyzer(mockAgent, "/tmp/prompts", "");
        const result = await analyzer.analyzeSession(["wrong"], "builtin");
        assert.strictEqual(result.source, "builtin");
        assert.ok(result.negative_ratio > 0.5);
    });

    test("off mode", async () => {
        const mockAgent = {} as unknown as OpencodeLLMAgent;
        const analyzer = new SentimentAnalyzer(mockAgent, "/tmp/prompts", "");
        const result = await analyzer.analyzeSession(["thanks"], "off");
        assert.strictEqual(result.source, "na");
        assert.strictEqual(result.negative_ratio, -1);
    });

    test("empty messages", async () => {
        const mockAgent = {} as unknown as OpencodeLLMAgent;
        const analyzer = new SentimentAnalyzer(mockAgent, "/tmp/prompts", "");
        const result = await analyzer.analyzeSession([], "agent");
        assert.strictEqual(result.source, "na");
        assert.strictEqual(result.negative_ratio, -1);
    });
});

reset();
const { passed, failed } = getResults();
console.log(`\nResults: ${passed} passed, ${failed} failed\n`);
process.exit(failed > 0 ? 1 : 0);
