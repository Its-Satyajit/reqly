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

package response

import (
	"encoding/json"
	"strconv"
	"strings"
)

// JSON returns the parsed JSON body, or nil if the body is not valid JSON.
func (r *Response) JSON() any {
	var v any
	if err := json.Unmarshal(r.Body, &v); err != nil {
		return nil
	}
	return v
}

// JSONValue extracts a value from the JSON body at the given JSONPath.
//
// Supported path syntax is a practical subset:
//
//	$.user.name
//	$.users[0].name
//	$['users'][0]['name']
//	users.name         (leading $ and dot are optional)
//
// It returns nil when the body is not JSON, the path is invalid, or the value
// is missing.
func (r *Response) JSONValue(path string) any {
	root := r.JSON()
	if root == nil {
		return nil
	}

	current := root
	steps := parseJSONPath(path)
	for _, step := range steps {
		switch t := current.(type) {
		case map[string]any:
			v, ok := t[step.key]
			if !ok {
				return nil
			}
			current = v
		case []any:
			idx, err := strconv.Atoi(step.key)
			if err != nil || idx < 0 || idx >= len(t) {
				return nil
			}
			current = t[idx]
		default:
			return nil
		}
	}
	return current
}

type pathStep struct {
	key string
}

// parseJSONPath splits a JSONPath string into lookup steps.
func parseJSONPath(path string) []pathStep {
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "$")
	path = strings.TrimPrefix(path, ".")

	var steps []pathStep
	for len(path) > 0 {
		path = strings.TrimLeft(path, ".")

		if strings.HasPrefix(path, "[") {
			end := strings.Index(path, "]")
			if end < 0 {
				break
			}
			inner := strings.Trim(path[1:end], " '")
			if inner != "" {
				steps = append(steps, pathStep{key: inner})
			}
			path = path[end+1:]
			continue
		}

		key := path
		if i := strings.IndexAny(key, ".["); i >= 0 {
			key = key[:i]
		}
		if key != "" {
			steps = append(steps, pathStep{key: key})
		}
		path = strings.TrimPrefix(path, key)
	}
	return steps
}
