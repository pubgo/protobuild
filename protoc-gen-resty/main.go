package main

import (
	"flag"

	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/protobuild/internal/protoutil"
	"github.com/pubgo/protobuild/protoc-gen-resty/internal"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	defer recovery.Exit()

	var flags flag.FlagSet
	flags.StringVar(&internal.PathTag, "path-tag", internal.PathTag, "router path params tag")
	flags.StringVar(&internal.QueryTag, "query-tag", internal.QueryTag, "router path query tag")

	if protoutil.IsHelp() {
		flags.PrintDefaults()
		return
	}

	opts := &protogen.Options{ParamFunc: flags.Set}
	opts.Run(func(gen *protogen.Plugin) error {
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
