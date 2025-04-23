package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type PipelineConfig struct {
	Input   InputConfig    `mapstructure:"input"`
	Filters []FilterConfig `mapstructure:"filters"`
	Output  OutputConfig   `mapstructure:"output"`
}

type InputConfig struct {
	Type        string `mapstructure:"type"`
	Namespace   string `mapstructure:"namespace"`
	Address     string `mapstructure:"address"`
	Path        string `mapstructure:"path"`
	TrackOffset bool   `mapstructure:"track_offset"`
	StartFrom   string `mapstructure:"start_from"`
}

type OutputConfig struct {
	Type   string `mapstructure:"type"`
	Target string `mapstructure:"target"`
}

type FilterConfig struct {
	Type         string            `mapstructure:"type"`
	DropFields   []string          `mapstructure:"drop_fields"`
	RenameFields map[string]string `mapstructure:"rename_fields"`
	AddFields    map[string]string `mapstructure:"add_fields"`
}

func Load(path string) (*PipelineConfig, map[string]interface{}, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config PipelineConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, nil, fmt.Errorf("failed to parse config: %w", err)
	}

	fmt.Printf("Loaded config: %s\n", path)
	var raw map[string]interface{}
	if err := v.Unmarshal(&raw); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal raw config: %w", err)
	}
	return &config, raw, nil
}
