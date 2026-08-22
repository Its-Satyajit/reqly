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
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/auth"
)

var rootCmd = &cobra.Command{
	Use:   "reqly",
	Short: "A local-first API development environment",
	Long: `A local-first, Git-native API development environment.

Requests, tests, schemas, mocks, environments, and documentation live together
as version-controlled project files. The CLI shares the same Go core as the
desktop application.`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Automatic authorization_code acquisition (a first request with no
	// usable cached token) opens the system browser via the shared launcher,
	// so reqly run/test/collection work without a separate login step.
	auth.SetOAuth2BrowserOpener(func(_ context.Context, authorizationURL string) error {
		return launchBrowser(authorizationURL)
	})
	// Automatic device-flow acquisition prints the verification URI + code to
	// stderr so the user can approve without a separate login step.
	auth.SetOAuth2DeviceStatus(func(line string) {
		fmt.Fprintln(os.Stderr, line)
	})
	rootCmd.AddCommand(
		runCmd,
		testCmd,
		collectionCmd,
		importCmd,
		exportCmd,
		wsCmd,
		sseCmd,
		mockCmd,
		validateCmd,
		diffCmd,
		docsCmd,
		versionCmd,
		jwtCmd,
		paginationCmd,
	)
}
