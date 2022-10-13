// Note: 本项目主要思路和代码来源于protoc-gen-gotag, 感谢srikrsna

package main

import (
	pgs "github.com/lyft/protoc-gen-star/v2"
	pgsgo "github.com/lyft/protoc-gen-star/v2/lang/go"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/protobuild/protoc-gen-retag/internal/retag"
)

func main() {
	pgs.Init(
		pgs.SupportedFeatures(generic.Ptr(uint64(1))),
	).RegisterModule(
		retag.New(),
	).RegisterPostProcessor(
		pgsgo.GoFmt(),
	).Render()
}
