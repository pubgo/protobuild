package format

import (
	"bytes"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
	"github.com/pubgo/funk/assert"
	"os"
)

// Format formats and writes the target module files into a read bucket.
func Format(path string) error {
	var data = assert.Must1(os.ReadFile(path))
	fileNode, err := parser.Parse(path, bytes.NewBuffer(data), reporter.NewHandler(nil))
	assert.Must(err)

	var buf bytes.Buffer
	assert.Must(newFormatter(&buf, fileNode).Run())
	assert.Must(os.WriteFile(path, buf.Bytes(), 0644))
	return nil
}
