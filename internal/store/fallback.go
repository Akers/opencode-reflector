package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FallbackWriter writes data to a JSON file as fallback when primary storage fails
type FallbackWriter struct {
	dir string // fallback directory path
}

// NewFallbackWriter creates a new FallbackWriter
func NewFallbackWriter(dir string) *FallbackWriter {
	return &FallbackWriter{dir: dir}
}

// Write writes data to a timestamped JSON file in the fallback directory
// Returns the path of the written file
func (w *FallbackWriter) Write(data interface{}) (string, error) {
	if err := w.EnsureDir(); err != nil {
		return "", fmt.Errorf("failed to ensure fallback directory: %w", err)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	filename := fmt.Sprintf("fallback-%s.json", time.Now().Format("20060102-150405.000000000"))
	filePath := filepath.Join(w.dir, filename)

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return "", fmt.Errorf("failed to write fallback file: %w", err)
	}

	return filePath, nil
}

// EnsureDir creates the fallback directory if it doesn't exist
func (w *FallbackWriter) EnsureDir() error {
	return os.MkdirAll(w.dir, 0755)
}
