// Signature is a generic structure for signed byte arrays

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.3
// 	protoc        (unknown)
// source: message_contents/signed_payload.proto

package message_contents

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// SignedPayload is a wrapper for a signature and a payload
type SignedPayload struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Payload       []byte                 `protobuf:"bytes,1,opt,name=payload,proto3" json:"payload,omitempty"`
	Signature     *Signature             `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SignedPayload) Reset() {
	*x = SignedPayload{}
	mi := &file_message_contents_signed_payload_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SignedPayload) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignedPayload) ProtoMessage() {}

func (x *SignedPayload) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_signed_payload_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignedPayload.ProtoReflect.Descriptor instead.
func (*SignedPayload) Descriptor() ([]byte, []int) {
	return file_message_contents_signed_payload_proto_rawDescGZIP(), []int{0}
}

func (x *SignedPayload) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *SignedPayload) GetSignature() *Signature {
	if x != nil {
		return x.Signature
	}
	return nil
}

var File_message_contents_signed_payload_proto protoreflect.FileDescriptor

var file_message_contents_signed_payload_proto_rawDesc = []byte{
	0x0a, 0x25, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x73, 0x2f, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x5f, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61,
	0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x1a, 0x20,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73,
	0x2f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x69, 0x0a, 0x0d, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61,
	0x64, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x3e, 0x0a, 0x09, 0x73,
	0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20,
	0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x42, 0xd9, 0x01, 0x0a, 0x19,
	0x63, 0x6f, 0x6d, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x42, 0x12, 0x53, 0x69, 0x67, 0x6e, 0x65,
	0x64, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a,
	0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x78, 0x6d, 0x74, 0x70,
	0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2d, 0x6e, 0x6f, 0x64, 0x65, 0x2d, 0x67, 0x6f, 0x2f, 0x70, 0x6b,
	0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xa2, 0x02, 0x03, 0x58, 0x4d, 0x58, 0xaa, 0x02,
	0x14, 0x58, 0x6d, 0x74, 0x70, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x73, 0xca, 0x02, 0x14, 0x58, 0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xe2, 0x02, 0x20, 0x58,
	0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x73, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea,
	0x02, 0x15, 0x58, 0x6d, 0x74, 0x70, 0x3a, 0x3a, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_message_contents_signed_payload_proto_rawDescOnce sync.Once
	file_message_contents_signed_payload_proto_rawDescData = file_message_contents_signed_payload_proto_rawDesc
)

func file_message_contents_signed_payload_proto_rawDescGZIP() []byte {
	file_message_contents_signed_payload_proto_rawDescOnce.Do(func() {
		file_message_contents_signed_payload_proto_rawDescData = protoimpl.X.CompressGZIP(file_message_contents_signed_payload_proto_rawDescData)
	})
	return file_message_contents_signed_payload_proto_rawDescData
}

var file_message_contents_signed_payload_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_message_contents_signed_payload_proto_goTypes = []any{
	(*SignedPayload)(nil), // 0: xmtp.message_contents.SignedPayload
	(*Signature)(nil),     // 1: xmtp.message_contents.Signature
}
var file_message_contents_signed_payload_proto_depIdxs = []int32{
	1, // 0: xmtp.message_contents.SignedPayload.signature:type_name -> xmtp.message_contents.Signature
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_message_contents_signed_payload_proto_init() }
func file_message_contents_signed_payload_proto_init() {
	if File_message_contents_signed_payload_proto != nil {
		return
	}
	file_message_contents_signature_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_message_contents_signed_payload_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_message_contents_signed_payload_proto_goTypes,
		DependencyIndexes: file_message_contents_signed_payload_proto_depIdxs,
		MessageInfos:      file_message_contents_signed_payload_proto_msgTypes,
	}.Build()
	File_message_contents_signed_payload_proto = out.File
	file_message_contents_signed_payload_proto_rawDesc = nil
	file_message_contents_signed_payload_proto_goTypes = nil
	file_message_contents_signed_payload_proto_depIdxs = nil
}
