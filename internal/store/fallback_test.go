package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type testData struct {
	Name    string `json:"name"`
	Value   int    `json:"value"`
	Message string `json:"message"`
}

func TestFallbackWrite(t *testing.T) {
	dir := t.TempDir()
	writer := NewFallbackWriter(dir)

	data := testData{
		Name:    "test",
		Value:   42,
		Message: "hello",
	}

	path, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if path == "" {
		t.Fatal("returned path is empty")
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("file was not created at %s", path)
	}

	// Verify content
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var parsed testData
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if parsed.Name != data.Name || parsed.Value != data.Value || parsed.Message != data.Message {
		t.Errorf("data mismatch: got %+v, want %+v", parsed, data)
	}

	// Verify filename format
	filename := filepath.Base(path)
	if !strings.HasPrefix(filename, "fallback-") || !strings.HasSuffix(filename, ".json") {
		t.Errorf("unexpected filename format: %s", filename)
	}
}

func TestFallbackWriteEnsureDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "path", "to", "fallback")
	writer := NewFallbackWriter(dir)

	data := testData{Name: "test", Value: 1, Message: "ensure dir"}

	path, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Write failed when directory doesn't exist: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("file was not created in newly created directory")
	}
}

func TestFallbackWriteMultiple(t *testing.T) {
	dir := t.TempDir()
	writer := NewFallbackWriter(dir)

	paths := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		data := testData{Name: "test", Value: i, Message: "multiple"}
		path, err := writer.Write(data)
		if err != nil {
			t.Fatalf("Write %d failed: %v", i, err)
		}
		paths = append(paths, path)
		time.Sleep(time.Millisecond * 10) // Ensure different timestamps
	}

	// Verify all paths are unique
	pathSet := make(map[string]bool)
	for _, p := range paths {
		if pathSet[p] {
			t.Errorf("duplicate path generated: %s", p)
		}
		pathSet[p] = true
	}

	// Verify all files exist
	for _, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("file missing: %s", p)
		}
	}
}

func TestFallbackWriteInvalidPath(t *testing.T) {
	// Use a path that cannot be created (empty string or invalid)
	writer := NewFallbackWriter("")
	data := testData{Name: "test", Value: 1, Message: "invalid"}

	_, err := writer.Write(data)
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}
