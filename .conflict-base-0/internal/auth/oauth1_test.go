// SPDX-License-Identifier: Apache-2.0
package auth

import (
	"net/http"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/variables"
)

func TestOAuth1ApplySetsAuthorizationHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://api.example.com/resource?b=2&a=1", nil)
	cfg := map[string]string{
		"consumerKey":    "ck",
		"consumerSecret": "cs",
		"token":          "tk",
		"tokenSecret":    "ts",
	}
	if err := Apply(req, "oauth1", cfg, variables.NewSet()); err != nil {
		t.Fatalf("Apply oauth1: %v", err)
	}
	h := req.Header.Get("Authorization")
	if !strings.HasPrefix(h, "OAuth ") {
		t.Fatalf("Authorization header = %q, want OAuth prefix", h)
	}
	for _, want := range []string{`oauth_consumer_key="ck"`, `oauth_token="tk"`, `oauth_signature_method="HMAC-SHA1"`, `oauth_signature="`, `oauth_timestamp="`, `oauth_nonce="`} {
		if !strings.Contains(h, want) {
			t.Errorf("header missing %q: %q", want, h)
		}
	}
}

func TestOAuth1ApplyRequiresConsumerKey(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://api.example.com/", nil)
	if err := Apply(req, "oauth1", map[string]string{"consumerSecret": "cs"}, variables.NewSet()); err == nil {
		t.Fatal("expected error for missing consumerKey")
	}
}

func TestOAuth1Interpolation(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://api.example.com/", nil)
	vars := variables.NewSet()
	vars.Set(variables.ScopeGlobal, "ck", "mykey")
	cfg := map[string]string{
		"consumerKey":    "{{ck}}",
		"consumerSecret": "cs",
	}
	if err := Apply(req, "oauth1", cfg, vars); err != nil {
		t.Fatalf("interpolated Apply: %v", err)
	}
	if !strings.Contains(req.Header.Get("Authorization"), `oauth_consumer_key="mykey"`) {
		t.Error("interpolation not applied to consumerKey")
	}
}
