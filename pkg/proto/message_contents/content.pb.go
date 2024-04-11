// Message content encoding structures

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: message_contents/content.proto

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

// Recognized compression algorithms
// protolint:disable ENUM_FIELD_NAMES_ZERO_VALUE_END_WITH
type Compression int32

const (
	Compression_COMPRESSION_DEFLATE Compression = 0
	Compression_COMPRESSION_GZIP    Compression = 1
)

// Enum value maps for Compression.
var (
	Compression_name = map[int32]string{
		0: "COMPRESSION_DEFLATE",
		1: "COMPRESSION_GZIP",
	}
	Compression_value = map[string]int32{
		"COMPRESSION_DEFLATE": 0,
		"COMPRESSION_GZIP":    1,
	}
)

func (x Compression) Enum() *Compression {
	p := new(Compression)
	*p = x
	return p
}

func (x Compression) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Compression) Descriptor() protoreflect.EnumDescriptor {
	return file_message_contents_content_proto_enumTypes[0].Descriptor()
}

func (Compression) Type() protoreflect.EnumType {
	return &file_message_contents_content_proto_enumTypes[0]
}

func (x Compression) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Compression.Descriptor instead.
func (Compression) EnumDescriptor() ([]byte, []int) {
	return file_message_contents_content_proto_rawDescGZIP(), []int{0}
}

// ContentTypeId is used to identify the type of content stored in a Message.
type ContentTypeId struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AuthorityId  string `protobuf:"bytes,1,opt,name=authority_id,json=authorityId,proto3" json:"authority_id,omitempty"`     // authority governing this content type
	TypeId       string `protobuf:"bytes,2,opt,name=type_id,json=typeId,proto3" json:"type_id,omitempty"`                    // type identifier
	VersionMajor uint32 `protobuf:"varint,3,opt,name=version_major,json=versionMajor,proto3" json:"version_major,omitempty"` // major version of the type
	VersionMinor uint32 `protobuf:"varint,4,opt,name=version_minor,json=versionMinor,proto3" json:"version_minor,omitempty"` // minor version of the type
}

func (x *ContentTypeId) Reset() {
	*x = ContentTypeId{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_contents_content_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ContentTypeId) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ContentTypeId) ProtoMessage() {}

func (x *ContentTypeId) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_content_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ContentTypeId.ProtoReflect.Descriptor instead.
func (*ContentTypeId) Descriptor() ([]byte, []int) {
	return file_message_contents_content_proto_rawDescGZIP(), []int{0}
}

func (x *ContentTypeId) GetAuthorityId() string {
	if x != nil {
		return x.AuthorityId
	}
	return ""
}

func (x *ContentTypeId) GetTypeId() string {
	if x != nil {
		return x.TypeId
	}
	return ""
}

func (x *ContentTypeId) GetVersionMajor() uint32 {
	if x != nil {
		return x.VersionMajor
	}
	return 0
}

func (x *ContentTypeId) GetVersionMinor() uint32 {
	if x != nil {
		return x.VersionMinor
	}
	return 0
}

