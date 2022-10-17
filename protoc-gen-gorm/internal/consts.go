package internal

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	gormImport         = "gorm.io/gorm"
	uuidImport         = "github.com/google/uuid"
	authImport         = "github.com/infobloxopen/atlas-app-toolkit/auth"
	gormpqImport       = "github.com/jinzhu/gorm/dialects/postgres"
	gtypesImport       = "github.com/infobloxopen/protoc-gen-gorm/types"
	resourceImport     = "github.com/infobloxopen/atlas-app-toolkit/gorm/resource"
	queryImport        = "github.com/infobloxopen/atlas-app-toolkit/query"
	ocTraceImport      = "go.opencensus.io/trace"
	gatewayImport      = "github.com/infobloxopen/atlas-app-toolkit/gateway"
	pqImport           = "github.com/lib/pq"
	gerrorsImport      = "github.com/infobloxopen/protoc-gen-gorm/errors"
	timestampImport    = "google.golang.org/protobuf/types/known/timestamppb"
	durationImport     = "google.golang.org/protobuf/types/known/durationpb"
	wktImport          = "google.golang.org/protobuf/types/known/wrapperspb"
	fmImport           = "google.golang.org/genproto/protobuf/field_mask"
	stdFmtImport       = "fmt"
	stdCtxImport       = "context"
	stdStringsImport   = "strings"
	stdTimeImport      = "time"
	encodingJsonImport = "encoding/json"
	bigintImport       = "math/big"
)

var protobufTypes = map[string]protogen.GoIdent{
	protoreflect.DoubleKind.String(): {GoName: "float64"},
	protoreflect.FloatKind.String():  {GoName: "float32"},
	protoreflect.Int64Kind.String():  {GoName: "int64"},
	protoreflect.Uint64Kind.String(): {GoName: "uint64"},
	protoreflect.Int32Kind.String():  {GoName: "int32"},
	protoreflect.Uint32Kind.String(): {GoName: "uint32"},
	protoreflect.BoolKind.String():   {GoName: "bool"},
	protoreflect.StringKind.String(): {GoName: "string"},
	//protoreflect.BytesKind.String():  {GoName: "[]byte"},
	//"google.protobuf.Timestamp": {
	//	GoName:       "Time",
	//	GoImportPath: "time",
	//},
	//"google.protobuf.Duration": {
	//	GoName:       "Duration",
	//	GoImportPath: "time",
	//},
}
