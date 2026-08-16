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

package main

import (
	"fmt"
	"os"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/core"
	"github.com/Its-Satyajit/reqly/internal/environments"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/variables"
)

// AppService is a Wails v3 service exposing the Go core to the frontend.
// It should stay thin: business logic belongs in the Go core (internal/core).
type AppService struct {
	requests *core.RequestService
}

// NewAppService creates a new AppService.
func NewAppService() *AppService {
	return &AppService{
		requests: core.NewRequestService(),
	}
}

// SendRequest executes an HTTP request through the core and returns the
// bridge-friendly response for the frontend to render. The active environment
// is resolved from the workspace rooted at the app's working directory and its
// variables are layered under the request for interpolation.
func (s *AppService) SendRequest(r request.Request) (*core.SendResponse, error) {
	vars, err := resolveAppEnvironment()
	if err != nil {
		return nil, err
	}
	return s.requests.Send(r, vars)
}

// resolveAppEnvironment loads the process-env scope plus the environment
// selected from the working directory by REQLY_ENV or the workspace
// descriptor's environment: field (the --env CLI flag is not available to the
// desktop). When no workspace or environment is present, the returned set
// still carries the process-env scope so OS variables always interpolate.
func resolveAppEnvironment() (*variables.Set, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("resolve working directory: %w", err)
	}
	sel := environments.Selection{
		EnvFlag:   os.Getenv("REQLY_ENV"),
		ConfigEnv: collections.WorkspaceEnvironment(dir),
	}
	set, _, err := environments.ResolveSet(dir, sel)
	if err != nil {
		return nil, err
	}
	return set, nil
}
