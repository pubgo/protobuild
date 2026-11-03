package protobuild

import (
	"strings"
	"testing"
)

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
