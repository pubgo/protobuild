package protobuild

import (
	"testing"
	"time"
)

func TestParseCleanDuration(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
		err  bool
	}{
		{"30d", 30 * 24 * time.Hour, false},
		{"1d", 24 * time.Hour, false},
		{"720h", 720 * time.Hour, false},
		{"45m", 45 * time.Minute, false},
		{"", 0, true},
		{"-5d", 0, true},
		{"abc", 0, true},
		{"0d", 0, true},
	}

	for _, c := range cases {
		got, err := parseCleanDuration(c.in)
		if c.err {
			if err == nil {
				t.Errorf("parseCleanDuration(%q): expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseCleanDuration(%q): unexpected error %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseCleanDuration(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseCleanSize(t *testing.T) {
	cases := []struct {
		in   string
		want int64
		err  bool
	}{
		{"500MB", 500e6, false},
		{"1GB", 1e9, false},
		{"10GiB", 10 << 30, false},
		{"1024KB", 1024e3, false},
		{"512B", 512, false},
		{"1048576", 1048576, false},
		{"500mb", 500e6, false}, // case-insensitive
		{"", 0, true},
		{"-1GB", 0, true},
		{"abc", 0, true},
		{"MB", 0, true},
	}

	for _, c := range cases {
		got, err := parseCleanSize(c.in)
		if c.err {
			if err == nil {
				t.Errorf("parseCleanSize(%q): expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseCleanSize(%q): unexpected error %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseCleanSize(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestPlanCacheRemoval(t *testing.T) {
	now := time.Now()
	entries := []cacheEntry{
		{name: "git/old", path: "git/old", size: 100, modTime: now.Add(-40 * 24 * time.Hour)},
		{name: "git/new", path: "git/new", size: 100, modTime: now.Add(-1 * time.Hour)},
		{name: "git/mid", path: "git/mid", size: 100, modTime: now.Add(-10 * 24 * time.Hour)},
	}

	// age policy only: 40d-old entry removed, 10d-old survives a 30d cutoff
	remove := planCacheRemoval(entries, now.Add(-30*24*time.Hour), 0)
	if len(remove) != 1 || remove[0].name != "git/old" {
		t.Fatalf("age policy: want [git/old], got %v", remove)
	}

	// size policy only: 300 total > 150 limit -> oldest entries removed,
	// result is deterministic by name for display
	remove = planCacheRemoval(entries, time.Time{}, 150)
	if len(remove) != 2 {
		t.Fatalf("size policy: want 2 removed, got %d: %v", len(remove), remove)
	}
	got := map[string]bool{}
	for _, e := range remove {
		got[e.name] = true
	}
	if !got["git/old"] || !got["git/mid"] {
		t.Fatalf("size policy: want git/old + git/mid removed, got %v", remove)
	}

	// combined: age removes git/old; remainder (200B) still exceeds the 150B
	// limit, so the oldest survivor (git/mid) is trimmed as well
	remove = planCacheRemoval(entries, now.Add(-30*24*time.Hour), 150)
	if len(remove) != 2 {
		t.Fatalf("combined policy: want 2 removed, got %d: %v", len(remove), remove)
	}
	got = map[string]bool{}
	for _, e := range remove {
		got[e.name] = true
	}
	if !got["git/old"] || !got["git/mid"] {
		t.Fatalf("combined policy: want git/old + git/mid removed, got %v", remove)
	}

	// no policy -> nothing removed
	if remove = planCacheRemoval(entries, time.Time{}, 0); len(remove) != 0 {
		t.Fatalf("no policy: want [], got %v", remove)
	}
}

func TestPickLatestSemver(t *testing.T) {
	cases := []struct {
		tags  []string
		want  string
		valid bool
	}{
		{[]string{"v0.0.5", "v0.0.10", "v0.0.2"}, "v0.0.10", true},
		{[]string{"v1.2.3-rc1", "v1.2.2"}, "v1.2.2", true},         // prefers stable
		{[]string{"v1.2.3-rc1"}, "v1.2.3-rc1", true},               // prerelease fallback
		{[]string{"1.2.3", "1.10.0"}, "1.10.0", true},              // no v prefix
		{[]string{"release-1", "main"}, "", false},                 // no semver tags
		{[]string{"v0.0.5", "v0.0.5^{}"}, "v0.0.5", true},          // peeled dup handled upstream
		{[]string{"v2.0.0", "v10.0.0", "v1.9.0"}, "v10.0.0", true}, // numeric order, not lexical
	}

	for _, c := range cases {
		got, err := pickLatestSemver(c.tags)
		if !c.valid {
			if err == nil {
				t.Errorf("pickLatestSemver(%v): expected error", c.tags)
			}
			continue
		}
		if err != nil {
			t.Errorf("pickLatestSemver(%v): unexpected error %v", c.tags, err)
			continue
		}
		if got != c.want {
			t.Errorf("pickLatestSemver(%v) = %q, want %q", c.tags, got, c.want)
		}
	}
}

func TestToGitURL(t *testing.T) {
	cases := map[string]string{
		"github.com/x/y":             "https://github.com/x/y",
		"git::github.com/x/y":        "https://github.com/x/y",
		"https://github.com/x/y":     "https://github.com/x/y",
		"git@github.com:x/y.git":     "git@github.com:x/y.git",
		"https://gitlab.com/x/y.git": "https://gitlab.com/x/y.git",
	}
	for in, want := range cases {
		if got := toGitURL(in); got != want {
			t.Errorf("toGitURL(%q) = %q, want %q", in, got, want)
		}
	}
}
