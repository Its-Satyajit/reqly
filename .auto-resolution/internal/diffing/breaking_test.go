package diffing

import (
	"testing"

	"github.com/Its-Satyajit/reqly/internal/openapi"
	"github.com/getkin/kin-openapi/openapi3"
)

func openapiLoadForTest(spec string) (*openapi3.T, error) {
	return openapi.Load([]byte(spec))
}

func openapiDoc(t *testing.T, spec string) *openapi3.T {
	t.Helper()
	doc, err := openapiLoadForTest(spec)
	if err != nil {
		t.Fatalf("load spec: %v", err)
	}
	return doc
}

const severityBase = `openapi: 3.0.3
info: {title: T, version: "1"}
paths:
  /pets:
    get:
      responses:
        "200": {description: ok}
`

func TestWithSeverityPathDeletedIsBreaking(t *testing.T) {
	specB := `openapi: 3.0.3
info: {title: T, version: "1"}
paths:
  /pets:
    get:
      responses:
        "200": {description: ok}
  /dogs:
    get:
      responses:
        "200": {description: ok}
`
	a := openapiDoc(t, severityBase)
	b := openapiDoc(t, specB)
	res, err := OpenAPI(b, a) // removing /dogs relative to b
	if err != nil {
		t.Fatalf("OpenAPI: %v", err)
	}
	out := WithSeverity(res)
	if !out.HasChanges {
		t.Fatal("expected changes")
	}
	found := false
	for _, c := range out.Changes {
		if c.Type == "delete" && c.Severity == SeverityBreaking {
			found = true
		}
	}
	if !found {
		t.Errorf("no breaking deletion found: %+v", out.Changes)
	}
}

func TestWithSeverityAdditionIsAdditive(t *testing.T) {
	specB := severityBase + `  /dogs:
    get:
      responses:
        "200": {description: ok}
`
	a := openapiDoc(t, severityBase)
	b := openapiDoc(t, specB)
	res, err := OpenAPI(a, b)
	if err != nil {
		t.Fatalf("OpenAPI: %v", err)
	}
	out := WithSeverity(res)
	additive := false
	for _, c := range out.Changes {
		if c.Type == "create" && c.Severity == SeverityNonBreaking {
			additive = true
		}
	}
	if !additive {
		t.Errorf("no addition severity found: %+v", out.Changes)
	}
}

func TestWithSeverityRequiredUpdateIsBreaking(t *testing.T) {
	specA := `openapi: 3.0.3
info: {title: T, version: "1"}
components:
  schemas:
    Pet:
      type: object
      required: [id]
paths: {}
`
	specB := `openapi: 3.0.3
info: {title: T, version: "1"}
components:
  schemas:
    Pet:
      type: object
      required: [id, name]
paths: {}
`
	a := openapiDoc(t, specA)
	b := openapiDoc(t, specB)
	res, err := OpenAPI(a, b)
	if err != nil {
		t.Fatalf("OpenAPI: %v", err)
	}
	out := WithSeverity(res)
	for _, c := range out.Changes {
		if lastSegment(c.Path) == "required" && c.Severity != SeverityBreaking {
			t.Errorf("required change classified %q, want breaking", c.Severity)
		}
	}
}

func TestWithSeverityDescriptionDeleteIsInfo(t *testing.T) {
	specA := `openapi: 3.0.3
info: {title: T, version: "1", description: hello}
paths: {}
`
	specB := `openapi: 3.0.3
info: {title: T, version: "1"}
paths: {}
`
	a := openapiDoc(t, specA)
	b := openapiDoc(t, specB)
	res, err := OpenAPI(a, b)
	if err != nil {
		t.Fatalf("OpenAPI: %v", err)
	}
	out := WithSeverity(res)
	for _, c := range out.Changes {
		if lastSegment(c.Path) == "description" && c.Severity != SeverityInfo {
			t.Errorf("description delete classified %q, want info", c.Severity)
		}
	}
}
