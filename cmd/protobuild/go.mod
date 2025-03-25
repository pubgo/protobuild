module github.com/pubgo/protobuild/cmd/protobuild

go 1.24

replace github.com/pubgo/protobuild => ../../

replace (
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20250324211829-b45e905df463
	google.golang.org/genproto/googleapis/rpc => google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463
)

require (
	github.com/a8m/envsubst v1.4.3
	github.com/cnf/structhash v0.0.0-20201127153200-e1b16c1ebc08
	github.com/deckarep/golang-set/v2 v2.6.0
	github.com/huandu/go-clone v1.5.1
	github.com/pubgo/funk v0.5.49
	github.com/pubgo/protobuild v0.0.0-00010101000000-000000000000
	github.com/samber/lo v1.47.0
	github.com/urfave/cli/v3 v3.0.0-alpha9.0.20240717192922-127cf54fac9f
	golang.org/x/mod v0.24.0
	google.golang.org/protobuf v1.36.6
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/alecthomas/repr v0.4.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/k0kubun/pp/v3 v3.2.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/phuslu/goid v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250313205543-e70fdf4c4cb4 // indirect
	google.golang.org/grpc v1.71.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
