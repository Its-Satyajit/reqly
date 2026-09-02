package importer

import (
	"testing"
)

func TestParseFetch_SimpleGET(t *testing.T) {
	code := `fetch("https://api.example.com/items");`
	req, err := ParseFetch(code)
	if err != nil {
		t.Fatalf("ParseFetch error: %v", err)
	}
	if string(req.Method) != "GET" {
		t.Errorf("want GET, got %s", req.Method)
	}
	if req.URL != "https://api.example.com/items" {
		t.Errorf("want https://api.example.com/items, got %s", req.URL)
	}
}

func TestParseFetch_POSTWithHeadersAndBody(t *testing.T) {
	code := `
		fetch("https://api.example.com/v1/users", {
			"headers": {
				"accept": "application/json, text/plain, */*",
				"authorization": "Bearer secret-token-123",
				"content-type": "application/json",
				"sec-fetch-mode": "cors"
			},
			"body": "{\"name\":\"Bob\",\"role\":\"admin\"}",
			"method": "POST"
		});
	`
	req, err := ParseFetch(code)
	if err != nil {
		t.Fatalf("ParseFetch error: %v", err)
	}

	if string(req.Method) != "POST" {
		t.Errorf("want POST, got %s", req.Method)
	}
	if req.URL != "https://api.example.com/v1/users" {
		t.Errorf("want https://api.example.com/v1/users, got %s", req.URL)
	}
	if req.Body != `{"name":"Bob","role":"admin"}` {
		t.Errorf("want body %s, got %s", `{"name":"Bob","role":"admin"}`, req.Body)
	}

	hasAuth := false
	for _, h := range req.Headers {
		if h.Key == "authorization" || h.Key == "Authorization" {
			hasAuth = true
			if h.Value != "Bearer secret-token-123" {
				t.Errorf("unexpected auth header value: %s", h.Value)
			}
		}
	}
	if !hasAuth {
		t.Errorf("missing authorization header")
	}
}
