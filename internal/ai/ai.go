// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/response"
)

// ExplainResponse returns a fixed-template explanation for a response.
func ExplainResponse(resp *response.Response) string {
	if resp == nil {
		return "no response"
	}
	timings := ""
	if resp.Timings != nil {
		timings = fmt.Sprintf(" dns=%dms connect=%dms tls=%dms", resp.Timings.DNS, resp.Timings.Connect, resp.Timings.TLS)
	}
	return fmt.Sprintf("response %d %s in %dms (%dB) proto %s%s", resp.StatusCode, resp.StatusText, resp.Duration.Milliseconds(), resp.Size, resp.Proto, timings)
}

// GenerateTests analyzes a response and generates Goja test assertions in JavaScript.
func GenerateTests(resp *response.Response) string {
	if resp == nil {
		return "// No response available to generate tests from.\n"
	}

	var sb strings.Builder
	sb.WriteString("// Auto-generated test assertions based on response metadata & schema\n\n")

	// 1. Status Code assertion
	sb.WriteString(fmt.Sprintf("reqly.test(\"Status code is %d\", function() {\n", resp.StatusCode))
	sb.WriteString(fmt.Sprintf("    return reqly.response.status === %d;\n", resp.StatusCode))
	sb.WriteString("});\n\n")

	// 2. Response Time assertion
	threshMs := resp.Duration.Milliseconds() * 2
	if threshMs < 500 {
		threshMs = 500
	}
	sb.WriteString(fmt.Sprintf("reqly.test(\"Response time is under %dms\", function() {\n", threshMs))
	sb.WriteString(fmt.Sprintf("    return reqly.response.duration < %d;\n", threshMs))
	sb.WriteString("});\n\n")

	// 3. Content-Type assertion
	ct := ""
	for k, vals := range resp.Headers {
		if strings.EqualFold(k, "Content-Type") && len(vals) > 0 {
			ct = vals[0]
			break
		}
	}
	if strings.Contains(ct, "application/json") || json.Valid(resp.Body) {
		sb.WriteString("reqly.test(\"Content-Type is JSON\", function() {\n")
		sb.WriteString("    const ct = reqly.response.headers[\"content-type\"] || reqly.response.headers[\"Content-Type\"];\n")
		sb.WriteString("    return ct && ct.includes(\"application/json\");\n")
		sb.WriteString("});\n\n")
	}

	// 4. JSON body property verification
	if len(resp.Body) > 0 && json.Valid(resp.Body) {
		var parsed any
		if err := json.Unmarshal(resp.Body, &parsed); err == nil {
			if obj, ok := parsed.(map[string]any); ok && len(obj) > 0 {
				keys := make([]string, 0, len(obj))
				for k := range obj {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				sb.WriteString("reqly.test(\"Response contains expected properties\", function() {\n")
				sb.WriteString("    const body = reqly.response.json();\n")
				sb.WriteString("    return body && typeof body === 'object' &&\n")
				for i, k := range keys {
					suffix := " &&"
					if i == len(keys)-1 {
						suffix = ";"
					}
					sb.WriteString(fmt.Sprintf("        Object.prototype.hasOwnProperty.call(body, %q)%s\n", k, suffix))
				}
				sb.WriteString("});\n\n")

				for _, k := range keys {
					sb.WriteString(fmt.Sprintf("reqly.test(\"Response property %q is defined\", function() {\n", k))
					sb.WriteString("    const body = reqly.response.json();\n")
					sb.WriteString(fmt.Sprintf("    return body && body[%q] !== undefined;\n", k))
					sb.WriteString("});\n\n")
				}
			}
		}
	}

	return strings.TrimSpace(sb.String())
}

// GenerateDocs synthesizes Markdown documentation for an endpoint.
func GenerateDocs(req *request.Request, resp *response.Response) string {
	var sb strings.Builder

	method := "GET"
	urlStr := "http://localhost"
	if req != nil {
		if req.Method != "" {
			method = strings.ToUpper(string(req.Method))
		}
		if req.URL != "" {
			urlStr = req.URL
		}
	}

	sb.WriteString(fmt.Sprintf("# `%s` %s\n\n", method, urlStr))

	if req != nil && len(req.Headers) > 0 {
		sb.WriteString("## Request Headers\n\n")
		sb.WriteString("| Header | Value |\n|---|---|\n")
		for _, h := range req.Headers {
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` |\n", h.Key, h.Value))
		}
		sb.WriteString("\n")
	}

	if req != nil && len(req.Body) > 0 {
		sb.WriteString("## Request Body\n\n")
		bodyBytes := []byte(req.Body)
		if json.Valid(bodyBytes) {
			var pretty json.RawMessage = bodyBytes
			formatted, err := json.MarshalIndent(pretty, "", "  ")
			if err == nil {
				sb.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", string(formatted)))
			} else {
				sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", req.Body))
			}
		} else {
			sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", req.Body))
		}
	}

	if resp != nil {
		sb.WriteString(fmt.Sprintf("## Response (`%d %s` - %dms)\n\n", resp.StatusCode, resp.StatusText, resp.Duration.Milliseconds()))
		if len(resp.Body) > 0 {
			if json.Valid(resp.Body) {
				var pretty json.RawMessage = resp.Body
				formatted, err := json.MarshalIndent(pretty, "", "  ")
				if err == nil {
					sb.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", string(formatted)))
				} else {
					sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", string(resp.Body)))
				}
			} else {
				sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", string(resp.Body)))
			}
		}
	}

	return strings.TrimSpace(sb.String())
}

// Diagnose analyzes error conditions or abnormal HTTP response status codes.
func Diagnose(resp *response.Response, err error) string {
	if err != nil {
		errStr := err.Error()
		switch {
		case strings.Contains(errStr, "connection refused"):
			return "### 🔴 Connection Refused\n\n**Cause:** The target server is not listening on the specified host/port.\n**Remediation:** Check that your local backend server is running and verify the port."
		case strings.Contains(errStr, "tls") || strings.Contains(errStr, "certificate"):
			return "### 🔴 TLS Handshake / Certificate Failure\n\n**Cause:** The SSL/TLS certificate could not be verified by the system certificate authority.\n**Remediation:** Verify custom CA configuration or enable insecure TLS mode if testing local self-signed certificates."
		case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded"):
			return "### 🔴 Network Timeout\n\n**Cause:** The request took longer than the configured timeout threshold.\n**Remediation:** Verify network connectivity, gateway firewalls, or increase timeout limits."
		default:
			return fmt.Sprintf("### 🔴 Network Error\n\n**Details:** `%s`\n**Remediation:** Inspect network route and proxy configuration.", errStr)
		}
	}

	if resp == nil {
		return "No response or error recorded to diagnose."
	}

	switch resp.StatusCode {
	case 200, 201, 204:
		return fmt.Sprintf("### 🟢 Success (%d %s)\n\nRequest completed normally in %dms.", resp.StatusCode, resp.StatusText, resp.Duration.Milliseconds())
	case 400:
		return "### ⚠️ Bad Request (400)\n\n**Cause:** The server could not understand the request due to malformed syntax or payload validation failure.\n**Remediation:** Check required JSON fields, query parameters, and Content-Type headers."
	case 401:
		return "### ⚠️ Unauthorized (401)\n\n**Cause:** Missing, invalid, or expired authentication credentials.\n**Remediation:** Provide a valid `Authorization` header (`Bearer <token>` / `Basic`) or run `reqly auth login` to refresh tokens."
	case 403:
		return "### ⚠️ Forbidden (403)\n\n**Cause:** The authenticated identity does not have sufficient permissions for this resource.\n**Remediation:** Verify role permissions, API key scopes, or organization tenant boundaries."
	case 404:
		return "### ⚠️ Not Found (404)\n\n**Cause:** The endpoint path or resource ID does not exist on the server.\n**Remediation:** Check route URL path, path variables, and base URL prefixes."
	case 429:
		return "### ⚠️ Rate Limited (429)\n\n**Cause:** Request volume exceeded API rate limits or quota.\n**Remediation:** Check `Retry-After` header and configure exponential backoff / retry policies (`request.retry`)."
	case 500:
		return "### 🚨 Internal Server Error (500)\n\n**Cause:** The server encountered an unhandled exception or crash while processing the request.\n**Remediation:** Inspect server logs and stack traces."
	case 502:
		return "### 🚨 Bad Gateway (502)\n\n**Cause:** Upstream server or reverse proxy received an invalid response from the backend service.\n**Remediation:** Check backend server availability behind the reverse proxy / ingress controller."
	case 503:
		return "### 🚨 Service Unavailable (503)\n\n**Cause:** The server is temporarily overloaded or down for maintenance.\n**Remediation:** Retry after delay or check server health status."
	case 504:
		return "### 🚨 Gateway Timeout (504)\n\n**Cause:** Upstream gateway or load balancer timed out waiting for backend response.\n**Remediation:** Check database query latency and backend processing time."
	default:
		return fmt.Sprintf("### ℹ️ Response Status %d %s\n\nDuration: %dms, Size: %dB", resp.StatusCode, resp.StatusText, resp.Duration.Milliseconds(), resp.Size)
	}
}
