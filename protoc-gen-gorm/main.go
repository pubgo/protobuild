package main

import (
	pgs "github.com/lyft/protoc-gen-star/v2"
	pgsgo "github.com/lyft/protoc-gen-star/v2/lang/go"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/protobuild/protoc-gen-gorm/internal"
)

func main() {
	pgs.Init(
		pgs.SupportedFeatures(generic.Ptr(uint64(1))),
	).RegisterModule(
		internal.New(),
	).RegisterPostProcessor(
		pgsgo.GoFmt(),
	).Render()
}
