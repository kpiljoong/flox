package filters_test

import (
	"testing"

	"github.com/kpiljoong/flox/internal/filters"
)

func TestJSONFilter_Process(t *testing.T) {
	filter := filters.NewJSONFilter(
		[]string{"secret", "password"},
		map[string]string{"msg": "message"},
		map[string]string{"env": "test"},
	)

	event := map[string]interface{}{
		"msg":    "hello",
		"secret": "supersecret",
		"user": "dev",
		"password": "1234",
	}

	processed := filter.Process(event)

	if _, exists := processed["secret"]; exists {
		t.Errorf("expected 'secret' to be dropped")
	}
	if _, exists := processed["password"]; exists {
		t.Errorf("expected 'password' to be dropped")
	}
	
	if processed["message"] != "hello" {
		t.Errorf("expected 'message' to be 'hello', got %v", processed["message"])
	}
	if processed["env"] != "test" {
		t.Errorf("expected 'env' to be 'test', got '%v'", processed["env"])
	}
	if processed["user"] != "dev" {
		t.Error("expected 'user' to remain unchanged")
	}
}
