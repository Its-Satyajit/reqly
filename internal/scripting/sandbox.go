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

package scripting

import (
	"context"
	"fmt"
	"os"

	"github.com/dop251/goja"

	"github.com/Its-Satyajit/reqly/internal/diffing"
	"github.com/Its-Satyajit/reqly/internal/graphql"
	grpcpkg "github.com/Its-Satyajit/reqly/internal/grpc"
	"github.com/Its-Satyajit/reqly/internal/importer"
	"github.com/Its-Satyajit/reqly/internal/jsonschema"
	"github.com/Its-Satyajit/reqly/internal/jwt"
	"github.com/Its-Satyajit/reqly/internal/mqtt"
	"github.com/Its-Satyajit/reqly/internal/request"
	"github.com/Its-Satyajit/reqly/internal/response"
	"github.com/Its-Satyajit/reqly/internal/socketio"
	"github.com/Its-Satyajit/reqly/internal/validation"
)

// Test is a named test registered by a post-request script via reqly.test().
type Test struct {
	// Name is the test label.
	Name string
	// Fn is the body to evaluate; it returns truthy on pass.
	Fn func() bool
}

// Sandbox is the JavaScript execution context for pre- and post-request
// scripts. It exposes a Postman-like `reqly` global plus console logging.
type Sandbox struct {
	vm *goja.Runtime

	// respView keeps track of the response body for assertions.
	respView *responseView

	request *goja.Object
	// response mirrors the received response for post-request scripts.
	response *goja.Object

	// getVariable / setVariable read and write the shared variable store.
	getVariable func(name string) (string, bool)
	setVariable func(name, value string)

	// tests collects tests registered during post-request scripts.
	tests []Test
	// logs collects console output in order.
	logs []string
}

// SandboxOptions configures a Sandbox.
type SandboxOptions struct {
	// GetVariable resolves a variable by name (nil means an empty lookup).
	GetVariable func(name string) (string, bool)
	// SetVariable stores a variable (nil disables writing).
	SetVariable func(name, value string)
}

// NewSandbox returns a sandbox with the reqly global bound.
func NewSandbox(opts SandboxOptions) *Sandbox {
	vm := goja.New()
	s := &Sandbox{
		vm:          vm,
		getVariable: opts.GetVariable,
		setVariable: opts.SetVariable,
	}
	if s.getVariable == nil {
		s.getVariable = func(string) (string, bool) { return "", false }
	}
	s.bindConsole()
	s.bindReqly()
	return s
}

// BindRequest exposes the outgoing request to the script. Values are read back
// through getter functions so mutations made by the script are observed by the
// caller.
func (s *Sandbox) BindRequest(req *requestView) {
	s.request = s.vm.NewObject()
	s.bindRequestProps(s.request, req)
	s.reqly().Set("request", s.request)
}

// NewRequestView builds a mutable script view from a request.Request.
func NewRequestView(req *request.Request) *requestView {
	return newRequestView(req)
}

// ApplyTo copies the mutated view back onto the request.
func (v *requestView) ApplyTo(req *request.Request) {
	v.applyTo(req)
}

// requestView is a mutable view of a request as seen by scripts.
type requestView struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

// newRequestView adapts a request.Request into a script view.
func newRequestView(req *request.Request) *requestView {
	headers := map[string]string{}
	for _, h := range req.Headers {
		headers[h.Key] = h.Value
	}
	return &requestView{
		Method:  string(req.Method),
		URL:     req.URL,
		Headers: headers,
		Body:    req.Body,
	}
}

// applyTo copies the mutated view back onto the request.
func (v *requestView) applyTo(req *request.Request) {
	if v.Method != "" {
		req.Method = request.Method(v.Method)
	}
	req.URL = v.URL
	req.Body = v.Body
	req.Headers = req.Headers[:0]
	for key, value := range v.Headers {
		req.Headers = append(req.Headers, request.Header{Key: key, Value: value})
	}
}

// BindResponse exposes the received response to the script. Test registration
// (reqly.test) is available in post-request scripts.
func (s *Sandbox) BindResponse(resp *responseView) {
	s.respView = resp
	s.response = s.vm.NewObject()
	if resp != nil {
		s.response.Set("status", resp.Status)
		s.response.Set("statusText", resp.StatusText)
		s.response.Set("body", resp.Body)
		headers := s.vm.NewObject()
		for k, v := range resp.Headers {
			headers.Set(k, v)
		}
		s.response.Set("headers", headers)
	}
	s.reqly().Set("response", s.response)
}

