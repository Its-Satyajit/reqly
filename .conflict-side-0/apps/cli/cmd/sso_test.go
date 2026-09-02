package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/jwt"
)

func TestSSOValidate(t *testing.T) {
	secret := []byte("test-secret-32-bytes-long-for-hs256")
	cfgIssuer := "https://auth.example.com"
	cfgClientID := "reqly"
	header := map[string]any{"alg": "HS256"}
	claims := map[string]any{"iss": cfgIssuer, "sub": "alice", "aud": cfgClientID}
	token, err := jwt.Sign(header, claims, secret)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"sso", "validate", "--issuer", cfgIssuer, "--client-id", cfgClientID, "--token", token, "--secret", string(secret)})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("sso validate: %v", err)
	}
	if !strings.Contains(buf.String(), "valid") {
		t.Fatalf("unexpected: %s", buf.String())
	}
}
