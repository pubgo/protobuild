WORKDIR=`pwd`
Project=protobuild
VersionBase=github.com/pubgo/protobuild
Tag=$(shell git describe --abbrev=0 --tags)
Version=$(shell git tag --sort=committerdate | tail -n 1)
BuildTime=$(shell date "+%F %T")
CommitID=$(shell git rev-parse --short=8 HEAD)
GOPATH=$(shell go env GOPATH )

LDFLAGS=-ldflags " \
-X '${VersionBase}/version.BuildTime=${BuildTime}' \
-X '${VersionBase}/version.CommitID=${CommitID}' \
-X '${VersionBase}/version.Version=${Version}' \
-X '${VersionBase}/version.Tag=${Tag}' \
-X '${VersionBase}/version.Project=${Project}' \
-X '${VersionBase}/version.Data=hello' \
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
