// Note: 本项目主要思路和代码来源于protoc-gen-go-tag
// https://github.com/searKing/golang/tree/master/tools/protoc-gen-go-tag

package main

import (
	"github.com/pubgo/protobuild/cmd/protoc-gen-retag/ast"

	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = gengo.SupportedFeatures
		var originFiles []*protogen.GeneratedFile
		for _, f := range gen.Files {
			if f.Generate {
				originFiles = append(originFiles, gengo.GenerateFile(gen, f))
			}
		}
		ast.Rewrite(gen)

		for _, f := range originFiles {
			f.Skip()
		}
		return nil
	})
}
