package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfigValues(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Port != 19870 {
		t.Errorf("default port should be 19870, got %d", cfg.Port)
	}
	if cfg.Trigger.Time.Schedule != "00:00" {
		t.Errorf("default schedule should be 00:00, got %s", cfg.Trigger.Time.Schedule)
	}
	if cfg.Retention.Days != 90 {
		t.Errorf("default retention days should be 90, got %d", cfg.Retention.Days)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("default log level should be info, got %s", cfg.LogLevel)
	}
	if !cfg.Sentiment.Enabled {
		t.Error("sentiment should be enabled by default")
	}
}

func TestLoadNoFile(t *testing.T) {
	cfg, err := Load("/nonexistent/reflector.yaml")
	if err != nil {
		t.Fatalf("Load with nonexistent file should not error: %v", err)
	}
	if cfg.Port != 19870 {
		t.Errorf("should return default config, got port %d", cfg.Port)
	}
}

func TestLoadFromYAML(t *testing.T) {
	content := `
port: 9999
trigger:
  time:
    enabled: false
    schedule: "08:30"
model:
  id: "custom-model"
logLevel: "debug"
`
	tmp := t.TempDir()
	path := filepath.Join(tmp, "reflector.yaml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Port != 9999 {
		t.Errorf("port should be 9999, got %d", cfg.Port)
	}
	if cfg.Trigger.Time.Enabled {
		t.Error("trigger.time.enabled should be false")
	}
	if cfg.Trigger.Time.Schedule != "08:30" {
		t.Errorf("schedule should be 08:30, got %s", cfg.Trigger.Time.Schedule)
	}
	if cfg.Model.ID != "custom-model" {
		t.Errorf("model id should be custom-model, got %s", cfg.Model.ID)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("log level should be debug, got %s", cfg.LogLevel)
	}
}

func TestLoadPartialOverride(t *testing.T) {
	content := `
port: 7777
`
	tmp := t.TempDir()
	path := filepath.Join(tmp, "reflector.yaml")
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Port != 7777 {
		t.Errorf("port should be overridden to 7777, got %d", cfg.Port)
	}
	// Other fields should retain defaults
	if cfg.Retention.Days != 90 {
		t.Errorf("retention should keep default 90, got %d", cfg.Retention.Days)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("log level should keep default info, got %s", cfg.LogLevel)
	}
}

func TestDefaultConfigYAMLRoundTrip(t *testing.T) {
	cfg := DefaultConfig()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var cfg2 Config
	if err := yaml.Unmarshal(data, &cfg2); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if cfg.Port != cfg2.Port {
		t.Errorf("port mismatch after round-trip: %d != %d", cfg.Port, cfg2.Port)
	}
	if cfg.LogLevel != cfg2.LogLevel {
		t.Errorf("logLevel mismatch after round-trip: %s != %s", cfg.LogLevel, cfg2.LogLevel)
	}
}
