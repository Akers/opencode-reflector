package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDefaultLayout(t *testing.T) {
	tmp := t.TempDir()
	reflectorDir := filepath.Join(tmp, ".reflector")

	if err := EnsureDefaultLayout(reflectorDir); err != nil {
		t.Fatalf("EnsureDefaultLayout failed: %v", err)
	}

	subdirs := []string{
		"data", "data/fallback", "reports", "logs", "prompts", "hooks",
	}
	for _, dir := range subdirs {
		p := filepath.Join(reflectorDir, dir)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("subdirectory %s should exist", dir)
		}
	}
}

func TestEnsureDefaultLayoutIdempotent(t *testing.T) {
	tmp := t.TempDir()
	reflectorDir := filepath.Join(tmp, ".reflector")

	if err := EnsureDefaultLayout(reflectorDir); err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	if err := EnsureDefaultLayout(reflectorDir); err != nil {
		t.Fatalf("second call (idempotent) failed: %v", err)
	}
}
