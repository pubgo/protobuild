package retag

import (
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/fatih/structtag"
	pgs "github.com/lyft/protoc-gen-star/v2"
	pgsgo "github.com/lyft/protoc-gen-star/v2/lang/go"
	"github.com/pubgo/funk/log"
)

var logger = log.GetLogger("retag")

type mod struct {
	*pgs.ModuleBase
	pgsgo.Context
}

func New() pgs.Module {
	return &mod{
		ModuleBase: new(pgs.ModuleBase),
	}
}

func (m *mod) InitContext(c pgs.BuildContext) {
	m.ModuleBase.InitContext(c)
	m.Context = pgsgo.InitContext(c.Parameters())
}

func (*mod) Name() string { return "retag" }

func (m *mod) Execute(targets map[string]pgs.File, packages map[string]pgs.Package) []pgs.Artifact {
	xtv := m.Parameters().Str("xxx")

	xtv = strings.Replace(xtv, "+", ":", -1)

	xt, err := structtag.Parse(xtv)
	m.CheckErr(err)

	autoTag := m.Parameters().Str("auto")
	var autoTags []string
	if autoTag != "" {
		autoTags = strings.Split(autoTag, "+")
	}

	module := m.Parameters().Str("module")

	extractor := newTagExtractor(m, m.Context, autoTags)

	for _, f := range targets {
		tags := extractor.Extract(f)

		tags.AddTagsToXXXFields(xt)

		gfName := m.Context.OutputPath(f).SetExt(".go").String()

		output := m.Parameters().Str("output")
		filename := gfName
		if output != "" {
			filename = filepath.Join(output, gfName)
		}

		if module != "" {
			filename = strings.ReplaceAll(filename, string(filepath.Separator), "/")
			trim := module + "/"
			if !strings.HasPrefix(filename, trim) {
				m.Debug(fmt.Sprintf("%v: generated file does not match prefix %q", filename, module))
				m.Exit(1)
			}
			filename = strings.TrimPrefix(filename, trim)
		}

		logger.Info().Msgf("retag file: %s", filename)

		fs := token.NewFileSet()
		fn, err := parser.ParseFile(fs, filename, nil, parser.ParseComments)
		m.CheckErr(err)

		m.CheckErr(Retag(fn, tags))

		var buf strings.Builder
		m.CheckErr(printer.Fprint(&buf, fs, fn))

		m.OverwriteGeneratorFile(filename, buf.String())
	}

	return m.Artifacts()
}
