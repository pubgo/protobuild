package main

import (
	"bytes"
	"github.com/pubgo/funk/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
	"io"
	"log"
	"os"
	"os/exec"
)

func main() {
	in := assert.Must1(io.ReadAll(os.Stdin))
	req := &pluginpb.CodeGeneratorRequest{}
	assert.Must(proto.Unmarshal(in, req))
	log.Println(req.GetParameter())

	cmd := exec.Command("/bin/sh", "-c", "protoc-gen-lava-errors")
	log.Println(cmd.Path)
	cmd.Env = os.Environ()
	var buf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &buf)
	cmd.Stdin = bytes.NewBuffer(in)
	cmd.Stderr = os.Stderr
	assert.Must(cmd.Run())
}
