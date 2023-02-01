package main

import (
	"os"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/protobuild/cmd"
)

func main() {
	assert.ExitFn(func() error {
		return cmd.Main().Run(os.Args)
	})
}
