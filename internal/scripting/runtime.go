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
