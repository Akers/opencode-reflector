import * as assert from "node:assert";
import { TaskClassifier } from "../task-classifier.js";
import { OpencodeLLMAgent } from "../llm-agent.js";
import { describe, test, getResults, reset } from "./runner.js";

describe("task-classifier", () => {
    test("successful classification with high confidence", async () => {
        const mockAgent = {
            callLLM: async () => ({
                text: '{"status":"COMPLETED","confidence":0.9}',
                tokens: { prompt: 5, completion: 10 },
            }),
        } as unknown as OpencodeLLMAgent;
        const classifier = new TaskClassifier(mockAgent, "/tmp/prompts", "");
        const result = await classifier.classify([{ role: "user", content: "do task" }]);
        assert.ok(result !== null);
        assert.strictEqual(result!.status, "COMPLETED");
        assert.strictEqual(result!.confidence, 0.9);
        assert.strictEqual(result!.source, "llm");
    });

    test("low confidence returns null", async () => {
        const mockAgent = {
            callLLM: async () => ({
                text: '{"status":"COMPLETED","confidence":0.5}',
                tokens: { prompt: 5, completion: 10 },
            }),
        } as unknown as OpencodeLLMAgent;
        const classifier = new TaskClassifier(mockAgent, "/tmp/prompts", "", 0.7);
        const result = await classifier.classify([{ role: "user", content: "do task" }]);
        assert.strictEqual(result, null);
    });

    test("LLM error returns null", async () => {
        const mockAgent = {
            callLLM: async () => { throw new Error("LLM error"); },
        } as unknown as OpencodeLLMAgent;
        const classifier = new TaskClassifier(mockAgent, "/tmp/prompts", "");
        const result = await classifier.classify([{ role: "user", content: "do task" }]);
        assert.strictEqual(result, null);
    });

    test("empty messages returns null", async () => {
        const mockAgent = {} as unknown as OpencodeLLMAgent;
        const classifier = new TaskClassifier(mockAgent, "/tmp/prompts", "");
        const result = await classifier.classify([]);
        assert.strictEqual(result, null);
    });
});

reset();
const { passed, failed } = getResults();
console.log(`\nResults: ${passed} passed, ${failed} failed\n`);
process.exit(failed > 0 ? 1 : 0);