// responseView is a read-only view of a response as seen by scripts.
type responseView struct {
	Status     int
	StatusText string
	Headers    map[string]string
	Body       string
}

// newResponseView adapts a response.Response into a script view.
func newResponseView(resp *response.Response) *responseView {
	headers := map[string]string{}
	for key, values := range resp.Headers {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return &responseView{
		Status:     resp.StatusCode,
		StatusText: resp.StatusText,
		Headers:    headers,
		Body:       resp.Text(),
	}
}

// NewResponseView builds a script view from a response.Response.
func NewResponseView(resp *response.Response) *responseView {
	return newResponseView(resp)
}

// Run evaluates source as a script. It returns the script's final value.
func (s *Sandbox) Run(source string) error {
	if _, err := s.vm.RunString(source); err != nil {
		return fmt.Errorf("script error: %w", err)
	}
	return nil
}

// Tests returns tests registered during script execution.
func (s *Sandbox) Tests() []Test { return s.tests }

// Logs returns console output lines.
func (s *Sandbox) Logs() []string { return s.logs }

// reqly returns the reqly global object, creating it on first use.
func (s *Sandbox) reqly() *goja.Object {
	val := s.vm.Get("reqly")
	if obj, ok := val.(*goja.Object); ok {
		return obj
	}
	obj := s.vm.NewObject()
	s.vm.Set("reqly", obj)
	return obj
}

// bindConsole exposes console.log / console.warn / console.error.
func (s *Sandbox) bindConsole() {
	console := s.vm.NewObject()
	log := func(call goja.FunctionCall) goja.Value {
		args := make([]any, 0, len(call.Arguments))
		for _, a := range call.Arguments {
			args = append(args, a.Export())
		}
		s.logs = append(s.logs, fmt.Sprint(args...))
		return goja.Undefined()
	}
	console.Set("log", log)
	console.Set("info", log)
	console.Set("warn", log)
	console.Set("error", log)
	s.vm.Set("console", console)
}

// bindReqly exposes variable access and test registration on reqly.
func (s *Sandbox) bindReqly() {
	r := s.reqly()

	r.Set("getVariable", func(call goja.FunctionCall) goja.Value {
		name := toString(call, 0)
		value, ok := s.getVariable(name)
		if !ok {
			return goja.Undefined()
		}
		return s.vm.ToValue(value)
	})

	r.Set("hasVariable", func(call goja.FunctionCall) goja.Value {
		name := toString(call, 0)
		_, ok := s.getVariable(name)
		return s.vm.ToValue(ok)
	})

	if s.setVariable != nil {
		r.Set("setVariable", func(call goja.FunctionCall) goja.Value {
			name := toString(call, 0)
			value := toString(call, 1)
			s.setVariable(name, value)
			return goja.Undefined()
		})
	}

	r.Set("assertXSD", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return s.vm.ToValue(false)
		}
		schemaPath := toString(call, 0)
		xsdData, err := os.ReadFile(schemaPath)
		if err != nil {
			return s.vm.ToValue(false)
		}
		var xmlData []byte
		if s.respView != nil {
			xmlData = []byte(s.respView.Body)
		}
		res, err := validation.ValidateXMLAgainstXSD(xmlData, xsdData, validation.ValidationOptions{})
		if err != nil || res == nil || !res.Valid {
			return s.vm.ToValue(false)
		}
		return s.vm.ToValue(true)
	})

	r.Set("introspectGraphQL", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		endpoint := toString(call, 0)
		sch, _, err := graphql.Introspect(context.Background(), endpoint, graphql.IntrospectOptions{})
		if err != nil || sch == nil {
			return goja.Undefined()
		}
		return s.vm.ToValue(sch)
	})

	r.Set("assertJSONSchema", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return s.vm.ToValue(false)
		}
		schemaPath := toString(call, 0)
		schemaData, err := os.ReadFile(schemaPath)
		if err != nil {
			schemaData = []byte(schemaPath)
		}
		sch, err := jsonschema.Compile(schemaData, "")
		if err != nil || sch == nil {
			return s.vm.ToValue(false)
		}
		var instanceData []byte
		if s.respView != nil {
			instanceData = []byte(s.respView.Body)
		}
		violations, err := jsonschema.Validate(sch, instanceData)
		if err != nil || len(violations) > 0 {
			return s.vm.ToValue(false)
		}
		return s.vm.ToValue(true)
	})

	r.Set("verifyJWT", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return s.vm.ToValue(false)
		}
		tokenStr := toString(call, 0)
		keyStr := toString(call, 1)
		alg := ""
		if len(call.Arguments) >= 3 {
			alg = toString(call, 2)
		}
		res, err := jwt.VerifyToken(tokenStr, []byte(keyStr), jwt.VerifyOptions{Algorithm: alg})
		if err != nil || res == nil || !res.Valid {
			return s.vm.ToValue(false)
		}
		return s.vm.ToValue(true)
	})

	r.Set("reflectGRPC", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		endpoint := toString(call, 0)
		services, err := grpcpkg.Discover(context.Background(), endpoint, grpcpkg.Transport{})
		if err != nil || len(services) == 0 {
			return goja.Undefined()
		}
		return s.vm.ToValue(services)
	})

	r.Set("replayHAR", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		harPath := toString(call, 0)
		res, err := importer.ReplayHAR(context.Background(), harPath, importer.HARReplayOptions{})
		if err != nil || res == nil {
			return goja.Undefined()
		}
		return s.vm.ToValue(res)
	})

	mqttObj := s.vm.NewObject()
	mqttObj.Set("publish", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			return s.vm.ToValue(false)
		}
		broker := toString(call, 0)
		topic := toString(call, 1)
		payload := []byte(toString(call, 2))
		err := mqtt.Publish(context.Background(), broker, topic, payload, mqtt.MQTTOptions{})
		return s.vm.ToValue(err == nil)
	})
	r.Set("mqtt", mqttObj)

	sioObj := s.vm.NewObject()
	sioObj.Set("emit", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			return s.vm.ToValue(false)
		}
		urlStr := toString(call, 0)
		event := toString(call, 1)
		data := call.Arguments[2].Export()
		err := socketio.Emit(context.Background(), urlStr, event, data, socketio.Options{})
		return s.vm.ToValue(err == nil)
	})
	r.Set("socketio", sioObj)

	r.Set("generateChangelog", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		oldSpec := toString(call, 0)
		newSpec := toString(call, 1)
		cl, err := diffing.GenerateChangelog([]byte(oldSpec), []byte(newSpec))
		if err != nil {
			return goja.Undefined()
		}
		outObj := s.vm.NewObject()
		outObj.Set("suggested_semver", cl.SuggestedSemver)
		outObj.Set("breaking", cl.Breaking)
		outObj.Set("additions", cl.Additions)
		outObj.Set("info", cl.Info)
		outObj.Set("markdown", cl.ToMarkdown())
		return outObj
	})

	r.Set("test", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(s.vm.ToValue("reqly.test(name, fn) requires two arguments"))
		}
		name := toString(call, 0)
		fn, ok := goja.AssertFunction(call.Arguments[1])
		if !ok {
			panic(s.vm.ToValue("reqly.test(name, fn) requires a function as the second argument"))
		}
		s.tests = append(s.tests, Test{
			Name: name,
			Fn: func() bool {
				result, err := fn(goja.Undefined())
				if err != nil {
					return false
				}
				return result.ToBoolean()
			},
		})
		return goja.Undefined()
	})
}

