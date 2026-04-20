package engine

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// HookPoint represents a specific point in the reflection lifecycle where hooks can be attached
type HookPoint string

const (
	HookBeforeReflect   HookPoint = "before-reflect"
	HookAfterReflect    HookPoint = "after-reflect"
	HookBeforeMetrics   HookPoint = "before-save-metrics"
	HookAfterMetrics    HookPoint = "after-save-metrics"
	HookBeforeSentiment HookPoint = "before-sentiment"
	HookAfterSentiment  HookPoint = "after-sentiment"
	HookBeforeReport    HookPoint = "before-report"
	HookAfterReport     HookPoint = "after-report"
	HookAfterClassify   HookPoint = "after-classify"
	HookOnError         HookPoint = "on-error"
)

// ScanHooks scans the hooks directory and returns a map of HookPoint to script paths
func ScanHooks(hooksDir string) (map[HookPoint][]string, error) {
	result := make(map[HookPoint][]string)

	// Check if directory exists
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		return result, nil
	}

	// Read directory entries
	entries, err := os.ReadDir(hooksDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		// Skip hidden files and directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		if entry.IsDir() {
			continue
		}

		// Get file name without extension to determine hook point
		ext := filepath.Ext(entry.Name())
		hookName := strings.TrimSuffix(entry.Name(), ext)

		// Validate hook point
		hookPoint := HookPoint(hookName)
		if !isValidHookPoint(hookPoint) {
			continue
		}

		// Add script path to result
		scriptPath := filepath.Join(hooksDir, entry.Name())
		result[hookPoint] = append(result[hookPoint], scriptPath)
	}

	return result, nil
}

// isValidHookPoint checks if the given hook point is valid
func isValidHookPoint(hp HookPoint) bool {
	switch hp {
	case HookBeforeReflect, HookAfterReflect, HookBeforeMetrics, HookAfterMetrics,
		HookBeforeSentiment, HookAfterSentiment, HookBeforeReport, HookAfterReport,
		HookAfterClassify, HookOnError:
		return true
	default:
		return false
	}
}

// ExecuteHook executes a hook script with the given input and captures output
func ExecuteHook(ctx context.Context, hookPath string, input interface{}, output interface{}) error {
	// Serialize input to JSON
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return err
	}

	// Execute the hook script
	cmd := exec.CommandContext(ctx, hookPath)
	cmd.Stdin = strings.NewReader(string(inputJSON))
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		// Check if it's an exit error
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &HookError{
				ExitCode: exitErr.ExitCode(),
				Err:      err,
			}
		}
		return err
	}

	// If output is not nil, deserialize the output
	if output != nil {
		if err := json.Unmarshal(out, output); err != nil {
			return err
		}
	}

	return nil
}

// HookError represents an error from a hook script execution
type HookError struct {
	ExitCode int
	Err      error
}

func (e *HookError) Error() string {
	return e.Err.Error()
}

func (e *HookError) Unwrap() error {
	return e.Err
}
