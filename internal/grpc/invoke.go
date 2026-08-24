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
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/bufbuild/protocompile"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"

	refv1 "google.golang.org/grpc/reflection/grpc_reflection_v1"
)

// Call addresses one gRPC method on one endpoint.
type Call struct {
	Target     string   `json:"target"`
	Service    string   `json:"service"`
	Method     string   `json:"method"`
	ProtoFiles []string `json:"protoFiles,omitempty"`
}

// InvokeOptions carry per-call settings: metadata (from the request file's
// headers) and the deadline duration.
type InvokeOptions struct {
	Metadata map[string]string `json:"metadata,omitempty"`
	Timeout  time.Duration     `json:"timeout,omitempty"`
}

// Result reports one completed unary call. OK is false when the server
// returned a non-OK status; then Code/CodeName/StatusMessage carry the
// failure and MessageJSON is empty.
type Result struct {
	MessageJSON   []byte          `json:"messageJson,omitempty"`
	OK            bool            `json:"ok"`
	DurationMS    int64           `json:"durationMs,omitempty"`
	Code          uint32          `json:"code,omitempty"`
	CodeName      string          `json:"codeName,omitempty"`
	StatusMessage string          `json:"statusMessage,omitempty"`
	StatusDetails []*StatusDetail `json:"statusDetails,omitempty"`
}

// StatusDetail is one error detail entry from a non-OK status.
type StatusDetail struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
}

// headerKey normalizes metadata keys to lowercase — gRPC requires it and
// request files may not bother.
func headerKey(k string) string { return strings.ToLower(strings.TrimSpace(k)) }

// Invoke performs one unary call against target/service/method. The message
// arrives as canonical-JSON bytes and is decoded into the method's input type
// via dynamic descriptors resolved from reflection or protoFiles.
func Invoke(ctx context.Context, call Call, messageJSON []byte, opts InvokeOptions) (*Result, error) {
	conn, err := dial(ctx, call.Target)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", call.Target, err)
	}
	defer conn.Close()

	methodDesc, err := resolveMethod(ctx, conn, call)
	if err != nil {
		return nil, err
	}
	if methodDesc.IsStreamingClient() || methodDesc.IsStreamingServer() {
		return nil, fmt.Errorf("%s is streaming; unary invoke does not apply", methodDesc.FullName())
	}

	in := dynamicpb.NewMessage(methodDesc.Input())
	dec := protojson.UnmarshalOptions{DiscardUnknown: true}
	if len(messageJSON) > 0 {
		if err := dec.Unmarshal(messageJSON, in); err != nil {
			return nil, fmt.Errorf("invalid message JSON for %s: %w", methodDesc.FullName(), err)
		}
	}
	out := dynamicpb.NewMessage(methodDesc.Output())

	callOpts := []grpc.CallOption{}
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}
	if len(opts.Metadata) > 0 {
		md := metadata.MD{}
		for k, v := range opts.Metadata {
			md.Set(headerKey(k), v)
		}
		callOpts = append(callOpts, grpc.Header(&md))
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	res := &Result{}
	start := time.Now()
	rpcMethod := "/" + string(methodDesc.FullName()[:strings.LastIndex(string(methodDesc.FullName()), ".")]) +
		"/" + string(methodDesc.Name())
	rpcErr := conn.Invoke(ctx, rpcMethod, in, out, callOpts...)
	res.DurationMS = time.Since(start).Milliseconds()
	if rpcErr != nil {
		st, ok := status.FromError(rpcErr)
		if !ok {
			return nil, fmt.Errorf("invoke %s: %w", methodDesc.FullName(), rpcErr)
		}
		res.OK = false
		res.Code = uint32(st.Code())
		res.CodeName = st.Code().String()
		res.StatusMessage = st.Message()
		for _, d := range st.Details() {
			res.StatusDetails = append(res.StatusDetails, detailToJSON(d))
		}
		return res, nil
	}
	msg, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("encode response message: %w", err)
	}
	return &Result{OK: true, MessageJSON: msg}, nil
}

// detailToJSON renders one status detail (a proto message) as type + JSON data.
func detailToJSON(d any) *StatusDetail {
	msg, ok := d.(proto.Message)
	if !ok {
		return &StatusDetail{Type: fmt.Sprintf("%T", d)}
	}
	data, err := protojson.MarshalOptions{}.Marshal(msg)
	if err != nil {
		return &StatusDetail{Type: string(msg.ProtoReflect().Descriptor().FullName())}
	}
	return &StatusDetail{Type: string(msg.ProtoReflect().Descriptor().FullName()), Data: string(data)}
}

