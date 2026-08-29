package theme

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		theme   Theme
		wantErr bool
	}{
		{
			name: "valid light",
			theme: Theme{
				ID:         "my-theme",
				Label:      "My Theme",
				Appearance: "light",
				Tokens:     map[string]string{"primary": "#ff0000", "background": "#ffffff"},
			},
			wantErr: false,
		},
		{
			name: "valid dark",
			theme: Theme{
				ID:         "dark-custom",
				Label:      "Dark Custom",
				Appearance: "dark",
				Tokens:     map[string]string{"primary": "#00ff00"},
			},
			wantErr: false,
		},
		{
			name: "missing id",
			theme: Theme{
				Label:      "No ID",
				Appearance: "light",
			},
			wantErr: true,
		},
		{
			name: "bad id",
			theme: Theme{
				ID:         "Bad_ID",
				Label:      "Bad",
				Appearance: "light",
			},
			wantErr: true,
		},
		{
			name: "missing label",
			theme: Theme{
				ID:         "my-theme",
				Appearance: "light",
			},
			wantErr: true,
		},
		{
			name: "bad appearance",
			theme: Theme{
				ID:         "my-theme",
				Label:      "My",
				Appearance: "blue",
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.theme)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate() err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestParseAndMarshal(t *testing.T) {
	yamlData := []byte(`
id: test-theme
label: Test Theme
appearance: dark
tokens:
  primary: "#123456"
  background: "#000000"
`)
	th, err := Parse(yamlData)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if th.ID != "test-theme" || th.Appearance != "dark" {
		t.Fatalf("unexpected theme %+v", th)
	}
	out, err := MarshalYAML(th)
	if err != nil {
		t.Fatalf("MarshalYAML: %v", err)
	}
	if !strings.Contains(string(out), "test-theme") {
		t.Fatalf("unexpected yaml %s", string(out))
	}
	j, err := MarshalJSON(th)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if !strings.Contains(string(j), "test-theme") {
		t.Fatalf("unexpected json %s", string(j))
	}
}

func TestToCSS(t *testing.T) {
	th := Theme{
		ID:         "my-theme",
		Label:      "My",
		Appearance: "light",
		Tokens:     map[string]string{"primary": "#ff0000", "background": "#ffffff"},
	}
	css, err := ToCSS(th)
	if err != nil {
		t.Fatalf("ToCSS: %v", err)
	}
	if !strings.Contains(css, `[data-theme="my-theme"]`) || !strings.Contains(css, "--primary: #ff0000") {
		t.Fatalf("unexpected css %s", css)
	}
}

func TestBuiltIn(t *testing.T) {
	builtins := BuiltInThemes()
	if len(builtins) != 2 {
		t.Fatalf("want 2 builtins, got %d", len(builtins))
	}
	if !IsBuiltIn("atlas-light") || !IsBuiltIn("atlas-dark") {
		t.Fatalf("expected atlas themes to be built-in")
	}
	if IsBuiltIn("custom") {
		t.Fatalf("custom should not be built-in")
	}
}

func TestParseJSON(t *testing.T) {
	jsonData := []byte(`{"id":"json-theme","label":"JSON","appearance":"light","tokens":{"primary":"#000000"}}`)
	th, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Parse JSON: %v", err)
	}
	if th.ID != "json-theme" {
		t.Fatalf("want json-theme, got %q", th.ID)
	}
}
