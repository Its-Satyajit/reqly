// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// awsNow returns the current time in AWS SigV4 format (YYYYMMDD'T'HHMMSS'Z').
// Overridable in tests for deterministic signatures.
var awsNow = func() string {
	return time.Now().UTC().Format("20060102T150405Z")
}

type awsScheme struct{}

func (awsScheme) SecretKeys() []string { return []string{"secretKey", "sessionToken"} }

func (awsScheme) Apply(req *http.Request, cfg map[string]string, vars Interpolator) error {
	accessKey, err := vars.Interpolate(cfg["accessKey"])
	if err != nil {
		return fmt.Errorf("aws accessKey: %w", err)
	}
	secretKey, err := vars.Interpolate(cfg["secretKey"])
	if err != nil {
		return fmt.Errorf("aws secretKey: %w", err)
	}
	region, err := vars.Interpolate(cfg["region"])
	if err != nil {
		return fmt.Errorf("aws region: %w", err)
	}
	service, err := vars.Interpolate(cfg["service"])
	if err != nil {
		return fmt.Errorf("aws service: %w", err)
	}
	sessionToken, err := vars.Interpolate(cfg["sessionToken"])
	if err != nil {
		return fmt.Errorf("aws sessionToken: %w", err)
	}
	if accessKey == "" {
		return fmt.Errorf("aws auth requires accessKey")
	}
	if secretKey == "" {
		return fmt.Errorf("aws auth requires secretKey")
	}
	if region == "" {
		return fmt.Errorf("aws auth requires region")
	}
	if service == "" {
		return fmt.Errorf("aws auth requires service")
	}

	amzDate := awsNow()
	dateStamp := amzDate[:8]

	payloadHash := hashPayload(req)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	if sessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", sessionToken)
	}

	// Canonical headers: host, x-amz-content-sha256, x-amz-date, + x-amz-security-token if present, sorted.
	host := req.URL.Host
	if host == "" {
		host = req.Host
	}
	if host == "" {
		host = req.Header.Get("Host")
	}
	if host == "" {
		host = strings.TrimPrefix(req.URL.String(), "https://")
		host = strings.TrimPrefix(host, "http://")
		if idx := strings.Index(host, "/"); idx != -1 {
			host = host[:idx]
		}
	}
	// Ensure host header is set for signing.
	if req.Header.Get("Host") == "" && host != "" {
		// Don't override if Host already set; signing uses the host value.
	}

	canonicalHeaders, signedHeaders := buildCanonicalHeaders(req, host, amzDate, payloadHash, sessionToken)
	canonicalURI := buildCanonicalURI(req.URL)
	canonicalQuery := buildCanonicalQuery(req.URL)
	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		canonicalQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	hashedCanonical := hashSHA256String(canonicalRequest)
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, region, service)
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		hashedCanonical,
	}, "\n")

	signingKey := deriveSigningKey(secretKey, dateStamp, region, service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKey, credentialScope, signedHeaders, signature)
	req.Header.Set("Authorization", authHeader)
	return nil
}

func hashSHA256String(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func hashPayload(req *http.Request) string {
	if req.Body == nil {
		return "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	}
	// Restore Body so transport can still read it.
	req.Body = io.NopCloser(bytes.NewReader(body))
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:])
}

func buildCanonicalURI(u *url.URL) string {
	path := u.EscapedPath()
	if path == "" {
		return "/"
	}
	return path
}

func buildCanonicalQuery(u *url.URL) string {
	if u.RawQuery == "" {
		return ""
	}
	values, _ := url.ParseQuery(u.RawQuery)
	var keys []string
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		vs := values[k]
		sort.Strings(vs)
		for _, v := range vs {
			parts = append(parts, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v)))
		}
	}
	// AWS requires space as %20 not +; QueryEscape uses %20 already for QueryEscape, but ParseQuery decodes +.
	return strings.Join(parts, "&")
}

func buildCanonicalHeaders(req *http.Request, host, amzDate, payloadHash, sessionToken string) (string, string) {
	headers := map[string]string{
		"host":                 host,
		"x-amz-content-sha256": payloadHash,
		"x-amz-date":           amzDate,
	}
	if sessionToken != "" {
		headers["x-amz-security-token"] = sessionToken
	}
	// Include any existing x-amz-* headers already on request (lowercase).
	for k, v := range req.Header {
		lk := strings.ToLower(k)
		if strings.HasPrefix(lk, "x-amz-") {
			if _, exists := headers[lk]; !exists {
				headers[lk] = strings.TrimSpace(v[0])
			}
		}
	}
	var keys []string
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var canonical string
	for _, k := range keys {
		canonical += fmt.Sprintf("%s:%s\n", k, headers[k])
	}
	return canonical, strings.Join(keys, ";")
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func deriveSigningKey(secret, dateStamp, region, service string) []byte {
	kSecret := []byte("AWS4" + secret)
	kDate := hmacSHA256(kSecret, []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	return kSigning
}

func init() {
	Register("aws", awsScheme{})
}
