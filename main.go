// Package main is the entry point for protobuild.
package main

import (
	_ "embed"

	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/buildinfo/version"

	"github.com/pubgo/protobuild/cmd/protobuild"
)

//go:embed .version/VERSION
var ver string
var _ = version.SetVersion(ver)

func main() {
	assert.Exit(protobuild.Main().Invoke().WithOS().Run())
}
