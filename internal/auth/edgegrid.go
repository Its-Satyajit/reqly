// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var edgeGridNow = func() string {
	return time.Now().UTC().Format("20060102T15:04:05+0000")
}

var edgeGridNonce = func() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type edgeGridScheme struct{}

func (edgeGridScheme) SecretKeys() []string { return []string{"clientSecret", "accessToken"} }

func (edgeGridScheme) Apply(req *http.Request, cfg map[string]string, vars Interpolator) error {
	clientToken, err := vars.Interpolate(cfg["clientToken"])
	if err != nil {
		return fmt.Errorf("edgegrid clientToken: %w", err)
	}
	clientSecret, err := vars.Interpolate(cfg["clientSecret"])
	if err != nil {
		return fmt.Errorf("edgegrid clientSecret: %w", err)
	}
	accessToken, err := vars.Interpolate(cfg["accessToken"])
	if err != nil {
		return fmt.Errorf("edgegrid accessToken: %w", err)
	}
	host, err := vars.Interpolate(cfg["host"])
	if err != nil {
		return fmt.Errorf("edgegrid host: %w", err)
	}
	if clientToken == "" {
		return fmt.Errorf("edgegrid auth requires clientToken")
	}
	if clientSecret == "" {
		return fmt.Errorf("edgegrid auth requires clientSecret")
	}
	if accessToken == "" {
		return fmt.Errorf("edgegrid auth requires accessToken")
	}
	if host == "" {
		return fmt.Errorf("edgegrid auth requires host")
	}

	timestamp := edgeGridNow()
	nonce := edgeGridNonce()

	// Signing key: base64(HMAC-SHA256(timestamp, clientSecret))
	mac := hmac.New(sha256.New, []byte(clientSecret))
	mac.Write([]byte(timestamp))
	signingKey := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Auth data: method + host + path+query
	path := req.URL.EscapedPath()
	if path == "" {
		path = "/"
	}
	if req.URL.RawQuery != "" {
		path += "?" + req.URL.RawQuery
	}
	// Simplified EdgeGrid signature base: timestamp + method + host + path
	data := strings.Join([]string{timestamp, req.Method, host, path}, "\t")
	mac2 := hmac.New(sha256.New, []byte(signingKey))
	mac2.Write([]byte(data))
	signature := base64.StdEncoding.EncodeToString(mac2.Sum(nil))

	authHeader := fmt.Sprintf("EG1-HMAC-SHA256 client_token=%s;access_token=%s;timestamp=%s;nonce=%s;signature=%s",
		clientToken, accessToken, timestamp, nonce, signature)
	req.Header.Set("Authorization", authHeader)
	// EdgeGrid also expects Host header to match config host.
	if req.Header.Get("Host") == "" {
		req.Header.Set("Host", host)
	}
	return nil
}

func init() {
	Register("edgegrid", edgeGridScheme{})
}
