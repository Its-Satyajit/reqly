package mcp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestToolsList(t *testing.T) {
	var buf bytes.Buffer
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	if err := Serve(strings.NewReader(input+"\n"), &buf); err != nil {
		t.Fatalf("Serve: %v", err)
	}
	var res Response
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if res.Error != nil {
		t.Fatalf("error: %v", res.Error)
	}
}

func TestToolsCallUnknown(t *testing.T) {
	var buf bytes.Buffer
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"unknown","arguments":{}}}`
	if err := Serve(strings.NewReader(input+"\n"), &buf); err != nil {
		t.Fatalf("Serve: %v", err)
	}
	var res Response
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if res.Error == nil {
		t.Fatalf("want error")
	}
}
