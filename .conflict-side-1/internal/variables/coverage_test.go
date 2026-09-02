package variables

import (
	"testing"
)

func TestCoverage_GetRangeClone(t *testing.T) {
	s := NewSet()
	s.Set(ScopeGlobal, "a", "1")
	s.Set(ScopeEnvironment, "b", "2")
	// Get
	if v, ok := s.Get(ScopeGlobal, "a"); !ok || v != "1" {
		t.Fatalf("Get global a: %v %v", v, ok)
	}
	if _, ok := s.Get(ScopeGlobal, "missing"); ok {
		t.Fatalf("should miss")
	}
	// Range
	count := 0
	s.Range(ScopeGlobal, func(k, v string) { count++ })
	if count != 1 {
		t.Fatalf("Range count %d", count)
	}
	// Clone
	c := s.Clone()
	c.Set(ScopeGlobal, "a", "9")
	if v, _ := s.Get(ScopeGlobal, "a"); v != "1" {
		t.Fatalf("Clone should not affect original")
	}
	if v, _ := c.Get(ScopeGlobal, "a"); v != "9" {
		t.Fatalf("Clone copy failed")
	}
}

func TestCoverage_UnknownDynamicTagsAndGenerate(t *testing.T) {
	// UnknownDynamicTags
	tags := UnknownDynamicTags("{{$uuid}} {{$unknown}} {{$timestamp}}")
	foundUnknown := false
	foundUUID := false
	for _, tag := range tags {
		if tag == "unknown" {
			foundUnknown = true
		}
		if tag == "uuid" {
			foundUUID = true
		}
	}
	if !foundUnknown {
		t.Fatalf("should find unknown, got %v", tags)
	}
	if foundUUID {
		t.Fatalf("uuid should not be unknown")
	}
	// Generate via defaultTagGenerator directly and via Interpolation
	gen := defaultTagGenerator{}
	if _, ok := gen.Generate("uuid", nil); !ok {
		t.Fatalf("uuid should generate")
	}
	if _, ok := gen.Generate("timestamp", nil); !ok {
		t.Fatalf("timestamp")
	}
	if _, ok := gen.Generate("isoTimestamp", nil); !ok {
		t.Fatalf("isoTimestamp")
	}
	if _, ok := gen.Generate("randomInt", nil); !ok {
		t.Fatalf("randomInt")
	}
	if _, ok := gen.Generate("randomString", nil); !ok {
		t.Fatalf("randomString")
	}
	if _, ok := gen.Generate("now", nil); !ok {
		t.Fatalf("now")
	}
	if _, ok := gen.Generate("unknownTag", nil); ok {
		t.Fatalf("unknown should not generate")
	}
	// Test SetTagGeneratorForTest
	fake := fixedGenerator{uuid: "fixed-uuid"}
	SetTagGeneratorForTest(fake)
	defer SetTagGeneratorForTest(nil)
	s := NewSet()
	got, _ := s.Interpolate("{{$uuid}}")
	if got != "fixed-uuid" {
		t.Fatalf("want fixed-uuid, got %q", got)
	}
}

type fixedGenerator struct {
	uuid string
}

func (f fixedGenerator) Generate(tag string, args []string) (string, bool) {
	if tag == "uuid" {
		return f.uuid, true
	}
	return "", false
}
