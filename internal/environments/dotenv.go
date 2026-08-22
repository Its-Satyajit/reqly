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

package environments

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ParseDotEnv parses dotenv file contents (KEY=VALUE lines, # comments, basic
// quoting, optional export prefix) into a map. Lines that are not valid
// KEY=VALUE pairs are ignored, matching the lenient convention of popular
// dotenv implementations.
func ParseDotEnv(input string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, line := range strings.Split(input, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		key, value, ok := parseDotEnvLine(trimmed)
		if !ok {
			continue
		}
		vars[key] = value
	}
	return vars, nil
}

// parseDotEnvLine parses a single KEY=VALUE line, handling an optional export
// prefix, an optional inline comment, and single/double quoting.
func parseDotEnvLine(line string) (key, value string, ok bool) {
	line = strings.TrimPrefix(line, "export ")
	line = strings.TrimPrefix(line, "export\t")
	eq := strings.Index(line, "=")
	if eq <= 0 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:eq])
	if key == "" {
		return "", "", false
	}
	raw := strings.TrimSpace(line[eq+1:])
	if strings.HasPrefix(raw, `"`) {
		if !strings.HasSuffix(raw, `"`) {
			return "", "", false
		}
		value = strings.ReplaceAll(raw[1:len(raw)-1], `\"`, `"`)
		return key, value, true
	}
	if strings.HasPrefix(raw, `'`) {
		if !strings.HasSuffix(raw, `'`) {
			return "", "", false
		}
		return key, raw[1 : len(raw)-1], true
	}
	// Unquoted: strip an inline comment only when preceded by whitespace.
	if idx := strings.Index(raw, " #"); idx >= 0 {
		raw = raw[:idx]
	}
	return key, strings.TrimSpace(raw), true
}

// LoadDotEnv returns the merged process-env scope for dir: the OS environment
// plus the nearest .env file discovered by walking up from dir. The OS
// environment wins when both define a key, so CI can override local .env
// values.
func LoadDotEnv(dir string) (map[string]string, error) {
	merged := make(map[string]string)
	for _, kv := range os.Environ() {
		key, value, found := strings.Cut(kv, "=")
		if found {
			merged[key] = value
		}
	}
	dotenvVars, err := DotEnvValues(dir)
	if err != nil {
		return nil, err
	}
	for key, value := range dotenvVars {
		if _, exists := merged[key]; !exists {
			merged[key] = value
		}
	}
	return merged, nil
}

// DotEnvValues returns the values defined in the nearest .env file discovered
// from dir, without merging the OS environment. The OS environment is never
// included, so callers can tell which values actually came from the file (and
// treat them as sensitive for masking).
func DotEnvValues(dir string) (map[string]string, error) {
	path, found, err := discoverDotEnv(dir)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read .env file %q: %w", path, err)
	}
	return ParseDotEnv(string(data))
}

// discoverDotEnv walks up from dir to the nearest .env file, returning its
// path and whether one was found.
func discoverDotEnv(dir string) (string, bool, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", false, fmt.Errorf("resolve %q: %w", dir, err)
	}
	for {
		candidate := filepath.Join(abs, ".env")
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, true, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", false, nil
		}
		abs = parent
	}
}
