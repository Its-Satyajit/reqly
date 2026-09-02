package history

import (
	"testing"
	"time"
)

func TestCookieMatching(t *testing.T) {
	cookies := []Cookie{
		{Name: "sess", Value: "abc", Domain: "example.com", Path: "/", Env: "dev", ExpiresAt: time.Now().Add(time.Hour)},
		{Name: "x", Value: "1", Domain: "other.com", Path: "/", Env: "dev", ExpiresAt: time.Now().Add(time.Hour)},
		{Name: "expired", Value: "z", Domain: "example.com", Path: "/", Env: "dev", ExpiresAt: time.Now().Add(-time.Hour)},
		{Name: "secure", Value: "s", Domain: "example.com", Path: "/", Secure: true, Env: "dev", ExpiresAt: time.Now().Add(time.Hour)},
	}
	got := FilterCookies(cookies, "https://api.example.com/users", true)
	if len(got) != 2 { // sess + secure (https)
		t.Fatalf("https match: got %d %v", len(got), got)
	}
	got = FilterCookies(cookies, "http://api.example.com/users", false)
	if len(got) != 1 { // sess only, secure excluded on http
		t.Fatalf("http match: got %d %v", len(got), got)
	}
	// path test
	cookies2 := []Cookie{{Name: "p", Value: "1", Domain: "example.com", Path: "/api", Env: "dev", ExpiresAt: time.Now().Add(time.Hour)}}
	got = FilterCookies(cookies2, "https://example.com/other", true)
	if len(got) != 0 {
		t.Fatalf("path mismatch: got %d", len(got))
	}
	got = FilterCookies(cookies2, "https://example.com/api/users", true)
	if len(got) != 1 {
		t.Fatalf("path match: got %d", len(got))
	}
}
