package main

import (
	"context"
	_ "embed"
	"os"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/protobuild/cmd/protobuild"
)

//go:embed .version/VERSION
var version string

func main() {
	assert.ExitFn(func() error {
		return protobuild.Main(version).
			Run(context.Background(), os.Args)
	})
}
