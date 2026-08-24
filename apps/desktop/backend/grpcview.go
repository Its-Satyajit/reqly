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

package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Its-Satyajit/reqly/internal/core"

	"github.com/Its-Satyajit/reqly/internal/grpc"
	"github.com/Its-Satyajit/reqly/internal/request"
)

// grpcEventName is the Wails event channel for one gRPC streaming session.
func grpcEventName(sessionID string) string {
	return "reqly.grpc." + sessionID
}

// GrpcServicesRequest targets one endpoint for reflection discovery.
type GrpcServicesRequest struct {
	Target     string   `json:"target"`
	ProtoFiles []string `json:"protoFiles,omitempty"`
	TLS        bool     `json:"tls,omitempty"`
}

// GrpcInvokeRequest describes one call. The full request file model rides
// along so auth/headers/grpc blocks behave exactly like a saved file.
type GrpcInvokeRequest struct {
	SessionID string          `json:"sessionId"`
	Request   request.Request `json:"request"`
}

// GrpcEvent is one streamed payload on reqly.grpc.<sessionID>.
type GrpcEvent struct {
	SessionID  string `json:"sessionId"`
	Type       string `json:"type"` // "message" | "done" | "error" | "cancelled"
	Seq        int    `json:"seq,omitempty"`
	Data       string `json:"data,omitempty"`
	CodeName   string `json:"codeName,omitempty"`
	DurationMS int64  `json:"durationMs,omitempty"`
}

var (
	grpcMu       sync.Mutex
	grpcSessions = make(map[string]context.CancelFunc)
)

func emitGrpcEvent(e *GrpcEvent) { emitRunEvent(grpcEventName(e.SessionID), e) }

// GrpcServices discovers services/methods via reflection (or protoFiles).
func (s *AppService) GrpcServices(req GrpcServicesRequest) ([]grpc.Service, error) {
	target := strings.TrimSpace(req.Target)
	if target == "" {
		return nil, fmt.Errorf("target is required")
	}
	return grpc.Discover(context.Background(), target, grpc.Transport{TLS: req.TLS, CAFile: ""})
}

// runGRPCRequest routes through the shared pipeline (interpolation,
// masking, history). The workspace must be open.
func (s *AppService) runGRPCRequest(ctx context.Context, r request.Request) (*grpc.Result, error) {
	if s == nil || s.root == "" {
		return nil, fmt.Errorf("no workspace found: open a reqly workspace to send")
	}
	res, err := s.requests.RunGRPC(ctx, r, core.RunRequestOptions{})
	if err != nil {
		return nil, err
	}
	return res.Result, nil
}

// GrpcInvoke performs one unary gRPC call.
func (s *AppService) GrpcInvoke(req GrpcInvokeRequest) (*grpc.Result, error) {
	res, err := s.runGRPCRequest(context.Background(), req.Request)
	if err != nil {
		return nil, err
	}
	if !res.OK {
		return res, fmt.Errorf("gRPC status %s (%d): %s", res.CodeName, res.Code, res.StatusMessage)
	}
	return res, nil
}

// GrpcStream opens a server-streaming call; messages stream on
// reqly.grpc.<sessionID> until the server closes or GrpcCancel is called.
func (s *AppService) GrpcStream(req GrpcInvokeRequest) error {
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		return fmt.Errorf("sessionId is required")
	}
	ctx, cancel := context.WithCancel(context.Background())
	grpcMu.Lock()
	if _, exists := grpcSessions[sessionID]; exists {
		grpcMu.Unlock()
		cancel()
		return fmt.Errorf("session %q already active", sessionID)
	}
	grpcSessions[sessionID] = cancel
	grpcMu.Unlock()

	go func() {
		defer func() {
			grpcMu.Lock()
			delete(grpcSessions, sessionID)
			grpcMu.Unlock()
		}()
		res, err := s.requests.RunGRPCStreamed(ctx, req.Request, core.RunRequestOptions{}, func(ev grpc.StreamEvent) error {
			emitGrpcEvent(&GrpcEvent{SessionID: sessionID, Type: "message", Seq: ev.Seq, Data: string(ev.MessageJSON)})
			return nil
		})
		if err != nil && ctx.Err() != nil {
			emitGrpcEvent(&GrpcEvent{SessionID: sessionID, Type: "cancelled"})
			return
		}
		if err != nil {
			emitGrpcEvent(&GrpcEvent{SessionID: sessionID, Type: "error", Data: err.Error()})
			return
		}
		done := &GrpcEvent{SessionID: sessionID, Type: "done", DurationMS: res.Result.DurationMS}
		if !res.Result.OK {
			done.CodeName = res.Result.CodeName
			done.Data = res.Result.StatusMessage
		}
		emitGrpcEvent(done)
	}()
	return nil
}

// GrpcCancel tears down one streaming session.
func (s *AppService) GrpcCancel(sessionID string) error {
	grpcMu.Lock()
	cancel, ok := grpcSessions[sessionID]
	grpcMu.Unlock()
	if !ok {
		return nil
	}
	cancel()
	return nil
}
