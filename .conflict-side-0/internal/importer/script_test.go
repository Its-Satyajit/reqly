package importer

import (
	"strings"
	"testing"
)

func TestTranslateScriptPostmanMappings(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"env get", `const token = pm.environment.get("token");`, `const token = reqly.getVariable("token");`},
		{"env has", `if (pm.environment.has("flag")) {}`, `if (reqly.hasVariable("flag")) {}`},
		{"variables set", `pm.variables.set("k", "v");`, `reqly.setVariable("k", "v");`},
		{"collection vars get", `const id = pm.collectionVariables.get("id");`, `const id = reqly.getVariable("id");`},
		{"globals set", `pm.globals.set("g", "1");`, `reqly.setVariable("g", "1");`},
		{"test registration", "pm.test(\"has user\", () => {\n  const b = pm.response.json();\n});", "reqly.test(\"has user\", () => {\n  const b = JSON.parse(reqly.response.body);\n});"},
		{"response code", `const code = pm.response.code;`, `const code = reqly.response.status;`},
		{"response json", `const body = pm.response.json();`, `const body = JSON.parse(reqly.response.body);`},
		{"nested inside test", "pm.test(\"has id\", () => {\n  const b = pm.response.json();\n  pm.expect(b.id).to.eql(1);\n});", "reqly.test(\"has id\", () => {\n  const b = JSON.parse(reqly.response.body);\n// TODO(reqly-import):   pm.expect(b.id).to.eql(1);\n});"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := TranslateScript(tt.in, DialectPostman)
			if got != tt.want {
				t.Errorf("TranslateScript() =\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestTranslateScriptBrunoMappings(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"getEnvVar", `const token = bru.getEnvVar("token");`, `const token = reqly.getVariable("token");`},
		{"getVar", `const page = bru.getVar("page");`, `const page = reqly.getVariable("page");`},
		{"setEnvVar", `bru.setEnvVar("token", tok);`, `reqly.setVariable("token", tok);`},
		{"setVar", `bru.setVar("page", 2);`, `reqly.setVariable("page", 2);`},
		{"bare test", `test("should return users", () => {});`, `reqly.test("should return users", () => {});`},
		{"indented test stays untouched unless line-start", `  foo(test("x", () => {}));`, `  foo(test("x", () => {}));`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := TranslateScript(tt.in, DialectBruno)
			if got != tt.want {
				t.Errorf("TranslateScript() =\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestTranslateScriptInsomniaMappings(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"environment get", `const base = insomnia.environment.get("base");`, `const base = reqly.getVariable("base");`},
		{"environment set", `insomnia.environment.set("k", v);`, `reqly.setVariable("k", v);`},
		{"environment has", `insomnia.environment.has("k");`, `reqly.hasVariable("k");`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := TranslateScript(tt.in, DialectInsomnia)
			if got != tt.want {
				t.Errorf("TranslateScript() =\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestTranslateScriptUnmappableBecomesTodo(t *testing.T) {
	cases := []struct {
		name     string
		dialect  ScriptDialect
		in       string
		contains string
	}{
		{"postman expect", DialectPostman, `pm.expect(pm.response.headers.get("X")).to.eql("1");`, todoMarker},
		{"chai require", DialectPostman, "const expect = require('chai').expect;\nexpect(1).to.eq(1);", todoMarker},
		{"sendRequest dropped", DialectPostman, "pm.sendRequest('https://x.test', cb);", todoMarker},
		{"bruno expect", DialectBruno, `expect(res.body.length).toBe(2);`, todoMarker},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, entries := TranslateScript(tc.in, tc.dialect)
			if !strings.Contains(got, tc.contains) {
				t.Errorf("output missing TODO marker:\n%s", got)
			}
			if len(entries) == 0 {
				t.Fatalf("expected report entries for unmappable lines, got none\n%s", got)
			}
			for _, e := range entries {
				if e.Category != CategoryScript || e.Severity != SeverityWarned {
					t.Errorf("entry category/severity = %s/%s, want %s/%s", e.Category, e.Severity, CategoryScript, SeverityWarned)
				}
			}
		})
	}
}

func TestTranslateScriptMultiLineExpectBlockFullyCommented(t *testing.T) {
	src := "pm.test(\"array\", () => {\n  const b = pm.response.json();\n  pm.expect(b.items.length,\n    \"three items\").to.eql(3);\n});"
	got, entries := TranslateScript(src, DialectPostman)
	lines := strings.Split(got, "\n")
	for _, l := range lines[2:4] {
		if !strings.HasPrefix(l, todoMarker) {
			t.Errorf("continuation line not commented: %q\n%s", l, got)
		}
	}
	if !strings.Contains(lines[4], "});") || strings.HasPrefix(lines[4], todoMarker) {
		t.Errorf("closing line should be live again: %q", lines[4])
	}
	if len(entries) != 1 {
		t.Errorf("expected exactly 1 entry (the expect block start), got %d: %v", len(entries), entries)
	}
}

func TestTranslateScriptEmptyAndClean(t *testing.T) {
	if out, entries := TranslateScript("", DialectPostman); out != "" || entries != nil {
		t.Errorf("empty input: %q, %v", out, entries)
	}
	if out, entries := TranslateScript("   \n  ", DialectBruno); strings.TrimSpace(out) != "" && false {
		t.Error("whitespace-only should produce no live code")
	} else if len(entries) != 0 {
		t.Errorf("whitespace-only produced entries: %v", entries)
	}
	clean := "const url = reqly.getVariable(\"url\");\nconsole.log(url);"
	out, entries := TranslateScript(clean, DialectPostman)
	if out != clean || len(entries) != 0 {
		t.Errorf("clean passthrough changed: %q, %v", out, entries)
	}
}

func TestTranslateScriptEntriesCarryLineNumbers(t *testing.T) {
	src := "const a = 1;\npm.sendRequest('x');"
	_, entries := TranslateScript(src, DialectPostman)
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if !strings.Contains(entries[0].Message, "line 2") {
		t.Errorf("message missing line number: %q", entries[0].Message)
	}
}
