package engine

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed prompts/*.md
var defaultPrompts embed.FS

// LoadPrompt loads a prompt by name (e.g., "sentiment", "classify")
// Priority: file system (.reflector/prompts/{name}.md) > embedded default
func LoadPrompt(name string) (string, error) {
	return loadPromptFromFile(name, defaultPrompts)
}

// LoadPromptFromDir loads a prompt with custom filesystem root
func LoadPromptFromDir(name string, customDir string) (string, error) {
	// Try custom dir first
	customPath := filepath.Join(customDir, name+".md")
	content, err := os.ReadFile(customPath)
	if err == nil {
		trimmed := strings.TrimSpace(string(content))
		if trimmed != "" {
			return trimmed, nil
		}
		// Empty file → fall through to embedded
	}

	// Fallback to embedded
	return loadPromptFromEmbed(name, defaultPrompts)
}

func loadPromptFromFile(name string, embeddedFS embed.FS) (string, error) {
	// Try .reflector/prompts/ directory first
	customPath := filepath.Join(".reflector", "prompts", name+".md")
	content, err := os.ReadFile(customPath)
	if err == nil {
		trimmed := strings.TrimSpace(string(content))
		if trimmed != "" {
			return trimmed, nil
		}
	}

	return loadPromptFromEmbed(name, embeddedFS)
}

func loadPromptFromEmbed(name string, embeddedFS embed.FS) (string, error) {
	path := "prompts/" + name + ".md"
	content, err := embeddedFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("prompt %q not found in embedded defaults: %w", name, err)
	}
	return strings.TrimSpace(string(content)), nil
}

// ListEmbeddedPrompts lists all embedded prompt names
func ListEmbeddedPrompts() ([]string, error) {
	entries, err := fs.ReadDir(defaultPrompts, "prompts")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".md")
		names = append(names, name)
	}
	return names, nil
}
