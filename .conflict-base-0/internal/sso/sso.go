// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package sso

import (
	"fmt"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/jwt"
)

// Config is a local SSO/OIDC config (e.g. .reqly/sso.yaml, 0600).
// It is Git-native when committed, but the file itself is 0600 for secrets.
// SSO is local-only, zero telemetry.
type Config struct {
	Issuer        string   `json:"issuer" yaml:"issuer"`
	ClientID      string   `json:"clientId" yaml:"clientId"`
	JWKSURL       string   `json:"jwksUrl,omitempty" yaml:"jwksUrl,omitempty"`
	AllowedGroups []string `json:"allowedGroups,omitempty" yaml:"allowedGroups,omitempty"`
}

// Validate checks a config for required fields.
func Validate(c Config) error {
	if strings.TrimSpace(c.Issuer) == "" {
		return fmt.Errorf("issuer is required")
	}
	if strings.TrimSpace(c.ClientID) == "" {
		return fmt.Errorf("clientId is required")
	}
	return nil
}

// ValidateToken validates an OIDC ID token using the configured issuer's key.
// For M73, it uses HMAC verification via jwt.Verify (RS256 via JWKS deferred).
// The token's iss claim must match Config.Issuer.
func ValidateToken(c Config, token string, key []byte) error {
	if err := Validate(c); err != nil {
		return err
	}
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("token is required")
	}
	tok, err := jwt.Decode(token)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	if iss, _ := tok.Payload["iss"].(string); iss != c.Issuer {
		return fmt.Errorf("issuer mismatch: got %q, want %q", iss, c.Issuer)
	}
	ok, err := jwt.Verify(token, key)
	if err != nil {
		return fmt.Errorf("verify: %w", err)
	}
	if !ok {
		return fmt.Errorf("invalid token")
	}
	return nil
}

// IsGroupAllowed checks if a user's groups intersect AllowedGroups.
// Empty AllowedGroups means allow all.
func IsGroupAllowed(c Config, userGroups []string) bool {
	if len(c.AllowedGroups) == 0 {
		return true
	}
	allowed := make(map[string]bool, len(c.AllowedGroups))
	for _, g := range c.AllowedGroups {
		allowed[g] = true
	}
	for _, g := range userGroups {
		if allowed[g] {
			return true
		}
	}
	return false
}
