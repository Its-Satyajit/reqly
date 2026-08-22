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

import "github.com/dop251/goja"

// Runtime wraps a lazily-initialized Goja JavaScript runtime.
//
// Goja is deliberately loaded on first use so the base API client never pays
// the runtime cost of the JavaScript engine unless scripting is used.
type Runtime struct {
	vm *goja.Runtime
}

// NewRuntime returns a Runtime that initializes Goja on first script run.
func NewRuntime() *Runtime {
	return &Runtime{}
}

// RunScript executes source and returns the resulting value.
func (r *Runtime) RunScript(source string) (goja.Value, error) {
	return r.ensureVM().RunString(source)
}

// ensureVM returns the underlying Goja runtime, initializing it on first use.
func (r *Runtime) ensureVM() *goja.Runtime {
	if r.vm == nil {
		r.vm = goja.New()
	}
	return r.vm
}
