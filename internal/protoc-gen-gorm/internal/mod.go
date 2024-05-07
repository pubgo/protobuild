package internal

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/protobuild/internal/protoutil"
	ormpb "github.com/pubgo/protobuild/pkg/orm"
	"google.golang.org/protobuf/compiler/protogen"
	_ "google.golang.org/protobuf/encoding/protojson"
	gp "google.golang.org/protobuf/proto"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm/schema"
)

func init() {
	_ = schema.ParseTagSetting
}

var logger = log.GetLogger("gorm")

// GenerateFile generates a .lava.pb.go file containing service definitions.
func GenerateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	filename := file.GeneratedFilenamePrefix + ".gorm.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)

	genFile := jen.NewFile(string(file.GoPackageName))
	genFile.HeaderComment("Code generated by protoc-gen-gorm. DO NOT EDIT.")
	genFile.HeaderComment("versions:")
	genFile.HeaderComment(fmt.Sprintf("- protoc-gen-gorm %s", version))
	genFile.HeaderComment(fmt.Sprintf("- protoc          %s", protocVersion(gen)))
	if file.Proto.GetOptions().GetDeprecated() {
		genFile.HeaderComment(fmt.Sprintf("%s is a deprecated file.", file.Desc.Path()))
	} else {
		genFile.HeaderComment(fmt.Sprintf("source: %s", file.Desc.Path()))
	}

	genFile.Comment("This is a compile-time assertion to ensure that this generated file")
	genFile.Comment("is compatible with the grpc package it is being compiled against.")
	genFile.Comment("Requires gRPC-Go v1.32.0 or later.")
	genFile.Id("const _ =").Qual("google.golang.org/grpc", "SupportPackageIsVersion7")

	type table struct {
		name   string
		pkName string
		pkType protogen.GoIdent
		fields map[string]*Field
	}

	tables := make(map[string]*table)
	for i := range file.Messages {
		m := file.Messages[i]

		if m.Desc.Options() == nil {
			continue
		}

		opts, ok := gp.GetExtension(m.Desc.Options(), ormpb.E_Opts).(*ormpb.GormMessageOptions)
		if !ok || opts == nil || !opts.Enabled {
			continue
		}

		tableName := string(m.Desc.Name())
		if opts.Table != "" {
			tableName = opts.Table
		}
		tableName = protoutil.Name(tableName).LowerSnakeCase().String()

		if tables[tableName] == nil {
			tables[tableName] = &table{
				name:   protoutil.Name(string(m.Desc.Name()) + "Model").UpperCamelCase().String(),
				fields: make(map[string]*Field),
			}
		}

		for j := range m.Fields {
			ff := NewField(m.Fields[j], gen)
			tables[tableName].fields[ff.GoName] = ff

			if ff.tag != nil && ff.tag.Pk {
				tables[tableName].pkName = ff.Name
				tables[tableName].pkType = ff.GoType
			}
		}
	}

	for name, fields := range tables {
		for field := range fields.fields {
			logger.Info().
				Str("field", field).
				Str("table", name).
				Msg("table info")
		}
	}

	for i := range file.Services {
		srv := file.Services[i]
		if srv.Desc.Options() == nil {
			continue
		}

		name := protoutil.Name(srv.Desc.Name()).UpperCamelCase().String()

		logger.Info().Msg(string(srv.Desc.FullName()))
		opts, ok := gp.GetExtension(srv.Desc.Options(), ormpb.E_Server).(*ormpb.GormMessageOptions)
		if !ok || opts == nil || !opts.Service {
			continue
		}

		tb := tables[opts.Table]
		if tb == nil {
			panic(fmt.Sprintf("table [%s] not found", opts.Table))
		}

		srvName := fmt.Sprintf("%sGormHandler", name)

		genFile.Add(
			jen.Type().Id(srvName).InterfaceFunc(func(g *jen.Group) {
				for j := range srv.Methods {
					m := srv.Methods[j]
					code := jen.Id(m.GoName).
						Params(
							jen.Id("ctx").Qual("context", "Context"),
							jen.Id("req").Op("*").Id(m.Input.GoIdent.GoName),
							jen.Id("where").Op("...").Id("func(db *gorm.DB)*gorm.DB"),
						)

					if strings.HasPrefix(m.GoName, "Create") {
						code.Id("error")
					}

					if strings.HasPrefix(m.GoName, "Delete") {
						code.Id("error")
					}

					if strings.HasPrefix(m.GoName, "Update") {
						code.Id("error")
					}

					if strings.HasPrefix(m.GoName, "Get") {
						code.Params(jen.Op("*").Id(tables[opts.Table].name), jen.Id("error"))
					}

					if strings.HasPrefix(m.GoName, "List") {
						code.Params(jen.Index().Op("*").Id(tables[opts.Table].name), jen.Id("error"))
					}

					g.Add(code)
				}
			}).Line(),
		)

		genFile.Add(
			jen.Func().
				Id(fmt.Sprintf("New%s", srvName)).
				Params(
					jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB"),
				).Id(fmt.Sprintf("%sServer", name)).BlockFunc(func(g *jen.Group) {
				g.If(jen.Id("db == nil")).Block(jen.Id(`panic("gorm handler panic: db is nil")`)).Line()
				g.Return(jen.Op("&").Id(protoutil.Name(srvName).LowerCamelCase().String()).Block(
					jen.Id("db:db,"),
				))
			}),
		)

		genFile.Add(
			jen.Type().
				Id(protoutil.Name(srvName).LowerCamelCase().String()).
				Struct(
					jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB"),
				),
		)

		for j := range srv.Methods {
			m := srv.Methods[j]

			logger.Info().Str("name", m.GoName).
				Msg("service method")

			genFile.Add(
				jen.Func().
					Params(
						jen.Id("h").Op("*").Id(protoutil.Name(srvName).LowerCamelCase().String()),
					).
					Id(m.GoName).
					Params(
						jen.Id("ctx").Qual("context", "Context"),
						jen.Id("req").Op("*").Id(m.Input.GoIdent.GoName),
					).
					Params(
						jen.Op("*").Id(m.Output.GoIdent.GoName),
						jen.Id("error"),
					).
					BlockFunc(func(g *jen.Group) {
						g.Id("var db = h.db.WithContext(ctx)")
						g.Id("_ = db")

						g.Var().Id("rsp").Op("=").New(jen.Id(m.Output.GoIdent.GoName))

						if strings.HasPrefix(m.GoName, "Create") {
						}

						if strings.HasPrefix(m.GoName, "Delete") {
						}

						if strings.HasPrefix(m.GoName, "Update") {
						}

						if strings.HasPrefix(m.GoName, "Get") {
						}

						if strings.HasPrefix(m.GoName, "List") {
							g.Var().Id("err = db.Find(&rsp.Data).Error")
							g.If(jen.Id("err").Op("!=").Nil()).Block(jen.Return(jen.Id("nil,err"))).Line()
						}

						g.Return(jen.Id("rsp"), jen.Nil())
					}).Line(),
			)
		}

	}

	for i := range file.Messages {
		m := file.Messages[i]
		opts, ok := gp.GetExtension(m.Desc.Options(), ormpb.E_Opts).(*ormpb.GormMessageOptions)
		if !ok || opts == nil || !opts.Enabled {
			continue
		}

		ormName := protoutil.Name(string(m.Desc.Name()) + "Model").UpperCamelCase().String()

		logger.Info().
			Str("orm", ormName).
			Msg(string(m.Desc.FullName()))
		tb := &table{fields: map[string]*Field{}, name: ormName}
		for j := range m.Fields {
			ff := NewField(m.Fields[j], gen)
			tb.fields[ff.GoName] = ff
			if ff.tag != nil && ff.tag.Pk {
				tb.pkName = ff.Name
				tb.pkType = ff.GoType
			}
		}

		_gen := jen.Commentf("%s gen from %s.%s", ormName, string(m.GoIdent.GoImportPath), m.GoIdent.GoName).Line()
		_gen = _gen.Type().Id(ormName).StructFunc(func(group *jen.Group) {
			for j := range m.Fields {
				ff := NewField(m.Fields[j], gen)
				group.Add(ff.genGormField())
			}
		}).Line().Line()

		createModel := protoutil.Name(string(m.Desc.Name()) + "CreateModel").UpperCamelCase().String()
		_gen = _gen.Type().Id(createModel).StructFunc(func(group *jen.Group) {
			for j := range m.Fields {
				ff := NewField(m.Fields[j], gen)
				if ff.tag == nil || (!ff.tag.AllowCreate && !ff.tag.AllowAll) {
					continue
				}

				group.Add(ff.genGormField())
			}
		}).Line().Line()

		updateModel := protoutil.Name(string(m.Desc.Name()) + "UpdateModel").UpperCamelCase().String()
		_gen = _gen.Type().Id(updateModel).StructFunc(func(group *jen.Group) {
			for j := range m.Fields {
				ff := NewField(m.Fields[j], gen)
				if ff.tag == nil || (!ff.tag.AllowUpdate && !ff.tag.AllowAll) {
					continue
				}

				group.Add(ff.genGormField())
			}
		}).Line().Line()

		detailModel := protoutil.Name(string(m.Desc.Name()) + "DetailModel").UpperCamelCase().String()
		_gen = _gen.Type().Id(detailModel).StructFunc(func(group *jen.Group) {
			for j := range m.Fields {
				ff := NewField(m.Fields[j], gen)
				if ff.tag == nil || (!ff.tag.AllowDetail && !ff.tag.AllowAll) {
					continue
				}

				group.Add(ff.genGormField())
			}
		}).Line().Line()

		listModel := protoutil.Name(string(m.Desc.Name()) + "ListModel").UpperCamelCase().String()
		_gen = _gen.Type().Id(listModel).StructFunc(func(group *jen.Group) {
			for j := range m.Fields {
				ff := NewField(m.Fields[j], gen)
				if ff.tag == nil || (!ff.tag.AllowList && !ff.tag.AllowAll) {
					continue
				}

				group.Add(ff.genGormField())
			}
		}).Line().Line()

		tableName := string(m.Desc.Name())
		if opts.Table != "" {
			tableName = opts.Table
		}

		_gen.Add(
			jen.Type().Id(ormName + "Srv").InterfaceFunc(func(g *jen.Group) {
				if tb.pkName != "" {
					g.Add(
						jen.Id("PkName").Params().String(),
					)

					g.Add(
						jen.Id("PkType").Params().Id(tb.pkType.GoName),
					)

					g.Add(
						jen.Id("CreateByPk").Params(
							jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB"),
							jen.Id("x").Op("*").Id(m.GoIdent.GoName),
						).Params(
							jen.Op("*").Id(m.GoIdent.GoName),
							jen.Error(),
						),
					)

					g.Add(
						jen.Id("DeleteByPk").Params(
							jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB"),
							jen.Id(tb.pkName).Id(tb.pkType.GoName),
						).Params(
							jen.Id(tb.pkType.GoName),
							jen.Error(),
						),
					)

					g.Add(
						jen.Id("UpdateByPk").Params(
							jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB"),
							jen.Id("x").Op("*").Id(m.GoIdent.GoName),
						).Params(
							jen.Id("int64"),
							jen.Error(),
						),
					)

					g.Add(
						jen.Id("GetByPk").Params(
							jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB"),
							jen.Id(tb.pkName).Id(tb.pkType.GoName),
						).Params(
							jen.Op("*").Id(m.GoIdent.GoName),
							jen.Error(),
						),
					)

					g.Add(
						jen.Id("ListByPk").Params(
							jen.Id("db").Op("*").Qual("gorm.io/gorm", "DB"),
						).Params(
							jen.Index().Op("*").Id(m.GoIdent.GoName),
							jen.Error(),
						),
					)
				}
			}).Line(),
		)

		_gen = _gen.Func().
			Params(jen.Id("m").Op("*").Id(ormName)).
			Id("TableName").
			Params().String().
			Block(jen.Return(jen.Lit(protoutil.Name(tableName).LowerSnakeCase().String()))).Line().Line()

		_gen = _gen.Func().
			Params(jen.Id("m").Op("*").Id(ormName)).
			Id("ToProto").
			Params().
			Op("*").Id(string(m.Desc.Name())).
			BlockFunc(func(g *jen.Group) {
				g.If(jen.Id("m").Op("==").Id("nil")).Block(jen.Return(jen.Id("nil"))).Line()
				g.Var().Id("x").Op("=").New(jen.Id(string(m.Desc.Name())))

				for j := range m.Fields {
					ff := NewField(m.Fields[j], gen)
					g.Add(ff.genModel2Protobuf())
				}

				g.Return(jen.Id("x"))
			}).Line()

		_gen = _gen.Func().
			Params(jen.Id("x").Op("*").Id(string(m.Desc.Name()))).
			Id("ToModel").
			Params().
			Op("*").Id(ormName).
			BlockFunc(func(g *jen.Group) {
				g.If(jen.Id("x").Op("==").Id("nil")).Block(jen.Return(jen.Id("nil"))).Line()
				g.Var().Id("m").Op("=").New(jen.Id(ormName))

				for j := range m.Fields {
					ff := NewField(m.Fields[j], gen)
					g.Add(ff.genProtobuf2Model())
				}

				g.Return(jen.Id("m"))
			}).Line()
		genFile.Add(_gen)
	}
	g.P(genFile.GoString())
	return g
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
