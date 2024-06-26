package format

import (
	"bytes"
	"os"

	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
	"github.com/pubgo/funk/assert"
)

// Format formats and writes the target module files into a read bucket.
func Format(path string) {
	data := assert.Must1(os.ReadFile(path))
	fileNode := assert.Must1(parser.Parse(path, bytes.NewBuffer(data), reporter.NewHandler(nil)))

	var buf bytes.Buffer
	assert.Must(newFormatter(&buf, fileNode).Run())
	assert.Must(os.WriteFile(path, buf.Bytes(), 0o644))
}