// EncodedContent bundles the content with metadata identifying its type
// and parameters required for correct decoding and presentation of the content.
type EncodedContent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// content type identifier used to match the payload with
	// the correct decoding machinery
	Type *ContentTypeId `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	// optional encoding parameters required to correctly decode the content
	Parameters map[string]string `protobuf:"bytes,2,rep,name=parameters,proto3" json:"parameters,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// optional fallback description of the content that can be used in case
	// the client cannot decode or render the content
	Fallback *string `protobuf:"bytes,3,opt,name=fallback,proto3,oneof" json:"fallback,omitempty"`
	// optional compression; the value indicates algorithm used to
	// compress the encoded content bytes
	Compression *Compression `protobuf:"varint,5,opt,name=compression,proto3,enum=xmtp.message_contents.Compression,oneof" json:"compression,omitempty"`
	// encoded content itself
	Content []byte `protobuf:"bytes,4,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *EncodedContent) Reset() {
	*x = EncodedContent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_contents_content_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EncodedContent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EncodedContent) ProtoMessage() {}

func (x *EncodedContent) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_content_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EncodedContent.ProtoReflect.Descriptor instead.
func (*EncodedContent) Descriptor() ([]byte, []int) {
	return file_message_contents_content_proto_rawDescGZIP(), []int{1}
}

func (x *EncodedContent) GetType() *ContentTypeId {
	if x != nil {
		return x.Type
	}
	return nil
}

func (x *EncodedContent) GetParameters() map[string]string {
	if x != nil {
		return x.Parameters
	}
	return nil
}

func (x *EncodedContent) GetFallback() string {
	if x != nil && x.Fallback != nil {
		return *x.Fallback
	}
	return ""
}

func (x *EncodedContent) GetCompression() Compression {
	if x != nil && x.Compression != nil {
		return *x.Compression
	}
	return Compression_COMPRESSION_DEFLATE
}

func (x *EncodedContent) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

// SignedContent attaches a signature to EncodedContent.
type SignedContent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// MUST contain EncodedContent
	Payload []byte                 `protobuf:"bytes,1,opt,name=payload,proto3" json:"payload,omitempty"`
	Sender  *SignedPublicKeyBundle `protobuf:"bytes,2,opt,name=sender,proto3" json:"sender,omitempty"`
	// MUST be a signature of a concatenation of
	// the message header bytes and the payload bytes,
	// signed by the sender's pre-key.
	Signature *Signature `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (x *SignedContent) Reset() {
	*x = SignedContent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_contents_content_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignedContent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignedContent) ProtoMessage() {}

func (x *SignedContent) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_content_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignedContent.ProtoReflect.Descriptor instead.
func (*SignedContent) Descriptor() ([]byte, []int) {
	return file_message_contents_content_proto_rawDescGZIP(), []int{2}
}

func (x *SignedContent) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *SignedContent) GetSender() *SignedPublicKeyBundle {
	if x != nil {
		return x.Sender
	}
	return nil
}

func (x *SignedContent) GetSignature() *Signature {
	if x != nil {
		return x.Signature
	}
	return nil
}

var File_message_contents_content_proto protoreflect.FileDescriptor

var file_message_contents_content_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x15, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x1a, 0x21, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63,
	0x5f, 0x6b, 0x65, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x73, 0x69, 0x67,
	0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x95, 0x01, 0x0a,
	0x0d, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x49, 0x64, 0x12, 0x21,
	0x0a, 0x0c, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x5f, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x49,
	0x64, 0x12, 0x17, 0x0a, 0x07, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x74, 0x79, 0x70, 0x65, 0x49, 0x64, 0x12, 0x23, 0x0a, 0x0d, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x6d, 0x61, 0x6a, 0x6f, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x0c, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x4d, 0x61, 0x6a, 0x6f, 0x72, 0x12,
	0x23, 0x0a, 0x0d, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x6d, 0x69, 0x6e, 0x6f, 0x72,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0c, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x4d,
	0x69, 0x6e, 0x6f, 0x72, 0x22, 0x83, 0x03, 0x0a, 0x0e, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x65, 0x64,
	0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x38, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x43, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x49, 0x64, 0x52, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x12, 0x55, 0x0a, 0x0a, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x35, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x45, 0x6e,
	0x63, 0x6f, 0x64, 0x65, 0x64, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2e, 0x50, 0x61, 0x72,
	0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0a, 0x70, 0x61,
	0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x12, 0x1f, 0x0a, 0x08, 0x66, 0x61, 0x6c, 0x6c,
	0x62, 0x61, 0x63, 0x6b, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x08, 0x66, 0x61,
	0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x88, 0x01, 0x01, 0x12, 0x49, 0x0a, 0x0b, 0x63, 0x6f, 0x6d,
	0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x22,
	0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x48, 0x01, 0x52, 0x0b, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x88, 0x01, 0x01, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x1a, 0x3d,
	0x0a, 0x0f, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x0b, 0x0a,
	0x09, 0x5f, 0x66, 0x61, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x42, 0x0e, 0x0a, 0x0c, 0x5f, 0x63,
	0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0xaf, 0x01, 0x0a, 0x0d, 0x53,
	0x69, 0x67, 0x6e, 0x65, 0x64, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x18, 0x0a, 0x07,
	0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x70,
	0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x44, 0x0a, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x53,
	0x69, 0x67, 0x6e, 0x65, 0x64, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x42, 0x75,
	0x6e, 0x64, 0x6c, 0x65, 0x52, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x12, 0x3e, 0x0a, 0x09,
	0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x20, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x2a, 0x3c, 0x0a, 0x0b,
	0x43, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x17, 0x0a, 0x13, 0x43,
	0x4f, 0x4d, 0x50, 0x52, 0x45, 0x53, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x44, 0x45, 0x46, 0x4c, 0x41,
	0x54, 0x45, 0x10, 0x00, 0x12, 0x14, 0x0a, 0x10, 0x43, 0x4f, 0x4d, 0x50, 0x52, 0x45, 0x53, 0x53,
	0x49, 0x4f, 0x4e, 0x5f, 0x47, 0x5a, 0x49, 0x50, 0x10, 0x01, 0x42, 0xd3, 0x01, 0x0a, 0x19, 0x63,
	0x6f, 0x6d, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x42, 0x0c, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2d, 0x6e,
	0x6f, 0x64, 0x65, 0x2d, 0x67, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x73, 0xa2, 0x02, 0x03, 0x58, 0x4d, 0x58, 0xaa, 0x02, 0x14, 0x58, 0x6d, 0x74, 0x70, 0x2e, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xca, 0x02,
	0x14, 0x58, 0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x73, 0xe2, 0x02, 0x20, 0x58, 0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x5c, 0x47, 0x50, 0x42,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x15, 0x58, 0x6d, 0x74, 0x70, 0x3a,
	0x3a, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_message_contents_content_proto_rawDescOnce sync.Once
	file_message_contents_content_proto_rawDescData = file_message_contents_content_proto_rawDesc
)

