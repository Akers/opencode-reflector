package engine

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestHookPointConstants(t *testing.T) {
	// Verify all 10 hook point constants exist with correct values
	expected := map[HookPoint]string{
		HookBeforeReflect:   "before-reflect",
		HookAfterReflect:    "after-reflect",
		HookBeforeMetrics:   "before-save-metrics",
		HookAfterMetrics:    "after-save-metrics",
		HookBeforeSentiment: "before-sentiment",
		HookAfterSentiment:  "after-sentiment",
		HookBeforeReport:    "before-report",
		HookAfterReport:     "after-report",
		HookAfterClassify:   "after-classify",
		HookOnError:         "on-error",
	}

	for hp, expectedValue := range expected {
		if string(hp) != expectedValue {
			t.Errorf("expected HookPoint %v to have value %q, got %q", hp, expectedValue, string(hp))
		}
	}

	if len(expected) != 10 {
		t.Errorf("expected 10 hook point constants, got %d", len(expected))
	}
}

func TestScanHooksDir(t *testing.T) {
	// Create temporary directory with test hooks
	tmpDir := t.TempDir()

	// Create test hook files
	testHooks := map[string]string{
		"before-save-metrics.sh": `#!/bin/bash
cat`,
		"after-report.py":        `#!/usr/bin/env python3
import sys
import json
print(json.dumps({"status": "ok"}))`,
	}

	for filename, content := range testHooks {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			t.Fatalf("failed to write test hook %s: %v", filename, err)
		}
	}

	// Scan hooks
	hooks, err := ScanHooks(tmpDir)
	if err != nil {
		t.Fatalf("ScanHooks failed: %v", err)
	}

	// Verify mapping
	if len(hooks) != 2 {
		t.Errorf("expected 2 hook points, got %d", len(hooks))
	}

	if len(hooks[HookBeforeMetrics]) != 1 {
		t.Errorf("expected 1 script for HookBeforeMetrics, got %d", len(hooks[HookBeforeMetrics]))
	}

	if len(hooks[HookAfterReport]) != 1 {
		t.Errorf("expected 1 script for HookAfterReport, got %d", len(hooks[HookAfterReport]))
	}
}

func TestScanHooksNonexistentDir(t *testing.T) {
	hooks, err := ScanHooks("/nonexistent/path/hooks")
	if err != nil {
		t.Fatalf("ScanHooks should not return error for nonexistent dir: %v", err)
	}

	if len(hooks) != 0 {
		t.Errorf("expected empty map for nonexistent dir, got %d entries", len(hooks))
	}
}

func TestExecuteHook(t *testing.T) {
	// Create a test script that echoes stdin to stdout
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "echo_hook.sh")
	scriptContent := `#!/bin/bash
cat`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to write test script: %v", err)
	}

	// Test input and expected output
	input := map[string]string{"key": "value", "test": "data"}
	var output map[string]string

	err := ExecuteHook(context.Background(), scriptPath, input, &output)
	if err != nil {
		t.Fatalf("ExecuteHook failed: %v", err)
	}

	// Verify output matches input (echo back)
	if output["key"] != "value" || output["test"] != "data" {
		t.Errorf("expected output to match input, got %v", output)
	}
}

func TestExecuteHookFailure(t *testing.T) {
	// Create a test script that exits with code 1
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "fail_hook.sh")
	scriptContent := `#!/bin/bash
exit 1`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to write test script: %v", err)
	}

	input := map[string]string{"test": "data"}
	var output map[string]string

	err := ExecuteHook(context.Background(), scriptPath, input, &output)

	// Should return error but not panic
	if err == nil {
		t.Error("expected error from failing hook, got nil")
	}

	// Verify it's a HookError with exit code 1
	hookErr, ok := err.(*HookError)
	if !ok {
		t.Errorf("expected *HookError, got %T", err)
	} else if hookErr.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", hookErr.ExitCode)
	}
}

func TestExecuteHookTimeout(t *testing.T) {
	// Create a test script that sleeps
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "slow_hook.sh")
	scriptContent := `#!/bin/bash
sleep 10`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to write test script: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*1000*1000) // 50ms
	defer cancel()

	input := map[string]string{"test": "data"}
	var output map[string]string

	err := ExecuteHook(ctx, scriptPath, input, &output)

	// Should return error due to context cancellation
	if err == nil {
		t.Error("expected error from cancelled context, got nil")
	}

	// Should be a context error
	if _, ok := err.(*exec.ExitError); ok {
		// This is also acceptable since the process was killed
		t.Logf("Got ExitError as expected: %v", err)
	} else if ctx.Err() != nil {
		// Or context error
		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("expected DeadlineExceeded, got %v", ctx.Err())
		}
	}
}

func TestExecuteHookNoOutput(t *testing.T) {
	// Create a test script that doesn't produce output
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "nooutput_hook.sh")
	scriptContent := `#!/bin/bash
exit 0`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to write test script: %v", err)
	}

	input := map[string]string{"test": "data"}

	// Execute with nil output
	err := ExecuteHook(context.Background(), scriptPath, input, nil)
	if err != nil {
		t.Fatalf("ExecuteHook with nil output failed: %v", err)
	}
}

func TestScanHooksIgnoresHiddenFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create hidden file
	hiddenPath := filepath.Join(tmpDir, ".hidden-hook.sh")
	if err := os.WriteFile(hiddenPath, []byte(`#!/bin/bash
cat`), 0755); err != nil {
		t.Fatalf("failed to write hidden file: %v", err)
	}

	// Create normal file
	normalPath := filepath.Join(tmpDir, "before-reflect.sh")
	if err := os.WriteFile(normalPath, []byte(`#!/bin/bash
cat`), 0755); err != nil {
		t.Fatalf("failed to write normal file: %v", err)
	}

	hooks, err := ScanHooks(tmpDir)
	if err != nil {
		t.Fatalf("ScanHooks failed: %v", err)
	}

	// Hidden file should be ignored
	if len(hooks) != 1 {
		t.Errorf("expected 1 hook (hidden ignored), got %d", len(hooks))
	}

	if len(hooks[HookBeforeReflect]) != 1 {
		t.Error("expected HookBeforeReflect to have 1 script")
	}
}
