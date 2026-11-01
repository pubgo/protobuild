package protobuild

import (
	"path/filepath"
	"testing"

	"github.com/pubgo/protobuild/internal/config"
)

func TestCopyToVendorWritesManifestWithSource(t *testing.T) {
	root := t.TempDir()
	// mirrors the real flow: the resolved dep path already points into the
	// dep directory, files are copied to vendor/<dep name>/...
	src := filepath.Join(root, "cache", "proto", "retag")
	writeTestFile(t, filepath.Join(src, "retag.proto"), "package retag;\n")
	vendorDir := filepath.Join(root, ".proto")

	svc := NewVendorService(&config.Config{Vendor: vendorDir, Checksum: "sum-1"})
	n, err := svc.CopyToVendor(map[string]string{"retag": src})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 copied file, got %d", n)
	}

	m, err := loadVendorManifest(vendorDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(m.Entries))
	}
	if m.Entries[0].Path != "retag/retag.proto" {
		t.Fatalf("expected path retag/retag.proto, got %q", m.Entries[0].Path)
	}
	if m.Entries[0].Source != "retag" {
		t.Fatalf("expected source=retag, got %q", m.Entries[0].Source)
	}
	if m.ConfigChecksum != "sum-1" {
		t.Fatalf("expected config checksum recorded, got %q", m.ConfigChecksum)
	}
}
