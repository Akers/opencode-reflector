package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEmbeddedPrompts(t *testing.T) {
	// Verify 3 embedded prompts can be read and are non-empty
	promptNames := []string{"sentiment", "classify"}

	for _, name := range promptNames {
		content, err := LoadPrompt(name)
		if err != nil {
			t.Errorf("Failed to load embedded prompt %q: %v", name, err)
			continue
		}
		if content == "" {
			t.Errorf("Embedded prompt %q is empty", name)
		}
	}
}

func TestLoadPromptFromFile(t *testing.T) {
	// Create temp dir with custom prompt file
	tmpDir := t.TempDir()
	customName := "custom_test"
	customContent := "This is a custom prompt content"

	tmpFile := filepath.Join(tmpDir, customName+".md")
	err := os.WriteFile(tmpFile, []byte(customContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// LoadPromptFromDir should return file content
	content, err := LoadPromptFromDir(customName, tmpDir)
	if err != nil {
		t.Errorf("LoadPromptFromDir failed: %v", err)
		return
	}
	if content != customContent {
		t.Errorf("Expected %q, got %q", customContent, content)
	}
}

func TestLoadPromptFallback(t *testing.T) {
	// When file doesn't exist, should return embedded default
	content, err := LoadPrompt("sentiment")
	if err != nil {
		t.Errorf("Failed to load embedded sentiment prompt: %v", err)
		return
	}
	if content == "" {
		t.Error("Embedded sentiment prompt should not be empty")
	}
}

func TestLoadPromptEmptyFile(t *testing.T) {
	// Create temp dir with empty prompt file using existing embedded prompt name
	tmpDir := t.TempDir()
	emptyName := "sentiment" // use existing prompt name with empty file

	tmpFile := filepath.Join(tmpDir, emptyName+".md")
	err := os.WriteFile(tmpFile, []byte("   "), 0644) // whitespace only
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Should fallback to embedded (empty file)
	content, err := LoadPromptFromDir(emptyName, tmpDir)
	if err != nil {
		t.Errorf("LoadPromptFromDir failed: %v", err)
		return
	}
	if content == "" {
		t.Error("Should fallback to embedded when file is empty")
	}
}

func TestListEmbeddedPrompts(t *testing.T) {
	names, err := ListEmbeddedPrompts()
	if err != nil {
		t.Errorf("ListEmbeddedPrompts failed: %v", err)
		return
	}
	if len(names) < 2 {
		t.Errorf("Expected at least 2 embedded prompts, got %d: %v", len(names), names)
	}
}