// resolveMethod finds the method descriptor via protoFiles (when given) or
// server reflection otherwise.
func resolveMethod(ctx context.Context, conn *grpc.ClientConn, call Call) (protoreflect.MethodDescriptor, error) {
	service := strings.TrimSpace(call.Service)
	method := strings.TrimSpace(call.Method)
	if service == "" || method == "" {
		return nil, fmt.Errorf("grpc.service and grpc.method are required")
	}
	symbol := service + "." + method
	fullName := "/" + service + "/" + method

	files, err := schemaFiles(ctx, conn, call.ProtoFiles, symbol)
	if err != nil {
		return nil, err
	}
	desc, err := files.FindDescriptorByName(protoreflect.FullName(symbol))
	if err != nil {
		return nil, fmt.Errorf("method %s not found in schema: %w", fullName, err)
	}
	md, ok := desc.(protoreflect.MethodDescriptor)
	if !ok {
		return nil, fmt.Errorf("symbol %s is not a method", symbol)
	}
	return md, nil
}

// schemaFiles builds a descriptor registry from explicit .proto source files,
// falling back to server reflection when none are given.
func schemaFiles(ctx context.Context, conn *grpc.ClientConn, protoFiles []string, symbol string) (*protoregistry.Files, error) {
	if len(protoFiles) > 0 {
		return parseProtoFiles(protoFiles)
	}
	raws, err := reflectFileDescriptors(ctx, conn, symbol)
	if err != nil {
		return nil, err
	}
	reg := &protoregistry.Files{}
	for _, raw := range raws {
		var fd descriptorpb.FileDescriptorProto
		if err := proto.Unmarshal(raw, &fd); err != nil {
			return nil, fmt.Errorf("decode descriptor: %w", err)
		}
		pf, perr := protodesc.NewFile(&fd, reg)
		if perr != nil {
			return nil, fmt.Errorf("build descriptor %q: %w", fd.GetName(), perr)
		}
		if rerr := reg.RegisterFile(pf); rerr != nil {
			return nil, fmt.Errorf("register descriptor %q: %w", fd.GetName(), rerr)
		}
	}
	return reg, nil
}

// reflectFileDescriptors collects every file descriptor needed to resolve
// symbol using the v1 reflection protocol.
func reflectFileDescriptors(ctx context.Context, conn *grpc.ClientConn, symbol string) ([][]byte, error) {
	client := refv1.NewServerReflectionClient(conn)
	stream, err := client.ServerReflectionInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("reflection not available: %w", err)
	}
	req := &refv1.ServerReflectionRequest{
		MessageRequest: &refv1.ServerReflectionRequest_FileContainingSymbol{FileContainingSymbol: symbol},
	}
	if err := stream.Send(req); err != nil {
		return nil, fmt.Errorf("reflection not available: %w", err)
	}
	resp, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("reflection not available: %w", err)
	}
	fdp := resp.GetFileDescriptorResponse()
	if fdp == nil || len(fdp.GetFileDescriptorProto()) == 0 {
		return nil, fmt.Errorf("reflection returned no descriptor for %s", symbol)
	}
	return fdp.GetFileDescriptorProto(), nil
}

// parseProtoFiles parses `.proto` source files into a registry. Imports
// resolve relative to each file's directory; dependencies are registered
// before their importers.
func parseProtoFiles(paths []string) (*protoregistry.Files, error) {
	reg := &protoregistry.Files{}
	var register func(protoreflect.FileDescriptor) error
	register = func(fd protoreflect.FileDescriptor) error {
		if fd == nil {
			return nil
		}
		imports := fd.Imports()
		for i := range imports.Len() {
			if err := register(imports.Get(i).FileDescriptor); err != nil {
				return err
			}
		}

		if _, err := reg.FindFileByPath(fd.Path()); err == nil {
			return nil // already present (shared import)
		}
		return reg.RegisterFile(fd)
	}
	for _, p := range paths {
		compiler := &protocompile.Compiler{
			Resolver: &protocompile.SourceResolver{ImportPaths: []string{filepath.Dir(p)}},
		}
		fd, err := compiler.Compile(context.Background(), filepath.Base(p))
		if err != nil {
			return nil, fmt.Errorf("read proto file %q: %w", p, err)
		}
		if err := register(fd[0]); err != nil {
			return nil, fmt.Errorf("register %q: %w", p, err)
		}
	}
	return reg, nil
}
