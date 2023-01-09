// Refer: https://github.com/emicklei/proto-contrib/tree/master/cmd/protofmt
// clang-format -style=google *.proto
package cmd

import (
	"bytes"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/emicklei/proto"
	"github.com/emicklei/proto-contrib/pkg/protofmt"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/protobuild/internal/typex"
	"github.com/pubgo/x/pathutil"
	"github.com/urfave/cli/v2"
)

var (
	overwrite = false
)

func fmtCmd() *cli.Command {
	return &cli.Command{
		Name:  "fmt",
		Usage: "格式化protobuf文件",
		Flags: typex.Flags{&cli.BoolFlag{
			Name:        "overwrite",
			Usage:       "write result to (source) file instead of stdout",
			Aliases:     typex.Strs{"w"},
			Value:       overwrite,
			Destination: &overwrite,
		}},
		Action: func(ctx *cli.Context) error {
			var protoList = make(map[string]bool)

			for i := range cfg.Root {
				if pathutil.IsNotExist(cfg.Root[i]) {
					logger.Info().Msgf("file %s not found", cfg.Root[i])
					continue
				}

				assert.Must(filepath.Walk(cfg.Root[i], func(path string, info fs.FileInfo, err error) error {
					if err != nil {
						return err
					}

					if info.IsDir() {
						return nil
					}

					if strings.HasSuffix(info.Name(), ".proto") {
						protoList[path] = true
						return nil
					}

					return nil
				}))
			}

			for name := range protoList {
				readFormatWrite(name)
			}

			return nil
		},
	}
}

func readFormatWrite(filename string) {
	// open for read
	file := assert.Must1(os.Open(filename))

	// buffer before write
	buf := new(bytes.Buffer)
	format(filename, file, buf)

	if overwrite {
		// write back to input
		assert.Must(ioutil.WriteFile(filename, buf.Bytes(), os.ModePerm))
	} else {
		// write to stdout
		buf.WriteString("\n================================================================================================\n")
		assert.Must1(io.Copy(os.Stdout, bytes.NewReader(buf.Bytes())))

	}
}

func format(filename string, input io.Reader, output io.Writer) {
	parser := proto.NewParser(input)
	parser.Filename(filename)
	def := assert.Must1(parser.Parse())
	protofmt.NewFormatter(output, "  ").Format(def) // 2 spaces
}
