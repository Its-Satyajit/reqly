package ai

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/response"
)

func TestExplainResponse(t *testing.T) {
	resp := &response.Response{StatusCode: 200, StatusText: "OK", Duration: 42 * time.Millisecond, Size: 123, Proto: "HTTP/1.1"}
	got := ExplainResponse(resp)
	if got == "" || got == "no response" {
		t.Fatalf("unexpected explanation: %q", got)
	}
	if ExplainResponse(nil) != "no response" {
		t.Fatalf("want no response")
	}
}

func TestGenerateTests(t *testing.T) {
	headers := map[string][]string{
		"Content-Type": {"application/json"},
	}
	resp := &response.Response{
		StatusCode: 200,
		StatusText: "OK",
		Duration:   85 * time.Millisecond,
		Headers:    headers,
		Body:       []byte(`{"id": 123, "name": "Reqly", "active": true}`),
	}

	tests := GenerateTests(resp)
	if !strings.Contains(tests, "Status code is 200") {
		t.Errorf("missing status code test: %s", tests)
	}
	if !strings.Contains(tests, "Content-Type is JSON") {
		t.Errorf("missing content-type test: %s", tests)
	}
	if !strings.Contains(tests, `"id"`) || !strings.Contains(tests, `"name"`) {
		t.Errorf("missing JSON properties tests: %s", tests)
	}
}

func TestGenerateDocs(t *testing.T) {
	req := &request.Request{
		Method: "POST",
		URL:    "https://api.example.com/v1/users",
		Headers: []request.Header{
			{Key: "Content-Type", Value: "application/json"},
		},
		Body: `{"name": "Alice"}`,
	}
	resp := &response.Response{
		StatusCode: 201,
		StatusText: "Created",
		Duration:   120 * time.Millisecond,
		Body:       []byte(`{"id": 1, "name": "Alice"}`),
	}

	docs := GenerateDocs(req, resp)
	if !strings.Contains(docs, "POST") || !strings.Contains(docs, "https://api.example.com/v1/users") {
		t.Errorf("missing endpoint header in docs: %s", docs)
	}
	if !strings.Contains(docs, "Alice") {
		t.Errorf("missing request/response bodies in docs: %s", docs)
	}
	if !strings.Contains(docs, "201 Created") {
		t.Errorf("missing status code in docs: %s", docs)
	}
}

func TestDiagnose(t *testing.T) {
	diag401 := Diagnose(&response.Response{StatusCode: 401}, nil)
	if !strings.Contains(diag401, "Unauthorized") || !strings.Contains(diag401, "Authorization") {
		t.Errorf("unexpected 401 diagnosis: %s", diag401)
	}

	diag429 := Diagnose(&response.Response{StatusCode: 429}, nil)
	if !strings.Contains(diag429, "Rate Limited") {
		t.Errorf("unexpected 429 diagnosis: %s", diag429)
	}

	diag500 := Diagnose(&response.Response{StatusCode: 500}, nil)
	if !strings.Contains(diag500, "Internal Server Error") {
		t.Errorf("unexpected 500 diagnosis: %s", diag500)
	}

	diagNet := Diagnose(nil, errors.New("connection refused"))
	if !strings.Contains(diagNet, "Connection Refused") {
		t.Errorf("unexpected network error diagnosis: %s", diagNet)
	}
}
