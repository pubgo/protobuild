// Refer: https://github.com/emicklei/proto-contrib/tree/master/cmd/protofmt
// https://github.com/bufbuild/buf/blob/main/private/buf/bufformat/bufformat.go

package format

import (
	"bytes"
	"io"
	"os"

	"github.com/emicklei/proto"
	"github.com/emicklei/proto-contrib/pkg/protofmt"
	"github.com/pubgo/funk/assert"
)

var overwrite = false

// func fmtCmd() *cli.Command {
// 	return &cli.Command{
// 		Name:  "fmt",
// 		Usage: "格式化protobuf文件",
// 		Flags: typex.Flags{&cli.BoolFlag{
// 			Name:        "overwrite",
// 			Usage:       "write result to (source) file instead of stdout",
// 			Aliases:     typex.Strs{"w"},
// 			Value:       overwrite,
// 			Destination: &overwrite,
// 		}},
// 		Before: func(context *cli.Context) error {
// 			return parseConfig()
// 		},
// 		Action: func(ctx *cli.Context) error {
// 			protoList := make(map[string]bool)

// 			for i := range cfg.Root {
// 				if pathutil.IsNotExist(cfg.Root[i]) {
// 					logger.Info().Msgf("file %s not found", cfg.Root[i])
// 					continue
// 				}

// 				assert.Must(filepath.Walk(cfg.Root[i], func(path string, info fs.FileInfo, err error) error {
// 					if err != nil {
// 						return err
// 					}

// 					if info.IsDir() {
// 						return nil
// 					}

// 					if strings.HasSuffix(info.Name(), ".proto") {
// 						protoList[path] = true
// 						return nil
// 					}

// 					return nil
// 				}))
// 			}

// 			for name := range protoList {
// 				//_ = shutil.MustRun("clang-format", "-i", fmt.Sprintf("-style=google %s", name))
// 				//readFormatWrite(name)
// 				format.Format(name)
// 			}

// 			return nil
// 		},
// 	}
// }

func readFormatWrite(filename string) {
	// open for read
	file := assert.Must1(os.Open(filename))

	// buffer before write
	buf := new(bytes.Buffer)
	format1(filename, file, buf)

	if overwrite {
		// write back to input
		assert.Must(os.WriteFile(filename, buf.Bytes(), os.ModePerm))
	} else {
		// write to stdout
		buf.WriteString("\n================================================================================================\n")
		assert.Must1(io.Copy(os.Stdout, bytes.NewReader(buf.Bytes())))

	}
}

func format1(filename string, input io.Reader, output io.Writer) {
	parser := proto.NewParser(input)
	parser.Filename(filename)
	def := assert.Must1(parser.Parse())
	protofmt.NewFormatter(output, "  ").Format(def) // 2 spaces
}
