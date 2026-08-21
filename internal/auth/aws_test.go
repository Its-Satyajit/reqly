// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/variables"
)

func TestAWSApplyMissingKeys(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	s := awsScheme{}
	if err := s.Apply(req, nil, variables.NewSet()); err == nil {
		t.Fatal("expected error when aws keys missing")
	}
	if err := s.Apply(req, map[string]string{"accessKey": "AKIA"}, variables.NewSet()); err == nil {
		t.Fatal("expected error when secretKey missing")
	} else if !strings.Contains(err.Error(), "secretKey") {
		t.Fatalf("expected error to mention secretKey, got %v", err)
	}
	if err := s.Apply(req, map[string]string{"accessKey": "AKIA", "secretKey": "sec", "region": "us-east-1"}, variables.NewSet()); err == nil {
		t.Fatal("expected error when service missing")
	}
}

func TestAWSApplyInterpolated(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/test", nil)
	vars := variables.NewSet()
	vars.Set(variables.ScopeRequest, "ak", "AKIAEXAMPLE")
	vars.Set(variables.ScopeRequest, "sk", "mysecret")
	vars.Set(variables.ScopeRequest, "rg", "us-west-2")
	vars.Set(variables.ScopeRequest, "svc", "execute-api")
	err := awsScheme{}.Apply(req, map[string]string{
		"accessKey": "{{ak}}",
		"secretKey": "{{sk}}",
		"region":    "{{rg}}",
		"service":   "{{svc}}",
	}, vars)
	if err != nil {
		t.Fatal(err)
	}
	if got := req.Header.Get("Authorization"); !strings.HasPrefix(got, "AWS4-HMAC-SHA256 ") {
		t.Fatalf("expected AWS4-HMAC-SHA256 header, got %q", got)
	}
	if got := req.Header.Get("X-Amz-Date"); got == "" {
		t.Fatal("expected X-Amz-Date header")
	}
}

func TestAWSApplyWithSessionToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	err := awsScheme{}.Apply(req, map[string]string{
		"accessKey":    "AKIA",
		"secretKey":    "sec",
		"region":       "us-east-1",
		"service":      "s3",
		"sessionToken": "tok123",
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	if got := req.Header.Get("X-Amz-Security-Token"); got != "tok123" {
		t.Fatalf("expected session token header, got %q", got)
	}
	if got := req.Header.Get("Authorization"); !strings.Contains(got, "Credential=AKIA/") {
		t.Fatalf("expected Credential with accessKey, got %q", got)
	}
}

func TestAWSApplyKnownVector(t *testing.T) {
	// AWS SigV4 test vector: IAM ListUsers example from AWS docs.
	// Using fixed time so signature is deterministic.
	oldNow := awsNow
	awsNow = func() string { return "20150830T123600Z" }
	defer func() { awsNow = oldNow }()

	req := httptest.NewRequest(http.MethodGet, "https://iam.amazonaws.com/?Action=ListUsers&Version=2010-05-08", nil)
	req.Header.Set("Host", "iam.amazonaws.com")
	err := awsScheme{}.Apply(req, map[string]string{
		"accessKey": "AKIAIOSFODNN7EXAMPLE",
		"secretKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"region":    "us-east-1",
		"service":   "iam",
	}, variables.NewSet())
	if err != nil {
		t.Fatal(err)
	}
	auth := req.Header.Get("Authorization")
	// Must contain the expected credential scope and signed headers.
	if !strings.Contains(auth, "Credential=AKIAIOSFODNN7EXAMPLE/20150830/us-east-1/iam/aws4_request") {
		t.Fatalf("credential scope missing in %q", auth)
	}
	if !strings.Contains(auth, "SignedHeaders=host;x-amz-content-sha256;x-amz-date") {
		t.Fatalf("signed headers missing in %q", auth)
	}
	// Signature must be non-empty hex (64 chars).
	parts := strings.Split(auth, "Signature=")
	if len(parts) != 2 || len(parts[1]) != 64 {
		t.Fatalf("expected 64-char hex signature, got %q", auth)
	}
}

func TestAWSSecretKeys(t *testing.T) {
	s := awsScheme{}
	keys := s.SecretKeys()
	if len(keys) != 2 || keys[0] != "secretKey" || keys[1] != "sessionToken" {
		t.Fatalf("SecretKeys: got %v", keys)
	}
	if got := MaskValues("aws", map[string]string{"secretKey": "s3", "sessionToken": "tok"}, variables.NewSet()); len(got) != 2 {
		t.Fatalf("MaskValues aws: got %v", got)
	}
}
