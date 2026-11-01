package protobuild

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/pubgo/funk/v2/assert"
	"gopkg.in/yaml.v3"
)

// vendorManifestName is the manifest file written inside the vendor directory.
const vendorManifestName = "manifest.yaml"

// files excluded from manifest hashing.
var manifestExcludedFiles = map[string]bool{
	vendorManifestName: true,
	"checksum":         true,
}

// manifestEntry records the expected content of a single vendored file.
type manifestEntry struct {
	// Path is relative to the vendor directory, slash-separated.
	Path string `yaml:"path"`
	// SHA256 hex digest of the file content.
	SHA256 string `yaml:"sha256"`
	Size   int64  `yaml:"size"`
	// Source is the dependency name the file was vendored from.
	Source string `yaml:"source,omitempty"`
}

// vendorManifest records the expected content of the vendor directory,
// written by 'protobuild vendor' and checked by 'protobuild verify'.
type vendorManifest struct {
	Version int `yaml:"version"`
	// GeneratedAt is the time the manifest was written.
	GeneratedAt time.Time `yaml:"generated_at"`
	// ConfigChecksum is the protobuf.yaml checksum at vendor time, used to
	// detect config drift (warning only, does not fail verification).
	ConfigChecksum string          `yaml:"config_checksum,omitempty"`
	Entries        []manifestEntry `yaml:"entries"`
}

// hashVendorFile returns the sha256 hex digest and size of a file.
func hashVendorFile(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	h := sha256.New()
	size, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}

	return hex.EncodeToString(h.Sum(nil)), size, nil
}

// buildVendorManifest walks the vendor directory and hashes every file,
// skipping the manifest and checksum files themselves. sourceOf attributes
// each relative path to the dependency it was vendored from.
func buildVendorManifest(vendorDir string, sourceOf func(relPath string) string, configChecksum string) (*vendorManifest, error) {
	m := &vendorManifest{
		Version:        1,
		GeneratedAt:    time.Now().UTC(),
		ConfigChecksum: configChecksum,
	}

	err := filepath.Walk(vendorDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(vendorDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if manifestExcludedFiles[rel] {
			return nil
		}

		sum, size, hashErr := hashVendorFile(path)
		if hashErr != nil {
			return hashErr
		}

		m.Entries = append(m.Entries, manifestEntry{
			Path:   rel,
			SHA256: sum,
			Size:   size,
			Source: optionalSource(sourceOf, rel),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(m.Entries, func(i, j int) bool {
		return m.Entries[i].Path < m.Entries[j].Path
	})

	return m, nil
}

// optionalSource resolves the dependency name for a relative path when an
// attribution callback is provided, otherwise returns an empty source.
func optionalSource(sourceOf func(relPath string) string, relPath string) string {
	if sourceOf == nil {
		return ""
	}
	return sourceOf(relPath)
}

// writeVendorManifest writes the manifest into the vendor directory.
func writeVendorManifest(vendorDir string, m *vendorManifest) error {
	data, err := yaml.Marshal(m)
	assert.Exit(err)

	return os.WriteFile(filepath.Join(vendorDir, vendorManifestName), data, 0o644)
}

// loadVendorManifest reads the manifest from the vendor directory.
func loadVendorManifest(vendorDir string) (*vendorManifest, error) {
	data, err := os.ReadFile(filepath.Join(vendorDir, vendorManifestName))
	if err != nil {
		return nil, err
	}

	var m vendorManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("解析 %s 失败: %w", vendorManifestName, err)
	}

	return &m, nil
}

// verifyVendor checks the vendor directory against its manifest and returns
// integrity problems (modified / missing / untracked files). stale reports
// config drift since the last vendor run, which is a warning, not a failure.
func verifyVendor(vendorDir, configChecksum string) (problems []string, stale bool, err error) {
	m, loadErr := loadVendorManifest(vendorDir)
	if loadErr != nil {
		return []string{fmt.Sprintf("%s 不存在或不可读，请重新执行 'protobuild vendor'", vendorManifestName)}, false, nil
	}

	stale = m.ConfigChecksum != "" && configChecksum != "" && m.ConfigChecksum != configChecksum

	expected := make(map[string]manifestEntry, len(m.Entries))
	for _, e := range m.Entries {
		expected[e.Path] = e
	}

	actual := make(map[string]bool)
	walkErr := filepath.Walk(vendorDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel, relErr := filepath.Rel(vendorDir, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		if manifestExcludedFiles[rel] {
			return nil
		}
		actual[rel] = true

		entry, ok := expected[rel]
		if !ok {
			problems = append(problems, fmt.Sprintf("未登记的文件: %s", rel))
			return nil
		}

		sum, _, hashErr := hashVendorFile(path)
		if hashErr != nil {
			return hashErr
		}
		if sum != entry.SHA256 {
			problems = append(problems, fmt.Sprintf("内容已变更: %s (来源: %s)", rel, entry.Source))
		}
		return nil
	})
	if walkErr != nil {
		return nil, false, walkErr
	}

	// deterministic problem order
	for _, e := range m.Entries {
		if !actual[e.Path] {
			problems = append(problems, fmt.Sprintf("文件缺失: %s (来源: %s)", e.Path, e.Source))
		}
	}
	sort.Strings(problems)

	return problems, stale, nil
}
