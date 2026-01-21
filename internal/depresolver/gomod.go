package depresolver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pubgo/funk/pathutil"
	"github.com/pubgo/protobuild/internal/modutil"
	"github.com/pubgo/protobuild/internal/shutil"
	"github.com/schollz/progressbar/v3"
)

// resolveGoMod resolves dependencies using Go modules
func (m *Manager) resolveGoMod(ctx context.Context, dep *Dependency) (*ResolveResult, error) {
	url := os.ExpandEnv(dep.URL)

	// Parse version from URL if specified as url@version
	if idx := strings.Index(url, "@"); idx > 0 {
		v := url[idx+1:]
		url = url[:idx]
		dep.Version = &v
	}

	// Check if URL is a local directory
	if pathutil.IsDir(url) {
		localPath := url
		if dep.Path != "" {
			localPath = filepath.Join(url, dep.Path)
		}
		return &ResolveResult{
			LocalPath: localPath,
			Changed:   false,
		}, nil
	}

	// Load versions from go.mod graph
	versions := modutil.LoadVersionGraph()

	// Resolve version
	version := m.resolveGoModVersion(dep, url, versions)

	// Check if we need to download
	changed := false
	modCachePath := filepath.Join(m.gomodPath, fmt.Sprintf("%s@%s", url, version))

	if version == "" || pathutil.IsNotExist(modCachePath) {
		changed = true

		// Create progress bar for visual feedback
		bar := progressbar.NewOptions(-1,
			progressbar.OptionSetDescription(fmt.Sprintf("  ↓ [Go Module] %s", dep.Name)),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionShowBytes(false),
			progressbar.OptionSetWidth(30),
			progressbar.OptionOnCompletion(func() { fmt.Println() }),
		)

		// Start spinner in background
		done := make(chan struct{})
		go func() {
			for {
				select {
				case <-done:
					return
				case <-ctx.Done():
					return
				default:
					_ = bar.Add(1)
				}
			}
		}()

		var err error
		if version == "" {
			err = shutil.Shell("go", "get", "-d", url+"/...").Run()
		} else {
			err = shutil.Shell("go", "get", "-d", fmt.Sprintf("%s@%s", url, version)).Run()
		}

		close(done)
		_ = bar.Finish()

		if err != nil {
			if dep.Optional != nil && *dep.Optional {
				return &ResolveResult{LocalPath: "", Changed: false}, nil
			}
			return nil, &ResolveError{
				Dependency: dep,
				Source:     SourceGoMod,
				URL:        url,
				Operation:  "download",
				Err:        err,
			}
		}

		// Reload versions after download
		versions = modutil.LoadVersions()
		version = versions[url]
		if version == "" {
			return nil, &ResolveError{
				Dependency: dep,
				Source:     SourceGoMod,
				URL:        url,
				Operation:  "resolve",
				Err:        fmt.Errorf("version not found after download, check if module path is correct"),
			}
		}
		modCachePath = filepath.Join(m.gomodPath, fmt.Sprintf("%s@%s", url, version))
	}

	// Build final path
	localPath := modCachePath
	if dep.Path != "" {
		localPath = filepath.Join(modCachePath, dep.Path)
	}

	// Check if path exists
	if pathutil.IsNotExist(localPath) {
		if dep.Optional != nil && *dep.Optional {
			return &ResolveResult{LocalPath: "", Changed: false}, nil
		}
		return nil, &ResolveError{
			Dependency: dep,
			Source:     SourceGoMod,
			URL:        url,
			Operation:  "validate",
			Err:        fmt.Errorf("subdirectory '%s' not found in module", dep.Path),
		}
	}

	// Update version in dependency
	dep.Version = &version

	return &ResolveResult{
		LocalPath: localPath,
		Version:   version,
		Changed:   changed,
	}, nil
}

// resolveGoModVersion resolves the version for a Go module dependency
func (m *Manager) resolveGoModVersion(dep *Dependency, url string, versions map[string]string) string {
	// Priority: version in versions map > explicit version > local cache scan
	if v, ok := versions[url]; ok {
		return v
	}

	if dep.Version != nil && *dep.Version != "" {
		return *dep.Version
	}

	// Try to find in local cache
	dir := filepath.Dir(filepath.Join(m.gomodPath, url))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	_, name := filepath.Split(url)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), name+"@") {
			return strings.TrimPrefix(entry.Name(), name+"@")
		}
	}

	return ""
}
