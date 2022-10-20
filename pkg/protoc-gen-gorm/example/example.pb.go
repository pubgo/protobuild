// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: protoc-gen-gorm/example/example.proto

package example

import (
	_ "github.com/pubgo/protobuild/pkg/orm"
	_ "github.com/pubgo/protobuild/pkg/retag"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/descriptorpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type User struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        uint64                 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	CreatedAt *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	Birthday  *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=birthday,proto3" json:"birthday,omitempty"`
	Num       uint32                 `protobuf:"varint,6,opt,name=num,proto3" json:"num,omitempty"`
}

func (x *User) Reset() {
	*x = User{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_gorm_example_example_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *User) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*User) ProtoMessage() {}

func (x *User) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_gorm_example_example_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use User.ProtoReflect.Descriptor instead.
func (*User) Descriptor() ([]byte, []int) {
	return file_protoc_gen_gorm_example_example_proto_rawDescGZIP(), []int{0}
}

func (x *User) GetId() uint64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *User) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *User) GetUpdatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdatedAt
	}
	return nil
}

func (x *User) GetBirthday() *timestamppb.Timestamp {
	if x != nil {
		return x.Birthday
	}
	return nil
}

func (x *User) GetNum() uint32 {
	if x != nil {
		return x.Num
	}
	return 0
}

