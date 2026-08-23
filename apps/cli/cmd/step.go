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

package cmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/response"
)

// printStep renders one runner step for the streaming step output shared by
// the Pagination and Bulk runners. Responses arrive pre-masked from the
// execution pipeline, so no per-field masking happens here.
func printStep(out, errOut io.Writer, index int, req request.Request, resp *response.Response, err error) {
	if err != nil {
		fmt.Fprintf(errOut, "step %d: error: %v\n", index, err)
		return
	}
	if resp == nil {
		return
	}
	fmt.Fprintf(out, "step %d: %d %s (%s) %s\n",
		index, resp.StatusCode, resp.StatusText,
		resp.Duration.Round(time.Millisecond), urlWithQuery(req))
}

// urlWithQuery reconstructs the display URL including query parameters.
func urlWithQuery(req request.Request) string {
	url := req.URL
	if len(req.Query) == 0 {
		return url
	}
	parts := make([]string, 0, len(req.Query))
	for _, p := range req.Query {
		parts = append(parts, p.Key+"="+p.Value)
	}
	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
	}
	return url + sep + strings.Join(parts, "&")
}
