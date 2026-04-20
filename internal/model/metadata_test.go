package model

import "testing"

func TestGetMetadataFloat(t *testing.T) {
	t.Run("nil map returns default", func(t *testing.T) {
		result := GetMetadataFloat(nil, "key", 99.0)
		if result != 99.0 {
			t.Errorf("expected 99.0, got %v", result)
		}
	})

	t.Run("missing key returns default", func(t *testing.T) {
		meta := map[string]interface{}{"other": 1.0}
		result := GetMetadataFloat(meta, "key", 99.0)
		if result != 99.0 {
			t.Errorf("expected 99.0, got %v", result)
		}
	})

	t.Run("float64 value", func(t *testing.T) {
		meta := map[string]interface{}{"key": 3.14}
		result := GetMetadataFloat(meta, "key", 0.0)
		if result != 3.14 {
			t.Errorf("expected 3.14, got %v", result)
		}
	})

	t.Run("int value", func(t *testing.T) {
		meta := map[string]interface{}{"key": 42}
		result := GetMetadataFloat(meta, "key", 0.0)
		if result != 42.0 {
			t.Errorf("expected 42.0, got %v", result)
		}
	})

	t.Run("int64 value", func(t *testing.T) {
		meta := map[string]interface{}{"key": int64(100)}
		result := GetMetadataFloat(meta, "key", 0.0)
		if result != 100.0 {
			t.Errorf("expected 100.0, got %v", result)
		}
	})

	t.Run("wrong type returns default", func(t *testing.T) {
		meta := map[string]interface{}{"key": "not a number"}
		result := GetMetadataFloat(meta, "key", 99.0)
		if result != 99.0 {
			t.Errorf("expected 99.0, got %v", result)
		}
	})
}

func TestGetMetadataString(t *testing.T) {
	t.Run("nil map returns default", func(t *testing.T) {
		result := GetMetadataString(nil, "key", "default")
		if result != "default" {
			t.Errorf("expected 'default', got %q", result)
		}
	})

	t.Run("missing key returns default", func(t *testing.T) {
		meta := map[string]interface{}{"other": "value"}
		result := GetMetadataString(meta, "key", "default")
		if result != "default" {
			t.Errorf("expected 'default', got %q", result)
		}
	})

	t.Run("string value", func(t *testing.T) {
		meta := map[string]interface{}{"key": "hello"}
		result := GetMetadataString(meta, "key", "default")
		if result != "hello" {
			t.Errorf("expected 'hello', got %q", result)
		}
	})

	t.Run("wrong type returns default", func(t *testing.T) {
		meta := map[string]interface{}{"key": 123}
		result := GetMetadataString(meta, "key", "default")
		if result != "default" {
			t.Errorf("expected 'default', got %q", result)
		}
	})
}

func TestGetMetadataMap(t *testing.T) {
	t.Run("nil map returns nil", func(t *testing.T) {
		result := GetMetadataMap(nil, "key")
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("missing key returns nil", func(t *testing.T) {
		meta := map[string]interface{}{"other": "value"}
		result := GetMetadataMap(meta, "key")
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("valid map", func(t *testing.T) {
		nested := map[string]interface{}{"nested_key": "nested_value"}
		meta := map[string]interface{}{"key": nested}
		result := GetMetadataMap(meta, "key")
		if result == nil {
			t.Fatal("expected non-nil map")
		}
		if result["nested_key"] != "nested_value" {
			t.Errorf("expected nested_value, got %v", result["nested_key"])
		}
	})

	t.Run("wrong type returns nil", func(t *testing.T) {
		meta := map[string]interface{}{"key": "not a map"}
		result := GetMetadataMap(meta, "key")
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})
}
