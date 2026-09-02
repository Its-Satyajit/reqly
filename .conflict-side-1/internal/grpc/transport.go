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

package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Transport selects the connection security for a gRPC call (M43 T4).
// Zero value is plaintext h2c.
type Transport struct {
	// TLS enables TLS against system roots.
	TLS bool
	// TLSSkipVerify disables certificate verification (dev only).
	TLSSkipVerify bool
	// CAFile trusts an explicit PEM CA bundle instead of system roots.
	CAFile string
}

// dial opens a client connection using the given transport (plaintext h2c by
// default).
func dial(_ context.Context, target string, tr Transport) (*grpc.ClientConn, error) {
	creds, err := transportCredentials(tr)
	if err != nil {
		return nil, err
	}
	return grpc.NewClient(target, grpc.WithTransportCredentials(creds))
}

// transportCredentials builds gRPC credentials from the transport selection.
func transportCredentials(tr Transport) (credentials.TransportCredentials, error) {
	if !tr.TLS {
		return insecure.NewCredentials(), nil
	}
	cfg := &tls.Config{}
	if tr.TLSSkipVerify {
		cfg.InsecureSkipVerify = true // dev-only escape hatch, explicitly requested
	}
	if tr.CAFile != "" {
		pem, err := os.ReadFile(tr.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read ca file %q: %w", tr.CAFile, err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("ca file %q contains no usable certificates", tr.CAFile)
		}
		cfg.RootCAs = pool
	}
	return credentials.NewTLS(cfg), nil
}
