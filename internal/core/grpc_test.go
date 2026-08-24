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

package core

import (
	"context"
	"strings"
	"testing"

	"github.com/Its-Satyajit/reqly/internal/grpc/testsrv"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/testsupport"
)

func TestRunGRPCRequiresGrpcBlock(t *testing.T) {
	dir := newRunWS(t)
	svc := NewRunService(dir)
	defer svc.Close()
	if _, err := svc.RunGRPC(context.Background(), request.Request{URL: "127.0.0.1:1"}, RunRequestOptions{}); err == nil {
		t.Fatal("expected error for missing grpc block")
	}
}

func newGRPCRunWorkspace(t *testing.T) *RequestService {
	t.Helper()
	dir := testsupport.Workspace(t, map[string]string{
		"reqly.yaml":            "name: ws\nenvironment: dev\n",
		"environments/dev.yaml": "name: dev\nvariables:\n  who: from-env\ndescription: \"\"\nsecrets:\n  api_key: super-secret-token\n",
	})
	return NewRunService(dir)
}

func TestRunGRPCInterpolatesMessageAndMetadata(t *testing.T) {
	srv := testsrv.Start(t)
	svc := newGRPCRunWorkspace(t)
	defer svc.Close()

	res, err := svc.RunGRPC(context.Background(), request.Request{
		URL:     srv.Addr,
		Headers: []request.Header{{Key: "authorization", Value: "Bearer {{api_key}}"}},
		GRPC: &request.GRPC{
			Service: "reqly.test.v1.EchoService",
			Method:  "Echo",
			Message: map[string]any{"text": "hello {{who}}"},
		},
	}, RunRequestOptions{})
	if err != nil {
		t.Fatalf("RunGRPC: %v", err)
	}
	if !res.Result.OK {
		t.Fatalf("expected OK: %+v", res.Result)
	}
	if !strings.Contains(string(res.Result.MessageJSON), "hello from-env") {
		t.Errorf("message not interpolated: %s", res.Result.MessageJSON)
	}
	// The secret must never appear in rendered output.
	if strings.Contains(string(res.Result.MessageJSON), "super-secret-token") {
		t.Errorf("secret leaked into response view")
	}
}

func TestRunGRPCHistoryRowRecorded(t *testing.T) {
	srv := testsrv.Start(t)
	svc := newGRPCRunWorkspace(t)
	defer svc.Close()

	if _, err := svc.RunGRPC(context.Background(), request.Request{
		URL: srv.Addr,
		GRPC: &request.GRPC{
			Service: "reqly.test.v1.EchoService",
			Method:  "Echo",
			Message: map[string]any{"text": "history me"},
		},
	}, RunRequestOptions{RequestPath: "collections/users/rpc.reqly.json"}); err != nil {
		t.Fatalf("RunGRPC: %v", err)
	}

	entries, err := svc.History().List(context.Background(), 10, 0, nil)
	if err != nil {
		t.Fatalf("list history: %v", err)
	}
	var found bool
	for _, e := range entries {
		if e.Method == "GRPC" && strings.Contains(e.URL, "reqly.test.v1.EchoService") {
			found = true
			if e.Status != 200 {
				t.Errorf("history status = %d, want 200", e.Status)
			}
			if !strings.Contains(string(e.RespBody), "history me") {
				t.Errorf("history response body = %q", e.RespBody)
			}
		}
	}
	if !found {
		t.Fatalf("no GRPC history row; entries = %+v", entries)
	}
}

func TestRunGRPCNonOKStatusMaskedAndRecorded(t *testing.T) {
	srv := testsrv.Start(t)
	svc := newGRPCRunWorkspace(t)
	defer svc.Close()

	res, err := svc.RunGRPC(context.Background(), request.Request{
		URL:     srv.Addr,
		Headers: []request.Header{{Key: "x-reqly", Value: "{{api_key}}"}},
		GRPC: &request.GRPC{
			Service: "reqly.test.v1.FailingService",
			Method:  "Boom",
		},
	}, RunRequestOptions{})
	if err != nil {
		t.Fatalf("RunGRPC (non-OK is a result): %v", err)
	}
	if res.Result.OK {
		t.Fatal("expected non-OK result")
	}
	for _, d := range res.Result.StatusDetails {
		if strings.Contains(d.Data, "super-secret-token") {
			t.Errorf("secret leaked into status details")
		}
	}
	entries, herr := svc.History().List(context.Background(), 10, 0, nil)
	if herr != nil {
		t.Fatalf("list history: %v", herr)
	}
	var found bool
	for _, e := range entries {
		if e.Method == "GRPC" && strings.Contains(e.URL, "FailingService") {
			found = true
			if e.Status != int(5) { // codes.NotFound
				t.Errorf("history status = %d, want 5 (NotFound)", e.Status)
			}
		}
	}
	if !found {
		t.Error("no history row for failed call")
	}
}
