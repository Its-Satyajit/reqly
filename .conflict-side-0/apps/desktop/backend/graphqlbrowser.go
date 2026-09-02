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
	"time"

	"github.com/Its-Satyajit/reqly/internal/graphql"
)

// GraphqlIntrospect runs the standard introspection query against an endpoint
// and returns the parsed schema — identical fidelity to `reqly graphql
// introspect`. Headers carry auth; timeout is in seconds (0 = none).
func (s *AppService) GraphqlIntrospect(endpoint string, headers []RealtimeHeader, timeoutSec int) (*graphql.Schema, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	opts := graphql.IntrospectOptions{
		Headers: headerPairs(headers),
	}
	if timeoutSec > 0 {
		opts.Timeout = time.Duration(timeoutSec) * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	schema, _, err := graphql.Introspect(ctx, endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("introspection failed: %w", err)
	}
	return schema, nil
}

func headerPairs(headers []RealtimeHeader) [][2]string {
	out := make([][2]string, 0, len(headers))
	for _, h := range headers {
		if h.Key != "" {
			out = append(out, [2]string{h.Key, h.Value})
		}
	}
	return out
}
