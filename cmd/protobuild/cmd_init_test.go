package protobuild

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInitProducesUsableProject(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	// syncBufConfig writes to the package-level pwd captured at process start;
	// align it with the chdir'ed temp dir (for the CLI these are the same).
	oldPwd := pwd
	pwd = dir
	t.Cleanup(func() { pwd = oldPwd })

	if err := os.WriteFile("go.mod", []byte("module example.com/demo\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runInit("basic", false); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	for _, want := range []string{"protobuf.yaml", "buf.yaml", "proto/example/v1/greeter.proto"} {
		if _, err := os.Stat(want); err != nil {
			t.Errorf("expected artifact %s: %v", want, err)
		}
	}
	content := string(mustReadFile(t, "protobuf.yaml"))
	if strings.Contains(content, "module:") {
		t.Errorf("source_relative config must NOT carry module (protoc-gen-go rejects module= with paths=source_relative), got:\n%s", content)
	}
	if strings.Contains(content, "protocolbuffers") {
		t.Error("config must not contain WKT github dep")
	}
}

func TestRunInitGatewayFillsModule(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	oldPwd := pwd
	pwd = dir
	t.Cleanup(func() { pwd = oldPwd })

	if err := os.WriteFile("go.mod", []byte("module smoke.example/gw\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runInit("grpc-gateway", false); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	content := string(mustReadFile(t, "protobuf.yaml"))
	if !strings.Contains(content, "module: smoke.example/gw/pkg") {
		t.Errorf("import-mode config must carry detected module, got:\n%s", content)
	}
	if !strings.Contains(content, "paths: import") {
		t.Errorf("gateway template must keep paths=import, got:\n%s", content)
	}
}

func TestVerifyPassesZeroDepProject(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	oldPwd := pwd
	pwd = dir
	t.Cleanup(func() { pwd = oldPwd })

	if err := os.WriteFile("go.mod", []byte("module smoke.example/verify\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runInit("minimal", false); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	if err := parseConfig(); err != nil {
		t.Fatalf("parseConfig: %v", err)
	}
	if err := runVerify(); err != nil {
		t.Errorf("zero-dep project must pass verify (new templates ship no deps), got: %v", err)
	}
}

func TestVerifyFailsWithStrayFilesNoDepsNoManifest(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	oldPwd := pwd
	pwd = dir
	t.Cleanup(func() { pwd = oldPwd })

	cfg := "vendor: .proto\nroot: [proto]\nincludes: [proto]\nplugins:\n  - name: go\n"
	if err := os.WriteFile("protobuf.yaml", []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := parseConfig(); err != nil {
		t.Fatalf("parseConfig: %v", err)
	}
	// drop a stray file into vendor without deps and without manifest:
	// the empty-set pass must NOT bless hand-placed vendor content
	if err := os.MkdirAll(filepath.Join(".proto", "google", "protobuf"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(".proto", "google", "protobuf", "any.proto"), []byte("syntax=\"proto3\";package google.protobuf;message Any{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runVerify(); err == nil {
		t.Error("stray vendor content without manifest must fail even for zero-dep projects")
	}
}

func TestVerifyFailsWithDepsNoVendor(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	oldPwd := pwd
	pwd = dir
	t.Cleanup(func() { pwd = oldPwd })

	src := filepath.Join(dir, "extsrc")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := "vendor: .proto\nroot: [proto]\nincludes: [proto]\ndeps:\n  - name: ext\n    source: local\n    url: " + src + "\nplugins:\n  - name: go\n"
	if err := os.WriteFile("protobuf.yaml", []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := parseConfig(); err != nil {
		t.Fatalf("parseConfig: %v", err)
	}
	if err := runVerify(); err == nil {
		t.Error("project with deps but never vendored must still fail verify (guard against over-broad zero-dep pass)")
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func TestTemplatesHaveBase(t *testing.T) {
	for _, name := range []string{"basic", "minimal", "grpc-gateway"} {
		tmpl := initTemplates[name]
		if tmpl.BasePlugin == nil {
			t.Fatalf("template %q: BasePlugin nil, generated config unusable for Go projects", name)
		}
		if tmpl.BasePlugin.Out == "" || tmpl.BasePlugin.Paths == "" {
			t.Errorf("template %q: BasePlugin out/paths empty", name)
		}
	}
}

func TestTemplatesNoWKTDep(t *testing.T) {
	for _, name := range []string{"basic", "minimal", "grpc-gateway"} {
		for _, dep := range initTemplates[name].Depends {
			if dep.Name == "google/protobuf" {
				t.Errorf("template %q: WKT dep %q must be dropped (protoc resolves WKT natively)", name, dep.Url)
			}
		}
	}
}

func TestGatewayTemplateOpenAPIV2SkipsBase(t *testing.T) {
	// protoc-gen-openapiv2 has no paths=/module= flags; protoc fails atomically
	// when runVendor injects base opts into it. The template must skip base for it.
	for _, p := range initTemplates["grpc-gateway"].Plugins {
		if p.Name == "openapiv2" {
			if !p.SkipBase {
				t.Fatalf("openapiv2 plugin must set SkipBase: protoc rejects paths=import for it (verified: 'cannot set flag paths=import: no such flag -paths')")
			}
			return
		}
	}
	t.Fatal("openapiv2 plugin missing from grpc-gateway template")
}

func TestGatewayTemplateHasLinterBlock(t *testing.T) {
	l := initTemplates["grpc-gateway"].Linter
	if l == nil || l.Rules == nil || len(l.Rules.EnabledRules) == 0 {
		t.Fatal("grpc-gateway template must ship the curated AIP rule block")
	}
	for _, want := range []string{"core::0131::http-method", "core::0131::http-body", "core::0235::plural-method-name"} {
		if !strings.Contains(strings.Join(l.Rules.EnabledRules, ","), want) {
			t.Errorf("enabled_rules missing %q", want)
		}
	}
}

func TestBasicMinimalNoLinter(t *testing.T) {
	for _, name := range []string{"basic", "minimal"} {
		if initTemplates[name].Linter != nil {
			t.Errorf("template %q must not ship disabled_rules:[all] (silences lint for generic projects)", name)
		}
	}
}
