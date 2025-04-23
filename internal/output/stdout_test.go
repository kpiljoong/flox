package output_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/kpiljoong/flox/internal/output"
)

func TestStdoutOutput_Send_Success(t *testing.T) {
	ctx := context.Background()

	var buf bytes.Buffer
	out := output.NewStdoutOutputWithWriter(ctx, &buf)

	event := map[string]interface{}{
		"msg": "hello test",
	}

	err := out.Send(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("hello test")) {
		t.Error("expected output to contain 'hello test'")
	}
}

func TestStdoutOutput_RespectsContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf bytes.Buffer
	out := output.NewStdoutOutputWithWriter(ctx, &buf)

	event := map[string]interface{}{
		"msg": "test log",
	}

	err := out.Send(event)
	if err == nil {
		t.Fatalf("expected error due to cancelled context, got nil")
	}
	if ctx.Err() != err {
		t.Fatalf("expected context error, got: %v", err)
	}
	if buf.Len() > 0 {
		t.Error("expected no output to be written after context cancel")
	}
}
