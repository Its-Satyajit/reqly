// Package exporter — code generation (M24)

package exporter

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/request"
)

// Generate returns a snippet for req in lang (curl, js, python, go). mask is applied to URL/headers/body when non-nil (secrets as [SECRET]).
func Generate(req request.Request, lang string, mask func(string) string) (string, error) {
	applyMask := func(s string) string {
		if mask != nil {
			return mask(s)
		}
		return s
	}
	method := string(req.Method)
	if method == "" {
		method = "GET"
	}
	// build URL with query
	rawURL := req.URL
	if rawURL == "" {
		rawURL = "https://example.com"
	}
	if len(req.Query) > 0 {
		u, err := url.Parse(rawURL)
		if err == nil {
			q := u.Query()
			for _, p := range req.Query {
				q.Set(p.Key, p.Value)
			}
			u.RawQuery = q.Encode()
			rawURL = u.String()
		}
	}
	rawURL = applyMask(rawURL)
	body := applyMask(req.Body)
	// headers map
	var headers []request.Header
	for _, h := range req.Headers {
		headers = append(headers, request.Header{Key: applyMask(h.Key), Value: applyMask(h.Value)})
	}
	// auth handling (basic/bearer/apikey/jwt as header, others TODO)
	var authHeaders []request.Header
	var curlUser string
	switch req.Auth.Type {
	case "basic":
		u := req.Auth.Config["username"]
		p := req.Auth.Config["password"]
		curlUser = applyMask(u + ":" + p)
		authHeaders = append(authHeaders, request.Header{Key: "Authorization", Value: "Basic " + applyMask(u+":"+p)})
	case "bearer":
		tok := req.Auth.Config["token"]
		authHeaders = append(authHeaders, request.Header{Key: "Authorization", Value: "Bearer " + applyMask(tok)})
	case "apikey":
		k := req.Auth.Config["key"]
		v := req.Auth.Config["value"]
		in := req.Auth.Config["in"]
		if in == "query" {
			// append to URL
			u, err := url.Parse(rawURL)
			if err == nil {
				q := u.Query()
				q.Set(k, v)
				u.RawQuery = q.Encode()
				rawURL = u.String()
			}
		} else {
			authHeaders = append(authHeaders, request.Header{Key: k, Value: applyMask(v)})
		}
	case "jwt":
		// jwt as bearer
		tok := req.Auth.Config["token"]
		if tok == "" {
			tok = "[JWT]"
		}
		authHeaders = append(authHeaders, request.Header{Key: "Authorization", Value: "Bearer " + applyMask(tok)})
	}
	headers = append(headers, authHeaders...)

	lang = strings.ToLower(lang)
	switch lang {
	case "curl", "cURL", "Curl":
		lang = "curl"
	}
	switch lang {
	case "curl":
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("curl --request %s '%s'", method, rawURL))
		if curlUser != "" {
			sb.WriteString(fmt.Sprintf(" --user '%s'", curlUser))
		}
		for _, h := range headers {
			sb.WriteString(fmt.Sprintf(" --header '%s: %s'", h.Key, h.Value))
		}
		if body != "" {
			// escape single quotes
			esc := strings.ReplaceAll(body, "'", "'\\''")
			sb.WriteString(fmt.Sprintf(" --data-raw '%s'", esc))
		}
		return sb.String(), nil
	case "js", "javascript", "node":
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("fetch('%s', {\n  method: '%s',\n", rawURL, method))
		if len(headers) > 0 {
			sb.WriteString("  headers: {\n")
			for _, h := range headers {
				sb.WriteString(fmt.Sprintf("    '%s': '%s',\n", h.Key, h.Value))
			}
			sb.WriteString("  },\n")
		}
		if body != "" {
			sb.WriteString(fmt.Sprintf("  body: '%s',\n", strings.ReplaceAll(body, "'", "\\'")))
		}
		sb.WriteString("});")
		return sb.String(), nil
	case "python", "py":
		var sb strings.Builder
		sb.WriteString("import requests\n\n")
		sb.WriteString(fmt.Sprintf("resp = requests.request('%s', '%s'", method, rawURL))
		if len(headers) > 0 {
			sb.WriteString(", headers={")
			for i, h := range headers {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("'%s': '%s'", h.Key, h.Value))
			}
			sb.WriteString("}")
		}
		if body != "" {
			sb.WriteString(fmt.Sprintf(", data='%s'", strings.ReplaceAll(body, "'", "\\'")))
		}
		sb.WriteString(")\n")
		return sb.String(), nil
	case "go", "golang":
		var sb strings.Builder
		sb.WriteString("package main\n\nimport (\n  \"net/http\"\n  \"strings\"\n)\n\n")
		sb.WriteString(fmt.Sprintf("req, _ := http.NewRequestWithContext(ctx, \"%s\", \"%s\", strings.NewReader(\"%s\"))\n", method, rawURL, strings.ReplaceAll(body, "\"", "\\\"")))
		for _, h := range headers {
			sb.WriteString(fmt.Sprintf("req.Header.Set(\"%s\", \"%s\")\n", h.Key, h.Value))
		}
		sb.WriteString("resp, _ := http.DefaultClient.Do(req)\n")
		return sb.String(), nil
	default:
		return "", fmt.Errorf("unknown lang %q (supported: curl, js, python, go)", lang)
	}
}