func file_message_contents_content_proto_rawDescGZIP() []byte {
	file_message_contents_content_proto_rawDescOnce.Do(func() {
		file_message_contents_content_proto_rawDescData = protoimpl.X.CompressGZIP(file_message_contents_content_proto_rawDescData)
	})
	return file_message_contents_content_proto_rawDescData
}

var file_message_contents_content_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_message_contents_content_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_message_contents_content_proto_goTypes = []interface{}{
	(Compression)(0),              // 0: xmtp.message_contents.Compression
	(*ContentTypeId)(nil),         // 1: xmtp.message_contents.ContentTypeId
	(*EncodedContent)(nil),        // 2: xmtp.message_contents.EncodedContent
	(*SignedContent)(nil),         // 3: xmtp.message_contents.SignedContent
	nil,                           // 4: xmtp.message_contents.EncodedContent.ParametersEntry
	(*SignedPublicKeyBundle)(nil), // 5: xmtp.message_contents.SignedPublicKeyBundle
	(*Signature)(nil),             // 6: xmtp.message_contents.Signature
}
var file_message_contents_content_proto_depIdxs = []int32{
	1, // 0: xmtp.message_contents.EncodedContent.type:type_name -> xmtp.message_contents.ContentTypeId
	4, // 1: xmtp.message_contents.EncodedContent.parameters:type_name -> xmtp.message_contents.EncodedContent.ParametersEntry
	0, // 2: xmtp.message_contents.EncodedContent.compression:type_name -> xmtp.message_contents.Compression
	5, // 3: xmtp.message_contents.SignedContent.sender:type_name -> xmtp.message_contents.SignedPublicKeyBundle
	6, // 4: xmtp.message_contents.SignedContent.signature:type_name -> xmtp.message_contents.Signature
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_message_contents_content_proto_init() }
func file_message_contents_content_proto_init() {
	if File_message_contents_content_proto != nil {
		return
	}
	file_message_contents_public_key_proto_init()
	file_message_contents_signature_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_message_contents_content_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ContentTypeId); i {
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
		file_message_contents_content_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EncodedContent); i {
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
		file_message_contents_content_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignedContent); i {
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
	file_message_contents_content_proto_msgTypes[1].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_message_contents_content_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_message_contents_content_proto_goTypes,
		DependencyIndexes: file_message_contents_content_proto_depIdxs,
		EnumInfos:         file_message_contents_content_proto_enumTypes,
		MessageInfos:      file_message_contents_content_proto_msgTypes,
	}.Build()
	File_message_contents_content_proto = out.File
	file_message_contents_content_proto_rawDesc = nil
	file_message_contents_content_proto_goTypes = nil
	file_message_contents_content_proto_depIdxs = nil
}
