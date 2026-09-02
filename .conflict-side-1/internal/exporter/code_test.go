package exporter

import (
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/request"
)

func TestGenerateCurl(t *testing.T) {
	req := request.Request{Method: "GET", URL: "https://api.example.com/users", Headers: []request.Header{{Key: "Accept", Value: "application/json"}}}
	got, err := Generate(req, "curl", nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(got, "curl") || !strings.Contains(got, "https://api.example.com/users") {
		t.Fatalf("curl: got %q", got)
	}
	if !strings.Contains(got, "Accept") {
		t.Fatalf("curl header: got %q", got)
	}
}

func TestGenerateJS(t *testing.T) {
	req := request.Request{Method: "POST", URL: "https://api.example.com/users", Headers: []request.Header{{Key: "Content-Type", Value: "application/json"}}, Body: `{"name":"Alice"}`}
	got, err := Generate(req, "js", nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(got, "fetch") || !strings.Contains(got, "Alice") {
		t.Fatalf("js: got %q", got)
	}
}

func TestGeneratePython(t *testing.T) {
	req := request.Request{Method: "GET", URL: "https://api.example.com/users"}
	got, err := Generate(req, "python", nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(got, "requests") {
		t.Fatalf("python: got %q", got)
	}
}

func TestGenerateGo(t *testing.T) {
	req := request.Request{Method: "GET", URL: "https://api.example.com/users"}
	got, err := Generate(req, "go", nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(got, "http.NewRequest") {
		t.Fatalf("go: got %q", got)
	}
}

func TestGenerateMasked(t *testing.T) {
	req := request.Request{Method: "GET", URL: "https://api.example.com/users", Headers: []request.Header{{Key: "Authorization", Value: "Bearer secret123"}}}
	mask := func(s string) string { return strings.ReplaceAll(s, "secret123", "[SECRET]") }
	got, err := Generate(req, "curl", mask)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if strings.Contains(got, "secret123") || !strings.Contains(got, "[SECRET]") {
		t.Fatalf("masked: got %q", got)
	}
}

func TestGenerateUnknownLang(t *testing.T) {
	if _, err := Generate(request.Request{URL: "https://a.com"}, "ruby", nil); err == nil {
		t.Fatal("expected error for unknown lang")
	}
}
