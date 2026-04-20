import * as assert from "node:assert";

let passed = 0;
let failed = 0;

export function test(name: string, fn: () => void): void {
    try {
        fn();
        console.log(`  ✓ ${name}`);
        passed++;
    } catch (err: any) {
        console.log(`  ✗ ${name}: ${err.message}`);
        failed++;
    }
}

export function describe(name: string, fn: () => void): void {
    console.log(`\n${name}`);
    fn();
}

export function getResults(): { passed: number; failed: number } {
    return { passed, failed };
}

export function reset(): void {
    passed = 0;
    failed = 0;
}
