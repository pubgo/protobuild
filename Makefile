WORKDIR=`pwd`
Project=protobuild
Base=github.com/pubgo/protobuild
Tag=$(shell git describe --abbrev=0 --tags)
Version=$(shell git tag --sort=committerdate | tail -n 1)
BuildTime=$(shell date "+%F %T")
CommitID=$(shell git rev-parse --short=8 HEAD)

LDFLAGS=-ldflags " \
-X '${Base}/pkg/version.BuildTime=${BuildTime}' \
-X '${Base}/pkg/version.CommitID=${CommitID}' \
-X '${Base}/pkg/version.Version=${Version}' \
-X '${Base}/pkg/version.Tag=${Tag}' \
-X '${Base}/pkg/version.Project=${Project}' \
-X '${Base}/pkg/version.Data=hello' \
"

.PHONY: build
build:
	go build ${LDFLAGS} -mod vendor -v -o main *.go

vet:
	@go vet ./...

generate:
	@go generate ./...

lint:
	@golangci-lint run --skip-dirs-use-default --timeout 3m0s
