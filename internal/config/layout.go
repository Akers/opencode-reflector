package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDefaultLayout creates the .reflector directory structure if it doesn't exist.
func EnsureDefaultLayout(basePath string) error {
	subdirs := []string{
		"data",
		"data/fallback",
		"reports",
		"logs",
		"prompts",
		"hooks",
	}
	for _, dir := range subdirs {
		p := filepath.Join(basePath, dir)
		if err := os.MkdirAll(p, 0755); err != nil {
			return fmt.Errorf("create dir %s: %w", p, err)
		}
	}
	return nil
}
