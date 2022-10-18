// Code generated by protoc-gen-gorm. DO NOT EDIT.
// versions:
// - protoc-gen-gorm v0.1.0
// - protoc          v3.19.4
// source: protoc-gen-gorm/example/example.proto

package example

import (
	generic "github.com/pubgo/funk/generic"
	grpc "google.golang.org/grpc"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ExampleModel gen from github.com/pubgo/protobuild/pkg/protoc-gen-gorm/example.Example
type ExampleModel struct {
	WithNewTags     string                         `graphql:"withNewTags,optional" json:"with_new_tags"`
	WithNewMultiple string                         `graphql:"withNewTags,optional" json:"with_new_multiple" xml:"multi,omitempty"`
	ReplaceDefault  string                         `json:"replacePrevious"`
	A               string                         `json:"A"`
	BJk             int32                          `json:"b_Jk"`
	Test_1          []*SecondMessageModel          `json:"test_1"`
	Test_2          *SecondMessageModel            `json:"test_2"`
	Test_61         *SecondMessageModel            `json:"test_61"`
	Test_31         map[string]*SecondMessageModel `json:"test_31"`
	Test_5          time.Time                      `json:"test_5"`
	Test_51         []time.Time                    `json:"test_51"`
	Test_17         *time.Time                     `json:"test_17"`
	Test_3          map[string]time.Time           `json:"test_3"`
	ListTest20      []string                       `json:"list_test_20"`
	ListTest21      map[string]string              `json:"list_test_21"`
}

func (m *ExampleModel) TableName() string {
	return "example"
}

func (m *ExampleModel) ToProto() *Example {
	if m == nil {
		return nil
	}

	var x = new(Example)
	x.WithNewTags = m.WithNewTags

	x.WithNewMultiple = m.WithNewMultiple

	x.ReplaceDefault = m.ReplaceDefault

	x.A = m.A

	x.BJk = m.BJk

	x.Test_1 = make([]*SecondMessage, len(m.Test_1))
	for i := range m.Test_1 {
		if m.Test_1[i] != nil {
			x.Test_1[i] = m.Test_1[i].ToProto()
		}
	}

	if m.Test_2 != nil {
		x.Test_2 = m.Test_2.ToProto()
	}

	if m.Test_61 != nil {
		x.Test_61 = m.Test_61.ToProto()
	}

	x.Test_31 = make(map[string]*SecondMessage, len(m.Test_31))
	for i := range m.Test_31 {
		if m.Test_31[i] != nil {
			x.Test_31[i] = m.Test_31[i].ToProto()
		}
	}

	if !m.Test_5.IsZero() {
		x.Test_5 = timestamppb.New(m.Test_5)
	}

	x.Test_51 = make([]*timestamppb.Timestamp, len(m.Test_51))
	for i := range m.Test_51 {
		if !m.Test_51[i].IsZero() {
			x.Test_51[i] = timestamppb.New(m.Test_51[i])
		}
	}

	if m.Test_17 != nil && !m.Test_17.IsZero() {
		x.Test_17 = timestamppb.New(*m.Test_17)
	}

	x.Test_3 = make(map[string]*timestamppb.Timestamp, len(m.Test_3))
	for i := range m.Test_3 {
		if !m.Test_3[i].IsZero() {
			x.Test_3[i] = timestamppb.New(m.Test_3[i])
		}
	}

	x.ListTest20 = m.ListTest20

	x.ListTest21 = m.ListTest21

	return x
}
func (x *Example) ToModel() *ExampleModel {
	if x == nil {
		return nil
	}

	var m = new(ExampleModel)
	m.WithNewTags = x.WithNewTags

	m.WithNewMultiple = x.WithNewMultiple

	m.ReplaceDefault = x.ReplaceDefault

	m.A = x.A

	m.BJk = x.BJk

	m.Test_1 = make([]*SecondMessageModel, len(x.Test_1))
	for i := range x.Test_1 {
		if x.Test_1[i] != nil {
			m.Test_1[i] = x.Test_1[i].ToModel()
		}
	}

	if x.Test_2 != nil {
		m.Test_2 = x.Test_2.ToModel()
	}

	if x.Test_61 != nil {
		m.Test_61 = x.Test_61.ToModel()
	}

	m.Test_31 = make(map[string]*SecondMessageModel, len(x.Test_31))
	for i := range x.Test_31 {
		if x.Test_31[i] != nil {
			m.Test_31[i] = x.Test_31[i].ToModel()
		}
	}

	if x.Test_5 != nil && x.Test_5.IsValid() {
		m.Test_5 = x.Test_5.AsTime()
	}

	m.Test_51 = make([]time.Time, len(x.Test_51))
	for i := range x.Test_51 {
		if x.Test_51[i] != nil && x.Test_51[i].IsValid() {
			m.Test_51[i] = x.Test_51[i].AsTime()
		}
	}

	if x.Test_17 != nil && x.Test_17.IsValid() {
		m.Test_17 = generic.Ptr(x.Test_17.AsTime())
	}

	m.Test_3 = make(map[string]time.Time, len(x.Test_3))
	for i := range x.Test_3 {
		if x.Test_3[i] != nil && x.Test_3[i].IsValid() {
			m.Test_3[i] = x.Test_3[i].AsTime()
		}
	}

	m.ListTest20 = x.ListTest20

	m.ListTest21 = x.ListTest21

	return m
}

// SecondMessageModel gen from github.com/pubgo/protobuild/pkg/protoc-gen-gorm/example.SecondMessage
type SecondMessageModel struct {
	WithNewTags     string `graphql:"withNewTags,optional" json:"with_new_tags"`
	WithNewMultiple string `graphql:"withNewTags,optional" json:"with_new_multiple" xml:"multi,omitempty"`
	ReplaceDefault  string `json:"replacePrevious"`
}

func (m *SecondMessageModel) TableName() string {
	return "second_message"
}

func (m *SecondMessageModel) ToProto() *SecondMessage {
	if m == nil {
		return nil
	}

	var x = new(SecondMessage)
	x.WithNewTags = m.WithNewTags

	x.WithNewMultiple = m.WithNewMultiple

	x.ReplaceDefault = m.ReplaceDefault

	return x
}
func (x *SecondMessage) ToModel() *SecondMessageModel {
	if x == nil {
		return nil
	}

	var m = new(SecondMessageModel)
	m.WithNewTags = x.WithNewTags

	m.WithNewMultiple = x.WithNewMultiple

	m.ReplaceDefault = x.ReplaceDefault

	return m
}

// ThirdExampleModel gen from github.com/pubgo/protobuild/pkg/protoc-gen-gorm/example.ThirdExample
type ThirdExampleModel struct {
	Test *string `json:"test"`
}

func (m *ThirdExampleModel) TableName() string {
	return "third_example"
}

func (m *ThirdExampleModel) ToProto() *ThirdExample {
	if m == nil {
		return nil
	}

	var x = new(ThirdExample)
	x.Test = m.Test

	return x
}
func (x *ThirdExample) ToModel() *ThirdExampleModel {
	if x == nil {
		return nil
	}

	var m = new(ThirdExampleModel)
	m.Test = x.Test

	return m
}
