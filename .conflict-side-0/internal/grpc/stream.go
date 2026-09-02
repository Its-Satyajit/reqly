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
	"io"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// StreamEvent is one streamed response message.
type StreamEvent struct {
	// Seq counts messages from 1 in delivery order.
	Seq         int    `json:"seq"`
	MessageJSON []byte `json:"messageJson"`
}

// InvokeStream performs one server-streaming call. onMessage runs once per
// message in delivery order before InvokeStream returns; returning an error
// from onMessage stops consumption and reports that error. The final Result
// carries the terminal status (OK after a clean stream end; non-OK statuses
// and mid-stream errors surface as result fields / transport error exactly
// like unary). Cancellation is the caller's ctx — cancelling tears down the
// RPC cleanly and surfaces as a context/deadline failure.
func InvokeStream(ctx context.Context, call Call, messageJSON []byte, opts InvokeOptions, onMessage func(StreamEvent) error) (*Result, error) {
	conn, err := dial(ctx, call.Target, opts.Transport)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", call.Target, err)
	}
	defer conn.Close()

	methodDesc, err := resolveMethod(ctx, conn, call)
	if err != nil {
		return nil, err
	}
	if !methodDesc.IsStreamingServer() || methodDesc.IsStreamingClient() {
		return nil, fmt.Errorf("%s is not server-streaming; use unary invoke", methodDesc.FullName())
	}

	in := dynamicpb.NewMessage(methodDesc.Input())
	if len(messageJSON) > 0 {
		if uerr := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(messageJSON, in); uerr != nil {
			return nil, fmt.Errorf("invalid message JSON for %s: %w", methodDesc.FullName(), uerr)
		}
	}

	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}
	callOpts := []grpc.CallOption{}
	if len(opts.Metadata) > 0 {
		md := metadata.MD{}
		for k, v := range opts.Metadata {
			md.Set(headerKey(k), v)
		}
		callOpts = append(callOpts, grpc.Header(&md))
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	stream, err := conn.NewStream(ctx, &grpc.StreamDesc{
		StreamName:    string(methodDesc.Name()),
		ServerStreams: true,
	}, rpcMethodPath(methodDesc), callOpts...)
	if err != nil {
		return nil, fmt.Errorf("open stream %s: %w", methodDesc.FullName(), err)
	}
	if serr := stream.SendMsg(in); serr != nil {
		return streamFailure(methodDesc.FullName(), serr)
	}
	if serr := stream.CloseSend(); serr != nil {
		return streamFailure(methodDesc.FullName(), serr)
	}

	res := &Result{}
	start := time.Now()
	for {
		out := dynamicpb.NewMessage(methodDesc.Output())
		rerr := stream.RecvMsg(out)
		if rerr == io.EOF {
			break
		}
		res.DurationMS = time.Since(start).Milliseconds()
		if rerr != nil {
			return streamFailure(methodDesc.FullName(), rerr)
		}
		msg, merr := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(out)
		if merr != nil {
			return nil, fmt.Errorf("encode response message: %w", merr)
		}
		res.seq++
		if oerr := onMessage(StreamEvent{Seq: res.seq, MessageJSON: msg}); oerr != nil {
			// Consumer stopped early — report what streamed so far alongside
			// the consumer's sentinel error.
			return res, oerr
		}
	}
	res.DurationMS = time.Since(start).Milliseconds()
	res.OK = true
	return res, nil
}

func rpcMethodPath(md protoreflect.MethodDescriptor) string {
	serviceName := string(md.FullName())[:strings.LastIndex(string(md.FullName()), ".")]
	return "/" + serviceName + "/" + string(md.Name())
}

// statusFromError reports whether err carries a gRPC status.
func statusFromError(err error) (*status.Status, bool) {
	return status.FromError(err)
}

// streamFailure converts a mid-stream RPC error into the standard result /
// wrapped-error duality shared with unary invokes.
func streamFailure(fullName protoreflect.FullName, rpcErr error) (*Result, error) {
	st, ok := statusFromError(rpcErr)
	if !ok {
		return nil, fmt.Errorf("stream %s: %w", fullName, rpcErr)
	}
	res := &Result{
		OK:            false,
		Code:          uint32(st.Code()),
		CodeName:      st.Code().String(),
		StatusMessage: st.Message(),
	}
	for _, d := range st.Details() {
		res.StatusDetails = append(res.StatusDetails, detailToJSON(d))
	}
	return res, nil
}
