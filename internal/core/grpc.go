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

// RunGRPCResult couples the invoke outcome with pipeline warnings.
type RunGRPCResult struct {
	Result  *grpc.Result
	Warning string
}

// RunGRPC sends a gRPC call through the same execution pipeline as HTTP
// (ADR 0025/0028): env precedence, variable interpolation across url /
// message / metadata / protoFiles, secret masking of metadata and status
// surfaces, and history recording. Non-OK gRPC statuses are results, not
// errors — they render masked like failed responses.
func (s *RequestService) RunGRPC(ctx context.Context, r request.Request, opts RunRequestOptions) (*RunGRPCResult, error) {
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

	var messageJSON []byte
	if g.Message != nil {
		raw, merr := json.Marshal(g.Message)
		if merr != nil {
			return nil, fmt.Errorf("encode grpc.message: %w", merr)
		}
		rendered, ierr := set.Interpolate(string(raw))
		if ierr != nil {
			return nil, fmt.Errorf("interpolate grpc.message: %w", ierr)
		}
		messageJSON = []byte(rendered)
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

	res, ierr := grpc.Invoke(ctx,
		grpc.Call{Target: target, Service: g.Service, Method: g.Method, ProtoFiles: protoFiles},
		messageJSON,
		grpc.InvokeOptions{
			Metadata:  metadata,
			Timeout:   timeout,
			Transport: grpc.Transport{TLS: g.TLS, TLSSkipVerify: g.TLSSkipVerify, CAFile: interp(g.CAFile)},
		},
	)
	if ierr != nil {
		return nil, maskErr(masker, ierr)
	}

	env := envName(envFlag, opts)
	warning := s.recordGRPC(opts, r, res, messageJSON, env)

	// Mask every user-visible surface; history already holds raw bytes.
	for i := range res.StatusDetails {
		res.StatusDetails[i].Data = masker.Mask(res.StatusDetails[i].Data)
	}
	res.StatusMessage = masker.Mask(res.StatusMessage)
	res.MessageJSON = []byte(masker.Mask(string(res.MessageJSON)))

	return &RunGRPCResult{Result: res, Warning: strings.TrimSpace(warning)}, nil
}

// recordGRPC writes one history row for a unary send. History stores raw
// bytes (masking happens at render); Status carries the gRPC code for
// non-OK statuses and 200-equivalent OK.
func (s *RequestService) recordGRPC(opts RunRequestOptions, r request.Request, res *grpc.Result, messageJSON []byte, env string) string {
	h := s.recorder()
	if h == nil {
		return ""
	}
	reqHdrs := map[string][]string{}
	for _, hd := range r.Headers {
		reqHdrs[hd.Key] = append(reqHdrs[hd.Key], hd.Value)
	}
	statusCode := int(res.Code)
	if res.OK {
		statusCode = 200
	}
	respBody := res.MessageJSON
	if !res.OK {
		respBody = []byte(fmt.Sprintf("%s (%d): %s", res.CodeName, res.Code, res.StatusMessage))
	}
	e := &history.Entry{
		RequestPath: opts.RequestPath,
		Method:      "GRPC",
		URL:         fmt.Sprintf("%s %s.%s", r.URL, r.GRPC.Service, r.GRPC.Method),
		Env:         env,
		Status:      statusCode,
		DurationMS:  res.DurationMS,
		Size:        int64(len(respBody)),
		ReqHeaders:  reqHdrs,
		ReqBody:     messageJSON,
		RespBody:    respBody,
	}
	if err := h.insert(context.Background(), e); err != nil {
		return "history recording failed: " + err.Error()
	}
	return ""
}
