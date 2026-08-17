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
	)
}
