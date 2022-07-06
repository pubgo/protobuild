package main

import (
	"os"

	"github.com/pubgo/funk"
	"github.com/pubgo/protobuild/cmd"
)

func main() {
	defer funk.RecoverAndExit()
	funk.Must(cmd.Main().Run(os.Args))
}
