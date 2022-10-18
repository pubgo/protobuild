package internal

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/pubgo/funk/logx"
	"github.com/pubgo/protobuild/internal/protoutil"
	ormpb "github.com/pubgo/protobuild/pkg/orm"
	"google.golang.org/protobuf/compiler/protogen"
	gp "google.golang.org/protobuf/proto"
	"gorm.io/gorm/schema"

	_ "google.golang.org/protobuf/encoding/protojson"
	_ "google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	_ = schema.ParseTagSetting
}

var logger = logx.WithName("gorm")

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

	for i := range file.Messages {
		m := file.Messages[i]

		if m.Desc.Options() == nil {
			continue
		}

		logger.Info(string(m.Desc.FullName()))
		var opts, ok = gp.GetExtension(m.Desc.Options(), ormpb.E_Opts).(*ormpb.GormMessageOptions)
		if !ok || opts == nil || !opts.Enabled {
			continue
		}

		var ormName = protoutil.Name(string(m.Desc.Name()) + "Model").UpperCamelCase().String()
		_gen := jen.Commentf("%s gen from %s.%s", ormName, string(m.GoIdent.GoImportPath), m.GoIdent.GoName).Line()
		_gen = _gen.Type().Id(ormName).StructFunc(func(group *jen.Group) {
			for j := range m.Fields {
				var ff = NewField(m.Fields[j], gen)
				group.Add(ff.genGormField())
			}
		}).Line()

		var tableName = string(m.Desc.Name())
		if opts.Table != "" {
			tableName = opts.Table
		}

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
					var ff = NewField(m.Fields[j], gen)
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
					var ff = NewField(m.Fields[j], gen)
					g.Add(ff.genProtobuf2Model())
				}

				g.Return(jen.Id("m"))
			}).Line()
		genFile.Add(_gen)
	}
	g.P(genFile.GoString())
	return g
}

