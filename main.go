// Package main is the entry point for protobuild.
package main

import (
	_ "embed"
	"runtime/debug"

	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/buildinfo/version"

	"github.com/pubgo/protobuild/cmd/protobuild"
)

//go:embed .version/VERSION
var ver string

// version meta comes from two sources:
//   - .version/VERSION, committed before tagging and embedded here
//   - go build vcs stamping (vcs.revision), present when building from a git
//     checkout, e.g. `go install` and goreleaser builds
//
// ldflags -X cannot inject anything here: funk v2 keeps these as unexported
// vars only reachable through the SetXxx setters.
func init() {
	version.SetVersion(ver)
	version.SetProject("protobuild")
	version.SetCommitID(commitID())
}

func commitID() string {
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			if s.Key == "vcs.revision" && len(s.Value) >= 7 {
				return s.Value[:7]
			}
		}
	}
	return ""
}

func main() {
	assert.Exit(protobuild.Main().Invoke().WithOS().Run())
}
