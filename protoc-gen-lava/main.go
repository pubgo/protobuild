package main

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	genLava "github.com/pubgo/protobuild/protoc-gen-lava/internal"
)

func main() {
	protogen.Options{ParamFunc: genLava.Flags.Set}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			genLava.GenerateFile(gen, f)
		}
		return nil
	})
}
