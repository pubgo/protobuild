// Note: 本项目主要思路和代码来源于protoc-gen-gotag, 感谢srikrsna

package main

import (
	"flag"
	"github.com/pubgo/protobuild/protoc-gen-gotag/internal/retag"

	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	var (
		flags flag.FlagSet
	)

	protogen.Options{ParamFunc: flags.Set}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = gengo.SupportedFeatures
		var originFiles []*protogen.GeneratedFile
		for _, f := range gen.Files {
			if f.Generate {
				originFiles = append(originFiles, gengo.GenerateFile(gen, f))
			}
		}

		retag.Rewrite(gen)

		for _, f := range originFiles {
			f.Skip()
		}
		return nil
	})
}
