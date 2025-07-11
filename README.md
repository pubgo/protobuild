# protobuild
> protobuf build and management tool

## install

go install github.com/pubgo/protobuild@latest

## example

[protobuf.yaml](./protobuf.test.yaml)

```yaml
protobuild -f protobuf.yaml vendor
protobuild -f protobuf.yaml gen
```

## lint

    protobuild lint -c protobuf.test.yaml

## format

     protobuild format -c protobuf.test.yaml

## protoc-gen-[plugin]

- istio.io/tools/cmd/protoc-gen-deepcopy@latest
- google.golang.org/protobuf/cmd/protoc-gen-go@latest
- github.com/mitchellh/protoc-gen-go-json@v1.1.0
- github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.10.2
- github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.10.2
- google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
- github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.5.1
- github.com/pubgo/protobuild/protoc-gen-retag@latest
- https://github.com/bufbuild/protobuf-es
- https://github.com/timostamm/protobuf-ts
- https://github.com/bufbuild/protoc-gen-validate
- https://github.com/istio/tools/tree/master/cmd/protoc-gen-golang-jsonshim
- https://github.com/istio/tools/tree/master/cmd/protoc-gen-golang-deepcopy
- https://github.com/istio/tools/tree/master/cmd/protoc-gen-docs
- https://github.com/solo-io/protoc-gen-openapi
- https://github.com/searKing/golang/tree/master/tools/protoc-gen-go-tag
- 