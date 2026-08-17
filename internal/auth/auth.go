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
