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
