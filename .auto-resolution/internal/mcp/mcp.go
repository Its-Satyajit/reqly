package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// Request is JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// Response is JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

// Error is JSON-RPC error.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Tool is MCP tool definition.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

var tools = []Tool{
	{Name: "list_requests", Description: "List requests", InputSchema: map[string]any{"type": "object", "properties": map[string]any{}}},
	{Name: "search_requests", Description: "Search requests", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"query": map[string]any{"type": "string"}}, "required": []string{"query"}}},
	{Name: "get_request", Description: "Get request", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}}, "required": []string{"id"}}},
	{Name: "run_request", Description: "Run request", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}, "env": map[string]any{"type": "string"}}}},
}

// Serve handles JSON-RPC over r/w line-delimited.
func Serve(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	enc := json.NewEncoder(w)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			_ = enc.Encode(Response{JSONRPC: "2.0", Error: &Error{Code: -32700, Message: "parse error"}})
			continue
		}
		res := handle(req)
		if err := enc.Encode(res); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func handle(req Request) Response {
	switch req.Method {
	case "initialize":
		return Response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]any{}}}}
	case "tools/list":
		return Response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"tools": tools}}
	case "tools/call":
		var params struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return Response{JSONRPC: "2.0", ID: req.ID, Error: &Error{Code: -32602, Message: "invalid params"}}
		}
		for _, t := range tools {
			if t.Name == params.Name {
				return Response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"content": []any{map[string]any{"type": "text", "text": fmt.Sprintf("tool %s called with %v", params.Name, params.Arguments)}}}}
			}
		}
		return Response{JSONRPC: "2.0", ID: req.ID, Error: &Error{Code: -32601, Message: "unknown tool"}}
	default:
		return Response{JSONRPC: "2.0", ID: req.ID, Error: &Error{Code: -32601, Message: "method not found"}}
	}
}
