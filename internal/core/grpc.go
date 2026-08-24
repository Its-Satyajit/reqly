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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Its-Satyajit/reqly/internal/auth"
	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/environments"
	"github.com/Its-Satyajit/reqly/internal/grpc"
	"github.com/Its-Satyajit/reqly/internal/history"
	"github.com/Its-Satyajit/reqly/internal/request"
)

// DefaultGRPCTimeout applies when a request file omits grpc.timeout.
const DefaultGRPCTimeout = 30 * time.Second

// ErrStopConsumption stops a streamed call after the caller has had enough
// messages (e.g. CLI --max-messages). It never reaches the user as an error.
var ErrStopConsumption = errors.New("consumption stopped")

// RunGRPCResult couples the invoke outcome with pipeline warnings.
type RunGRPCResult struct {
	Result  *grpc.Result
	Warning string
}

// grpcCallPrep carries everything resolved by prepareGRPCCall.
type grpcCallPrep struct {
	call       grpc.Call
	message    []byte
	invokeOpts grpc.InvokeOptions
	masker     *environments.Masker
	env        string
}

// prepareGRPCCall runs the shared pipeline front-half for gRPC sends: env
// precedence, variable interpolation across url/message/metadata/protoFiles,
// auth masking seeds, and timeout resolution (ADR 0025/0028).
func (s *RequestService) prepareGRPCCall(ctx context.Context, r request.Request, opts RunRequestOptions) (*grpcCallPrep, error) {
	if s == nil || s.client == nil {
		return nil, errNoWorkspace()
	}
	g := r.GRPC
	if g == nil || strings.TrimSpace(g.Service) == "" || strings.TrimSpace(g.Method) == "" {
		return nil, fmt.Errorf("request has no grpc block (grpc.service / grpc.method)")
	}
	dir := s.root
	if dir == "" {
		dir = "."
	}

	envFlag := os.Getenv("REQLY_ENV")
	if envFlag == "" {
		envFlag = opts.EnvFlag
	}
	set, masker, err := environments.ResolveSet(dir, environments.Selection{
		EnvFlag:   envFlag,
		FileEnv:   opts.FileEnv,
		ConfigEnv: collections.WorkspaceEnvironment(dir),
	})
	if err != nil {
		return nil, err
	}
	layerScope(set, opts.FileVars)
	layerScope(set, opts.RuntimeVars)
	masker.Add(auth.MaskValues(r.Auth.Type, r.Auth.Config, set)...)

	interp := func(s string) string {
		out, ierr := set.Interpolate(s)
		if ierr != nil {
			return ""
		}
		return out
	}
	target := interp(r.URL)
	if target == "" {
		return nil, fmt.Errorf("request url (host:port) is required")
	}

	var message []byte
	if g.Message != nil {
		raw, merr := json.Marshal(g.Message)
		if merr != nil {
			return nil, fmt.Errorf("encode grpc.message: %w", merr)
		}
		rendered, ierr := set.Interpolate(string(raw))
		if ierr != nil {
			return nil, fmt.Errorf("interpolate grpc.message: %w", ierr)
		}
		message = []byte(rendered)
	}

	metadata := map[string]string{}
	for _, h := range r.Headers {
		key := strings.TrimSpace(h.Key)
		if key == "" {
			continue
		}
		metadata[key] = interp(h.Value)
	}

	protoFiles := make([]string, 0, len(g.ProtoFiles))
	for _, p := range g.ProtoFiles {
		protoFiles = append(protoFiles, interp(p))
	}

	timeout := DefaultGRPCTimeout
	if strings.TrimSpace(g.Timeout) != "" {
		parsed, perr := time.ParseDuration(g.Timeout)
		if perr != nil {
			return nil, fmt.Errorf("invalid grpc.timeout %q: %w", g.Timeout, perr)
		}
		timeout = parsed
	}

	return &grpcCallPrep{
		call:       grpc.Call{Target: target, Service: g.Service, Method: g.Method, ProtoFiles: protoFiles},
		message:    message,
		invokeOpts: grpc.InvokeOptions{Metadata: metadata, Timeout: timeout, Transport: grpc.Transport{TLS: g.TLS, TLSSkipVerify: g.TLSSkipVerify, CAFile: interp(g.CAFile)}},
		masker:     masker,
		env:        envName(envFlag, opts),
	}, nil
}

// maskResult applies secret masking to every user-visible surface of a
// result; history already holds raw bytes.
func maskResult(masker *environments.Masker, res *grpc.Result) *grpc.Result {
	for i := range res.StatusDetails {
		res.StatusDetails[i].Data = masker.Mask(res.StatusDetails[i].Data)
	}
	res.StatusMessage = masker.Mask(res.StatusMessage)
	res.MessageJSON = []byte(masker.Mask(string(res.MessageJSON)))
	return res
}

