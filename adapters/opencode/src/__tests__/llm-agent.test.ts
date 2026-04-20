import * as assert from "node:assert";
import { OpencodeLLMAgent } from "../llm-agent.js";
import { describe, test, getResults, reset } from "./runner.js";

describe("llm-agent", () => {
    test("resolveModel uses override when provided", async () => {
        const mockClient = { config: { get: async () => ({ data: { model: "other/model" } }) } };
        const agent = new OpencodeLLMAgent(mockClient);
        const result = await agent.resolveModel("custom/provider-model");
        assert.strictEqual(result.providerID, "custom");
        assert.strictEqual(result.modelID, "provider-model");
    });

    test("resolveModel gets from config when no override", async () => {
        const mockClient = { config: { get: async () => ({ data: { model: "someprovider/somemodel" } }) } };
        const agent = new OpencodeLLMAgent(mockClient);
        const result = await agent.resolveModel();
        assert.strictEqual(result.providerID, "someprovider");
        assert.strictEqual(result.modelID, "somemodel");
    });

    test("resolveModel falls back on error", async () => {
        const mockClient = { config: { get: async () => { throw new Error("fail"); } } };
        const agent = new OpencodeLLMAgent(mockClient);
        const result = await agent.resolveModel();
        assert.strictEqual(result.providerID, "minimax");
        assert.strictEqual(result.modelID, "MiniMax-M2.7-highspeed");
    });

    test("callLLM creates session, calls prompt, and deletes session", async () => {
        const sessionId = "test-session-123";
        const mockClient = {
            session: {
                create: async () => ({ data: { id: sessionId } }),
                prompt: async () => ({
                    data: {
                        parts: [{ type: "text", text: '{"result":"ok"}' }],
                        info: { tokens: { input: 10, output: 20 } },
                    },
                }),
                delete: async () => ({}),
            },
        };
        const agent = new OpencodeLLMAgent(mockClient);
        const result = await agent.callLLM("system prompt", "user content");
        assert.strictEqual(result.text, '{"result":"ok"}');
        assert.strictEqual(result.tokens.prompt, 10);
        assert.strictEqual(result.tokens.completion, 20);
    });
});

reset();
const { passed, failed } = getResults();
console.log(`\nResults: ${passed} passed, ${failed} failed\n`);
process.exit(failed > 0 ? 1 : 0);
