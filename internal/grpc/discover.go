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

// Package grpc provides gRPC client support: proto loading, server reflection,
// service discovery, and unary/streaming requests (M43). The package is pure:
// it speaks plaintext h2c today and owns no pipeline concerns — callers
// (core services, CLI, desktop bridge) handle interpolation, masking, and
// history per Send Fidelity.
package grpc

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	refv1 "google.golang.org/grpc/reflection/grpc_reflection_v1"
)

// Method describes one method of a discovered service.
type Method struct {
	// Name is the method name without the service prefix ("Echo").
	Name string `json:"name"`
	// FullName is "/package.Service/Name", the address a call targets.
	FullName        string `json:"fullName"`
	InputType       string `json:"inputType"`
	OutputType      string `json:"outputType"`
	ServerStreaming bool   `json:"serverStreaming"`
}

// Service is one discovered gRPC service.
type Service struct {
	Name    string   `json:"name"` // fully-qualified "package.Service"
	Methods []Method `json:"methods"`
}

// dial opens a plaintext client connection (h2c); TLS options arrive with M43
// ticket 04 as additional Options fields.
func dial(ctx context.Context, target string) (*grpc.ClientConn, error) {
	return grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// Discover lists the target's services and methods over server reflection.
// A server without reflection fails with an error naming it.
func Discover(ctx context.Context, target string) ([]Service, error) {
	conn, err := dial(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", target, err)
	}
	defer conn.Close()

	client := refv1.NewServerReflectionClient(conn)
	stream, err := client.ServerReflectionInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("reflection not available on %s: %w", target, err)
	}
	if err := stream.Send(&refv1.ServerReflectionRequest{MessageRequest: &refv1.ServerReflectionRequest_ListServices{}}); err != nil {
		return nil, fmt.Errorf("reflection request failed: %w", err)
	}
	resp, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("reflection not available on %s: %w", target, err)
	}
	list := resp.GetListServicesResponse()
	if list == nil {
		return nil, fmt.Errorf("reflection on %s returned no service list", target)
	}

	services := make([]Service, 0, len(list.GetService()))
	for _, svc := range list.GetService() {
		s, err := discoverService(ctx, stream, svc.GetName())
		if err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	sortServices(services)
	return services, nil
}

// discoverService resolves one service symbol to its full descriptor via the
// same reflection stream (FileContainingSymbol).
func discoverService(ctx context.Context, stream refv1.ServerReflection_ServerReflectionInfoClient, name string) (Service, error) {
	fdResp, err := fileContainingSymbol(stream, name)
	if err != nil {
		return Service{}, fmt.Errorf("resolve %s: %w", name, err)
	}
	descriptor, err := buildDescriptor(fdResp)
	if err != nil {
		return Service{}, fmt.Errorf("resolve %s: %w", name, err)
	}
	sd, err := descriptor.FindDescriptorByName(protoreflect.FullName(name))
	if err != nil {
		return Service{}, fmt.Errorf("descriptor for %s: %w", name, err)
	}
	serviceDesc, ok := sd.(protoreflect.ServiceDescriptor)
	if !ok {
		return Service{}, fmt.Errorf("symbol %s is not a service", name)
	}
	svc := Service{Name: name, Methods: make([]Method, 0, serviceDesc.Methods().Len())}
	for i := range serviceDesc.Methods().Len() {
		m := serviceDesc.Methods().Get(i)
		full := "/" + name + "/" + string(m.Name())
		svc.Methods = append(svc.Methods, Method{
			Name:            string(m.Name()),
			FullName:        full,
			InputType:       string(m.Input().FullName()),
			OutputType:      string(m.Output().FullName()),
			ServerStreaming: m.IsStreamingServer(),
		})
	}
	return svc, nil
}

func fileContainingSymbol(stream refv1.ServerReflection_ServerReflectionInfoClient, symbol string) (*refv1.ServerReflectionResponse, error) {
	req := &refv1.ServerReflectionRequest{
		MessageRequest: &refv1.ServerReflectionRequest_FileContainingSymbol{FileContainingSymbol: symbol},
	}
	if err := stream.Send(req); err != nil {
		return nil, err
	}
	return stream.Recv()
}

// buildDescriptor converts a reflection FileDescriptorResponse into a
// resolvable protoregistry.Files set.
func buildDescriptor(resp *refv1.ServerReflectionResponse) (*protodescFiles, error) {
	fdp := resp.GetFileDescriptorResponse()
	if fdp == nil {
		return nil, fmt.Errorf("no descriptor returned")
	}
	files := newFiles()
	for _, raw := range fdp.GetFileDescriptorProto() {
		var fd descriptorpb.FileDescriptorProto
		if err := proto.Unmarshal(raw, &fd); err != nil {
			return nil, err
		}
		if err := files.add(&fd); err != nil {
			return nil, err
		}
	}
	return files, nil
}

func sortServices(services []Service) {
	for i := 1; i < len(services); i++ {
		for j := i; j > 0 && strings.Compare(services[j].Name, services[j-1].Name) < 0; j-- {
			services[j], services[j-1] = services[j-1], services[j]
		}
	}
}

// protodescFiles is a mutable registry accumulating the transitive file
// descriptors reflection returns, so cross-file type references resolve.
type protodescFiles struct {
	reg *protoregistry.Files
}

func newFiles() *protodescFiles {
	return &protodescFiles{reg: &protoregistry.Files{}}
}

func (f *protodescFiles) add(fd *descriptorpb.FileDescriptorProto) error {
	pf, err := protodesc.NewFile(fd, f.reg)
	if err != nil {
		return err
	}
	return f.reg.RegisterFile(pf)
}

func (f *protodescFiles) FindDescriptorByName(name protoreflect.FullName) (protoreflect.Descriptor, error) {
	return f.reg.FindDescriptorByName(name)
}
