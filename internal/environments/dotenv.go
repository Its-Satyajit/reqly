// Reqly - A local-first, Git-native API development environment.
// Copyright (C) 2026 It's Satyajit
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
