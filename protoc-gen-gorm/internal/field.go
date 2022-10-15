package internal

import (
	"github.com/dave/jennifer/jen"
	"github.com/pubgo/protobuild/internal/protoutil"
	retagpb "github.com/pubgo/protobuild/pkg/retag"
	"google.golang.org/protobuf/compiler/protogen"
	gp "google.golang.org/protobuf/proto"
)

func NewField(field *protogen.Field, gen *protogen.Plugin) *Field {
	var f = &Field{OrmTag: make(map[string]string)}
	f.Name = field.Desc.TextName()
	f.GoName = field.GoName
	f.IsList = field.Desc.IsList()
	f.IsOptional = field.Desc.HasOptionalKeyword()
	f.IsMap = field.Desc.IsMap()
	f.IsMessage = field.Message != nil
	f.OrmTag = getTags(field)
	f.Type = field.Desc.Kind().String()
	f.OrmType = protobufTypes[f.Type]

	if f.IsMessage {
		f.Type = string(field.Message.Desc.FullName())
		f.IsMessage = protobufTypes[f.Type].GoName == ""
		f.OrmType = protobufTypes[f.Type]
		if f.OrmType.GoName == "" {
			f.OrmType = field.Message.GoIdent
		}
	}

	if f.IsMap {
		f.IsMessage = false
		f.MapKeyType = protobufTypes[field.Desc.MapKey().Kind().String()]
		f.OrmType = protobufTypes[field.Desc.MapValue().Kind().String()]
		if f.OrmType.GoName == "" {
			f.OrmType = protobufTypes[string(field.Desc.MapValue().Message().FullName())]
		}

		if f.OrmType.GoName == "" {
			f.IsMessage = true
			f.OrmType = protogen.GoIdent{
				GoName:       protoutil.Name(field.Desc.MapValue().Message().Name()).UpperCamelCase().String(),
				GoImportPath: gen.FilesByPath[field.Desc.MapValue().Message().ParentFile().Path()].GoImportPath,
			}
		}
	}

	if f.OrmType.GoImportPath == field.GoIdent.GoImportPath {
		f.OrmType.GoImportPath = ""
		f.IsSelfPackage = true
	}

	return f
}

type Field struct {
	IsMessage     bool
	IsList        bool
	IsMap         bool
	IsOptional    bool
	IsSelfPackage bool
	Name          string
	Type          string

	MapKeyType protogen.GoIdent

	GoName string
	GoType protogen.GoIdent

	OrmType protogen.GoIdent
	OrmTag  map[string]string
}

func getTags(field *protogen.Field) map[string]string {
	var tagMap = map[string]string{"json": protoutil.Name(field.GoName).LowerSnakeCase().String()}
	if tags, ok := gp.GetExtension(field.Desc.Options(), retagpb.E_Tags).([]*retagpb.Tag); ok && tags != nil {
		for i := range tags {
			tagMap[tags[i].Name] = tags[i].Value
		}
	}
	return tagMap
}

func (f *Field) genGoGormField() *jen.Statement {
	switch f.Type {
	case "google.protobuf.Timestamp", "google.protobuf.Duration":
	}

	var g = jen.Id(f.GoName)
	if f.IsList {
		g = g.Index()
	}

	if f.IsMap {
		if f.MapKeyType.GoImportPath == "" {
			g = g.Map(jen.Id(f.MapKeyType.GoName))
		} else {
			g = g.Map(jen.Qual(string(f.MapKeyType.GoImportPath), f.MapKeyType.GoName))
		}
	}

	if f.IsOptional || f.IsMessage {
		g = g.Op("*")
	}

	if f.OrmType.GoImportPath == "" {
		if f.IsSelfPackage {
			g = g.Id(f.OrmType.GoName + "Model")
		} else {
			g = g.Id(f.OrmType.GoName)
		}

	} else {
		if f.IsMessage {
			g = g.Qual(string(f.OrmType.GoImportPath), f.OrmType.GoName+"Model")
		} else {
			g = g.Qual(string(f.OrmType.GoImportPath), f.OrmType.GoName)
		}
	}

	return g.Tag(f.OrmTag)
}

func (f *Field) genModel2Protobuf() *jen.Statement {
	switch f.Type {
	case "google.protobuf.Timestamp":
		return jen.If(
			jen.Op("!").Id("m").Dot(f.GoName).Dot("IsZero").Call(),
		).BlockFunc(func(g *jen.Group) {
			g.Id("x").Dot(f.GoName).Op("=").
				Qual("google.golang.org/protobuf/types/known/timestamppb", "New").Call(jen.Id("m").Dot(f.GoName))
		}).Line()
	case "google.protobuf.Duration":
		return jen.If(
			jen.Id("m").Dot(f.GoName).Op("==").Id("0"),
		).BlockFunc(func(g *jen.Group) {
			g.Id("x").Dot(f.GoName).Op("=").
				Qual("google.golang.org/protobuf/types/known/durationpb", "New").Call(jen.Id("m").Dot(f.GoName))
		}).Line()
	default:
		return jen.Id("x").Dot(f.GoName).Op("=").Id("m").Dot(f.GoName).Line()
	}
}

func (f *Field) genProtobuf2Model() *jen.Statement {
	switch f.Type {
	case "google.protobuf.Timestamp":
		return jen.If(
			jen.Id("x").Dot(f.GoName).Op("!=").Nil(),
		).BlockFunc(func(g *jen.Group) {
			g.Id("m").Dot(f.GoName).Op("=").
				Id("x").Dot(f.GoName).Dot("AsTime").Call()
		}).Line()
	case "google.protobuf.Duration":
		return jen.If(
			jen.Id("x").Dot(f.GoName).Op("!=").Nil(),
		).BlockFunc(func(g *jen.Group) {
			g.Id("m").Dot(f.GoName).Op("=").
				Id("x").Dot(f.GoName).Dot("AsDuration").Call()
		}).Line()
	default:
		return jen.Id("m").Dot(f.GoName).Op("=").Id("x").Dot(f.GoName).Line()
	}
}
