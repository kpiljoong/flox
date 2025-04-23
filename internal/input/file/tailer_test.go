package file

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func createTempLogFile(t *testing.T, lines []string) string {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.log")

	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("failed to create temp log file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Logf("failed to close tmp file: %v", err)
		}
	}()

	for _, line := range lines {
		_, err := f.WriteString(line + "\n")
		if err != nil {
			t.Fatalf("failed to write to temp log file: %v", err)
		}
	}
	return tmpFile
}

func TestTailerReadsNewLines(t *testing.T) {
	tmpFile := createTempLogFile(t, []string{
		`{"msg":"hello world","level":"info"}`,
		`{"msg":"another log","level":"debug"}`,
	})

	tailer := &Tailer{
		files:       make(map[string]*os.File),
		trackOffset: false,
		startFrom:   "beginning",
		path:        tmpFile,
		ctx:         context.Background(),
		cancel:      func() {},
		lock:        sync.Mutex{},
	}

	// tailer := file.NewTailer(tmpFile, "default", false, "beginning")

	var events []map[string]interface{}
	handler := func(event map[string]interface{}) {
		events = append(events, event)
	}

	go tailer.openFile(tmpFile, handler)

	time.Sleep(2 * time.Second)

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if events[0]["msg"] != "hello world" {
		t.Errorf("first event mismatch: got %s", events[0]["msg"])
	}
	if events[1]["level"] != "debug" {
		t.Errorf("second event mismatch: got %s", events[1]["level"])
	}
}

func TestTailerIgnoresInvalidJSON(t *testing.T) {
	tmpFile := createTempLogFile(t, []string{
		`INVALID JSON LINE`,
		`{"msg":"valid log","level":"info"}`,
	})

	tailer := &Tailer{
		files:       make(map[string]*os.File),
		trackOffset: false,
		startFrom:   "beginning",
		path:        tmpFile,
		ctx:         context.Background(),
		cancel:      func() {},
		lock:        sync.Mutex{},
	}

	var events []map[string]interface{}
	handler := func(event map[string]interface{}) {
		events = append(events, event)
	}

	go tailer.openFile(tmpFile, handler)

	time.Sleep(2 * time.Second)

	if len(events) != 1 {
		t.Fatalf("expected 1 valid event, got %d", len(events))
	}

	if events[0]["msg"] != "valid log" {
		t.Errorf("unexpected event content: got %s", events[0]["msg"])
	}
}
