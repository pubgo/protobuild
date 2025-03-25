module github.com/pubgo/protobuild

go 1.23.0

replace (
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20250324211829-b45e905df463
	google.golang.org/genproto/googleapis/api => google.golang.org/genproto/googleapis/api v0.0.0-20250324211829-b45e905df463
	google.golang.org/genproto/googleapis/rpc => google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463
)

require (
	github.com/bufbuild/protocompile v0.5.1
	github.com/dave/jennifer v1.6.0
	github.com/emicklei/proto v1.11.0
	github.com/emicklei/proto-contrib v0.11.0
	github.com/fatih/structtag v1.2.0
	github.com/flosch/pongo2/v5 v5.0.0
	github.com/golang/protobuf v1.5.4
	github.com/hashicorp/go-version v1.6.0
	github.com/lyft/protoc-gen-star/v2 v2.0.4-0.20230330145011-496ad1ac90a4
	github.com/open2b/scriggo v0.56.1
	github.com/pubgo/funk v0.5.49
	github.com/searKing/golang/go v1.2.115
	github.com/spf13/cast v1.5.0
	github.com/urfave/cli/v3 v3.0.0-alpha9.0.20240717192922-127cf54fac9f
	github.com/yuin/goldmark v1.4.13
	go.uber.org/multierr v1.11.0
	golang.org/x/mod v0.24.0
	google.golang.org/genproto/googleapis/api v0.0.0-20250227231956-55c901821b1e
	google.golang.org/grpc v1.71.0
	google.golang.org/protobuf v1.36.6
	gorm.io/gorm v1.24.5
)

require (
	github.com/alecthomas/repr v0.4.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/k0kubun/pp/v3 v3.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/phuslu/goid v1.0.0 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	github.com/spf13/afero v1.10.0 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/tools v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250313205543-e70fdf4c4cb4 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
