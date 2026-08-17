// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

// basicScheme sends a username/password pair as HTTP Basic credentials.
type basicScheme struct{}

// Apply sets the request's basic auth from username and password.
func (basicScheme) Apply(req *http.Request, cfg map[string]string, vars Interpolator) error {
	username, err := vars.Interpolate(cfg["username"])
	if err != nil {
		return fmt.Errorf("basic username: %w", err)
	}
	password, err := vars.Interpolate(cfg["password"])
	if err != nil {
		return fmt.Errorf("basic password: %w", err)
	}
	req.SetBasicAuth(username, password)
	return nil
}

// apikeyScheme sends an API key either as a header or a query parameter.
type apikeyScheme struct{}

// Apply sets the configured key (header or query, default header) to value.
func (apikeyScheme) Apply(req *http.Request, cfg map[string]string, vars Interpolator) error {
	key, err := vars.Interpolate(cfg["key"])
	if err != nil {
		return fmt.Errorf("apikey key: %w", err)
	}
	value, err := vars.Interpolate(cfg["value"])
	if err != nil {
		return fmt.Errorf("apikey value: %w", err)
	}
	if cfg["in"] == "query" {
		q := req.URL.Query()
		q.Set(key, value)
		req.URL.RawQuery = q.Encode()
		return nil
	}
	req.Header.Set(key, value)
	return nil
}

func init() {
	Register("bearer", bearerScheme{})
	Register("basic", basicScheme{})
	Register("apikey", apikeyScheme{})
}
