package output

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type LokiOutput struct {
	endpoint string
	labels   map[string]string
	client   *http.Client
	retries  int
	backoff  time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewLokiOutput(ctx context.Context, endpoint string, labels map[string]string) *LokiOutput {
	ctx, cancel := context.WithCancel(ctx)
	return &LokiOutput{
		endpoint: endpoint,
		labels:   labels,
		client:   &http.Client{Timeout: 5 * time.Second},
		retries:  3,
		backoff:  2 * time.Second,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (o *LokiOutput) Send(event map[string]interface{}) error {
	payload, err := o.preparePayload(event)
	if err != nil {
		log.Printf("failed to prepare payload: %v", err)
		return err
	}

	var lastErr error
	for attempt := 0; attempt < o.retries; attempt++ {
		select {
		case <-o.ctx.Done():
			return o.ctx.Err()
		default:
			err = o.sendOnce(payload)
			if err == nil {
				log.Printf("Successfully sent event to Loki: %s\n", string(payload))
				return nil
			}
			lastErr = err
			time.Sleep(o.backoff * time.Duration(attempt+1))
		}
	}
	return fmt.Errorf("all retries failed: %w", lastErr)
}

func (o *LokiOutput) sendOnce(payload []byte) error {
	req, err := http.NewRequestWithContext(o.ctx, http.MethodPost, o.endpoint+"/loki/api/v1/push", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("loki responded with status: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (o *LokiOutput) preparePayload(event map[string]interface{}) ([]byte, error) {
	stream := map[string]interface{}{
		"stream": o.getLabels(),
		"values": [][]string{
			{
				fmt.Sprintf("%d000000", time.Now().UnixMilli()), // nanoseconds required
				toJSONString(event),
			},
		},
	}

	payload := map[string]interface{}{
		"streams": []interface{}{stream},
	}

	return json.Marshal(payload)
}

func (o *LokiOutput) getLabels() map[string]string {
	labels := make(map[string]string)
	for k, v := range o.labels {
		labels[k] = v
	}
	if _, ok := labels["host"]; !ok {
		if hostname, err := os.Hostname(); err == nil {
			labels["host"] = hostname
		}
	}
	if _, ok := labels["env"]; !ok {
		labels["env"] = os.Getenv("FLOX_ENV")
	}
	return labels
}

func toJSONString(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
