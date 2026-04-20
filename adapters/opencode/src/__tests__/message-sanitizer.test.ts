import * as assert from "node:assert";
import { sanitizeMessages } from "../message-sanitizer.js";
import { describe, test, getResults, reset } from "./runner.js";

describe("message-sanitizer", () => {
    test("API key redaction", () => {
        const input = ["sk-abc123def456xyz789uvw012"];
        const result = sanitizeMessages(input);
        assert.strictEqual(result[0], "[REDACTED_API_KEY]");
    });

    test("Bearer token redaction", () => {
        const input = ["Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"];
        const result = sanitizeMessages(input);
        assert.strictEqual(result[0], "Bearer [REDACTED]");
    });

    test("Password in URL redaction", () => {
        const input = ["https://user:secret123@example.com/path"];
        const result = sanitizeMessages(input);
        assert.strictEqual(result[0], "https://[REDACTED]:[REDACTED]@example.com/path");
    });

    test("Generic password redaction (case insensitive)", () => {
        const input = [
            "password=abc123",
            "PASSWD=xyz789",
            "pwd=secret",
        ];
        const result = sanitizeMessages(input);
        assert.strictEqual(result[0], "password=[REDACTED]");
        assert.strictEqual(result[1], "PASSWD=[REDACTED]");
        assert.strictEqual(result[2], "pwd=[REDACTED]");
    });

    test("Normal text unchanged", () => {
        const input = ["Hello world, this is a normal message."];
        const result = sanitizeMessages(input);
        assert.strictEqual(result[0], "Hello world, this is a normal message.");
    });

    test("Multiple patterns in single message", () => {
        const input = ["Bearer token and sk-apiKey12345678901234567890 and password=secret"];
        const result = sanitizeMessages(input);
        // Bearer token  matches BEARER_TOKEN_PATTERN, replaced with "Bearer [REDACTED]"
        assert.strictEqual(result[0], "Bearer [REDACTED] and [REDACTED_API_KEY] and password=[REDACTED]");
    });

    test("Empty array", () => {
        const input: string[] = [];
        const result = sanitizeMessages(input);
        assert.deepStrictEqual(result, []);
    });
});

reset();
const { passed, failed } = getResults();
console.log(`\nResults: ${passed} passed, ${failed} failed\n`);
process.exit(failed > 0 ? 1 : 0);
