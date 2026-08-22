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
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	jwtpkg "github.com/Its-Satyajit/reqly/internal/jwt"
)

var jwtCmd = &cobra.Command{
	Use:   "jwt",
	Short: "JWT tooling",
	Long:  `Offline JWT inspection tools (no network, no verification).`,
}

var jwtDecodeJSON bool

var jwtDecodeCmd = &cobra.Command{
	Use:   "decode <token> [--json]",
	Short: "Decode a JWT and show header, payload, and expiry",
	Long: `Decode a JWT without verification and show header, payload, and expiry.

  reqly jwt decode eyJhbG... --json
  echo $JWT | reqly jwt decode -
  reqly jwt decode "Bearer eyJ..."`,

	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		raw := args[0]
		if raw == "-" {
			b, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return fmt.Errorf("read stdin: %w", err)
			}
			raw = strings.TrimSpace(string(b))
			if raw == "" {
				return fmt.Errorf("invalid token: empty")
			}
		}
		tok, fieldErr := jwtpkg.Decode(raw)
		if tok == nil {
			// hard failure: malformed token
			if fieldErr != nil {
				return fieldErr
			}
			return fmt.Errorf("invalid token")
		}
		// Surface field-level expiry errors as warnings but still print.
		if fieldErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", fieldErr)
		}

		if jwtDecodeJSON {
			out := map[string]any{
				"header":    tok.Header,
				"payload":   tok.Payload,
				"signature": tok.Signature,
				"alg":       tok.Alg,
				"expiry":    tok.Expiry,
			}
			data, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		}

		// Human pretty.
		headerPretty, err := json.MarshalIndent(tok.Header, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal header: %w", err)
		}
		payloadPretty, err := json.MarshalIndent(tok.Payload, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Header:")
		fmt.Fprintln(cmd.OutOrStdout(), string(headerPretty))
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "Payload:")
		fmt.Fprintln(cmd.OutOrStdout(), string(payloadPretty))
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintf(cmd.OutOrStdout(), "Expiry: %s\n", formatJWTExpiry(tok.Expiry))
		if tok.Signature == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "Signature: (none)")
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Signature: %s\n", tok.Signature)
		}
		if tok.Alg != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Alg: %s\n", tok.Alg)
		}
		return nil
	},
}

func formatJWTExpiry(e jwtpkg.ExpiryStatus) string {
	base := ""
	switch e.Status {
	case "no_expiry":
		base = "no expiry"
	case "not_yet_valid":
		base = fmt.Sprintf("not yet valid (nbf in %s)", time.Duration(e.Remaining)*time.Second)
	case "expired":
		base = fmt.Sprintf("expired %s ago", time.Duration(-e.Remaining)*time.Second)
	case "valid":
		if e.Exp != nil {
			base = fmt.Sprintf("valid for %s", time.Duration(e.Remaining)*time.Second)
		} else {
			base = "valid"
		}
	default:
		base = e.Status
	}
	if e.Iat != nil {
		iatTime := time.Unix(*e.Iat, 0)
		age := jwtpkg.Now().Sub(iatTime).Round(time.Second)
		if age < 0 {
			age = -age
		}
		if e.Status == "no_expiry" {
			return fmt.Sprintf("%s (issued %s ago)", base, age)
		}
		if base != "" && age != 0 {
			return fmt.Sprintf("%s (issued %s ago)", base, age)
		}
	}
	return base
}

func init() {
	jwtDecodeCmd.Flags().BoolVar(&jwtDecodeJSON, "json", false, "output machine JSON")
	jwtCmd.AddCommand(jwtDecodeCmd)
}
