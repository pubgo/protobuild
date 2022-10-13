// Note: 本项目主要思路和代码来源于protoc-gen-gotag, 感谢srikrsna

package main

import (
	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
	"github.com/pubgo/protobuild/protoc-gen-retag/internal/retag"
)

func main() {
	pgs.Init().
		RegisterModule(retag.New()).
		RegisterPostProcessor(pgsgo.GoFmt()).
		Render()
}