// bindRequestProps wires the request object's getter/setter pairs.
func (s *Sandbox) bindRequestProps(obj *goja.Object, req *requestView) {
	getter := func(dst *string) goja.Value {
		return s.vm.ToValue(func(goja.FunctionCall) goja.Value { return s.vm.ToValue(*dst) })
	}
	setter := func(dst *string) goja.Value {
		return s.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if call.Argument(0) != goja.Undefined() {
				*dst = toString(call, 0)
			}
			return goja.Undefined()
		})
	}

	_ = obj.DefineAccessorProperty("method", getter(&req.Method), setter(&req.Method), goja.FLAG_TRUE, goja.FLAG_TRUE)
	_ = obj.DefineAccessorProperty("url", getter(&req.URL), setter(&req.URL), goja.FLAG_TRUE, goja.FLAG_TRUE)
	_ = obj.DefineAccessorProperty("body", getter(&req.Body), setter(&req.Body), goja.FLAG_TRUE, goja.FLAG_TRUE)

	headers := s.vm.NewObject()
	for k, v := range req.Headers {
		headers.Set(k, v)
	}
	headers.Set("set", func(call goja.FunctionCall) goja.Value {
		key := toString(call, 0)
		value := toString(call, 1)
		req.Headers[key] = value
		headers.Set(key, value)
		return goja.Undefined()
	})
	obj.Set("headers", headers)
}

// toString returns the nth argument rendered as a string.
func toString(call goja.FunctionCall, n int) string {
	if n >= len(call.Arguments) {
		return ""
	}
	return call.Arguments[n].String()
}