type Example struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	WithNewTags     string                            `protobuf:"bytes,1,opt,name=with_new_tags,json=withNewTags,proto3" json:"with_new_tags,omitempty" graphql:"withNewTags,optional"`
	WithNewMultiple string                            `protobuf:"bytes,2,opt,name=with_new_multiple,json=withNewMultiple,proto3" json:"with_new_multiple,omitempty" graphql:"withNewTags,optional" xml:"multi,omitempty"`
	ReplaceDefault  string                            `protobuf:"bytes,3,opt,name=replace_default,json=replaceDefault,proto3" json:"replacePrevious"`
	A               string                            `protobuf:"bytes,5,opt,name=a,proto3" json:"A"`
	BJk             int32                             `protobuf:"varint,6,opt,name=b_jk,json=bJk,proto3" json:"b_Jk"`
	Test_1          []*SecondMessage                  `protobuf:"bytes,10,rep,name=test_1,json=test1,proto3" json:"test_1,omitempty"`
	Test_2          *SecondMessage                    `protobuf:"bytes,11,opt,name=test_2,json=test2,proto3,oneof" json:"test_2,omitempty"`
	Test_61         *SecondMessage                    `protobuf:"bytes,19,opt,name=test_61,json=test61,proto3" json:"test_61,omitempty"`
	Test_31         map[string]*SecondMessage         `protobuf:"bytes,13,rep,name=test_31,json=test31,proto3" json:"test_31,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Test_5          *timestamppb.Timestamp            `protobuf:"bytes,15,opt,name=test_5,json=test5,proto3" json:"test_5,omitempty"`
	Test_51         []*timestamppb.Timestamp          `protobuf:"bytes,18,rep,name=test_51,json=test51,proto3" json:"test_51,omitempty"`
	Test_17         *timestamppb.Timestamp            `protobuf:"bytes,17,opt,name=test_17,json=test17,proto3,oneof" json:"test_17,omitempty"`
	Test_3          map[string]*timestamppb.Timestamp `protobuf:"bytes,12,rep,name=test_3,json=test3,proto3" json:"test_3,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	ListTest20      []string                          `protobuf:"bytes,20,rep,name=list_test20,json=listTest20,proto3" json:"list_test20,omitempty"`
	ListTest21      map[string]string                 `protobuf:"bytes,21,rep,name=list_test21,json=listTest21,proto3" json:"list_test21,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Example) Reset() {
	*x = Example{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_gorm_example_example_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Example) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Example) ProtoMessage() {}

func (x *Example) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_gorm_example_example_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Example.ProtoReflect.Descriptor instead.
func (*Example) Descriptor() ([]byte, []int) {
	return file_protoc_gen_gorm_example_example_proto_rawDescGZIP(), []int{1}
}

func (x *Example) GetWithNewTags() string {
	if x != nil {
		return x.WithNewTags
	}
	return ""
}

func (x *Example) GetWithNewMultiple() string {
	if x != nil {
		return x.WithNewMultiple
	}
	return ""
}

func (x *Example) GetReplaceDefault() string {
	if x != nil {
		return x.ReplaceDefault
	}
	return ""
}

func (x *Example) GetA() string {
	if x != nil {
		return x.A
	}
	return ""
}

func (x *Example) GetBJk() int32 {
	if x != nil {
		return x.BJk
	}
	return 0
}

func (x *Example) GetTest_1() []*SecondMessage {
	if x != nil {
		return x.Test_1
	}
	return nil
}

func (x *Example) GetTest_2() *SecondMessage {
	if x != nil {
		return x.Test_2
	}
	return nil
}

func (x *Example) GetTest_61() *SecondMessage {
	if x != nil {
		return x.Test_61
	}
	return nil
}

func (x *Example) GetTest_31() map[string]*SecondMessage {
	if x != nil {
		return x.Test_31
	}
	return nil
}

func (x *Example) GetTest_5() *timestamppb.Timestamp {
	if x != nil {
		return x.Test_5
	}
	return nil
}

func (x *Example) GetTest_51() []*timestamppb.Timestamp {
	if x != nil {
		return x.Test_51
	}
	return nil
}

func (x *Example) GetTest_17() *timestamppb.Timestamp {
	if x != nil {
		return x.Test_17
	}
	return nil
}

func (x *Example) GetTest_3() map[string]*timestamppb.Timestamp {
	if x != nil {
		return x.Test_3
	}
	return nil
}

func (x *Example) GetListTest20() []string {
	if x != nil {
		return x.ListTest20
	}
	return nil
}

func (x *Example) GetListTest21() map[string]string {
	if x != nil {
		return x.ListTest21
	}
	return nil
}

type SecondMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	WithNewTags     string `protobuf:"bytes,1,opt,name=with_new_tags,json=withNewTags,proto3" json:"with_new_tags,omitempty" graphql:"withNewTags,optional"`
	WithNewMultiple string `protobuf:"bytes,2,opt,name=with_new_multiple,json=withNewMultiple,proto3" json:"with_new_multiple,omitempty" graphql:"withNewTags,optional" xml:"multi,omitempty"`
	ReplaceDefault  string `protobuf:"bytes,3,opt,name=replace_default,json=replaceDefault,proto3" json:"replacePrevious"`
}

func (x *SecondMessage) Reset() {
	*x = SecondMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_gorm_example_example_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SecondMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SecondMessage) ProtoMessage() {}

func (x *SecondMessage) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_gorm_example_example_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SecondMessage.ProtoReflect.Descriptor instead.
func (*SecondMessage) Descriptor() ([]byte, []int) {
	return file_protoc_gen_gorm_example_example_proto_rawDescGZIP(), []int{2}
}

func (x *SecondMessage) GetWithNewTags() string {
	if x != nil {
		return x.WithNewTags
	}
	return ""
}

func (x *SecondMessage) GetWithNewMultiple() string {
	if x != nil {
		return x.WithNewMultiple
	}
	return ""
}

func (x *SecondMessage) GetReplaceDefault() string {
	if x != nil {
		return x.ReplaceDefault
	}
	return ""
}

type ThirdExample struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Test *string `protobuf:"bytes,2,opt,name=test,proto3,oneof" json:"test,omitempty"`
}

func (x *ThirdExample) Reset() {
	*x = ThirdExample{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_gorm_example_example_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ThirdExample) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ThirdExample) ProtoMessage() {}

func (x *ThirdExample) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_gorm_example_example_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ThirdExample.ProtoReflect.Descriptor instead.
func (*ThirdExample) Descriptor() ([]byte, []int) {
	return file_protoc_gen_gorm_example_example_proto_rawDescGZIP(), []int{3}
}

func (x *ThirdExample) GetTest() string {
	if x != nil && x.Test != nil {
		return *x.Test
	}
	return ""
}

type CreateExampleRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name *string `protobuf:"bytes,1,opt,name=name,proto3,oneof" json:"name,omitempty"`
}

func (x *CreateExampleRequest) Reset() {
	*x = CreateExampleRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_gorm_example_example_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateExampleRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateExampleRequest) ProtoMessage() {}

func (x *CreateExampleRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_gorm_example_example_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateExampleRequest.ProtoReflect.Descriptor instead.
func (*CreateExampleRequest) Descriptor() ([]byte, []int) {
	return file_protoc_gen_gorm_example_example_proto_rawDescGZIP(), []int{4}
}

func (x *CreateExampleRequest) GetName() string {
	if x != nil && x.Name != nil {
		return *x.Name
	}
	return ""
}

type CreateExampleResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name *string `protobuf:"bytes,1,opt,name=name,proto3,oneof" json:"name,omitempty"`
}

func (x *CreateExampleResponse) Reset() {
	*x = CreateExampleResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_gorm_example_example_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateExampleResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateExampleResponse) ProtoMessage() {}

func (x *CreateExampleResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_gorm_example_example_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateExampleResponse.ProtoReflect.Descriptor instead.
func (*CreateExampleResponse) Descriptor() ([]byte, []int) {
	return file_protoc_gen_gorm_example_example_proto_rawDescGZIP(), []int{5}
}

func (x *CreateExampleResponse) GetName() string {
	if x != nil && x.Name != nil {
		return *x.Name
	}
	return ""
}

type ListExampleRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ListExampleRequest) Reset() {
	*x = ListExampleRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_gorm_example_example_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListExampleRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListExampleRequest) ProtoMessage() {}

func (x *ListExampleRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_gorm_example_example_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListExampleRequest.ProtoReflect.Descriptor instead.
func (*ListExampleRequest) Descriptor() ([]byte, []int) {
	return file_protoc_gen_gorm_example_example_proto_rawDescGZIP(), []int{6}
}

type ListExampleResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []*Example `protobuf:"bytes,1,rep,name=data,proto3" json:"data,omitempty"`
}

func (x *ListExampleResponse) Reset() {
	*x = ListExampleResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_gorm_example_example_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListExampleResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListExampleResponse) ProtoMessage() {}

func (x *ListExampleResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_gorm_example_example_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListExampleResponse.ProtoReflect.Descriptor instead.
func (*ListExampleResponse) Descriptor() ([]byte, []int) {
	return file_protoc_gen_gorm_example_example_proto_rawDescGZIP(), []int{7}
}

func (x *ListExampleResponse) GetData() []*Example {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_protoc_gen_gorm_example_example_proto protoreflect.FileDescriptor

var file_protoc_gen_gorm_example_example_proto_rawDesc = []byte{
	0x0a, 0x25, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x67, 0x6f, 0x72,
	0x6d, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x1a, 0x11, 0x72, 0x65, 0x74, 0x61, 0x67, 0x2f, 0x72, 0x65, 0x74, 0x61, 0x67, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x0e, 0x6f, 0x72, 0x6d, 0x2f, 0x67, 0x6f, 0x72, 0x6d, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xe6, 0x01, 0x0a, 0x04, 0x55, 0x73, 0x65, 0x72, 0x12,
	0x16, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x42, 0x06, 0x8a, 0xea, 0x30,
	0x02, 0x28, 0x01, 0x52, 0x02, 0x69, 0x64, 0x12, 0x39, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64,
	0x41, 0x74, 0x12, 0x39, 0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x36, 0x0a,
	0x08, 0x62, 0x69, 0x72, 0x74, 0x68, 0x64, 0x61, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x08, 0x62, 0x69, 0x72,
	0x74, 0x68, 0x64, 0x61, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6e, 0x75, 0x6d, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x03, 0x6e, 0x75, 0x6d, 0x3a, 0x06, 0x8a, 0xea, 0x30, 0x02, 0x08, 0x01, 0x22,
	0xd1, 0x08, 0x0a, 0x07, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x12, 0x47, 0x0a, 0x0d, 0x77,
	0x69, 0x74, 0x68, 0x5f, 0x6e, 0x65, 0x77, 0x5f, 0x74, 0x61, 0x67, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x23, 0x82, 0xea, 0x30, 0x1f, 0x0a, 0x07, 0x67, 0x72, 0x61, 0x70, 0x68, 0x71,
	0x6c, 0x12, 0x14, 0x77, 0x69, 0x74, 0x68, 0x4e, 0x65, 0x77, 0x54, 0x61, 0x67, 0x73, 0x2c, 0x6f,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x52, 0x0b, 0x77, 0x69, 0x74, 0x68, 0x4e, 0x65, 0x77,
	0x54, 0x61, 0x67, 0x73, 0x12, 0x69, 0x0a, 0x11, 0x77, 0x69, 0x74, 0x68, 0x5f, 0x6e, 0x65, 0x77,
	0x5f, 0x6d, 0x75, 0x6c, 0x74, 0x69, 0x70, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42,
	0x3d, 0x82, 0xea, 0x30, 0x1f, 0x0a, 0x07, 0x67, 0x72, 0x61, 0x70, 0x68, 0x71, 0x6c, 0x12, 0x14,
	0x77, 0x69, 0x74, 0x68, 0x4e, 0x65, 0x77, 0x54, 0x61, 0x67, 0x73, 0x2c, 0x6f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x61, 0x6c, 0x82, 0xea, 0x30, 0x16, 0x0a, 0x03, 0x78, 0x6d, 0x6c, 0x12, 0x0f, 0x6d,
	0x75, 0x6c, 0x74, 0x69, 0x2c, 0x6f, 0x6d, 0x69, 0x74, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x52, 0x0f,
	0x77, 0x69, 0x74, 0x68, 0x4e, 0x65, 0x77, 0x4d, 0x75, 0x6c, 0x74, 0x69, 0x70, 0x6c, 0x65, 0x12,
	0x44, 0x0a, 0x0f, 0x72, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x5f, 0x64, 0x65, 0x66, 0x61, 0x75,
	0x6c, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x1b, 0x82, 0xea, 0x30, 0x17, 0x0a, 0x04,
	0x6a, 0x73, 0x6f, 0x6e, 0x12, 0x0f, 0x72, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x50, 0x72, 0x65,
	0x76, 0x69, 0x6f, 0x75, 0x73, 0x52, 0x0e, 0x72, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x44, 0x65,
	0x66, 0x61, 0x75, 0x6c, 0x74, 0x12, 0x1b, 0x0a, 0x01, 0x61, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09,
	0x42, 0x0d, 0x82, 0xea, 0x30, 0x09, 0x0a, 0x04, 0x6a, 0x73, 0x6f, 0x6e, 0x12, 0x01, 0x41, 0x52,
	0x01, 0x61, 0x12, 0x23, 0x0a, 0x04, 0x62, 0x5f, 0x6a, 0x6b, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05,
	0x42, 0x10, 0x82, 0xea, 0x30, 0x0c, 0x0a, 0x04, 0x6a, 0x73, 0x6f, 0x6e, 0x12, 0x04, 0x62, 0x5f,
	0x4a, 0x6b, 0x52, 0x03, 0x62, 0x4a, 0x6b, 0x12, 0x2d, 0x0a, 0x06, 0x74, 0x65, 0x73, 0x74, 0x5f,
	0x31, 0x18, 0x0a, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c,
	0x65, 0x2e, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52,
	0x05, 0x74, 0x65, 0x73, 0x74, 0x31, 0x12, 0x32, 0x0a, 0x06, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x32,
	0x18, 0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x2e, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x00,
	0x52, 0x05, 0x74, 0x65, 0x73, 0x74, 0x32, 0x88, 0x01, 0x01, 0x12, 0x2f, 0x0a, 0x07, 0x74, 0x65,
	0x73, 0x74, 0x5f, 0x36, 0x31, 0x18, 0x13, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x65, 0x78,
	0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x52, 0x06, 0x74, 0x65, 0x73, 0x74, 0x36, 0x31, 0x12, 0x35, 0x0a, 0x07, 0x74,
	0x65, 0x73, 0x74, 0x5f, 0x33, 0x31, 0x18, 0x0d, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x65,
	0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x54,
	0x65, 0x73, 0x74, 0x33, 0x31, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x74, 0x65, 0x73, 0x74,
	0x33, 0x31, 0x12, 0x31, 0x0a, 0x06, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x35, 0x18, 0x0f, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x05,
	0x74, 0x65, 0x73, 0x74, 0x35, 0x12, 0x33, 0x0a, 0x07, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x35, 0x31,
	0x18, 0x12, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x52, 0x06, 0x74, 0x65, 0x73, 0x74, 0x35, 0x31, 0x12, 0x38, 0x0a, 0x07, 0x74, 0x65,
	0x73, 0x74, 0x5f, 0x31, 0x37, 0x18, 0x11, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x48, 0x01, 0x52, 0x06, 0x74, 0x65, 0x73, 0x74, 0x31,
	0x37, 0x88, 0x01, 0x01, 0x12, 0x32, 0x0a, 0x06, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x33, 0x18, 0x0c,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x45,
	0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x54, 0x65, 0x73, 0x74, 0x33, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x05, 0x74, 0x65, 0x73, 0x74, 0x33, 0x12, 0x1f, 0x0a, 0x0b, 0x6c, 0x69, 0x73, 0x74,
	0x5f, 0x74, 0x65, 0x73, 0x74, 0x32, 0x30, 0x18, 0x14, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0a, 0x6c,
	0x69, 0x73, 0x74, 0x54, 0x65, 0x73, 0x74, 0x32, 0x30, 0x12, 0x41, 0x0a, 0x0b, 0x6c, 0x69, 0x73,
	0x74, 0x5f, 0x74, 0x65, 0x73, 0x74, 0x32, 0x31, 0x18, 0x15, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20,
	0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x2e, 0x4c, 0x69, 0x73, 0x74, 0x54, 0x65, 0x73, 0x74, 0x32, 0x31, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x0a, 0x6c, 0x69, 0x73, 0x74, 0x54, 0x65, 0x73, 0x74, 0x32, 0x31, 0x1a, 0x51, 0x0a, 0x0b,
	0x54, 0x65, 0x73, 0x74, 0x33, 0x31, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x65,
	0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a,
	0x54, 0x0a, 0x0a, 0x54, 0x65, 0x73, 0x74, 0x33, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x30, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x3d, 0x0a, 0x0f, 0x4c, 0x69, 0x73, 0x74, 0x54, 0x65, 0x73,
	0x74, 0x32, 0x31, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x3a, 0x06, 0x8a, 0xea, 0x30, 0x02, 0x08, 0x01, 0x42, 0x09, 0x0a, 0x07,
	0x5f, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x32, 0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x74, 0x65, 0x73, 0x74,
	0x5f, 0x31, 0x37, 0x22, 0x91, 0x02, 0x0a, 0x0d, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x47, 0x0a, 0x0d, 0x77, 0x69, 0x74, 0x68, 0x5f, 0x6e, 0x65,
	0x77, 0x5f, 0x74, 0x61, 0x67, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x23, 0x82, 0xea,
	0x30, 0x1f, 0x0a, 0x07, 0x67, 0x72, 0x61, 0x70, 0x68, 0x71, 0x6c, 0x12, 0x14, 0x77, 0x69, 0x74,
	0x68, 0x4e, 0x65, 0x77, 0x54, 0x61, 0x67, 0x73, 0x2c, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61,
	0x6c, 0x52, 0x0b, 0x77, 0x69, 0x74, 0x68, 0x4e, 0x65, 0x77, 0x54, 0x61, 0x67, 0x73, 0x12, 0x69,
	0x0a, 0x11, 0x77, 0x69, 0x74, 0x68, 0x5f, 0x6e, 0x65, 0x77, 0x5f, 0x6d, 0x75, 0x6c, 0x74, 0x69,
	0x70, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x3d, 0x82, 0xea, 0x30, 0x1f, 0x0a,
	0x07, 0x67, 0x72, 0x61, 0x70, 0x68, 0x71, 0x6c, 0x12, 0x14, 0x77, 0x69, 0x74, 0x68, 0x4e, 0x65,
	0x77, 0x54, 0x61, 0x67, 0x73, 0x2c, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x82, 0xea,
	0x30, 0x16, 0x0a, 0x03, 0x78, 0x6d, 0x6c, 0x12, 0x0f, 0x6d, 0x75, 0x6c, 0x74, 0x69, 0x2c, 0x6f,
	0x6d, 0x69, 0x74, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x52, 0x0f, 0x77, 0x69, 0x74, 0x68, 0x4e, 0x65,
	0x77, 0x4d, 0x75, 0x6c, 0x74, 0x69, 0x70, 0x6c, 0x65, 0x12, 0x44, 0x0a, 0x0f, 0x72, 0x65, 0x70,
	0x6c, 0x61, 0x63, 0x65, 0x5f, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x1b, 0x82, 0xea, 0x30, 0x17, 0x0a, 0x04, 0x6a, 0x73, 0x6f, 0x6e, 0x12, 0x0f,
	0x72, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x50, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73, 0x52,
	0x0e, 0x72, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x44, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x3a,
	0x06, 0x8a, 0xea, 0x30, 0x02, 0x08, 0x01, 0x22, 0x4b, 0x0a, 0x0c, 0x54, 0x68, 0x69, 0x72, 0x64,
	0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x12, 0x17, 0x0a, 0x04, 0x74, 0x65, 0x73, 0x74, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x04, 0x74, 0x65, 0x73, 0x74, 0x88, 0x01, 0x01,
	0x3a, 0x19, 0x8a, 0xea, 0x30, 0x15, 0x08, 0x01, 0x12, 0x11, 0x73, 0x79, 0x73, 0x5f, 0x74, 0x68,
	0x69, 0x72, 0x64, 0x5f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x42, 0x07, 0x0a, 0x05, 0x5f,
	0x74, 0x65, 0x73, 0x74, 0x22, 0x38, 0x0a, 0x14, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x45, 0x78,
	0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x88, 0x01, 0x01, 0x42, 0x07, 0x0a, 0x05, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x39,
	0x0a, 0x15, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x17, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x88, 0x01, 0x01,
	0x42, 0x07, 0x0a, 0x05, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x14, 0x0a, 0x12, 0x4c, 0x69, 0x73,
	0x74, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22,
	0x3b, 0x0a, 0x13, 0x4c, 0x69, 0x73, 0x74, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x24, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x45,
	0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x32, 0x8f, 0x03, 0x0a,
	0x0e, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x49, 0x0a, 0x06, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x1d, 0x2e, 0x65, 0x78, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c,
	0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70,
	0x6c, 0x65, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x49, 0x0a, 0x06, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x12, 0x1d, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x49, 0x0a, 0x06, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12,
	0x1d, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e,
	0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x45,
	0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x12, 0x46, 0x0a, 0x03, 0x47, 0x65, 0x74, 0x12, 0x1d, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c,
	0x65, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x43, 0x0a, 0x04, 0x4c, 0x69, 0x73, 0x74,
	0x12, 0x1b, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x45,
	0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e,
	0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x45, 0x78, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x1a, 0x0f, 0x8a,
	0xea, 0x30, 0x0b, 0x12, 0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x18, 0x01, 0x42, 0x41,
	0x5a, 0x3f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x75, 0x62,
	0x67, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2f, 0x70, 0x6b,
	0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x67, 0x6f, 0x72,
	0x6d, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x3b, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c,
	0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protoc_gen_gorm_example_example_proto_rawDescOnce sync.Once
	file_protoc_gen_gorm_example_example_proto_rawDescData = file_protoc_gen_gorm_example_example_proto_rawDesc
)

func file_protoc_gen_gorm_example_example_proto_rawDescGZIP() []byte {
	file_protoc_gen_gorm_example_example_proto_rawDescOnce.Do(func() {
		file_protoc_gen_gorm_example_example_proto_rawDescData = protoimpl.X.CompressGZIP(file_protoc_gen_gorm_example_example_proto_rawDescData)
	})
	return file_protoc_gen_gorm_example_example_proto_rawDescData
}

var file_protoc_gen_gorm_example_example_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_protoc_gen_gorm_example_example_proto_goTypes = []interface{}{
	(*User)(nil),                  // 0: example.User
	(*Example)(nil),               // 1: example.Example
	(*SecondMessage)(nil),         // 2: example.SecondMessage
	(*ThirdExample)(nil),          // 3: example.ThirdExample
	(*CreateExampleRequest)(nil),  // 4: example.CreateExampleRequest
	(*CreateExampleResponse)(nil), // 5: example.CreateExampleResponse
	(*ListExampleRequest)(nil),    // 6: example.ListExampleRequest
	(*ListExampleResponse)(nil),   // 7: example.ListExampleResponse
	nil,                           // 8: example.Example.Test31Entry
	nil,                           // 9: example.Example.Test3Entry
	nil,                           // 10: example.Example.ListTest21Entry
	(*timestamppb.Timestamp)(nil), // 11: google.protobuf.Timestamp
}
var file_protoc_gen_gorm_example_example_proto_depIdxs = []int32{
	11, // 0: example.User.created_at:type_name -> google.protobuf.Timestamp
	11, // 1: example.User.updated_at:type_name -> google.protobuf.Timestamp
	11, // 2: example.User.birthday:type_name -> google.protobuf.Timestamp
	2,  // 3: example.Example.test_1:type_name -> example.SecondMessage
	2,  // 4: example.Example.test_2:type_name -> example.SecondMessage
	2,  // 5: example.Example.test_61:type_name -> example.SecondMessage
	8,  // 6: example.Example.test_31:type_name -> example.Example.Test31Entry
	11, // 7: example.Example.test_5:type_name -> google.protobuf.Timestamp
	11, // 8: example.Example.test_51:type_name -> google.protobuf.Timestamp
	11, // 9: example.Example.test_17:type_name -> google.protobuf.Timestamp
	9,  // 10: example.Example.test_3:type_name -> example.Example.Test3Entry
	10, // 11: example.Example.list_test21:type_name -> example.Example.ListTest21Entry
	1,  // 12: example.ListExampleResponse.data:type_name -> example.Example
	2,  // 13: example.Example.Test31Entry.value:type_name -> example.SecondMessage
	11, // 14: example.Example.Test3Entry.value:type_name -> google.protobuf.Timestamp
	4,  // 15: example.ExampleService.Create:input_type -> example.CreateExampleRequest
	4,  // 16: example.ExampleService.Delete:input_type -> example.CreateExampleRequest
	4,  // 17: example.ExampleService.Update:input_type -> example.CreateExampleRequest
	4,  // 18: example.ExampleService.Get:input_type -> example.CreateExampleRequest
	6,  // 19: example.ExampleService.List:input_type -> example.ListExampleRequest
	5,  // 20: example.ExampleService.Create:output_type -> example.CreateExampleResponse
	5,  // 21: example.ExampleService.Delete:output_type -> example.CreateExampleResponse
	5,  // 22: example.ExampleService.Update:output_type -> example.CreateExampleResponse
	5,  // 23: example.ExampleService.Get:output_type -> example.CreateExampleResponse
	7,  // 24: example.ExampleService.List:output_type -> example.ListExampleResponse
	20, // [20:25] is the sub-list for method output_type
	15, // [15:20] is the sub-list for method input_type
	15, // [15:15] is the sub-list for extension type_name
	15, // [15:15] is the sub-list for extension extendee
	0,  // [0:15] is the sub-list for field type_name
}

func init() { file_protoc_gen_gorm_example_example_proto_init() }
func file_protoc_gen_gorm_example_example_proto_init() {
	if File_protoc_gen_gorm_example_example_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protoc_gen_gorm_example_example_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*User); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protoc_gen_gorm_example_example_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Example); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protoc_gen_gorm_example_example_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SecondMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protoc_gen_gorm_example_example_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ThirdExample); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protoc_gen_gorm_example_example_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateExampleRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protoc_gen_gorm_example_example_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateExampleResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protoc_gen_gorm_example_example_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListExampleRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protoc_gen_gorm_example_example_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListExampleResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_protoc_gen_gorm_example_example_proto_msgTypes[1].OneofWrappers = []interface{}{}
	file_protoc_gen_gorm_example_example_proto_msgTypes[3].OneofWrappers = []interface{}{}
	file_protoc_gen_gorm_example_example_proto_msgTypes[4].OneofWrappers = []interface{}{}
	file_protoc_gen_gorm_example_example_proto_msgTypes[5].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_protoc_gen_gorm_example_example_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_protoc_gen_gorm_example_example_proto_goTypes,
		DependencyIndexes: file_protoc_gen_gorm_example_example_proto_depIdxs,
		MessageInfos:      file_protoc_gen_gorm_example_example_proto_msgTypes,
	}.Build()
	File_protoc_gen_gorm_example_example_proto = out.File
	file_protoc_gen_gorm_example_example_proto_rawDesc = nil
	file_protoc_gen_gorm_example_example_proto_goTypes = nil
	file_protoc_gen_gorm_example_example_proto_depIdxs = nil
}
