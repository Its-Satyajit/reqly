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
	"fmt"
	"net/http"
)

// bearerScheme sends the configured token as an Authorization: Bearer header.
type bearerScheme struct{}

// Apply sets Authorization: Bearer <token>.
func (bearerScheme) Apply(req *http.Request, cfg map[string]string, vars Interpolator) error {
	token, err := vars.Interpolate(cfg["token"])
	if err != nil {
		return fmt.Errorf("bearer token: %w", err)
	}
	if token == "" {
		return fmt.Errorf("bearer auth requires a token")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

// SecretKeys reports the token as the sensitive config value.
func (bearerScheme) SecretKeys() []string { return []string{"token"} }

// basicScheme sends a username/password pair as HTTP Basic credentials.
type basicScheme struct{}

// Apply sets the request's basic auth from username and password.
func (basicScheme) Apply(req *http.Request, cfg map[string]string, vars Interpolator) error {
	username, err := vars.Interpolate(cfg["username"])
	if err != nil {
		return fmt.Errorf("basic username: %w", err)
	}
	if username == "" {
		return fmt.Errorf("basic auth requires a username")
	}
	password, err := vars.Interpolate(cfg["password"])
	if err != nil {
		return fmt.Errorf("basic password: %w", err)
	}
	if password == "" {
		return fmt.Errorf("basic auth requires a password")
	}
	req.SetBasicAuth(username, password)
	return nil
}

// SecretKeys reports the password as the sensitive config value.
func (basicScheme) SecretKeys() []string { return []string{"password"} }

// apikeyScheme sends an API key either as a header or a query parameter.
type apikeyScheme struct{}

// Apply sets the configured key (header or query, default header) to value.
func (apikeyScheme) Apply(req *http.Request, cfg map[string]string, vars Interpolator) error {
	key, err := vars.Interpolate(cfg["key"])
	if err != nil {
		return fmt.Errorf("apikey key: %w", err)
	}
	if key == "" {
		return fmt.Errorf("apikey auth requires a key")
	}
	value, err := vars.Interpolate(cfg["value"])
	if err != nil {
		return fmt.Errorf("apikey value: %w", err)
	}
	if value == "" {
		return fmt.Errorf("apikey auth requires a value")
	}
	switch in := cfg["in"]; in {
	case "", "header":
		req.Header.Set(key, value)
	case "query":
		q := req.URL.Query()
		q.Set(key, value)
		req.URL.RawQuery = q.Encode()
	default:
		return fmt.Errorf("apikey auth: unsupported in %q (want header or query)", in)
	}
	return nil
}

// SecretKeys reports the key value as the sensitive config value.
func (apikeyScheme) SecretKeys() []string { return []string{"value"} }

func init() {
	Register("bearer", bearerScheme{})
	Register("basic", basicScheme{})
	Register("apikey", apikeyScheme{})
}
