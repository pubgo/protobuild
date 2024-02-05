// Note: 本项目主要思路和代码来源于protoc-gen-gotag, 感谢srikrsna

package main

// https://github.com/searKing/golang/blob/master/tools/protoc-gen-go-tag/main.go

import (
	pgs "github.com/lyft/protoc-gen-star/v2"
	pgsgo "github.com/lyft/protoc-gen-star/v2/lang/go"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/protobuild/cmd/protoc-gen-gotag/internal/retag"
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
