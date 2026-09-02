package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Its-Satyajit/reqly/internal/collections"
	"github.com/Its-Satyajit/reqly/internal/exporter"
)

// Generate writes Markdown docs for ws to outDir (index.md + per-collection <coll>.md).
// env is reserved for future --env resolution (currently unused, raw {{var}} shown).
func Generate(outDir string, ws *collections.Workspace, env string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	// index.md
	var idx strings.Builder
	idx.WriteString(fmt.Sprintf("# %s\n\n", ws.Config.Name))
	if ws.Config.Name == "" {
		idx.WriteString("# Workspace\n\n")
	}
	idx.WriteString("Collections:\n\n")
	for _, c := range ws.Collections {
		idx.WriteString(fmt.Sprintf("- [%s](%s.md) (%d requests)\n", c.Name, c.Name, len(c.Requests)+countFolderRequests(c.Folders)))
	}
	if err := writeFile(filepath.Join(outDir, "index.md"), idx.String()); err != nil {
		return err
	}
	// per collection
	for _, c := range ws.Collections {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# %s\n\n", c.Name))
		if c.Config.BaseURL != "" {
			sb.WriteString(fmt.Sprintf("Base URL: %s\n\n", c.Config.BaseURL))
		}
		// requests in collection
		for _, r := range c.Requests {
			writeRequest(&sb, r)
		}
		// folders recursively
		var walk func(folders []*collections.Folder, prefix string)
		walk = func(folders []*collections.Folder, prefix string) {
			for _, f := range folders {
				sb.WriteString(fmt.Sprintf("## %s%s\n\n", prefix, f.Name))
				for _, r := range f.Requests {
					writeRequest(&sb, r)
				}
				walk(f.Folders, prefix+f.Name+"/")
			}
		}
		walk(c.Folders, "")
		if err := writeFile(filepath.Join(outDir, c.Name+".md"), sb.String()); err != nil {
			return err
		}
	}
	return nil
}

func countFolderRequests(folders []*collections.Folder) int {
	n := 0
	for _, f := range folders {
		n += len(f.Requests)
		n += countFolderRequests(f.Folders)
	}
	return n
}

func writeRequest(sb *strings.Builder, r *collections.RequestEntry) {
	req := r.File.Request
	method := string(req.Method)
	if method == "" {
		method = "GET"
	}
	sb.WriteString(fmt.Sprintf("### %s\n\n", r.Name))
	sb.WriteString(fmt.Sprintf("- **Method:** %s\n", method))
	sb.WriteString(fmt.Sprintf("- **URL:** %s\n", req.URL))
	if len(req.Headers) > 0 {
		sb.WriteString("- **Headers:**\n")
		for _, h := range req.Headers {
			sb.WriteString(fmt.Sprintf("  - %s: %s\n", h.Key, h.Value))
		}
	}
	if len(req.Query) > 0 {
		sb.WriteString("- **Query:**\n")
		for _, q := range req.Query {
			sb.WriteString(fmt.Sprintf("  - %s=%s\n", q.Key, q.Value))
		}
	}
	if req.Body != "" {
		sb.WriteString(fmt.Sprintf("- **Body:**\n\n```\n%s\n```\n", req.Body))
	}
	// cURL via exporter
	curl, err := exporter.Generate(req, "curl", nil)
	if err == nil {
		sb.WriteString(fmt.Sprintf("- **cURL:**\n\n```bash\n%s\n```\n", curl))
	}
	// template for future: use text/template for more
	_ = template.New
}

func writeFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "docs-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	tmp.Close()
	if err := os.Chmod(tmpName, 0o644); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}
