// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/sso"
)

var ssoCmd = &cobra.Command{
	Use:   "sso",
	Short: "Enterprise SSO (OIDC token validation, local)",
	Long:  `Validate OIDC ID tokens against a local SSO config (issuer + HMAC).`,
}

var ssoValidateCmd = &cobra.Command{
	Use:   "validate --issuer <url> --client-id <id> --token <jwt> --secret <secret>",
	Short: "Validate an OIDC token",
	RunE: func(cmd *cobra.Command, args []string) error {
		issuer, _ := cmd.Flags().GetString("issuer")
		clientID, _ := cmd.Flags().GetString("client-id")
		token, _ := cmd.Flags().GetString("token")
		secret, _ := cmd.Flags().GetString("secret")
		cfg := sso.Config{Issuer: issuer, ClientID: clientID}
		if err := sso.ValidateToken(cfg, token, []byte(secret)); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Token is valid")
		return nil
	},
}

func init() {
	ssoValidateCmd.Flags().String("issuer", "", "OIDC issuer URL")
	ssoValidateCmd.Flags().String("client-id", "", "OIDC client ID")
	ssoValidateCmd.Flags().String("token", "", "JWT to validate")
	ssoValidateCmd.Flags().String("secret", "", "HMAC secret for HS256")
	_ = ssoValidateCmd.MarkFlagRequired("issuer")
	_ = ssoValidateCmd.MarkFlagRequired("client-id")
	_ = ssoValidateCmd.MarkFlagRequired("token")
	_ = ssoValidateCmd.MarkFlagRequired("secret")
	ssoCmd.AddCommand(ssoValidateCmd)
	rootCmd.AddCommand(ssoCmd)
}
