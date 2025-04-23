package output_test

import (
	"context"
	"testing"
	"time"

	"github.com/kpiljoong/flox/internal/output"
)

func TestKafkaOutput_Send_SkipIfUnavailable(t *testing.T) {
	// This test is for structure only and will not assert delivery.
	// Useful for checking that KafkaOutput.Send() does not panic.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out := output.NewKafkaOutput(
		ctx,
		[]string{"localhost:9092"},
		"non-existent-topic",
		"flox-test-client",
	)

	event := map[string]interface{}{
		"msg":  "test message",
		"env":  "test",
		"type": "unit",
	}

	ch := make(chan error, 1)
	go func() {
		ch <- out.Send(event)
	}()

	select {
	case <-ctx.Done():
		t.Log("Send timeout exceeded - Kafka likely not running, as expected")
	case err := <-ch:
		if err != nil {
			t.Logf("Kafka send returned error as expected: %v", err)
		} else {
			t.Log("Kafka send returned success unexpectedly")
		}
	}
}

func TestKafkaOutput_Send_RespectsContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	out := output.NewKafkaOutput(
		ctx,
		[]string{"localhost:9092"},
		"non-existent-topic",
		"flox-test-client",
	)

	err := out.Send(map[string]interface{}{
		"msg": "should not send",
	})

	if err != context.Canceled {
		t.Fatalf("expected context.Canceled error, got:  %v", err)
	}
}
