package skeleton_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryStructure(t *testing.T) {
	base := ".."
	dirs := []string{
		"cmd/reflector",
		"internal/api",
		"internal/engine",
		"internal/model",
		"internal/store",
		"internal/config",
		"adapters/opencode/src",
		"prompts",
		"scripts",
	}
	for _, dir := range dirs {
		p := filepath.Join(base, dir)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("directory %s should exist", dir)
		}
	}
}

func TestGoModExists(t *testing.T) {
	if _, err := os.Stat(filepath.Join("..", "go.mod")); os.IsNotExist(err) {
		t.Fatal("go.mod should exist at project root")
	}
}

func TestMainGoExists(t *testing.T) {
	if _, err := os.Stat(filepath.Join("..", "cmd", "reflector", "main.go")); os.IsNotExist(err) {
		t.Fatal("cmd/reflector/main.go should exist")
	}
}
