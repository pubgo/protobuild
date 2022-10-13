package internal

import "flag"

const version = "v0.1.0"

var path string
var testDir string
var genGin bool
var enableLava bool
var Flags flag.FlagSet

func init() {
	Flags.BoolVar(&genGin, "gin", false, "generate gin api")
	Flags.BoolVar(&enableLava, "lava", false, "enable lava gen")
	Flags.StringVar(&path, "path", "", "*.pb.go root dir")
	Flags.StringVar(&testDir, "testDir", "docs/http", "*.http root dir")
}
