package main

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	_ "github.com/jhump/protoreflect/desc/builder"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/desc/protoprint"
	_ "github.com/jhump/protoreflect/desc/protoprint"
	"github.com/pubgo/funk"

	_ "github.com/golang/protobuf/proto"
	_ "google.golang.org/protobuf/types/descriptorpb"
)

func main() {
	defer funk.RecoverAndExit()
	fd := funk.Must1(DescriptorSourceFromProtoFiles([]string{"proto/retag"}, "retag.proto"))
	//fmt.Println(fd[0].AsProto())
	//fmt.Println(fd[0].String())

	var pr = &protoprint.Printer{}

	var buf bytes.Buffer
	var ddd = fd[0].AsFileDescriptorProto()

	// 构建新的message
	var hh = funk.Must1(builder.NewMessage("hhhh").
		AddField(builder.NewField("id", builder.FieldTypeInt64())).
		AddField(builder.NewField("name", builder.FieldTypeString())).
		AddField(builder.NewField("options", builder.FieldTypeFixed64()).SetRepeated()).Build()).AsDescriptorProto()
	ddd.MessageType = append(ddd.MessageType, hh)

	md := funk.Must1(desc.LoadMessageDescriptorForMessage((*empty.Empty)(nil)))
	sb := builder.NewService("FooService").
		AddMethod(builder.NewMethod("DoSomething", builder.RpcTypeMessage(hh, false), builder.RpcTypeMessage(hh, false))).
		AddMethod(builder.NewMethod("ReturnThings", builder.RpcTypeImportedMessage(md, false), builder.RpcTypeMessage(hh, true)))

	ddd.Service = append(ddd.Service, funk.Must1(sb.Build()).AsServiceDescriptorProto())
	newFd := funk.Must1(desc.CreateFileDescriptor(ddd, fd[0].GetDependencies()...))
	funk.Must(pr.PrintProtoFile(newFd, &buf))
	fmt.Println(buf.String())
}

func DescriptorSourceFromProtoFiles(importPaths []string, fileNames ...string) ([]*desc.FileDescriptor, error) {
	p := protoparse.Parser{
		ImportPaths:      importPaths,
		InferImportPaths: len(importPaths) == 0,
	}
	return p.ParseFiles(fileNames...)
}
