package internal

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/protobuild/internal/protoutil"
	ormpb "github.com/pubgo/protobuild/pkg/orm"
	retagpb "github.com/pubgo/protobuild/pkg/retag"
	"google.golang.org/protobuf/compiler/protogen"
	gp "google.golang.org/protobuf/proto"
)

func NewField(field *protogen.Field, gen *protogen.Plugin) *Field {
	var f = &Field{GoTag: make(map[string]string)}
	f.IsList = field.Desc.IsList()
	f.IsOptional = field.Desc.HasOptionalKeyword()
	f.IsMap = field.Desc.IsMap()
	f.IsMessage = field.Message != nil

	var tag, ok = gp.GetExtension(field.Desc.Options(), ormpb.E_Field).(*ormpb.GormTag)
	if !ok || tag != nil {
		f.tag = tag
	}

	f.Name = field.Desc.TextName()
	f.GoName = field.GoName
	f.GoTag = getTags(field)

	f.Type = field.Desc.Kind().String()
	f.GoType = protobufTypes[f.Type]

	if f.IsMessage {
		f.Type = string(field.Message.Desc.FullName())
		f.GoType = field.Message.GoIdent
	}

	if f.IsMap {
		f.IsMessage = field.Desc.MapValue().Message() != nil
		f.MapKeyType = protobufTypes[field.Desc.MapKey().Kind().String()]
		f.Type = field.Desc.MapValue().Kind().String()
		f.GoType = protobufTypes[f.Type]

		if f.IsMessage {
			f.Type = string(field.Desc.MapValue().Message().FullName())
			f.GoType = protogen.GoIdent{
				GoName:       protoutil.Name(field.Desc.MapValue().Message().Name()).UpperCamelCase().String(),
				GoImportPath: gen.FilesByPath[field.Desc.MapValue().Message().ParentFile().Path()].GoImportPath,
			}
		}
	}

	if f.GoType.GoImportPath == field.GoIdent.GoImportPath {
		f.GoType.GoImportPath = ""
		f.IsSelfPackage = true
	}

	logger.WithFields(log.Map{
		"go-type":  f.GoType.GoName,
		"import":   f.GoType.GoImportPath,
		"map":      f.IsMap,
		"list":     f.IsList,
		"message":  f.IsMessage,
		"optional": f.IsOptional,
	}).Info().Msg(f.Type)

	return f
}

