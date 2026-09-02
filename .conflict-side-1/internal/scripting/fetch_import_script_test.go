package scripting

import (
	"testing"
)

func TestScripting_ImportFetch(t *testing.T) {
	var method, url string
	sb := NewSandbox(SandboxOptions{
		SetVariable: func(key, val string) {
			if key == "method" {
				method = val
			} else if key == "url" {
				url = val
			}
		},
	})

	script := `
		const snippet = 'fetch("https://api.example.com/products", { method: "POST" })';
		const parsed = reqly.importFetch(snippet);
		if (parsed) {
			reqly.setVariable("method", parsed.method);
			reqly.setVariable("url", parsed.url);
		}
	`

	if err := sb.Run(script); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if method != "POST" || url != "https://api.example.com/products" {
		t.Fatalf("want POST https://api.example.com/products, got %s %s", method, url)
	}
}
