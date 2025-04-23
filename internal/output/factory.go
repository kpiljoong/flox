package output

import (
	"context"
	"fmt"

	"github.com/go-viper/mapstructure/v2"
)

type Output interface {
	Send(event map[string]interface{}) error
}

// NewOutput creates -the appropriate output based on type and config.
func NewOutput(ctx context.Context, outputType string, config map[string]interface{}) (Output, error) {
	switch outputType {
	case "stdout":
		return NewStdoutOutput(ctx), nil

	case "file":
		target, ok := config["target"].(string)
		if !ok || target == "" {
			return nil, fmt.Errorf("target not specified for file output")
		}
		return NewFileOutput(ctx, target)

	case "loki":
		var target string
		if t, ok := config["target"].(string); ok {
			target = t
		} else {
			return nil, fmt.Errorf("target not specified for loki output")
		}
		labels := map[string]string{
			"job":  "flox",
			"host": "local",
		}
		if raw, ok := config["labels"].(map[string]interface{}); ok {
			for k, v := range raw {
				if str, ok := v.(string); ok {
					labels[k] = str
				}
			}
		}
		return NewLokiOutput(ctx, target, labels), nil

	case "kafka":
		var kafkaCfg struct {
			Brokers  []string `json:"brokers"`
			Topic    string   `json:"topic"`
			ClientID string   `json:"client_id"`
		}
		if err := mapstructure.Decode(config, &kafkaCfg); err != nil {
			return nil, fmt.Errorf("invalid kafka output config: %w", err)
		}
		return NewKafkaOutput(ctx, kafkaCfg.Brokers, kafkaCfg.Topic, kafkaCfg.ClientID), nil
	default:
		return nil, fmt.Errorf("unsupported output type: %s", outputType)
	}
}
