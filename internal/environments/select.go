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

package environments

import (
	"os"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/variables"
)

// Selection captures the possible sources of an active environment, ordered by
// precedence (highest wins): an explicit flag or REQLY_ENV, then a request or
// test file's environment: field, then the workspace descriptor's
// environment: field.
type Selection struct {
	// EnvFlag is the environment selected via the --env flag or REQLY_ENV.
	EnvFlag string
	// FileEnv is the environment: field of a request or test file.
	FileEnv string
	// ConfigEnv is the environment: field of the workspace descriptor.
	ConfigEnv string
}

// Active returns the environment name selected with the highest precedence,
// or "" when no selection is present.
func (s Selection) Active() string {
	for _, candidate := range []string{s.EnvFlag, s.FileEnv, s.ConfigEnv} {
		if candidate != "" {
			return candidate
		}
	}
	return ""
}

// ResolveSet builds the execution variable set for dir: the process-env scope
// (OS env + nearest .env, always populated) plus the active environment's
// variables and secrets in the environment scope. It also returns a Masker
// covering the sensitive values for output: the active environment's secrets
// plus every value loaded from the .env file, so both are redacted even when a
// key name does not look secret. The masker is never nil. A selected-but-
// missing environment is a hard error.
func ResolveSet(dir string, sel Selection) (*variables.Set, *Masker, error) {
	set := variables.NewSet()

	// Only values that came from the .env file (not the OS environment) are
	// treated as sensitive for masking, matching the dotenv convention that
	// local credentials are sensitive regardless of key name.
	dotenvVars, err := DotEnvValues(dir)
	if err != nil {
		return nil, nil, err
	}

	// Process-env scope: OS environment plus the nearest .env, with the OS
	// winning on key conflicts so CI can override local values.
	for _, kv := range os.Environ() {
		key, value, found := strings.Cut(kv, "=")
		if !found {
			continue
		}
		set.Set(variables.ScopeProcessEnv, key, value)
	}
	maskValues := make([]string, 0, len(dotenvVars))
	for key, value := range dotenvVars {
		if _, exists := set.Get(variables.ScopeProcessEnv, key); !exists {
			set.Set(variables.ScopeProcessEnv, key, value)
		}
		maskValues = append(maskValues, value)
	}

	active := sel.Active()
	if active != "" {
		env, err := Read(active, dir)
		if err != nil {
			return nil, nil, err
		}
		for key, value := range env.Variables {
			set.Set(variables.ScopeEnvironment, key, value)
		}
		for key, value := range env.Secrets {
			set.Set(variables.ScopeEnvironment, key, value)
		}
		maskValues = append(maskValues, env.SecretValues()...)
	}
	return set, NewMasker(maskValues...), nil
}
