package output_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kpiljoong/flox/internal/output"
)

func TestLokiOutput_Send_Success(t *testing.T) {
	var received bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ctx := context.Background()
	out := output.NewLokiOutput(ctx, server.URL, map[string]string{"job": "test"})

	err := out.Send(map[string]interface{}{"msg": "test log"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !received {
		t.Error("expected Loki to receive request")
	}
}

func TestLokiOutput_Send_RespectsContextCancel(t *testing.T) {
	var received bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Cancel the context before Send
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	out := output.NewLokiOutput(ctx, server.URL, map[string]string{"job": "test"})
	err := out.Send(map[string]interface{}{"msg": "should not send"})

	if err == nil {
		t.Fatal("expected error due to context cancel, got nil")
	}
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled error, got: %v", err)
	}
	if received {
		t.Error("expected no request to be sent to Loki after cancellation")
	}
}