//func renderGormTag(field *Field) string {
//	var gormRes, atlasRes string
//	tag := field.GetTag()
//	if tag == nil {
//		tag = &gorm.GormTag{}
//	}
//
//	if len(tag.Column) > 0 {
//		gormRes += fmt.Sprintf("column:%s;", tag.GetColumn())
//	}
//	if len(tag.Type) > 0 {
//		gormRes += fmt.Sprintf("type:%s;", tag.GetType())
//	}
//	if tag.GetSize() > 0 {
//		gormRes += fmt.Sprintf("size:%d;", tag.GetSize())
//	}
//	if tag.Precision > 0 {
//		gormRes += fmt.Sprintf("precision:%d;", tag.GetPrecision())
//	}
//	if tag.GetPrimaryKey() {
//		gormRes += "primary_key;"
//	}
//	if tag.GetUnique() {
//		gormRes += "unique;"
//	}
//	if len(tag.Default) > 0 {
//		gormRes += fmt.Sprintf("default:%s;", tag.GetDefault())
//	}
//	if tag.GetNotNull() {
//		gormRes += "not null;"
//	}
//	if tag.GetAutoIncrement() {
//		gormRes += "auto_increment;"
//	}
//	if len(tag.Index) > 0 {
//		if tag.GetIndex() == "" {
//			gormRes += "index;"
//		} else {
//			gormRes += fmt.Sprintf("index:%s;", tag.GetIndex())
//		}
//	}
//	if len(tag.UniqueIndex) > 0 {
//		if tag.GetUniqueIndex() == "" {
//			gormRes += "unique_index;"
//		} else {
//			gormRes += fmt.Sprintf("unique_index:%s;", tag.GetUniqueIndex())
//		}
//	}
//	if tag.GetEmbedded() {
//		gormRes += "embedded;"
//	}
//	if len(tag.EmbeddedPrefix) > 0 {
//		gormRes += fmt.Sprintf("embedded_prefix:%s;", tag.GetEmbeddedPrefix())
//	}
//	if tag.GetIgnore() {
//		gormRes += "-;"
//	}
//
//	var foreignKey, associationForeignKey, joinTable, joinTableForeignKey, associationJoinTableForeignKey string
//	var associationAutoupdate, associationAutocreate, associationSaveReference, preload, replace, append, clear bool
//	if hasOne := field.GetHasOne(); hasOne != nil {
//		foreignKey = hasOne.Foreignkey
//		associationForeignKey = hasOne.AssociationForeignkey
//		associationAutoupdate = hasOne.AssociationAutoupdate
//		associationAutocreate = hasOne.AssociationAutocreate
//		associationSaveReference = hasOne.AssociationSaveReference
//		preload = hasOne.Preload
//		clear = hasOne.Clear
//		replace = hasOne.Replace
//		append = hasOne.Append
//	} else if belongsTo := field.GetBelongsTo(); belongsTo != nil {
//		foreignKey = belongsTo.Foreignkey
//		associationForeignKey = belongsTo.AssociationForeignkey
//		associationAutoupdate = belongsTo.AssociationAutoupdate
//		associationAutocreate = belongsTo.AssociationAutocreate
//		associationSaveReference = belongsTo.AssociationSaveReference
//		preload = belongsTo.Preload
//	} else if hasMany := field.GetHasMany(); hasMany != nil {
//		foreignKey = hasMany.Foreignkey
//		associationForeignKey = hasMany.AssociationForeignkey
//		associationAutoupdate = hasMany.AssociationAutoupdate
//		associationAutocreate = hasMany.AssociationAutocreate
//		associationSaveReference = hasMany.AssociationSaveReference
//		clear = hasMany.Clear
//		preload = hasMany.Preload
//		replace = hasMany.Replace
//		append = hasMany.Append
//		if len(hasMany.PositionField) > 0 {
//			atlasRes += fmt.Sprintf("position:%s;", hasMany.GetPositionField())
//		}
//	} else if mtm := field.GetManyToMany(); mtm != nil {
//		foreignKey = mtm.Foreignkey
//		associationForeignKey = mtm.AssociationForeignkey
//		joinTable = mtm.Jointable
//		joinTableForeignKey = mtm.JointableForeignkey
//		associationJoinTableForeignKey = mtm.AssociationJointableForeignkey
//		associationAutoupdate = mtm.AssociationAutoupdate
//		associationAutocreate = mtm.AssociationAutocreate
//		associationSaveReference = mtm.AssociationSaveReference
//		preload = mtm.Preload
//		clear = mtm.Clear
//		replace = mtm.Replace
//		append = mtm.Append
//	} else {
//		foreignKey = tag.Foreignkey
//		associationForeignKey = tag.AssociationForeignkey
//		joinTable = tag.ManyToMany
//		joinTableForeignKey = tag.JointableForeignkey
//		associationJoinTableForeignKey = tag.AssociationJointableForeignkey
//		associationAutoupdate = tag.AssociationAutoupdate
//		associationAutocreate = tag.AssociationAutocreate
//		associationSaveReference = tag.AssociationSaveReference
//		preload = tag.Preload
//	}
//
//	if len(foreignKey) > 0 {
//		gormRes += fmt.Sprintf("foreignkey:%s;", foreignKey)
//	}
//
//	if len(associationForeignKey) > 0 {
//		gormRes += fmt.Sprintf("association_foreignkey:%s;", associationForeignKey)
//	}
//
//	if len(joinTable) > 0 {
//		gormRes += fmt.Sprintf("many2many:%s;", joinTable)
//	}
//	if len(joinTableForeignKey) > 0 {
//		gormRes += fmt.Sprintf("jointable_foreignkey:%s;", joinTableForeignKey)
//	}
//	if len(associationJoinTableForeignKey) > 0 {
//		gormRes += fmt.Sprintf("association_jointable_foreignkey:%s;", associationJoinTableForeignKey)
//	}
//
//	if associationAutoupdate {
//		gormRes += fmt.Sprintf("association_autoupdate:%s;", strconv.FormatBool(associationAutoupdate))
//	}
//
//	if associationAutocreate {
//		gormRes += fmt.Sprintf("association_autocreate:%s;", strconv.FormatBool(associationAutocreate))
//	}
//
//	if associationSaveReference {
//		gormRes += fmt.Sprintf("association_save_reference:%s;", strconv.FormatBool(associationSaveReference))
//	}
//
//	if preload {
//		gormRes += fmt.Sprintf("preload:%s;", strconv.FormatBool(preload))
//	}
//
//	if clear {
//		gormRes += fmt.Sprintf("clear:%s;", strconv.FormatBool(clear))
//	} else if replace {
//		gormRes += fmt.Sprintf("replace:%s;", strconv.FormatBool(replace))
//	} else if append {
//		gormRes += fmt.Sprintf("append:%s;", strconv.FormatBool(append))
//	}
//
//	var gormTag, atlasTag string
//	if gormRes != "" {
//		gormTag = fmt.Sprintf("gorm:\"%s\"", strings.TrimRight(gormRes, ";"))
//	}
//	if atlasRes != "" {
//		atlasTag = fmt.Sprintf("atlas:\"%s\"", strings.TrimRight(atlasRes, ";"))
//	}
//	finalTag := strings.TrimSpace(strings.Join([]string{gormTag, atlasTag}, " "))
//	if finalTag == "" {
//		return ""
//	} else {
//		return fmt.Sprintf("`%s`", finalTag)
//	}
//}

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
