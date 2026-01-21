package protobuild

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProtoWalker_GetProtoFiles(t *testing.T) {
	// Create temp directory with proto files
	tmpDir := t.TempDir()

	// Create test proto files
	testFiles := []string{"test1.proto", "test2.proto", "other.txt"}
	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("syntax = \"proto3\";"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	walker := NewProtoWalker([]string{tmpDir}, nil)
	files := walker.GetProtoFiles(tmpDir)

	if len(files) != 2 {
		t.Errorf("expected 2 proto files, got %d", len(files))
	}

	for _, f := range files {
		if filepath.Ext(f) != ".proto" {
			t.Errorf("expected .proto extension, got %s", f)
		}
	}
}

func TestProtoWalker_HasProtoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	walker := NewProtoWalker([]string{tmpDir}, nil)

	// Empty directory should return false
	if walker.HasProtoFiles(tmpDir) {
		t.Error("expected no proto files in empty directory")
	}

	// Create a proto file
	protoFile := filepath.Join(tmpDir, "test.proto")
	if err := os.WriteFile(protoFile, []byte("syntax = \"proto3\";"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Should now return true
	if !walker.HasProtoFiles(tmpDir) {
		t.Error("expected proto files after creating one")
	}
}

func TestProtoWalker_WalkDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	subDir1 := filepath.Join(tmpDir, "api", "v1")
	subDir2 := filepath.Join(tmpDir, "internal")

	if err := os.MkdirAll(subDir1, 0755); err != nil {
		t.Fatalf("failed to create subdir1: %v", err)
	}
	if err := os.MkdirAll(subDir2, 0755); err != nil {
		t.Fatalf("failed to create subdir2: %v", err)
	}

	// Create proto files in subDir1 only
	if err := os.WriteFile(filepath.Join(subDir1, "test.proto"), []byte("syntax = \"proto3\";"), 0644); err != nil {
		t.Fatalf("failed to create proto file: %v", err)
	}

	walker := NewProtoWalker([]string{tmpDir}, nil)
	dirs := walker.WalkDirs()

	if len(dirs) != 1 {
		t.Errorf("expected 1 directory with proto files, got %d", len(dirs))
	}

	if _, ok := dirs[subDir1]; !ok {
		t.Errorf("expected subDir1 in result")
	}
}

func TestProtoWalker_WithExcludes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	includedDir := filepath.Join(tmpDir, "api")
	excludedDir := filepath.Join(tmpDir, "vendor")

	if err := os.MkdirAll(includedDir, 0755); err != nil {
		t.Fatalf("failed to create includedDir: %v", err)
	}
	if err := os.MkdirAll(excludedDir, 0755); err != nil {
		t.Fatalf("failed to create excludedDir: %v", err)
	}

	// Create proto files in both directories
	if err := os.WriteFile(filepath.Join(includedDir, "test.proto"), []byte("syntax = \"proto3\";"), 0644); err != nil {
		t.Fatalf("failed to create proto file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(excludedDir, "vendor.proto"), []byte("syntax = \"proto3\";"), 0644); err != nil {
		t.Fatalf("failed to create proto file: %v", err)
	}

	walker := NewProtoWalker([]string{tmpDir}, []string{excludedDir})
	dirs := walker.WalkDirs()

	if len(dirs) != 1 {
		t.Errorf("expected 1 directory (excluding vendor), got %d", len(dirs))
	}

	if _, ok := dirs[excludedDir]; ok {
		t.Error("vendor directory should be excluded")
	}
}

func TestCountProtoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some proto files
	for i := 0; i < 5; i++ {
		path := filepath.Join(tmpDir, "test"+string(rune('0'+i))+".proto")
		if err := os.WriteFile(path, []byte("syntax = \"proto3\";"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Create non-proto file
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	paths := map[string]string{"test": tmpDir}
	count := CountProtoFiles(paths)

	if count != 5 {
		t.Errorf("expected 5 proto files, got %d", count)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %s, expected %s", tt.bytes, result, tt.expected)
		}
	}
}
