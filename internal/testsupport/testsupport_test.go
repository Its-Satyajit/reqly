package testsupport

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceCreatesDescriptorAndFiles(t *testing.T) {
	dir := Workspace(t, map[string]string{
		"environments/dev.yaml":       "variables:\n  REGION: us-west-2\n",
		"collections/users/list.yaml": "request: {method: GET, url: /list}\n",
	})
	for _, name := range []string{"reqly.yaml", "environments/dev.yaml", "collections/users/list.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(name))); err != nil {
			t.Fatalf("expected %s to exist: %v", name, err)
		}
	}
}

func TestWorkspaceRespectsSuppliedDescriptor(t *testing.T) {
	dir := Workspace(t, map[string]string{"reqly.yaml": "name: custom\nenvironment: dev\n"})
	data, err := os.ReadFile(filepath.Join(dir, "reqly.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "name: custom\nenvironment: dev\n" {
		t.Fatalf("descriptor overwritten: %q", data)
	}
}