type Field struct {
	tag           *ormpb.GormTag
	IsMessage     bool
	IsList        bool
	IsMap         bool
	IsOptional    bool
	IsSelfPackage bool

	Name       string
	Type       string
	MapKeyType protogen.GoIdent
	GoName     string
	GoType     protogen.GoIdent
	GoTag      map[string]string
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

func (f *Field) genGormField() *jen.Statement {
	var g = jen.Id(f.GoName)
	if f.IsList {
		g = g.Index()
	}

	if f.IsMap {
		g = g.Map(jen.Id(f.MapKeyType.GoName))
	}

	if f.IsMessage {
		var ormType = f.GoType
		switch f.Type {
		case "google.protobuf.Timestamp":
			if f.IsOptional {
				g = g.Op("*")
			}
			ormType = protogen.GoIdent{
				GoName:       "Time",
				GoImportPath: "time",
			}
		case "google.protobuf.Duration":
			if f.IsOptional {
				g = g.Op("*")
			}
			ormType = protogen.GoIdent{
				GoName:       "Duration",
				GoImportPath: "time",
			}
		default:
			g = g.Op("*")
			ormType = protogen.GoIdent{
				GoName:       ormType.GoName + "Model",
				GoImportPath: ormType.GoImportPath,
			}
		}

		if f.IsSelfPackage {
			g = g.Id(ormType.GoName)
		} else {
			g = g.Qual(string(ormType.GoImportPath), ormType.GoName)
		}
	} else {
		if f.IsOptional {
			g = g.Op("*")
		}

		g = g.Id(f.GoType.GoName)
	}

	return g.Tag(f.GoTag)
}

func (f *Field) genGormCond() string {
	var name = f.Name
	if strings.HasSuffix(name, "_from") {
		name = strings.TrimSuffix(name, "_from")
		return fmt.Sprintf("%s >= ?", name)
	}

	if strings.HasSuffix(name, "_to") {
		name = strings.TrimSuffix(name, "_to")
		return fmt.Sprintf("%s < ?", name)
	}

	if f.IsList {
		return fmt.Sprintf("%s in ?", name)
	}

	return fmt.Sprintf("%s = ?", name)
}

func (f *Field) genModel2Protobuf() *jen.Statement {
	if !f.IsMessage {
		return jen.Id("x").Dot(f.GoName).Op("=").Id("m").Dot(f.GoName).Line()
	}

	switch f.Type {
	case "google.protobuf.Timestamp":
		if f.IsList || f.IsMap {
			var v = jen.Op("*").Qual("google.golang.org/protobuf/types/known/timestamppb", "Timestamp")
			var gen = jen.Id("x").Dot(f.GoName).
				Op("=").
				Make(generic.Ternary(f.IsMap, jen.Map(jen.Id(f.MapKeyType.GoName)), jen.Index()).Add(v), jen.Len(jen.Id("m").Dot(f.GoName))).Line()
			return gen.For(
				jen.Id("i").Op(":=").Range().Id("m").Dot(f.GoName),
			).Block(
				jen.If(
					jen.Op("!").Id("m").Dot(f.GoName).Index(jen.Id("i")).Dot("IsZero").Call(),
				).Block(
					jen.Id("x").Dot(f.GoName).Index(jen.Id("i")).
						Op("=").
						Qual("google.golang.org/protobuf/types/known/timestamppb", "New").Call(jen.Id("m").Dot(f.GoName).Index(jen.Id("i"))),
				),
			).Line()
		}

		return jen.If(
			generic.Ternary(f.IsOptional, jen.Id("m").Dot(f.GoName).Op("!=").Nil().Op("&&"), jen.Empty()).Op("!").Id("m").Dot(f.GoName).Dot("IsZero").Call(),
		).BlockFunc(func(g *jen.Group) {
			g.Id("x").Dot(f.GoName).
				Op("=").
				Qual("google.golang.org/protobuf/types/known/timestamppb", "New").Call(generic.Ternary(f.IsOptional, jen.Op("*"), jen.Empty()).Id("m").Dot(f.GoName))
		}).Line()
	case "google.protobuf.Duration":
		return jen.If(
			jen.Id("m").Dot(f.GoName).Op("==").Id("0"),
		).BlockFunc(func(g *jen.Group) {
			g.Id("x").Dot(f.GoName).Op("=").
				Qual("google.golang.org/protobuf/types/known/durationpb", "New").Call(jen.Id("m").Dot(f.GoName))
		}).Line()
	default:
		if f.IsList || f.IsMap {
			var gen = jen.Id("x").Dot(f.GoName).
				Op("=").
				Make(generic.Ternary(f.IsList, jen.Index(), jen.Map(jen.Id(f.MapKeyType.GoName))).Op("*").Qual(string(f.GoType.GoImportPath), f.GoType.GoName), jen.Len(jen.Id("m").Dot(f.GoName))).Line()
			return gen.For(
				jen.Id("i").Op(":=").Range().Id("m").Dot(f.GoName),
			).Block(
				jen.If(
					jen.Id("m").Dot(f.GoName).Index(jen.Id("i")).Op("!=").Nil(),
				).Block(
					jen.Id("x").Dot(f.GoName).Index(jen.Id("i")).
						Op("=").
						Id("m").Dot(f.GoName).Index(jen.Id("i")).Dot("ToProto").Call(),
				),
			).Line()
		}

		return jen.If(
			jen.Id("m").Dot(f.GoName).Op("!=").Nil(),
		).Block(
			jen.Id("x").Dot(f.GoName).Op("=").Id("m").Dot(f.GoName).Dot("ToProto").Call(),
		).Line()
	}
}

func (f *Field) genProtobuf2Model() *jen.Statement {
	if !f.IsMessage {
		return jen.Id("m").Dot(f.GoName).Op("=").Id("x").Dot(f.GoName).Line()
	}

	switch f.Type {
	case "google.protobuf.Timestamp":
		if f.IsList || f.IsMap {
			var v = jen.Qual("time", "Time")
			var gen = jen.Id("m").Dot(f.GoName).
				Op("=").
				Make(generic.Ternary(f.IsMap, jen.Map(jen.Id(f.MapKeyType.GoName)), jen.Index()).Add(v), jen.Len(jen.Id("x").Dot(f.GoName))).Line()
			return gen.For(
				jen.Id("i").Op(":=").Range().Id("x").Dot(f.GoName),
			).Block(
				jen.If(
					jen.Id("x").Dot(f.GoName).Index(jen.Id("i")).Op("!=").Nil().
						Op("&&").Id("x").Dot(f.GoName).Index(jen.Id("i")).Dot("IsValid").Call(),
				).Block(
					jen.Id("m").Dot(f.GoName).Index(jen.Id("i")).
						Op("=").
						Id("x").Dot(f.GoName).Index(jen.Id("i")).Dot("AsTime").Call(),
				),
			).Line()
		}

		return jen.If(
			jen.Id("x").Dot(f.GoName).Op("!=").Nil().
				Op("&&").
				Id("x").Dot(f.GoName).Dot("IsValid").Call(),
		).BlockFunc(func(g *jen.Group) {
			g.Id("m").Dot(f.GoName).
				Op("=").
				Add(generic.Ternary(f.IsOptional, jen.Qual("github.com/pubgo/funk/generic", "Ptr").Call(jen.Id("x").Dot(f.GoName).Dot("AsTime").Call()), jen.Id("x").Dot(f.GoName).Dot("AsTime").Call()))
		}).Line()
	case "google.protobuf.Duration":
		return jen.If(
			jen.Id("x").Dot(f.GoName).Op("==").Id("0"),
		).BlockFunc(func(g *jen.Group) {
			g.Id("m").Dot(f.GoName).Op("=").
				Qual("google.golang.org/protobuf/types/known/durationpb", "New").Call(jen.Id("m").Dot(f.GoName))
		}).Line()
	default:
		if f.IsList || f.IsMap {
			var gen = jen.Id("m").Dot(f.GoName).
				Op("=").
				Make(generic.Ternary(f.IsList, jen.Index(), jen.Map(jen.Id(f.MapKeyType.GoName))).Op("*").Qual(string(f.GoType.GoImportPath), f.GoType.GoName+"Model"), jen.Len(jen.Id("x").Dot(f.GoName))).Line()
			return gen.For(
				jen.Id("i").Op(":=").Range().Id("x").Dot(f.GoName),
			).Block(
				jen.If(
					jen.Id("x").Dot(f.GoName).Index(jen.Id("i")).Op("!=").Nil(),
				).Block(
					jen.Id("m").Dot(f.GoName).Index(jen.Id("i")).
						Op("=").
						Id("x").Dot(f.GoName).Index(jen.Id("i")).Dot("ToModel").Call(),
				),
			).Line()
		}

		return jen.If(
			jen.Id("x").Dot(f.GoName).Op("!=").Nil(),
		).Block(
			jen.Id("m").Dot(f.GoName).Op("=").Id("x").Dot(f.GoName).Dot("ToModel").Call(),
		).Line()
	}
}
