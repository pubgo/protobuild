package internal

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pubgo/funk/errors"

	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/protobuild/internal/protoutil"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
)

var (
	contextCall = protoutil.Import("context")
	reflectCall = protoutil.Import("reflect")
	restyCall   = protoutil.Import("github.com/go-resty/resty/v2")
	jsonCall    = protoutil.Import("github.com/goccy/go-json")
	stringsCall = protoutil.Import("strings")
	lavaCall    = protoutil.Import("github.com/pubgo/lava/proto/lava")
)

// GenerateFile generates a .lava.pb.go file containing service definitions.
func GenerateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Services) == 0 {
		return nil
	}

	filename := file.GeneratedFilenamePrefix + ".resty.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-resty. DO NOT EDIT.")
	g.P("// versions:")
	g.P("// - protoc-gen-resty ", version)
	g.P("// - protoc           ", protocVersion(gen))
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()

	g.QualifiedGoIdent(restyCall(""))
	g.QualifiedGoIdent(reflectCall(""))
	g.QualifiedGoIdent(jsonCall(""))

	generateFileContent(gen, file, g)
	return g
}

// generateFileContent generates the service definitions, excluding the package statement.
func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	for _, srv := range file.Services {
		genClient(gen, file, g, srv)
	}
}

func genClient(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, srv *protogen.Service) (ret bool) {
	defer recovery.Raise(func(err errors.XErr) {
		ret = false
	})

	clientName := srv.GoName + "Resty"
	g.P("type ", clientName, " interface {")
	for _, method := range srv.Methods {
		if method.Desc.Options().(*descriptorpb.MethodOptions).GetDeprecated() {
			g.P(deprecationComment)
		}
		g.P(method.Comments.Leading, clientSignature(g, method))
	}
	g.P("}")
	g.P()

	g.P("func New", clientName, " (client *", restyCall("Client"), ") ", clientName, " {")
	g.P(`client.SetContentLength(true)`)
	g.P("return &", unExport(clientName), "{client: client}")
	g.P("}")
	g.P()

	// Client structure.
	g.P("type ", unExport(clientName), " struct {")
	g.P("client *", restyCall("Client"))
	g.P("}")
	g.P()

	for _, method := range srv.Methods {
		genClientMethod(gen, file, g, method, 0)
	}
	return
}

func genClientMethod(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, method *protogen.Method, index int) {
	service := method.Parent
	hr, err := protoutil.ExtractAPIOptions(method.Desc)
	if err != nil || hr == nil {
		replacer := strings.NewReplacer(".", "/", "-", "/")
		hr = protoutil.DefaultAPIOptions(replacer.Replace(string(file.Desc.Package())), service.GoName, method.GoName)
	}
	mth, path := protoutil.ExtractHttpMethod(hr)

	if method.Desc.Options().(*descriptorpb.MethodOptions).GetDeprecated() {
		g.P(deprecationComment)
	}

	mth = strings.ToUpper(mth)

	g.P("func (c *", unExport(service.GoName), "Resty) ", clientSignature(g, method), "{")
	g.P(`var req = c.client.R()
	if ctx != nil {
		req.SetContext(ctx)
	}`)
	g.P("for i := range opts {")
	g.P("opts[i](req)")
	g.P("}")

	switch mth {
	case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		g.P(`var body map[string]interface{}`)
	}

	g.P("if in != nil {")
	g.P(`var rv = reflect.ValueOf(in).Elem()`)
	g.P(`var rt = reflect.TypeOf(in).Elem()`)
	g.P(`for i := 0; i < rt.NumField(); i++ {`)

	// url param
	g.P(`if val,ok := rt.Field(i).Tag.Lookup("`, PathTag, `"); ok && val != ""{
			req.SetPathParam(val, rv.Field(i).String())
			continue
		}`)

	// url query
	g.P(`if val,ok := rt.Field(i).Tag.Lookup("`, QueryTag, `"); ok && val != ""{
			req.SetQueryParam(val, rv.Field(i).String())
			continue
		}`)

	switch mth {
	case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		g.P(`if body == nil {
				body = make(map[string]interface{})
			}`)
		g.P(`if val, ok := rt.Field(i).Tag.Lookup("json"); ok && val != "" {
				body[val] = rv.Field(i).String()
			}`)
	default:
		g.P(`if val,ok := rt.Field(i).Tag.Lookup("json"); ok && val != "" {
			req.SetQueryParam(val, rv.Field(i).String())
		}`)
	}

	g.P(`}}`)

	switch mth {
	case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		g.P("req.SetBody(body)")
	}

	g.P(`var resp, err = req.Execute("`, mth, `","`, path, `")`)
	g.P(`if err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	g.P("out := new(", method.Output.GoIdent, ")")
	for i := range method.Output.Fields {
		if method.Output.Fields[i].GoName == "Response" {
			g.P(`
			var headers = make(map[string]string)
			for k, v := range resp.Header() {
				headers[k] = `, stringsCall("Join"), `(v, ",")
			}
			out.Response = &`, lavaCall("Response"), `{
				Code:    int32(resp.StatusCode()),
				Headers: headers,
			}`)
		}
	}

	g.P(`if err := `, jsonCall("Unmarshal"), `(resp.Body(), out); err != nil {
		return nil, err
	}`)

	g.P(`return out, nil`)
	g.P("}")
	g.P()
}

func clientSignature(g *protogen.GeneratedFile, method *protogen.Method) string {
	g.QualifiedGoIdent(contextCall(""))
	return protoutil.Template(`
{{method}}(ctx context.Context, in *{{methodIn}}, opts ...func(req *{{methodOpt}})) (*{{methodOut}}, error)
`, protoutil.Context{
		"method":    method.GoName,
		"methodOpt": g.QualifiedGoIdent(restyCall("Request")),
		"methodIn":  g.QualifiedGoIdent(method.Input.GoIdent),
		"methodOut": g.QualifiedGoIdent(method.Output.GoIdent),
	})
}

func protocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}
