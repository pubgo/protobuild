package main

import (
	"os"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/protobuild/cmd/protobuild"
)

func main() {
	assert.ExitFn(func() error {
		return protobuild.Main().Run(os.Args)
	})
}
