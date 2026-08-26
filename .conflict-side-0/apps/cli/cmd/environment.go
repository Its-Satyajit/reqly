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

package cmd

import (
	"os"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/environments"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// envFlag holds the --env flag value shared by run/test/collection commands.
// It is registered by each command that supports it.
var envFlag string

// activeEnvironment resolves the active environment for a run and returns the
// populated variable set plus a Masker for output redaction. The environment
// scope is layered with the process-env scope at the bottom of the precedence
// chain.
//
// Selection precedence (highest wins): REQLY_ENV → --env flag → request/test
// file environment: field → workspace descriptor environment: field.
func activeEnvironment(dir, fileEnv string) (*environments.Masker, *variables.Set, error) {
	sel := environments.Selection{
		EnvFlag:   envSelection(os.Getenv("REQLY_ENV"), envFlag),
		FileEnv:   fileEnv,
		ConfigEnv: collections.WorkspaceEnvironment(dir),
	}
	set, masker, err := environments.ResolveSet(dir, sel)
	if err != nil {
		return nil, nil, err
	}
	return masker, set, nil
}

// envSelection applies selection precedence to the REQLY_ENV process variable
// and the --env CLI flag: REQLY_ENV wins, then the flag.
func envSelection(processVar, flagValue string) string {
	if processVar != "" {
		return processVar
	}
	return flagValue
}

// mergeEnvScope copies the process-env and environment scopes of src into dst,
// so a resolved environment layers under the file/request scopes already in
// dst without disturbing them.
func mergeEnvScope(dst, src *variables.Set) {
	for _, scope := range []variables.Scope{variables.ScopeProcessEnv, variables.ScopeEnvironment} {
		src.Range(scope, func(key, value string) {
			dst.Set(scope, key, value)
		})
	}
}
