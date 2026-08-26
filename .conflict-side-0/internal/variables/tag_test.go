package variables

import (
	"strings"
	"testing"
)

func TestInterpolateDynamicTags(t *testing.T) {
	// Use fixed generator for determinism
	SetTagGeneratorForTest(fixedTagGenerator{})
	defer SetTagGeneratorForTest(nil)

	s := NewSet()
	// uuid
	got, err := s.Interpolate("a {{$uuid}} b")
	if err != nil {
		t.Fatalf("Interpolate: %v", err)
	}
	if got != "a 00000000-0000-4000-a000-000000000000 b" {
		t.Fatalf("uuid: got %q", got)
	}
	// timestamp
	got, _ = s.Interpolate("{{$timestamp}}")
	if got != "1700000000" {
		t.Fatalf("timestamp: got %q", got)
	}
	// isoTimestamp
	got, _ = s.Interpolate("{{$isoTimestamp}}")
	if got != "2023-11-14T22:13:20Z" {
		t.Fatalf("isoTimestamp: got %q", got)
	}
	// randomInt
	got, _ = s.Interpolate("{{$randomInt}}")
	if got != "42" {
		t.Fatalf("randomInt: got %q", got)
	}
	// randomString
	got, _ = s.Interpolate("{{$randomString}}")
	if got != "abcd1234" {
		t.Fatalf("randomString: got %q", got)
	}
	// per occurrence fresh with fixed generator still same (fixed), but with default would be two different; test fixed yields same twice
	got, _ = s.Interpolate("{{$uuid}} {{$uuid}}")
	if got != "00000000-0000-4000-a000-000000000000 00000000-0000-4000-a000-000000000000" {
		t.Fatalf("per occurrence fixed: got %q", got)
	}
	// unknown left literal
	got, _ = s.Interpolate("x {{$unknown}} y")
	if got != "x {{$unknown}} y" {
		t.Fatalf("unknown: got %q", got)
	}
	// args ignored
	got, _ = s.Interpolate("{{$randomInt 1 100}}")
	if got != "42" {
		t.Fatalf("args ignored: got %q", got)
	}
	// variables still work alongside tags
	s.Set(ScopeGlobal, "myVar", "hello")
	got, _ = s.Interpolate("{{myVar}} {{$uuid}}")
	if got != "hello 00000000-0000-4000-a000-000000000000" {
		t.Fatalf("mix: got %q", got)
	}
	// unknown tag with args still literal
	got, _ = s.Interpolate("{{$unknown foo bar}}")
	if !strings.Contains(got, "{{$unknown") {
		t.Fatalf("unknown with args: got %q", got)
	}
}

type fixedTagGenerator struct{}

func (f fixedTagGenerator) Generate(tag string, args []string) (string, bool) {
	switch tag {
	case "uuid":
		return "00000000-0000-4000-a000-000000000000", true
	case "timestamp":
		return "1700000000", true
	case "isoTimestamp":
		return "2023-11-14T22:13:20Z", true
	case "randomInt":
		return "42", true
	case "randomString":
		return "abcd1234", true
	default:
		return "", false
	}
}
