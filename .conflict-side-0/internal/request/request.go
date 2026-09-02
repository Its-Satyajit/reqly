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

package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Request defines a single API request. The engine layer (transport, protocol
// dispatch, authentication, variable interpolation, scripting) builds on this
// definition.
//
// The transport remains abstract so additional protocols (WebSocket, gRPC,
// SSE, MQTT, ...) can be added without changing the application architecture.
type Request struct {
	ID     string `json:"id,omitempty" yaml:"id,omitempty"`
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`
	Method Method `json:"method,omitempty" yaml:"method,omitempty"`
	URL    string `json:"url,omitempty" yaml:"url,omitempty"`

	Headers Headers     `json:"headers,omitempty" yaml:"headers,omitempty"`
	Query   []Parameter `json:"query,omitempty" yaml:"query,omitempty"`
	Body    string      `json:"body,omitempty" yaml:"body,omitempty"`

	Auth       Auth        `json:"auth,omitempty" yaml:"auth,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty" yaml:"pagination,omitempty"`
	Retry      *Retry      `json:"retry,omitempty" yaml:"retry,omitempty"`
	Timeout    int64       `json:"timeout,omitempty" yaml:"timeout,omitempty"` // milliseconds; 0 means "no explicit timeout"
	// FollowRedirects overrides the default follow behavior when explicitly
	// false: the first response is returned as-is (3xx and all). nil keeps
	// the standard net/http follow behavior.
	FollowRedirects *bool `json:"followRedirects,omitempty" yaml:"followRedirects,omitempty"`

	// Proxy is an optional per-request proxy URL (http://proxy:8080). When set it
	// overrides ProxyFromEnvironment for this send; empty means environment proxy.
	Proxy string `json:"proxy,omitempty" yaml:"proxy,omitempty"`
	// TLS configures per-request TLS (M47). When nil, system roots are used.
	TLS *TLSConfig `json:"tls,omitempty" yaml:"tls,omitempty"`

	// HTTPVersion configures protocol negotiation ("auto", "http1.1", "http2", "http3") (M56).
	HTTPVersion string `json:"httpVersion,omitempty" yaml:"httpVersion,omitempty"`
	// DisableKeepAlives disables HTTP connection pooling per-request (M56).
	DisableKeepAlives bool `json:"disableKeepAlives,omitempty" yaml:"disableKeepAlives,omitempty"`

	// GRPC configures a gRPC call (M43). When set, url carries host:port,
	// headers act as metadata, and body is ignored in favor of Message.
	GRPC *GRPC `json:"grpc,omitempty" yaml:"grpc,omitempty"`
}

// GRPC configures one gRPC call (ADR 0028). Service/Method address the call
// ("/package.Service/Method"); Message holds canonical-JSON; Timeout is a Go
// duration string ("30s"); ProtoFiles are workspace-relative fallback schemas
// for reflection-disabled servers.
type GRPC struct {
	Service    string   `json:"service" yaml:"service"`
	Method     string   `json:"method" yaml:"method"`
	Message    any      `json:"message,omitempty" yaml:"message,omitempty"`
	Timeout    string   `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	ProtoFiles []string `json:"protoFiles,omitempty" yaml:"protoFiles,omitempty"`

	// Transport (M43 T4): plaintext h2c by default; TLS enables TLS against
	// system roots or CAFile, optionally skipping verification.
	TLS           bool   `json:"tls,omitempty" yaml:"tls,omitempty"`
	TLSSkipVerify bool   `json:"tlsSkipVerify,omitempty" yaml:"tlsSkipVerify,omitempty"`
	CAFile        string `json:"caFile,omitempty" yaml:"caFile,omitempty"`
}

// Retry configures automatic re-sending of a failed request. Count is the
// number of retries after the initial attempt; a nil Retry or Count <= 0
// disables retrying entirely.
type Retry struct {
	Count      int    `json:"count,omitempty" yaml:"count,omitempty"`
	DelayMs    int64  `json:"delayMs,omitempty" yaml:"delayMs,omitempty"`
	Strategy   string `json:"strategy,omitempty" yaml:"strategy,omitempty"` // fixed | exponential
	MaxDelayMs int64  `json:"maxDelayMs,omitempty" yaml:"maxDelayMs,omitempty"`
	RetryOn    []int  `json:"retryOn,omitempty" yaml:"retryOn,omitempty"`
}

// Pagination configures iterative fetching for a paginated endpoint.
type Pagination struct {
	Strategy      string `json:"strategy" yaml:"strategy"` // page|offset|cursor|link-header
	PageParam     string `json:"pageParam,omitempty" yaml:"pageParam,omitempty"`
	PageSizeParam string `json:"pageSizeParam,omitempty" yaml:"pageSizeParam,omitempty"`
	OffsetParam   string `json:"offsetParam,omitempty" yaml:"offsetParam,omitempty"`
	LimitParam    string `json:"limitParam,omitempty" yaml:"limitParam,omitempty"`
	CursorParam   string `json:"cursorParam,omitempty" yaml:"cursorParam,omitempty"`
	NextPath      string `json:"nextPath,omitempty" yaml:"nextPath,omitempty"`
	MaxPages      int    `json:"maxPages,omitempty" yaml:"maxPages,omitempty"`
	PageSize      int    `json:"pageSize,omitempty" yaml:"pageSize,omitempty"`
	Limit         int    `json:"limit,omitempty" yaml:"limit,omitempty"`
}

// Method is an HTTP method.
type Method string

const (
	MethodGet     Method = "GET"
	MethodPost    Method = "POST"
	MethodPut     Method = "PUT"
	MethodPatch   Method = "PATCH"
	MethodDelete  Method = "DELETE"
	MethodHead    Method = "HEAD"
	MethodOptions Method = "OPTIONS"
)

// Header is a single request header.
type Header struct {
	Key   string `json:"key" yaml:"key,omitempty"`
	Value string `json:"value" yaml:"value,omitempty"`
}

// Headers is a slice of Header values supporting flexible JSON and YAML unmarshaling (sequence or mapping).
type Headers []Header

// UnmarshalJSON supports unmarshaling from either a JSON array of Header objects or a JSON object of key-value pairs.
func (h *Headers) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*h = nil
		return nil
	}
	if data[0] == '[' {
		var list []Header
		type rawHeader Header
		var rawList []rawHeader
		if err := json.Unmarshal(data, &rawList); err != nil {
			return err
		}
		for _, item := range rawList {
			list = append(list, Header(item))
		}
		*h = list
		return nil
	}
	if data[0] == '{' {
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		var list []Header
		for k, v := range m {
			list = append(list, Header{Key: k, Value: fmt.Sprintf("%v", v)})
		}
		*h = list
		return nil
	}
	return fmt.Errorf("cannot unmarshal %s into Headers", string(data))
}

