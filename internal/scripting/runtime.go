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
