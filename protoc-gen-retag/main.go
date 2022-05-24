package main

import (
	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"

	"github.com/pubgo/protobuild/protoc-gen-retag/retag"
)

func main() {
	pgs.Init(pgs.DebugEnv("DEBUG")).
		RegisterModule(retag.New()).
		RegisterPostProcessor(pgsgo.GoFmt()).
		Render()
}