// UnmarshalYAML supports unmarshaling from either a YAML sequence of Header objects or a YAML mapping of key-value pairs.
func (h *Headers) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		*h = nil
		return nil
	}
	switch node.Kind {
	case yaml.SequenceNode:
		var list []Header
		type rawHeader Header
		for _, item := range node.Content {
			var hdr rawHeader
			if err := item.Decode(&hdr); err != nil {
				return err
			}
			list = append(list, Header(hdr))
		}
		*h = list
		return nil
	case yaml.MappingNode:
		var list []Header
		for i := 0; i < len(node.Content); i += 2 {
			k := node.Content[i].Value
			var v string
			if err := node.Content[i+1].Decode(&v); err != nil {
				v = node.Content[i+1].Value
			}
			list = append(list, Header{Key: k, Value: v})
		}
		*h = list
		return nil
	default:
		return fmt.Errorf("cannot unmarshal YAML kind %d into Headers", node.Kind)
	}
}

// Parameter is a query or path parameter.
type Parameter struct {
	Key   string `json:"key" yaml:"key,omitempty"`
	Value string `json:"value" yaml:"value,omitempty"`
}

// TLSConfig configures per-request TLS (M47).
type TLSConfig struct {
	InsecureSkipVerify bool   `json:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty"`
	CAFile             string `json:"caFile,omitempty" yaml:"caFile,omitempty"`
}

// Auth describes the authentication configuration attached to a request.
// Implementations live in the auth package and are dispatched by type.
type Auth struct {
	Type   string            `json:"type,omitempty" yaml:"type,omitempty"`
	Config map[string]string `json:"config,omitempty" yaml:"config,omitempty"`
}

