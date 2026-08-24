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

// Package testsrv provides the shared in-process gRPC fixture used by M43
// tests across packages: a plaintext server with reflection enabled, one
// unary echo method, and one server-streaming echo method.
package testsrv

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Server is the running fixture; Addr is the plaintext host:port clients dial.
type Server struct {
	Addr string
	srv  *grpc.Server
	// CAPath is the trusted CA bundle for TLS fixtures (empty for plaintext).
	CAPath string
}

// Greeter is implemented by tests to observe incoming calls.
type Greeter interface {
	Echo(ctx context.Context, req *EchoRequest) (*EchoResponse, error)
}

// Start runs the fixture on 127.0.0.1:0 with reflection enabled and registers
// cleanup on t.
func Start(t testing.TB) *Server {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("testsrv listen: %v", err)
	}
	srv := grpc.NewServer()
	RegisterEchoServiceServer(srv, &fixture{})
	RegisterFailingServiceServer(srv, &fixture{})
	RegisterSlowServiceServer(srv, &fixture{})
	reflection.Register(srv)
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(srv.Stop)
	return &Server{Addr: ln.Addr().String(), srv: srv}
}

// fixture echoes requests back; streaming responses carry their sequence
// number so ordering assertions are possible.
type fixture struct {
	UnimplementedEchoServiceServer
	UnimplementedFailingServiceServer
	UnimplementedSlowServiceServer
}

func (f *fixture) Echo(_ context.Context, req *EchoRequest) (*EchoResponse, error) {
	return &EchoResponse{Text: req.GetText(), Labels: req.GetLabels()}, nil
}

func (f *fixture) Slow(_ context.Context, _ *EchoRequest) (*EchoResponse, error) {
	time.Sleep(500 * time.Millisecond)
	return &EchoResponse{Text: "finally"}, nil
}

func (f *fixture) Boom(_ context.Context, _ *EchoRequest) (*EchoResponse, error) {
	return nil, status.Errorf(codes.NotFound, "thing missing")
}

func (f *fixture) StreamEcho(req *EchoRequest, stream EchoService_StreamEchoServer) error {
	for i := range 3 {
		if err := stream.Send(&EchoResponse{Text: req.GetText(), Sequence: int32(i + 1)}); err != nil {
			return err
		}
	}
	return nil
}

// StartPlain runs the fixture's echo services WITHOUT reflection — the
// negative-path fixture for reflection-disabled servers.
func StartPlain(t testing.TB) *Server {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("testsrv listen: %v", err)
	}
	srv := grpc.NewServer()
	RegisterEchoServiceServer(srv, &fixture{})
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(srv.Stop)
	return &Server{Addr: ln.Addr().String(), srv: srv}
}

// StartTLS runs the fixture behind TLS with a self-signed certificate
// written to a temp dir (CA bundle at CAPath). Clients trust it via
// Transport{TLS: true, CAFile: s.CAPath}.
func StartTLS(t testing.TB) *Server {
	t.Helper()
	caPEM, certPEM, keyPEM := selfSigned(t)
	caFile := filepath.Join(t.TempDir(), "ca.pem")
	certFile := filepath.Join(t.TempDir(), "cert.pem")
	keyFile := filepath.Join(t.TempDir(), "key.pem")
	write(t, caFile, caPEM)
	write(t, certFile, certPEM)
	write(t, keyFile, keyPEM)

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("key pair: %v", err)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("testsrv listen: %v", err)
	}
	srv := grpc.NewServer(grpc.Creds(credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})))
	RegisterEchoServiceServer(srv, &fixture{})
	reflection.Register(srv)
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(srv.Stop)
	return &Server{Addr: ln.Addr().String(), srv: srv, CAPath: caFile}
}

func write(t testing.TB, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// selfSigned generates a self-signed certificate for 127.0.0.1 at test time.
func selfSigned(t testing.TB) (caPEM, certPEM, keyPEM []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa key: %v", err)
	}
	tmpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "127.0.0.1"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost"},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	caPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	certPEM = caPEM
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return caPEM, certPEM, keyPEM
}
