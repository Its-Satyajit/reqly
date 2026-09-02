// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package importer

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"

	"github.com/Its-Satyajit/reqly/internal/request"
)

// ParseFetch evaluates a browser DevTools "Copy as fetch" snippet and converts it into a native request.Request.
func ParseFetch(code string) (*request.Request, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("empty fetch snippet")
	}

	vm := goja.New()

	var capturedURL string
	var capturedMethod string = "GET"
	capturedHeaders := make([]request.Header, 0)
	var capturedBody string

	fetchFn := func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			capturedURL = call.Arguments[0].String()
		}
		if len(call.Arguments) > 1 {
			optsObj := call.Arguments[1].ToObject(vm)
			if optsObj != nil {
				// 1. Method
				if m := optsObj.Get("method"); m != nil && m != goja.Undefined() && m != goja.Null() {
					methodStr := strings.TrimSpace(m.String())
					if methodStr != "" {
						capturedMethod = strings.ToUpper(methodStr)
					}
				}

				// 2. Body
				if b := optsObj.Get("body"); b != nil && b != goja.Undefined() && b != goja.Null() {
					capturedBody = b.String()
				}

				// 3. Headers
				if h := optsObj.Get("headers"); h != nil && h != goja.Undefined() && h != goja.Null() {
					hObj := h.ToObject(vm)
					if hObj != nil {
						for _, k := range hObj.Keys() {
							val := hObj.Get(k)
							if val != nil && val != goja.Undefined() && val != goja.Null() {
								// Filter browser noise headers if desired, or preserve them
								capturedHeaders = append(capturedHeaders, request.Header{
									Key:   k,
									Value: val.String(),
								})
							}
						}
					}
				}
			}
		}
		return goja.Undefined()
	}

	if err := vm.Set("fetch", fetchFn); err != nil {
		return nil, fmt.Errorf("setup fetch runtime: %w", err)
	}

	if _, err := vm.RunString(code); err != nil {
		return nil, fmt.Errorf("invalid fetch expression: %w", err)
	}

	if capturedURL == "" {
		return nil, fmt.Errorf("no URL found in fetch call")
	}

	return &request.Request{
		Method:  request.Method(capturedMethod),
		URL:     capturedURL,
		Headers: capturedHeaders,
		Body:    capturedBody,
	}, nil
}
