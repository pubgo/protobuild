package internal

import (
	"fmt"
	"text/template"

	pgs "github.com/lyft/protoc-gen-star/v2"
	pgsgo "github.com/lyft/protoc-gen-star/v2/lang/go"
	"github.com/pubgo/funk/logx"
)

var logger = logx.WithName("gorm")

type mod struct {
	*pgs.ModuleBase
	ctx pgsgo.Context
	tpl *template.Template
}

func New() pgs.Module {
	return &mod{
		ModuleBase: new(pgs.ModuleBase),
	}
}

func (m *mod) InitContext(c pgs.BuildContext) {
	m.ModuleBase.InitContext(c)
	m.ctx = pgsgo.InitContext(c.Parameters())

	m.tpl = template.New("").Funcs(map[string]interface{}{
		"package":     m.ctx.PackageName,
		"name":        m.ctx.Name,
		"marshaler":   m.marshaler,
		"unmarshaler": m.unmarshaler,
	})
	m.tpl = template.Must(m.tpl.Parse(jsonifyTpl))
}

func (m *mod) marshaler(m1 pgs.Message) pgs.Name {
	return m.ctx.Name(m1) + "JSONMarshaler"
}

func (m *mod) unmarshaler(m1 pgs.Message) pgs.Name {
	return m.ctx.Name(m1) + "JSONUnmarshaler"
}

func (*mod) Name() string { return "gorm" }

func (m *mod) Execute(targets map[string]pgs.File, packages map[string]pgs.Package) []pgs.Artifact {
	for _, f := range targets {
		if len(f.Messages()) == 0 {
			continue
		}

		filename := m.ctx.OutputPath(f).SetExt(".gorm.go")
		logger.Info(fmt.Sprintf("gorm %s", filename))
		m.OverwriteGeneratorTemplateFile(filename.String(), m.tpl, f)
	}
	return m.Artifacts()
}

const jsonifyTpl = `package {{ package . }}
import (
	"bytes"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
)
{{ range .AllMessages }}
// {{ marshaler . }} describes the default jsonpb.Marshaler used by all 
// instances of {{ name . }}. This struct is safe to replace or modify but 
// should not be done so concurrently.
var {{ marshaler . }} = new(jsonpb.Marshaler)
// MarshalJSON satisfies the encoding/json Marshaler interface. This method 
// uses the more correct jsonpb package to correctly marshal the message.
func (m *{{ name . }}) MarshalJSON() ([]byte, error) {
	if m == nil {
		return json.Marshal(nil)
	}
	buf := &bytes.Buffer{}
	if err := {{ marshaler . }}.Marshal(buf, m); err != nil {
	  return nil, err
	}
	return buf.Bytes(), nil
}
var _ json.Marshaler = (*{{ name . }})(nil)
// {{ unmarshaler . }} describes the default jsonpb.Unmarshaler used by all 
// instances of {{ name . }}. This struct is safe to replace or modify but 
// should not be done so concurrently.
var {{ unmarshaler . }} = new(jsonpb.Unmarshaler)
// UnmarshalJSON satisfies the encoding/json Unmarshaler interface. This method 
// uses the more correct jsonpb package to correctly unmarshal the message.
func (m *{{ name . }}) UnmarshalJSON(b []byte) error {
	return {{ unmarshaler . }}.Unmarshal(bytes.NewReader(b), m)
}
var _ json.Unmarshaler = (*{{ name . }})(nil)
{{ end }}
`
