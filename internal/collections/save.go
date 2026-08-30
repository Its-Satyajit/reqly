package collections

import (
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"

	"github.com/Its-Satyajit/reqly/internal/requestfile"
)

// SaveWorkspace writes ws to root (bulk save, pruning deleted). It creates
// missing dirs, writes reqly.yaml + per-collection/folder reqly.yaml + per-request
// via requestfile.Save (format-preserving, atomic), and prunes collections/
// dirs/files that no longer exist in ws.
func SaveWorkspace(root string, ws *Workspace) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	if err := saveConfig(filepath.Join(root, configFileName), ws.Config); err != nil {
		return err
	}
	collectionsDir := filepath.Join(root, "collections")
	if err := os.MkdirAll(collectionsDir, 0o755); err != nil {
		return err
	}
	expectedColls := map[string]bool{}
	expectedFiles := map[string]bool{}
	expectedFolders := map[string]bool{}
	for _, coll := range ws.Collections {
		collDir := filepath.Join(collectionsDir, coll.Name)
		expectedColls[collDir] = true
		if err := os.MkdirAll(collDir, 0o755); err != nil {
			return err
		}
		if err := saveConfig(filepath.Join(collDir, configFileName), coll.Config); err != nil {
			return err
		}
		if err := saveRequests(collDir, coll.Requests, expectedFiles); err != nil {
			return err
		}
		if err := saveFolders(collDir, coll.Folders, expectedFiles, expectedFolders); err != nil {
			return err
		}
	}
	// prune deleted collections
	if err := pruneCollections(collectionsDir, expectedColls); err != nil {
		return err
	}
	// prune deleted request files and folders
	return pruneFiles(collectionsDir, expectedFiles, expectedFolders)
}

func saveConfig(path string, cfg Config) error {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("encode config %q: %w", path, err)
	}
	// atomic via temp file
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "reqly.yaml.tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
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

func saveRequests(dir string, reqs []*RequestEntry, expected map[string]bool) error {
	for _, r := range reqs {
		// r.Path may be absolute; if not, use dir + name + ext from original or .yaml
		path := r.Path
		if path == "" || !filepath.IsAbs(path) {
			ext := ".yaml"
			if r.File != nil && r.Path != "" {
				ext = filepath.Ext(r.Path)
				if ext == "" {
					ext = ".yaml"
				}
			}
			path = filepath.Join(dir, r.Name+ext)
		}
		// ensure path is under dir
		if !filepath.IsAbs(path) {
			path = filepath.Join(dir, path)
		}
		expected[path] = true
		if r.File == nil {
			r.File = &requestfile.File{}
		}
		// ensure File has request
		if err := requestfile.Save(path, r.File); err != nil {
			return err
		}
	}
	return nil
}

func saveFolders(parentDir string, folders []*Folder, expectedFiles map[string]bool, expectedFolders map[string]bool) error {
	for _, f := range folders {
		dir := filepath.Join(parentDir, f.Name)
		expectedFolders[dir] = true
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		if err := saveConfig(filepath.Join(dir, configFileName), f.Config); err != nil {
			return err
		}
		if err := saveRequests(dir, f.Requests, expectedFiles); err != nil {
			return err
		}
		if err := saveFolders(dir, f.Folders, expectedFiles, expectedFolders); err != nil {
			return err
		}
	}
	return nil
}

func pruneCollections(collectionsDir string, expected map[string]bool) error {
	entries, err := os.ReadDir(collectionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(collectionsDir, e.Name())
		if !expected[path] {
			if err := os.RemoveAll(path); err != nil {
				return err
			}
		}
	}
	return nil
}

func pruneFiles(collectionsDir string, expectedFiles map[string]bool, expectedFolders map[string]bool) error {
	return filepath.Walk(collectionsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Base(path) == configFileName {
			return nil
		}
		// request file?
		if isRequestFile(filepath.Base(path)) {
			if !expectedFiles[path] {
				_ = os.Remove(path)
			}
		}
		return nil
	})
}