// RunGRPC sends a unary gRPC call through the same execution pipeline as HTTP
// (ADR 0025/0028). Non-OK gRPC statuses are results, not errors — they render
// masked like failed responses. For server-streaming methods use
// RunGRPCStreamed; this function rejects them with a clear error.
func (s *RequestService) RunGRPC(ctx context.Context, r request.Request, opts RunRequestOptions) (*RunGRPCResult, error) {
	prep, err := s.prepareGRPCCall(ctx, r, opts)
	if err != nil {
		return nil, err
	}
	res, ierr := grpc.Invoke(ctx, prep.call, prep.message, prep.invokeOpts)
	if ierr != nil {
		return nil, maskErr(prep.masker, ierr)
	}
	warning := s.recordGRPC(opts, r, res, prep.message, prep.env)
	return &RunGRPCResult{
		Result:  maskResult(prep.masker, res),
		Warning: strings.TrimSpace(warning),
	}, nil
}

// RunGRPCStreamed sends a server-streaming gRPC call through the pipeline.
// onMessage receives each masked response message in delivery order; return
// ErrStopConsumption to stop early (the call reports the terminal status).
// Unary methods dispatch through RunGRPC automatically.
func (s *RequestService) RunGRPCStreamed(ctx context.Context, r request.Request, opts RunRequestOptions, onMessage func(grpc.StreamEvent) error) (*RunGRPCResult, error) {
	prep, err := s.prepareGRPCCall(ctx, r, opts)
	if err != nil {
		return nil, err
	}

	messageCount := 0
	res, ierr := grpc.InvokeStream(ctx, prep.call, prep.message, prep.invokeOpts, func(ev grpc.StreamEvent) error {
		ev.MessageJSON = []byte(prep.masker.Mask(string(ev.MessageJSON)))
		messageCount++
		return onMessage(ev)
	})
	if ierr != nil && !errors.Is(ierr, ErrStopConsumption) && !strings.Contains(strings.ToLower(ierr.Error()), "not server-streaming") {
		return nil, maskErr(prep.masker, ierr)
	}
	if ierr != nil && strings.Contains(strings.ToLower(ierr.Error()), "not server-streaming") {
		// Unary method requested through the streaming surface.
		unary, uerr := s.RunGRPC(ctx, r, opts)
		if uerr != nil {
			return nil, uerr
		}
		if errors.Is(ierr, ErrStopConsumption) {
			return unary, ErrStopConsumption
		}
		return unary, nil
	}

	warning := s.recordStreamGRPC(opts, r, res, messageCount, prep.env)
	return &RunGRPCResult{
		Result:  maskResult(prep.masker, res),
		Warning: strings.TrimSpace(warning),
	}, ierr // ErrStopConsumption propagates to the caller as a sentinel
}

// recordGRPC writes one history row for a unary send. History stores raw
// bytes (masking happens at render); Status carries the gRPC code for
// non-OK statuses and 200-equivalent OK.
func (s *RequestService) recordGRPC(opts RunRequestOptions, r request.Request, res *grpc.Result, messageJSON []byte, env string) string {
	respBody := res.MessageJSON
	statusCode := 200
	if !res.OK {
		statusCode = int(res.Code)
		respBody = []byte(fmt.Sprintf("%s (%d): %s", res.CodeName, res.Code, res.StatusMessage))
	}
	return s.insertGRPCRow(opts, r, statusCode, res.DurationMS, messageJSON, respBody, env)
}

// recordStreamGRPC writes one summary history row for a streamed call.
func (s *RequestService) recordStreamGRPC(opts RunRequestOptions, r request.Request, res *grpc.Result, count int, env string) string {
	statusCode := 200
	if !res.OK {
		statusCode = int(res.Code)
	}
	body := fmt.Sprintf("%d messages", count)
	if !res.OK {
		body = fmt.Sprintf("%s (%d): %s after %d messages", res.CodeName, res.Code, res.StatusMessage, count)
	}
	return s.insertGRPCRow(opts, r, statusCode, res.DurationMS, []byte(body), []byte(body), env)
}

func (s *RequestService) insertGRPCRow(opts RunRequestOptions, r request.Request, status int, durationMS int64, reqBody, respBody []byte, env string) string {
	h := s.recorder()
	if h == nil {
		return ""
	}
	reqHdrs := map[string][]string{}
	for _, hd := range r.Headers {
		reqHdrs[hd.Key] = append(reqHdrs[hd.Key], hd.Value)
	}
	e := &history.Entry{
		RequestPath: opts.RequestPath,
		Method:      "GRPC",
		URL:         fmt.Sprintf("%s %s.%s", r.URL, r.GRPC.Service, r.GRPC.Method),
		Env:         env,
		Status:      status,
		DurationMS:  durationMS,
		Size:        int64(len(respBody)),
		ReqHeaders:  reqHdrs,
		ReqBody:     reqBody,
		RespBody:    respBody,
	}
	if err := h.insert(context.Background(), e); err != nil {
		return "history recording failed: " + err.Error()
	}
	return ""
}
