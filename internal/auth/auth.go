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

// Package auth provides Git-native authentication schemes. Each scheme knows
// its own config keys and applies itself to an outgoing request; a registry
// dispatches by the request's auth.type string.
package auth

import (
	"fmt"
	"net/http"
	"sync"
)

// Interpolator resolves {{key}} placeholders in config values. It is the
// variables.Set interface, kept small so auth does not depend on the full
// variables package internals.
type Interpolator interface {
	Interpolate(input string) (string, error)
}

// Scheme applies an authentication configuration to an outgoing request.
type Scheme interface {
	// Apply mutates req based on cfg, interpolating values against vars.
	Apply(req *http.Request, cfg map[string]string, vars Interpolator) error
}

// SecretKeyScheme is implemented by schemes whose config holds credential
// values that must never surface in output. SecretKeys returns the config
// keys whose values are secrets.
type SecretKeyScheme interface {
	SecretKeys() []string
}

// MaskValues returns the resolved secret config values for the auth type,
// for feeding into an output Masker. Unknown types and schemes without
// secrets yield no values.
func MaskValues(typ string, cfg map[string]string, vars Interpolator) []string {
	s, ok := Lookup(typ)
	if !ok {
		return nil
	}
	sks, ok := s.(SecretKeyScheme)
	if !ok {
		return nil
	}
	var values []string
	for _, key := range sks.SecretKeys() {
		raw, ok := cfg[key]
		if !ok || raw == "" {
			continue
		}
		resolved, err := vars.Interpolate(raw)
		if err != nil || resolved == "" {
			continue
		}
		values = append(values, resolved)
	}
	return values
}

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Scheme)
)

// Register adds a scheme to the registry under name. It panics on duplicate
// registration so a misspelled scheme name surfaces at startup.
func Register(name string, s Scheme) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, dup := registry[name]; dup {
		panic(fmt.Sprintf("auth: scheme %q already registered", name))
	}
	registry[name] = s
}

// Apply dispatches to the scheme registered for typ. An empty type is a
// no-op (no auth configured). The "none" scheme explicitly clears inherited
// auth. Any other unregistered type is an error so typos surface instead of
// silently sending an unauthenticated request.
func Apply(req *http.Request, typ string, cfg map[string]string, vars Interpolator) error {
	if typ == "" {
		return nil
	}
	registryMu.RLock()
	s, ok := registry[typ]
	registryMu.RUnlock()
	if !ok {
		return fmt.Errorf("unknown auth type %q", typ)
	}
	return s.Apply(req, cfg, vars)
}

// Lookup returns the scheme registered for typ and whether it exists.
func Lookup(typ string) (Scheme, bool) {
	if typ == "" {
		return nil, false
	}
	registryMu.RLock()
	defer registryMu.RUnlock()
	s, ok := registry[typ]
	return s, ok
}

// Challenge applies a 401 WWW-Authenticate challenge to req for the scheme
// registered under typ. It reports whether a retry should follow (true when
// the scheme is challenge-based and applied the credentials).
func Challenge(req *http.Request, typ, challenge string, cfg map[string]string, vars Interpolator) (bool, error) {
	s, ok := Lookup(typ)
	if !ok {
		return false, nil
	}
	cs, ok := s.(ChallengedScheme)
	if !ok {
		return false, nil
	}
	if err := cs.Challenge(req, challenge, cfg, vars); err != nil {
		return false, err
	}
	return true, nil
}
