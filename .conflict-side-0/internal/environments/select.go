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
