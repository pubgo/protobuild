package protobuild

import (
	"os"
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
	if !strings.Contains(content, "module: example.com/demo/pkg") {
		t.Errorf("config must contain detected module, got:\n%s", content)
	}
	if strings.Contains(content, "protocolbuffers") {
		t.Error("config must not contain WKT github dep")
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
