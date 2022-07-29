package plugin

import (
	"bytes"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

// StripParam strips a named param from req.
func StripParam(req *pluginpb.CodeGeneratorRequest, p string) {
	if req.Parameter == nil {
		return
	}

	v := stripParam(*req.Parameter, p)
	req.Parameter = &v

}

func stripParam(s, p string) string {
	var b strings.Builder
	for _, param := range strings.Split(s, ",") {
		if strings.SplitN(param, "=", 2)[0] != p {
			if b.Len() > 0 {
				b.WriteString(",")
			}
			b.WriteString(param)
		}
	}
	return b.String()
}

// RunPlugin runs a protoc plugin named "protoc-gen-$plugin"
// and returns the generated CodeGeneratorResponse or an error.
// Supply a non-nil stderr to override stderr on the called plugin.
func RunPlugin(plugin string, req *pluginpb.CodeGeneratorRequest, stderr io.Writer) (*pluginpb.CodeGeneratorResponse, error) {
	if stderr == nil {
		stderr = os.Stderr
	}

	// Marshal the CodeGeneratorRequest.
	b, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Call the plugin with the modified CodeGeneratorRequest.
	var buf bytes.Buffer
	cmd := exec.Command("protoc-gen-" + plugin)
	cmd.Stdin = bytes.NewReader(b)
	cmd.Stdout = &buf
	cmd.Stderr = stderr
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	// Read the CodeGeneratorResponse.
	var res pluginpb.CodeGeneratorResponse
	err = proto.Unmarshal(buf.Bytes(), &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// ReadRequest Unmarshal CodeGeneratorRequest.
func ReadRequest() (*pluginpb.CodeGeneratorRequest, error) {
	in, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	req := &pluginpb.CodeGeneratorRequest{}
	err = proto.Unmarshal(in, req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// WriteResponse marshals and writes CodeGeneratorResponse res to w.
func WriteResponse(w io.Writer, res *pluginpb.CodeGeneratorResponse) error {
	out, err := proto.Marshal(res)
	if err != nil {
		return err
	}

	_, err = w.Write(out)
	return err
}

func readDesc(path string) (*descriptor.FileDescriptorSet, error) {
	var desc descriptor.FileDescriptorSet

	p, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := proto.Unmarshal(p, &desc); err != nil {
		log.Fatalln(err)
	}

	return &desc, nil
}

// stabilize outputs the merged protobuf descriptor set into the provided writer.
//
// This is equivalent to the following command:
//
// cat merged.pb | protoc -I /path/to --decode google.protobuf.FileDescriptorSet /path/to/google/protobuf/descriptor.proto
//func marshalTo(w io.Writer) error {
//	p, err := proto.Marshal(&d.merged)
//	if err != nil {
//		return err
//	}
//
//	args := []string{
//		"protoc",
//		"-I",
//		d.includeDir,
//		"--decode",
//		"google.protobuf.FileDescriptorSet",
//		d.descProto,
//	}
//
//	cmd := exec.Command(args[0], args[1:]...)
//	cmd.Stdin = bytes.NewReader(p)
//	cmd.Stdout = w
//	cmd.Stderr = os.Stderr
//
//	if !quiet {
//		fmt.Println(strings.Join(args, " "))
//	}
//	return cmd.Run()
//}
