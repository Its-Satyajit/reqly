package sso

import (
	"testing"

	"github.com/Its-Satyajit/reqly/internal/jwt"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{name: "valid", cfg: Config{Issuer: "https://auth.example.com", ClientID: "reqly"}, wantErr: false},
		{name: "missing issuer", cfg: Config{ClientID: "reqly"}, wantErr: true},
		{name: "missing clientId", cfg: Config{Issuer: "https://auth.example.com"}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.cfg)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	cfg := Config{Issuer: "https://auth.example.com", ClientID: "reqly"}
	secret := []byte("test-secret-32-bytes-long-for-hs256")
	header := map[string]any{"alg": "HS256"}
	claims := map[string]any{"iss": cfg.Issuer, "sub": "alice", "aud": cfg.ClientID}
	token, err := jwt.Sign(header, claims, secret)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if err := ValidateToken(cfg, token, secret); err != nil {
		t.Fatalf("ValidateToken valid: %v", err)
	}
	badClaims := map[string]any{"iss": "https://other.example.com", "sub": "alice"}
	badToken, _ := jwt.Sign(header, badClaims, secret)
	if err := ValidateToken(cfg, badToken, secret); err == nil {
		t.Fatalf("expected issuer mismatch")
	}
	if err := ValidateToken(cfg, token, []byte("wrong-secret")); err == nil {
		t.Fatalf("expected invalid signature")
	}
}

func TestIsGroupAllowed(t *testing.T) {
	cfg := Config{Issuer: "https://a", ClientID: "c", AllowedGroups: []string{"eng", "admin"}}
	if !IsGroupAllowed(cfg, []string{"eng"}) {
		t.Fatalf("should allow eng")
	}
	if IsGroupAllowed(cfg, []string{"other"}) {
		t.Fatalf("should deny other")
	}
	// Empty allowed means allow all
	cfg2 := Config{Issuer: "https://a", ClientID: "c"}
	if !IsGroupAllowed(cfg2, []string{"any"}) {
		t.Fatalf("empty allowed should allow all")
	}
}
