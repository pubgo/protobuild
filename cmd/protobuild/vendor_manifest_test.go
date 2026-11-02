package protobuild

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func setupVendorDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "retag", "retag.proto"), "syntax = \"proto3\";\npackage retag;\n")
	writeTestFile(t, filepath.Join(dir, "google", "protobuf", "timestamp.proto"), "syntax = \"proto3\";\npackage google.protobuf;\n")
	// the checksum file is written by saveConfig, excluded from the manifest
	writeTestFile(t, filepath.Join(dir, "checksum"), "deadbeef")
	return dir
}

func TestVendorManifestRoundtrip(t *testing.T) {
	dir := setupVendorDir(t)

	m, err := buildVendorManifest(dir, func(rel string) string { return "test-dep" }, "cfgsum-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m.Entries))
	}

	if err := writeVendorManifest(dir, m); err != nil {
		t.Fatal(err)
	}

	problems, stale, err := verifyVendor(dir, "cfgsum-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 {
		t.Fatalf("expected no problems, got %v", problems)
	}
	if stale {
		t.Fatal("expected stale=false for matching config checksum")
	}
}

func TestVendorManifestDetectsModification(t *testing.T) {
	dir := setupVendorDir(t)

	m, err := buildVendorManifest(dir, nil, "cfgsum-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeVendorManifest(dir, m); err != nil {
		t.Fatal(err)
	}

	// simulate the hand-edited vendored file from the retag incident
	writeTestFile(t, filepath.Join(dir, "retag", "retag.proto"), "package .proto/retag;\n")

	problems, _, err := verifyVendor(dir, "cfgsum-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 {
		t.Fatalf("expected 1 problem, got %d: %v", len(problems), problems)
	}
}

func TestVendorManifestDetectsMissingAndUntracked(t *testing.T) {
	dir := setupVendorDir(t)

	m, err := buildVendorManifest(dir, nil, "cfgsum-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeVendorManifest(dir, m); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(filepath.Join(dir, "google", "protobuf", "timestamp.proto")); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(dir, "extra", "extra.proto"), "syntax = \"proto3\";\n")

	problems, _, err := verifyVendor(dir, "cfgsum-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 2 {
		t.Fatalf("expected 2 problems (missing + untracked), got %d: %v", len(problems), problems)
	}
}

func TestVendorManifestStaleConfig(t *testing.T) {
	dir := setupVendorDir(t)

	m, err := buildVendorManifest(dir, nil, "cfgsum-old")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeVendorManifest(dir, m); err != nil {
		t.Fatal(err)
	}

	problems, stale, err := verifyVendor(dir, "cfgsum-new")
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 {
		t.Fatalf("config drift must not fail integrity, got %v", problems)
	}
	if !stale {
		t.Fatal("expected stale=true when config checksum differs")
	}
}

func TestVerifyVendorWithoutManifest(t *testing.T) {
	dir := t.TempDir()

	problems, _, err := verifyVendor(dir, "cfgsum-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 {
		t.Fatalf("expected missing-manifest problem, got %v", problems)
	}
}
