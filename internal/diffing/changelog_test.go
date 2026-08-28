package diffing

import (
	"strings"
	"testing"
)

func TestGenerateChangelog_BreakingAndAdditions(t *testing.T) {
	oldSpec := []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "Users API", "version": "1.0.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "getUsers",
					"responses": {"200": {"description": "OK"}}
				},
				"delete": {
					"operationId": "deleteUser",
					"responses": {"204": {"description": "No Content"}}
				}
			}
		}
	}`)

	newSpec := []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "Users API", "version": "2.0.0", "description": "Updated API"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "getUsers",
					"responses": {"200": {"description": "OK"}}
				},
				"post": {
					"operationId": "createUser",
					"responses": {"201": {"description": "Created"}}
				}
			},
			"/orders": {
				"get": {
					"operationId": "getOrders",
					"responses": {"200": {"description": "Orders list"}}
				}
			}
		}
	}`)

	cl, err := GenerateChangelog(oldSpec, newSpec)
	if err != nil {
		t.Fatalf("GenerateChangelog unexpected error: %v", err)
	}

	if cl.SuggestedSemver != "major" {
		t.Errorf("expected suggested semver 'major', got %q", cl.SuggestedSemver)
	}

	if len(cl.Breaking) == 0 {
		t.Errorf("expected breaking changes for deleted endpoint, got 0")
	}

	if len(cl.Additions) == 0 {
		t.Errorf("expected additions for new endpoints, got 0")
	}

	md := cl.ToMarkdown()
	if !strings.Contains(md, "Breaking Changes") {
		t.Errorf("markdown missing 'Breaking Changes' section: %s", md)
	}
	if !strings.Contains(md, "Additions") {
		t.Errorf("markdown missing 'Additions' section: %s", md)
	}
	if !strings.Contains(md, "Suggested Version Bump") || !strings.Contains(md, "`major`") {
		t.Errorf("markdown missing suggested semver: %s", md)
	}

	jsonStr, err := cl.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}
	if !strings.Contains(jsonStr, `"suggested_semver": "major"`) {
		t.Errorf("JSON missing suggested semver: %s", jsonStr)
	}
}

func TestGenerateChangelog_MinorOnly(t *testing.T) {
	oldSpec := []byte(`{"openapi": "3.0.0", "paths": {"/users": {"get": {"responses": {"200": {"description": "OK"}}}}}}`)
	newSpec := []byte(`{"openapi": "3.0.0", "paths": {"/users": {"get": {"responses": {"200": {"description": "OK"}}}, "post": {"responses": {"201": {"description": "Created"}}}}}}`)

	cl, err := GenerateChangelog(oldSpec, newSpec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cl.SuggestedSemver != "minor" {
		t.Errorf("expected semver 'minor', got %q", cl.SuggestedSemver)
	}
	if len(cl.Breaking) != 0 {
		t.Errorf("expected 0 breaking changes, got %d", len(cl.Breaking))
	}
	if len(cl.Additions) == 0 {
		t.Errorf("expected additions > 0, got 0")
	}
}

func TestGenerateChangelog_NoChanges(t *testing.T) {
	spec := []byte(`{"openapi": "3.0.0", "info": {"title": "API", "version": "1.0.0"}}`)
	cl, err := GenerateChangelog(spec, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cl.SuggestedSemver != "none" {
		t.Errorf("expected semver 'none', got %q", cl.SuggestedSemver)
	}
	if len(cl.Breaking)+len(cl.Additions)+len(cl.Info) != 0 {
		t.Errorf("expected 0 changes, got breaking=%d additions=%d info=%d", len(cl.Breaking), len(cl.Additions), len(cl.Info))
	}
}
