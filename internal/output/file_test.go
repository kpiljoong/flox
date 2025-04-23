package output_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kpiljoong/flox/internal/output"
)

func TestFileOutput_Send(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "flox_test_output_*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			t.Logf("failed to close tmp file: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out, err := output.NewFileOutput(ctx, tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to create FileOutput: %v", err)
	}

	event := map[string]interface{}{
		"level": "info",
		"msg":   "hello from test",
	}

	if err := out.Send(event); err != nil {
		t.Fatalf("failed to read written log: %v", err)
	}

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read written log: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("written log is not valid JSON: %v", err)
	}

	if !strings.Contains(string(data), "hello from test") {
		t.Errorf("log file doesn't contain expected message: %s", string(data))
	}
}

func TestFileOutput_RespectsContextCancellation(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "flox_test_output_shutdown_*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			t.Logf("failed to close tmp file: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	out, err := output.NewFileOutput(ctx, tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to create FileOutput: %v", err)
	}
	cancel()

	event := map[string]interface{}{
		"msg": "this should not be written",
	}

	err = out.Send(event)
	if err == nil {  
		t.Error("expected error when sending after context cancellation, got nil")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled error, got %v", err)
	}

	// Ensure file is still empty
	data, _ := os.ReadFile(tmpFile.Name())
	if len(data) != 0 {
		t.Errorf("expected no data to be written after shutdown, got %s", string(data))
	}
}
