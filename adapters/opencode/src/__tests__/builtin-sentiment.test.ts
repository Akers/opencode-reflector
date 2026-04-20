import * as assert from "node:assert";
import { builtinAnalyze } from "../builtin-sentiment.js";
import { describe, test, getResults, reset } from "./runner.js";

describe("builtin-sentiment", () => {
    test("positive messages → low negative_ratio", () => {
        const input = ["谢谢你的帮助，完美！", "great job, excellent work!"];
        const result = builtinAnalyze(input);
        assert.ok(result.negative_ratio < 0.5, `Expected low negative_ratio, got ${result.negative_ratio}`);
        assert.ok(result.approval_ratio > 0.5, `Expected high approval_ratio, got ${result.approval_ratio}`);
    });

    test("negative messages → high negative_ratio", () => {
        const input = ["不对，错误", "this is wrong and broken"];
        const result = builtinAnalyze(input);
        assert.ok(result.negative_ratio > 0.5, `Expected high negative_ratio, got ${result.negative_ratio}`);
    });

    test("mixed messages", () => {
        const input = ["谢谢 but also wrong", "好的 but error"];
        const result = builtinAnalyze(input);
        assert.ok(result.negative_ratio > 0, `Expected some negative ratio, got ${result.negative_ratio}`);
    });

    test("neutral returns default values", () => {
        const input = ["hello world"];
        const result = builtinAnalyze(input);
        assert.strictEqual(result.negative_ratio, 0);
        assert.strictEqual(result.attitude_score, 5);
        assert.strictEqual(result.approval_ratio, 0.5);
    });

    test("empty messages returns -1 values", () => {
        const input: string[] = [];
        const result = builtinAnalyze(input);
        assert.strictEqual(result.negative_ratio, -1);
        assert.strictEqual(result.attitude_score, -1);
        assert.strictEqual(result.approval_ratio, -1);
    });
});

reset();
const { passed, failed } = getResults();
console.log(`\nResults: ${passed} passed, ${failed} failed\n`);
process.exit(failed > 0 ? 1 : 0);
