package scripting

import "testing"

func TestRuntimeIsLazy(t *testing.T) {
	r := NewRuntime()
	if r.vm != nil {
		t.Fatal("expected Goja runtime to remain uninitialized until first use")
	}
}

func TestRunScriptReturnsValue(t *testing.T) {
	r := NewRuntime()
	value, err := r.RunScript("1 + 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := value.ToInteger(); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}

func TestRunScriptReportsSyntaxError(t *testing.T) {
	r := NewRuntime()
	if _, err := r.RunScript("const =="); err == nil {
		t.Fatal("expected syntax error")
	}
}

func TestRuntimeRetainsStateAcrossRuns(t *testing.T) {
	r := NewRuntime()
	if _, err := r.RunScript("var counter = 0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := r.RunScript("counter = counter + 1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	value, err := r.RunScript("counter")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := value.ToInteger(); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}
