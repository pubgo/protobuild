package main

import (
	"os"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"

	"github.com/pubgo/protobuild/cmd"
)

func main() {
	defer recovery.Exit()
	assert.Must(cmd.Main().Run(os.Args))
}
