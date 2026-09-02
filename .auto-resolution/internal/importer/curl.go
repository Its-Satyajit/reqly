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

// Package importer converts external API artifacts (cURL, OpenAPI, Postman,
// Insomnia, Swagger, HAR, ...) into native Reqly structures.
package importer

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/request"
)

// curlFlags whose value is consumed by the parser (value = next token) or
// boolean (no value).
type flagSpec struct {
	hasValue bool
}

var curlFlags = map[string]flagSpec{
	"-X":                {true}, // method
	"--request":         {true},
	"-H":                {true}, // header
	"--header":          {true},
	"-d":                {true}, // data body
	"--data":            {true},
	"--data-raw":        {true},
	"--data-binary":     {true},
	"--data-urlencode":  {true},
	"-u":                {true}, // basic auth
	"--user":            {true},
	"--url":             {true},
	"-A":                {true}, // user-agent
	"--user-agent":      {true},
	"-b":                {true}, // cookie (as header)
	"--cookie":          {true},
	"--connect-timeout": {true}, // ignored
	"--max-time":        {true}, // ignored
	"--max-redirs":      {true}, // ignored
	"-o":                {true}, // output file, ignored
	"--output":          {true},
	"-G":                {false}, // data becomes query params
	"--get":             {false},
	"-L":                {false}, // follow redirects, ignored
	"--location":        {false},
	"-s":                {false}, // silent, ignored
	"--silent":          {false},
	"-v":                {false}, // verbose, ignored
	"--verbose":         {false},
	"-k":                {false}, // insecure, ignored
	"--insecure":        {false},
	"--compressed":      {false},
	"--fail":            {false},
	"-i":                {false}, // include headers, ignored
	"--include":         {false},
	"-I":                {false}, // HEAD
	"--head":            {false},
}

// ParseCurl converts a single curl command line into a request.Request. The
// command may be a raw string copied from a terminal (quotes preserved) or a
// shell-tokenized slice joined with spaces.
//
// Supported: method, headers, JSON/raw/data bodies, basic auth, user-agent,
// cookies, GET-style data as query params. Unsupported curl features (forms,
// multipart, certs, proxies, uploads) are reported as errors.
func ParseCurl(command string) (*request.Request, error) {
	tokens, err := tokenizeCurl(command)
	if err != nil {
		return nil, err
	}

	req := &request.Request{
		Method: request.MethodGet,
		Auth:   request.Auth{},
	}
	var dataParts []string
	var dataInQuery bool
	var methodFlag bool
	var rawURL string

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == "curl" {
			continue
		}
		spec, ok := curlFlags[tok]
		if !ok {
			if rawURL == "" && !strings.HasPrefix(tok, "-") {
				rawURL = tok
				continue
			}
			return nil, fmt.Errorf("unsupported curl token %q", tok)
		}

		if !spec.hasValue {
			switch tok {
			case "-G", "--get":
				dataInQuery = true
			case "-I", "--head":
				req.Method = request.MethodHead
				methodFlag = true
			}
			continue
		}

		if i+1 >= len(tokens) {
			return nil, fmt.Errorf("curl flag %q missing a value", tok)
		}
		i++
		value := tokens[i]

		switch tok {
		case "-X", "--request":
			req.Method = request.Method(strings.ToUpper(value))
			methodFlag = true
		case "-H", "--header":
			key, val, ok := strings.Cut(value, ":")
			if !ok {
				return nil, fmt.Errorf("invalid header %q: expected 'Key: Value'", value)
			}
			req.Headers = append(req.Headers, request.Header{Key: strings.TrimSpace(key), Value: strings.TrimSpace(val)})
		case "-d", "--data", "--data-raw", "--data-binary", "--data-urlencode":
			dataParts = append(dataParts, value)
		case "-u", "--user":
			username, password, _ := strings.Cut(value, ":")
			req.Auth.Type = "basic"
			req.Auth.Config = map[string]string{"username": username, "password": password}
		case "--url":
			rawURL = value
		case "-A", "--user-agent":
			req.Headers = append(req.Headers, request.Header{Key: "User-Agent", Value: value})
		case "-b", "--cookie":
			req.Headers = append(req.Headers, request.Header{Key: "Cookie", Value: value})
		case "--connect-timeout", "--max-time", "--max-redirs", "-o", "--output":
			// accepted and ignored
		}
	}

	if rawURL == "" {
		return nil, fmt.Errorf("curl command has no URL")
	}
	req.URL = rawURL

	if len(dataParts) > 0 {
		body := strings.Join(dataParts, "&")
		if dataInQuery {
			if err := appendQuery(req, body); err != nil {
				return nil, err
			}
		} else {
			req.Body = body
			if !methodFlag {
				req.Method = request.MethodPost
			}
		}
	}

	req.Query = splitQueryParams(req.URL)
	return req, nil
}

// appendQuery adds data (a=1&b=2) as query parameters to the request URL.
func appendQuery(req *request.Request, data string) error {
	u, err := url.Parse(req.URL)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", req.URL, err)
	}
	q := u.Query()
	for _, pair := range strings.Split(data, "&") {
		if pair == "" {
			continue
		}
		key, value, _ := strings.Cut(pair, "=")
		decoded, err := url.QueryUnescape(value)
		if err != nil {
			decoded = value
		}
		q.Set(key, decoded)
	}
	u.RawQuery = q.Encode()
	req.URL = u.String()
	return nil
}

// splitQueryParams extracts query parameters from the URL into req.Query.
func splitQueryParams(rawURL string) []request.Parameter {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}
	var params []request.Parameter
	for key, values := range u.Query() {
		for _, value := range values {
			params = append(params, request.Parameter{Key: key, Value: value})
		}
	}
	return params
}

// tokenizeCurl splits a curl command line into tokens, honoring single and
// double quotes and backslash escapes (shell-style).
func tokenizeCurl(line string) ([]string, error) {
	var tokens []string
	var cur strings.Builder
	inSingle, inDouble := false, false

	flush := func() {
		if cur.Len() > 0 {
			tokens = append(tokens, cur.String())
			cur.Reset()
		}
	}

	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch {
		case inSingle:
			if r == '\'' {
				inSingle = false
			} else {
				cur.WriteRune(r)
			}
		case inDouble:
			if r == '"' {
				inDouble = false
			} else if r == '\\' && i+1 < len(runes) {
				i++
				cur.WriteRune(runes[i])
			} else {
				cur.WriteRune(r)
			}
		case r == '\'':
			inSingle = true
		case r == '"':
			inDouble = true
		case r == '\\' && i+1 < len(runes):
			i++
			cur.WriteRune(runes[i])
		case r == ' ' || r == '\t':
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	if inSingle || inDouble {
		return nil, fmt.Errorf("unterminated quote in curl command")
	}
	flush()
	return tokens, nil
}
