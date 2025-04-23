package config_test

import (
	"testing"

	"github.com/spf13/viper"

	"github.com/kpiljoong/flox/internal/config"
)

func TestLoadValidConfig(t *testing.T) {
	// Define a valid config file path
	configFilePath := "testdata/valid_config.yaml"
	v := viper.New()
	v.SetConfigFile(configFilePath)

	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	// Load the configuration
	var cfg config.PipelineConfig
	if err := v.Unmarshal(&cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Input.Type == "" {
		t.Error("input.type should not be empty")
	}
}
