package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for opencode-reflector.
type Config struct {
	Port      int            `yaml:"port"`
	Trigger   TriggerConfig  `yaml:"trigger"`
	Model     ModelConfig    `yaml:"model"`
	Sentiment SentimentConfig `yaml:"sentiment"`
	Retention RetentionConfig `yaml:"retention"`
	Report    ReportConfig   `yaml:"report"`
	LogLevel  string         `yaml:"logLevel"`
}

type TriggerConfig struct {
	Time   TriggerTimeConfig   `yaml:"time"`
	Events TriggerEventsConfig `yaml:"events"`
}

type TriggerTimeConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Schedule string `yaml:"schedule"`
}

type TriggerEventsConfig struct {
	Enabled         bool     `yaml:"enabled"`
	Types           []string `yaml:"types"`
	MessageInterval int      `yaml:"messageInterval"`
}

type ModelConfig struct {
	ID string `yaml:"id"`
}

type SentimentConfig struct {
	Enabled             bool `yaml:"enabled"`
	SkipOnMessageTrigger bool `yaml:"skipOnMessageTrigger"`
}

type RetentionConfig struct {
	Days int `yaml:"days"`
}

type ReportConfig struct {
	Template string `yaml:"template"`
}

// DefaultConfig returns the built-in default configuration.
func DefaultConfig() *Config {
	return &Config{
		Port: 19870,
		Trigger: TriggerConfig{
			Time: TriggerTimeConfig{
				Enabled:  true,
				Schedule: "00:00",
			},
			Events: TriggerEventsConfig{
				Enabled:         true,
				Types:           []string{"TASK_FINISHED", "N_MESSAGES"},
				MessageInterval: 10,
			},
		},
		Model: ModelConfig{
			ID: "minimax-cn-coding-plan/MiniMax-M2.7-highspeed",
		},
		Sentiment: SentimentConfig{
			Enabled:             true,
			SkipOnMessageTrigger: true,
		},
		Retention: RetentionConfig{
			Days: 90,
		},
		Report: ReportConfig{
			Template: "default",
		},
		LogLevel: "info",
	}
}

// Load reads configuration from a YAML file.
// If the file doesn't exist, returns DefaultConfig.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}
	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
