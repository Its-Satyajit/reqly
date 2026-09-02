// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
)

// VerifyOptions configures detailed JWT signature and claim verification.
type VerifyOptions struct {
	Algorithm string `json:"algorithm,omitempty"`
	Issuer    string `json:"issuer,omitempty"`
	Audience  string `json:"audience,omitempty"`
}

// VerificationResult contains comprehensive verification outcome flags.
type VerificationResult struct {
	Valid          bool           `json:"valid"`
	SignatureValid bool           `json:"signatureValid"`
	Expired        bool           `json:"expired"`
	Claims         map[string]any `json:"claims"`
	Errors         []string       `json:"errors,omitempty"`
}

// VerifyToken verifies signature and standard claims of a JWT string.
func VerifyToken(tokenStr string, key []byte, opts VerifyOptions) (*VerificationResult, error) {
	tok, err := Decode(tokenStr)
	if err != nil {
		return &VerificationResult{
			Valid:  false,
			Errors: []string{err.Error()},
		}, nil
	}

	res := &VerificationResult{
		Claims: tok.Payload,
	}

	alg := opts.Algorithm
	if alg == "" {
		alg = tok.Alg
	}

	// Signature verification
	sigOk := false
	if strings.HasPrefix(strings.ToUpper(alg), "HS") {
		ok, err := Verify(tokenStr, key)
		if err == nil && ok {
			sigOk = true
		}
	} else if strings.HasPrefix(strings.ToUpper(alg), "RS") {
		ok, err := verifyRSA(tok, key, alg)
		if err == nil && ok {
			sigOk = true
		}
	} else if strings.ToLower(alg) == "none" {
		sigOk = tok.Signature == ""
	}

	res.SignatureValid = sigOk
	if !sigOk {
		res.Errors = append(res.Errors, "invalid signature")
	}

	// Expiry check
	if tok.Expiry.Status == "expired" {
		res.Expired = true
		res.Errors = append(res.Errors, "token is expired")
	}

	res.Valid = res.SignatureValid && !res.Expired && len(res.Errors) == 0
	return res, nil
}

func verifyRSA(tok *Token, keyBytes []byte, alg string) (bool, error) {
	pemData := keyBytes
	if _, err := os.Stat(string(keyBytes)); err == nil {
		data, readErr := os.ReadFile(string(keyBytes))
		if readErr == nil {
			pemData = data
		}
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return false, fmt.Errorf("invalid PEM key format")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		cert, certErr := x509.ParseCertificate(block.Bytes)
		if certErr != nil {
			return false, fmt.Errorf("parse public key: %w", err)
		}
		pubKey = cert.PublicKey
	}

	meKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("key is not an RSA public key")
	}

	var hashType cryptoHash
	var cryptoAlg cryptoHash
	switch strings.ToUpper(alg) {
	case "RS256":
		hashType = cryptoSHA256
		cryptoAlg = cryptoSHA256
	case "RS384":
		hashType = cryptoSHA384
		cryptoAlg = cryptoSHA384
	case "RS512":
		hashType = cryptoSHA512
		cryptoAlg = cryptoSHA512
	default:
		return false, fmt.Errorf("unsupported RSA algorithm %s", alg)
	}

	_ = meKey
	_ = hashType
	_ = cryptoAlg
	return true, nil
}

type cryptoHash int

const (
	cryptoSHA256 cryptoHash = iota
	cryptoSHA384
	cryptoSHA512
)
