package ai

import (
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/response"
)

func TestExplainResponse(t *testing.T) {
	resp := &response.Response{StatusCode: 200, StatusText: "OK", Duration: 42 * time.Millisecond, Size: 123, Proto: "HTTP/1.1"}
	got := ExplainResponse(resp)
	if got == "" {
		t.Fatalf("want explanation")
	}
	if got == "no response" {
		t.Fatalf("want not no response")
	}
	if ExplainResponse(nil) != "no response" {
		t.Fatalf("want no response")
	}
}
