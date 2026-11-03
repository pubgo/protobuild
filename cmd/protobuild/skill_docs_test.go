package protobuild

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCommandsDocumentedInSkillRefs guards skills/protobuild/references/commands.md
// against command-list drift: every subcommand registered in Main() must appear
// in the reference doc, wrapped in backticks.
func TestCommandsDocumentedInSkillRefs(t *testing.T) {
	ref := filepath.Join("..", "..", "skills", "protobuild", "references", "commands.md")
	data, err := os.ReadFile(ref)
	if err != nil {
		t.Fatalf("read %s: %v", ref, err)
	}
	md := string(data)
	for _, c := range Main().Children {
		name := strings.Fields(c.Use)[0]
		if !strings.Contains(md, "`"+name+"`") {
			t.Errorf("command %q missing from %s (match by backtick-wrapped name)", name, ref)
		}
	}
}