// UnmarshalJSON supports flexible body shapes for hand-written request files.
// Body may be a plain string, a typed object {type: json/graphql/binary, data/file/query/variables},
// an ADR-style object {file} or {query,variables}, or a form-data array.
func (r *Request) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	bodyRaw, hasBody := raw["body"]
	if hasBody {
		delete(raw, "body")
	}
	// Decode remaining fields into the receiver via alias to avoid recursion.
	if len(raw) > 0 {
		remaining, _ := json.Marshal(raw)
		type Alias Request
		var a Alias
		if err := json.Unmarshal(remaining, &a); err != nil {
			return err
		}
		*r = Request(a)
	} else {
		// No remaining fields: zero other fields.
		type Alias Request
		var a Alias
		*r = Request(a)
	}
	if !hasBody {
		r.Body = ""
		return nil
	}
	var s string
	if err := json.Unmarshal(bodyRaw, &s); err == nil {
		r.Body = s
		return nil
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(bodyRaw, &obj); err == nil {
		var typeStr string
		if v, ok := obj["type"]; ok {
			_ = json.Unmarshal(v, &typeStr)
			typeStr = strings.ToLower(strings.TrimSpace(typeStr))
		}
		switch typeStr {
		case "json":
			if v, ok := obj["data"]; ok {
				var dataStr string
				if err := json.Unmarshal(v, &dataStr); err == nil {
					r.Body = dataStr
				} else {
					r.Body = string(v)
				}
				return nil
			}
			r.Body = ""
			return nil
		case "graphql":
			var query string
			if v, ok := obj["query"]; ok {
				_ = json.Unmarshal(v, &query)
			}
			var variables json.RawMessage
			if v, ok := obj["variables"]; ok {
				var varsStr string
				if err := json.Unmarshal(v, &varsStr); err == nil {
					var parsed interface{}
					if err := json.Unmarshal([]byte(varsStr), &parsed); err == nil {
						variables = json.RawMessage(varsStr)
					} else {
						variables = json.RawMessage(`"` + varsStr + `"`)
					}
				} else {
					variables = v
				}
			}
			out := map[string]interface{}{"query": query}
			if len(variables) > 0 && string(variables) != "null" {
				var vars interface{}
				if err := json.Unmarshal(variables, &vars); err == nil {
					out["variables"] = vars
				} else {
					out["variables"] = string(variables)
				}
			} else {
				out["variables"] = map[string]interface{}{}
			}
			b, _ := json.Marshal(out)
			r.Body = string(b)
			return nil
		case "binary":
			if v, ok := obj["file"]; ok {
				var fileStr string
				if err := json.Unmarshal(v, &fileStr); err == nil {
					r.Body = fileStr
					return nil
				}
			}
			if v, ok := obj["data"]; ok {
				var dataStr string
				if err := json.Unmarshal(v, &dataStr); err == nil {
					r.Body = dataStr
				} else {
					r.Body = string(v)
				}
				return nil
			}
			if v, ok := obj["file"]; ok {
				r.Body = string(v)
				return nil
			}
			r.Body = ""
			return nil
		case "":
			if v, ok := obj["file"]; ok {
				var fileStr string
				if err := json.Unmarshal(v, &fileStr); err == nil {
					r.Body = fileStr
					return nil
				}
				r.Body = string(v)
				return nil
			}
			if _, hasQuery := obj["query"]; hasQuery {
				var query string
				_ = json.Unmarshal(obj["query"], &query)
				var variables json.RawMessage
				if v, ok := obj["variables"]; ok {
					variables = v
					var varsStr string
					if err := json.Unmarshal(v, &varsStr); err == nil {
						var parsed interface{}
						if err := json.Unmarshal([]byte(varsStr), &parsed); err == nil {
							variables = json.RawMessage(varsStr)
						}
					}
				}
				out := map[string]interface{}{"query": query}
				if len(variables) > 0 && string(variables) != "null" {
					var vars interface{}
					if err := json.Unmarshal(variables, &vars); err == nil {
						out["variables"] = vars
					} else {
						out["variables"] = string(variables)
					}
				} else {
					out["variables"] = map[string]interface{}{}
				}
				b, _ := json.Marshal(out)
				r.Body = string(b)
				return nil
			}
			r.Body = string(bodyRaw)
			return nil
		default:
			r.Body = string(bodyRaw)
			return nil
		}
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(bodyRaw, &arr); err == nil {
		r.Body = string(bodyRaw)
		return nil
	}
	r.Body = string(bodyRaw)
	return nil
}

// UnmarshalYAML supports the same flexible body shapes for YAML files.
func (r *Request) UnmarshalYAML(node *yaml.Node) error {
	var raw map[string]yaml.Node
	if err := node.Decode(&raw); err != nil {
		return err
	}
	bodyNode, hasBody := raw["body"]
	if hasBody {
		delete(raw, "body")
	}
	// Decode remaining fields.
	if len(raw) > 0 {
		// Re-encode remaining map to YAML and decode into alias.
		b, err := yaml.Marshal(raw)
		if err != nil {
			return err
		}
		type Alias Request
		var a Alias
		if err := yaml.Unmarshal(b, &a); err != nil {
			return err
		}
		// Preserve body for later; copy other fields.
		body := r.Body
		*r = Request(a)
		r.Body = body
	} else {
		type Alias Request
		var a Alias
		*r = Request(a)
	}
	if !hasBody {
		r.Body = ""
		return nil
	}
	switch bodyNode.Kind {
	case yaml.ScalarNode:
		var s string
		if err := bodyNode.Decode(&s); err == nil {
			r.Body = s
			return nil
		}
		var raw interface{}
		if err := bodyNode.Decode(&raw); err == nil {
			b, _ := json.Marshal(raw)
			r.Body = string(b)
			return nil
		}
	case yaml.MappingNode:
		var m map[string]interface{}
		if err := bodyNode.Decode(&m); err == nil {
			b, _ := json.Marshal(m)
			var obj map[string]json.RawMessage
			if err := json.Unmarshal(b, &obj); err == nil {
				var typeStr string
				if v, ok := obj["type"]; ok {
					_ = json.Unmarshal(v, &typeStr)
					typeStr = strings.ToLower(strings.TrimSpace(typeStr))
				}
				switch typeStr {
				case "json":
					if v, ok := obj["data"]; ok {
						var dataStr string
						if err := json.Unmarshal(v, &dataStr); err == nil {
							r.Body = dataStr
						} else {
							var dataVal interface{}
							if err := json.Unmarshal(v, &dataVal); err == nil {
								bb, _ := json.Marshal(dataVal)
								r.Body = string(bb)
							} else {
								r.Body = string(v)
							}
						}
						return nil
					}
					r.Body = ""
					return nil
				case "graphql":
					var query string
					if v, ok := obj["query"]; ok {
						_ = json.Unmarshal(v, &query)
					}
					var variables json.RawMessage
					if v, ok := obj["variables"]; ok {
						var varsStr string
						if err := json.Unmarshal(v, &varsStr); err == nil {
							var parsed interface{}
							if err := json.Unmarshal([]byte(varsStr), &parsed); err == nil {
								variables = json.RawMessage(varsStr)
							} else {
								variables = json.RawMessage(`"` + varsStr + `"`)
							}
						} else {
							variables = v
						}
					}
					out := map[string]interface{}{"query": query}
					if len(variables) > 0 && string(variables) != "null" {
						var vars interface{}
						if err := json.Unmarshal(variables, &vars); err == nil {
							out["variables"] = vars
						} else {
							out["variables"] = string(variables)
						}
					} else {
						out["variables"] = map[string]interface{}{}
					}
					bb, _ := json.Marshal(out)
					r.Body = string(bb)
					return nil
				case "binary":
					if v, ok := obj["file"]; ok {
						var fileStr string
						if err := json.Unmarshal(v, &fileStr); err == nil {
							r.Body = fileStr
							return nil
						}
						r.Body = string(v)
						return nil
					}
					if v, ok := obj["data"]; ok {
						var dataStr string
						if err := json.Unmarshal(v, &dataStr); err == nil {
							r.Body = dataStr
						} else {
							var dataVal interface{}
							if err := json.Unmarshal(v, &dataVal); err == nil {
								bb, _ := json.Marshal(dataVal)
								r.Body = string(bb)
							} else {
								r.Body = string(v)
							}
						}
						return nil
					}
					r.Body = ""
					return nil
				case "":
					if v, ok := obj["file"]; ok {
						var fileStr string
						if err := json.Unmarshal(v, &fileStr); err == nil {
							r.Body = fileStr
							return nil
						}
						r.Body = string(v)
						return nil
					}
					if _, hasQuery := obj["query"]; hasQuery {
						var query string
						_ = json.Unmarshal(obj["query"], &query)
						var variables json.RawMessage
						if v, ok := obj["variables"]; ok {
							variables = v
							var varsStr string
							if err := json.Unmarshal(v, &varsStr); err == nil {
								var parsed interface{}
								if err := json.Unmarshal([]byte(varsStr), &parsed); err == nil {
									variables = json.RawMessage(varsStr)
								}
							}
						}
						out := map[string]interface{}{"query": query}
						if len(variables) > 0 && string(variables) != "null" {
							var vars interface{}
							if err := json.Unmarshal(variables, &vars); err == nil {
								out["variables"] = vars
							} else {
								out["variables"] = string(variables)
							}
						} else {
							out["variables"] = map[string]interface{}{}
						}
						bb, _ := json.Marshal(out)
						r.Body = string(bb)
						return nil
					}
					r.Body = string(b)
					return nil
				default:
					r.Body = string(b)
					return nil
				}
			}
			r.Body = string(b)
			return nil
		}
		var s string
		if err := bodyNode.Decode(&s); err == nil {
			r.Body = s
			return nil
		}
	case yaml.SequenceNode:
		var arr []interface{}
		if err := bodyNode.Decode(&arr); err == nil {
			b, _ := json.Marshal(arr)
			r.Body = string(b)
			return nil
		}
	}
	var s string
	if err := bodyNode.Decode(&s); err == nil {
		r.Body = s
		return nil
	}
	return nil
}
