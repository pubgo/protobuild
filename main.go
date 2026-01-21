// Package main is the entry point for protobuild.
package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/pubgo/protobuild/cmd/protobuild"
)

//go:embed .version/VERSION
var version string

func main() {
	err := protobuild.Main(version).Invoke().WithOS().Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
