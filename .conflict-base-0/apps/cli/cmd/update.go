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
	"runtime"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/update"
	"github.com/Its-Satyajit/reqly/internal/version"
)

var (
	updateCheckOnly bool
	updateTargetVer string
)

var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"upgrade"},
	Short:   "Check for and install Reqly updates",
	Long: `Check GitHub for the latest Reqly release and automatically download
and replace the current executable binary.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		checker := update.NewChecker()

		fmt.Fprintln(cmd.OutOrStdout(), "Checking for updates...")
		info, err := checker.Check(ctx, version.Version)
		if err != nil {
			return fmt.Errorf("check for updates: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", info.CurrentVersion)
		fmt.Fprintf(cmd.OutOrStdout(), "Latest version:  %s\n", info.LatestVersion)

		if !info.HasUpdate && updateTargetVer == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "Reqly is already up to date!")
			return nil
		}

		if updateCheckOnly {
			if info.HasUpdate {
				fmt.Fprintf(cmd.OutOrStdout(), "\nA newer version (%s) is available!\nRun 'reqly update' to install.\n", info.LatestVersion)
			}
			return nil
		}

		asset, found := update.FindAssetForPlatform(info.Assets, runtime.GOOS, runtime.GOARCH)
		if !found {
			return fmt.Errorf("no matching binary asset found for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, info.LatestVersion)
		}

		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("locate running executable: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\nDownloading %s...\n", asset.Name)
		if err := checker.ApplyBinaryUpdate(ctx, asset.DownloadURL, execPath); err != nil {
			return fmt.Errorf("apply update: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Successfully updated Reqly to %s!\n", info.LatestVersion)
		return nil
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheckOnly, "check", false, "Check for updates without downloading or installing")
	updateCmd.Flags().StringVar(&updateTargetVer, "version", "", "Target a specific release version")
}
