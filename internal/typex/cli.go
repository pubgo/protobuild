package typex

import (
	"github.com/urfave/cli/v3"
)

type (
	Strs     = []string
	Flags    = []cli.Flag
	Command  = cli.Command
	Commands = []*cli.Command
)
