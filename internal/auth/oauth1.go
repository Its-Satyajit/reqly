// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// newOAuth1Nonce generates a random nonce for OAuth 1.0. Overridable for tests.
var newOAuth1Nonce = func() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("oauth1: generating nonce: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// oauth1Now returns current Unix timestamp. Overridable for tests.
var oauth1Now = func() int64 { return time.Now().Unix() }

type oauth1Scheme struct{}

func (oauth1Scheme) Apply(req *http.Request, cfg map[string]string, vars Interpolator) error {
	consumerKey, err := vars.Interpolate(cfg["consumerKey"])
	if err != nil {
		return fmt.Errorf("oauth1 consumerKey: %w", err)
	}
	if consumerKey == "" {
		consumerKey, err = vars.Interpolate(cfg["consumer_key"])
		if err != nil {
			return fmt.Errorf("oauth1 consumerKey: %w", err)
		}
	}
	if consumerKey == "" {
		return fmt.Errorf("oauth1 auth requires consumerKey")
	}
	consumerSecret, err := vars.Interpolate(cfg["consumerSecret"])
	if err != nil {
		return fmt.Errorf("oauth1 consumerSecret: %w", err)
	}
	if consumerSecret == "" {
		consumerSecret, err = vars.Interpolate(cfg["consumer_secret"])
		if err != nil {
			return fmt.Errorf("oauth1 consumerSecret: %w", err)
		}
	}
	if consumerSecret == "" {
		return fmt.Errorf("oauth1 auth requires consumerSecret")
	}
	token, _ := vars.Interpolate(cfg["token"])
	tokenSecret, _ := vars.Interpolate(cfg["tokenSecret"])
	if tokenSecret == "" {
		tokenSecret, _ = vars.Interpolate(cfg["token_secret"])
	}
	sigMethod, _ := vars.Interpolate(cfg["signatureMethod"])
	if sigMethod == "" {
		sigMethod, _ = vars.Interpolate(cfg["signature_method"])
	}
	if sigMethod == "" {
		sigMethod = "HMAC-SHA1"
	}
	sigMethod = strings.ToUpper(sigMethod)
	if sigMethod != "HMAC-SHA1" {
		return fmt.Errorf("oauth1: unsupported signature method %q (only HMAC-SHA1)", sigMethod)
	}

	nonce, err := newOAuth1Nonce()
	if err != nil {
		return err
	}
	timestamp := fmt.Sprintf("%d", oauth1Now())

	// Collect OAuth and query params — RFC 5849 §3.4.1.3.1 requires each value
	// for a repeated key to appear as a separate pair, sorted by encoded value.
	type kv struct{ k, v string }
	var pairs []kv
	// OAuth params
	pairs = append(pairs,
		kv{percentEncode("oauth_consumer_key"), percentEncode(consumerKey)},
		kv{percentEncode("oauth_nonce"), percentEncode(nonce)},
		kv{percentEncode("oauth_signature_method"), percentEncode(sigMethod)},
		kv{percentEncode("oauth_timestamp"), percentEncode(timestamp)},
		kv{percentEncode("oauth_version"), percentEncode("1.0")},
	)
	if token != "" {
		pairs = append(pairs, kv{percentEncode("oauth_token"), percentEncode(token)})
	}
	// Query params — each value produces its own pair (multi-value correct)
	for k, vs := range req.URL.Query() {
		for _, v := range vs {
			pairs = append(pairs, kv{percentEncode(k), percentEncode(v)})
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].k == pairs[j].k {
			return pairs[i].v < pairs[j].v
		}
		return pairs[i].k < pairs[j].k
	})
	var normParams strings.Builder
	for i, p := range pairs {
		if i > 0 {
			normParams.WriteByte('&')
		}
		normParams.WriteString(p.k)
		normParams.WriteByte('=')
		normParams.WriteString(p.v)
	}

	// Normalized URL: scheme://host/path (no query) — RFC 5849 §3.4.1.2
	scheme := strings.ToLower(req.URL.Scheme)
	if scheme == "" {
		scheme = "https"
	}
	host := strings.ToLower(req.URL.Host)
	if host == "" {
		host = strings.ToLower(req.Host)
	}
	if host == "" {
		return fmt.Errorf("oauth1: request URL must have host")
	}
	// Strip default ports per RFC
	if strings.HasSuffix(host, ":443") && scheme == "https" {
		host = strings.TrimSuffix(host, ":443")
	} else if strings.HasSuffix(host, ":80") && scheme == "http" {
		host = strings.TrimSuffix(host, ":80")
	}
	path := req.URL.EscapedPath()
	if path == "" {
		path = "/"
	}
	normalizedURL := fmt.Sprintf("%s://%s%s", scheme, host, path)

	baseString := strings.ToUpper(req.Method) + "&" + percentEncode(normalizedURL) + "&" + percentEncode(normParams.String())
	signingKey := percentEncode(consumerSecret) + "&" + percentEncode(tokenSecret)
	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(baseString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Build Authorization header
	// Keep ordering stable for tests: consumer_key, nonce, signature, signature_method, timestamp, token, version
	var headerParams []string
	headerParams = append(headerParams, fmt.Sprintf(`oauth_consumer_key="%s"`, percentEncode(consumerKey)))
	headerParams = append(headerParams, fmt.Sprintf(`oauth_nonce="%s"`, percentEncode(nonce)))
	headerParams = append(headerParams, fmt.Sprintf(`oauth_signature="%s"`, percentEncode(signature)))
	headerParams = append(headerParams, fmt.Sprintf(`oauth_signature_method="%s"`, percentEncode(sigMethod)))
	headerParams = append(headerParams, fmt.Sprintf(`oauth_timestamp="%s"`, percentEncode(timestamp)))
	if token != "" {
		headerParams = append(headerParams, fmt.Sprintf(`oauth_token="%s"`, percentEncode(token)))
	}
	headerParams = append(headerParams, `oauth_version="1.0"`)

	req.Header.Set("Authorization", "OAuth "+strings.Join(headerParams, ", "))

	return nil
}

func (oauth1Scheme) SecretKeys() []string {
	return []string{"consumerSecret", "consumer_secret", "tokenSecret", "token_secret"}
}

func percentEncode(s string) string {
	// OAuth percent-encode per RFC 5849 §3.6
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

func init() { Register("oauth1", oauth1Scheme{}) }
