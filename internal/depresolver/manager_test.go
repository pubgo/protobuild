package depresolver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager("", "")

	if m.cacheDir == "" {
		t.Error("cacheDir should not be empty")
	}

	if m.gomodPath == "" {
		t.Error("gomodPath should not be empty")
	}
}

func TestNewManagerCustomPaths(t *testing.T) {
	cacheDir := "/tmp/test-cache"
	gomodPath := "/tmp/test-gomod"

	m := NewManager(cacheDir, gomodPath)

	if m.cacheDir != cacheDir {
		t.Errorf("cacheDir = %q, want %q", m.cacheDir, cacheDir)
	}

	if m.gomodPath != gomodPath {
		t.Errorf("gomodPath = %q, want %q", m.gomodPath, gomodPath)
	}
}

func TestDetectSource(t *testing.T) {
	tests := []struct {
		name     string
		dep      *Dependency
		expected Source
	}{
		{
			name:     "explicit source",
			dep:      &Dependency{Source: SourceGit, URL: "example.com/repo"},
			expected: SourceGit,
		},
		{
			name:     "git URL with .git suffix",
			dep:      &Dependency{URL: "https://github.com/user/repo.git"},
			expected: SourceGit,
		},
		{
			name:     "git:: prefix",
			dep:      &Dependency{URL: "git::https://github.com/user/repo"},
			expected: SourceGit,
		},
		{
			name:     "git@ SSH URL",
			dep:      &Dependency{URL: "git@github.com:user/repo.git"},
			expected: SourceGit,
		},
		{
			name:     "HTTP URL",
			dep:      &Dependency{URL: "https://example.com/file.tar.gz"},
			expected: SourceHTTP,
		},
		{
			name:     "S3 URL",
			dep:      &Dependency{URL: "s3://bucket/path"},
			expected: SourceS3,
		},
		{
			name:     "GCS URL",
			dep:      &Dependency{URL: "gs://bucket/path"},
			expected: SourceGCS,
		},
		{
			name:     "Go module path",
			dep:      &Dependency{URL: "github.com/user/repo"},
			expected: SourceGoMod,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use public DetectSource for URL-based detection
			// For explicit source, directly test the expected value
			if tt.dep.Source != SourceAuto {
				if tt.dep.Source != tt.expected {
					t.Errorf("explicit source = %q, want %q", tt.dep.Source, tt.expected)
				}
			} else {
				result := DetectSource(tt.dep.URL)
				if result != tt.expected {
					t.Errorf("DetectSource() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

func TestSourceDisplayName(t *testing.T) {
	tests := []struct {
		source   Source
		expected string
	}{
		{SourceGoMod, "Go Module"},
		{SourceGit, "Git"},
		{SourceHTTP, "HTTP"},
		{SourceS3, "AWS S3"},
		{SourceGCS, "Google Cloud Storage"},
		{SourceLocal, "Local"},
		{SourceAuto, "Auto"},
	}

	for _, tt := range tests {
		t.Run(string(tt.source), func(t *testing.T) {
			result := tt.source.DisplayName()
			if result != tt.expected {
				t.Errorf("DisplayName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestResolveLocal(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "depresolver-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.proto")
	if err := os.WriteFile(testFile, []byte("syntax = \"proto3\";"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	m := NewManager("", "")

	t.Run("existing path", func(t *testing.T) {
		dep := &Dependency{
			Name:   "test",
			Source: SourceLocal,
			URL:    tmpDir,
		}
		result, err := m.resolveLocal(dep)
		if err != nil {
			t.Errorf("resolveLocal() error = %v", err)
		}
		if result.LocalPath == "" {
			t.Error("LocalPath should not be empty")
		}
	})

	t.Run("non-existing path", func(t *testing.T) {
		dep := &Dependency{
			Name:   "test",
			Source: SourceLocal,
			URL:    "/non/existing/path",
		}
		_, err := m.resolveLocal(dep)
		if err == nil {
			t.Error("resolveLocal() should return error for non-existing path")
		}
	})

	t.Run("optional non-existing path", func(t *testing.T) {
		optional := true
		dep := &Dependency{
			Name:     "test",
			Source:   SourceLocal,
			URL:      "/non/existing/path",
			Optional: &optional,
		}
		result, err := m.resolveLocal(dep)
		if err != nil {
			t.Errorf("resolveLocal() should not return error for optional: %v", err)
		}
		if result.LocalPath != "" {
			t.Error("LocalPath should be empty for optional non-existing")
		}
	})
}

func TestResolveError(t *testing.T) {
	dep := &Dependency{
		Name: "test-dep",
		URL:  "https://example.com/repo.git",
		Ref:  "v1.0.0",
		Path: "proto",
	}

	err := &ResolveError{
		Dependency: dep,
		Source:     SourceGit,
		URL:        "git::https://example.com/repo.git?ref=v1.0.0",
		Operation:  "download",
		Err:        os.ErrNotExist,
	}

	errStr := err.Error()

	checks := []string{
		"test-dep",
		"Git",
		"download",
	}

	for _, check := range checks {
		if !strings.Contains(errStr, check) {
			t.Errorf("Error message should contain %q", check)
		}
	}
}

func TestBuildGetterURL(t *testing.T) {
	m := NewManager("", "")

	tests := []struct {
		name     string
		dep      *Dependency
		source   Source
		contains string
	}{
		{
			name:     "git with ref",
			dep:      &Dependency{URL: "https://github.com/user/repo.git", Ref: "v1.0.0"},
			source:   SourceGit,
			contains: "ref=v1.0.0",
		},
		{
			name:     "git SSH URL",
			dep:      &Dependency{URL: "git@github.com:user/repo.git"},
			source:   SourceGit,
			contains: "git::git@github.com",
		},
		{
			name:     "S3 URL",
			dep:      &Dependency{URL: "s3://bucket/path"},
			source:   SourceS3,
			contains: "s3://bucket/path",
		},
		{
			name:     "GCS gs:// URL",
			dep:      &Dependency{URL: "gs://bucket/path"},
			source:   SourceGCS,
			contains: "gcs://bucket/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.buildGetterURL(tt.dep, tt.source)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("buildGetterURL() = %q, want contains %q", result, tt.contains)
			}
		})
	}
}

func TestHashString(t *testing.T) {
	hash1 := hashString("test-input")
	hash2 := hashString("test-input")
	if hash1 != hash2 {
		t.Error("Same input should produce same hash")
	}

	hash3 := hashString("different-input")
	if hash1 == hash3 {
		t.Error("Different input should produce different hash")
	}

	if len(hash1) != 24 {
		t.Errorf("Hash length = %d, want 24", len(hash1))
	}
}

func TestCleanCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "depresolver-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	m := NewManager(tmpDir, "")

	testDir := filepath.Join(tmpDir, "git", "test123")
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	if err := m.CleanCache(); err != nil {
		t.Errorf("CleanCache() error = %v", err)
	}

	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		t.Error("Cache directory should be removed")
	}
}
