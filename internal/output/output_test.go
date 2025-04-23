package output_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/kpiljoong/flox/internal/output"
)

func TestStdoutOutput_Send(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var buf bytes.Buffer
	out := output.NewStdoutOutputWithWriter(ctx, &buf)

	event := map[string]interface{}{
		"msg": "hello",
		"env": "test",
	}

	err := out.Send(event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("expected output to contain 'hello', got %s", buf.String())
	}
}
