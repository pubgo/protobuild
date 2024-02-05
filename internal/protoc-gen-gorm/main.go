package main

import (
	"github.com/pubgo/protobuild/internal/protoc-gen-gorm/internal"
	_ "github.com/spf13/cast"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	_ "google.golang.org/protobuf/types/known/wrapperspb"
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			internal.GenerateFile(gen, f)
		}
		return nil
	})
}
