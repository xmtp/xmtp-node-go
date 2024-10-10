// Message content encoding structures
// Copied from V2 code so that we can eventually retire all V2 message content

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        (unknown)
// source: mls/message_contents/content.proto

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
	return file_mls_message_contents_content_proto_enumTypes[0].Descriptor()
}

func (Compression) Type() protoreflect.EnumType {
	return &file_mls_message_contents_content_proto_enumTypes[0]
}

func (x Compression) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Compression.Descriptor instead.
func (Compression) EnumDescriptor() ([]byte, []int) {
	return file_mls_message_contents_content_proto_rawDescGZIP(), []int{0}
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
	mi := &file_mls_message_contents_content_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ContentTypeId) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ContentTypeId) ProtoMessage() {}

func (x *ContentTypeId) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_content_proto_msgTypes[0]
	if x != nil {
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
	return file_mls_message_contents_content_proto_rawDescGZIP(), []int{0}
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
	Compression *Compression `protobuf:"varint,5,opt,name=compression,proto3,enum=xmtp.mls.message_contents.Compression,oneof" json:"compression,omitempty"`
	// encoded content itself
	Content []byte `protobuf:"bytes,4,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *EncodedContent) Reset() {
	*x = EncodedContent{}
	mi := &file_mls_message_contents_content_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EncodedContent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EncodedContent) ProtoMessage() {}

func (x *EncodedContent) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_content_proto_msgTypes[1]
	if x != nil {
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
	return file_mls_message_contents_content_proto_rawDescGZIP(), []int{1}
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

// A PlaintextEnvelope is the outermost payload that gets encrypted by MLS
type PlaintextEnvelope struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Selector which declares which version of the EncodedContent this
	// PlaintextEnvelope is
	//
	// Types that are assignable to Content:
	//
	//	*PlaintextEnvelope_V1_
	//	*PlaintextEnvelope_V2_
	Content isPlaintextEnvelope_Content `protobuf_oneof:"content"`
}

func (x *PlaintextEnvelope) Reset() {
	*x = PlaintextEnvelope{}
	mi := &file_mls_message_contents_content_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PlaintextEnvelope) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PlaintextEnvelope) ProtoMessage() {}

func (x *PlaintextEnvelope) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_content_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PlaintextEnvelope.ProtoReflect.Descriptor instead.
func (*PlaintextEnvelope) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_content_proto_rawDescGZIP(), []int{2}
}

func (m *PlaintextEnvelope) GetContent() isPlaintextEnvelope_Content {
	if m != nil {
		return m.Content
	}
	return nil
}

func (x *PlaintextEnvelope) GetV1() *PlaintextEnvelope_V1 {
	if x, ok := x.GetContent().(*PlaintextEnvelope_V1_); ok {
		return x.V1
	}
	return nil
}

func (x *PlaintextEnvelope) GetV2() *PlaintextEnvelope_V2 {
	if x, ok := x.GetContent().(*PlaintextEnvelope_V2_); ok {
		return x.V2
	}
	return nil
}

type isPlaintextEnvelope_Content interface {
	isPlaintextEnvelope_Content()
}

type PlaintextEnvelope_V1_ struct {
	V1 *PlaintextEnvelope_V1 `protobuf:"bytes,1,opt,name=v1,proto3,oneof"`
}

type PlaintextEnvelope_V2_ struct {
	V2 *PlaintextEnvelope_V2 `protobuf:"bytes,2,opt,name=v2,proto3,oneof"`
}

func (*PlaintextEnvelope_V1_) isPlaintextEnvelope_Content() {}

func (*PlaintextEnvelope_V2_) isPlaintextEnvelope_Content() {}

// Initiator or new installation id requesting a history will send a request
type MessageHistoryRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Unique identifier for each request
	RequestId string `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	// Ensures a human is in the loop
	PinCode string `protobuf:"bytes,2,opt,name=pin_code,json=pinCode,proto3" json:"pin_code,omitempty"`
}

func (x *MessageHistoryRequest) Reset() {
	*x = MessageHistoryRequest{}
	mi := &file_mls_message_contents_content_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MessageHistoryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageHistoryRequest) ProtoMessage() {}

func (x *MessageHistoryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_content_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageHistoryRequest.ProtoReflect.Descriptor instead.
func (*MessageHistoryRequest) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_content_proto_rawDescGZIP(), []int{3}
}

func (x *MessageHistoryRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

func (x *MessageHistoryRequest) GetPinCode() string {
	if x != nil {
		return x.PinCode
	}
	return ""
}

// Pre-existing installation id capable of supplying a history sends this reply
type MessageHistoryReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Must match an existing request_id from a message history request
	RequestId string `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	// Where the messages can be retrieved from
	Url string `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
	// Generated input 'secret' for the AES Key used to encrypt the message-bundle
	EncryptionKey *MessageHistoryKeyType `protobuf:"bytes,3,opt,name=encryption_key,json=encryptionKey,proto3" json:"encryption_key,omitempty"`
}

func (x *MessageHistoryReply) Reset() {
	*x = MessageHistoryReply{}
	mi := &file_mls_message_contents_content_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MessageHistoryReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageHistoryReply) ProtoMessage() {}

func (x *MessageHistoryReply) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_content_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageHistoryReply.ProtoReflect.Descriptor instead.
func (*MessageHistoryReply) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_content_proto_rawDescGZIP(), []int{4}
}

func (x *MessageHistoryReply) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

func (x *MessageHistoryReply) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *MessageHistoryReply) GetEncryptionKey() *MessageHistoryKeyType {
	if x != nil {
		return x.EncryptionKey
	}
	return nil
}

// Key used to encrypt the message-bundle
type MessageHistoryKeyType struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Key:
	//
	//	*MessageHistoryKeyType_Chacha20Poly1305
	Key isMessageHistoryKeyType_Key `protobuf_oneof:"key"`
}

func (x *MessageHistoryKeyType) Reset() {
	*x = MessageHistoryKeyType{}
	mi := &file_mls_message_contents_content_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MessageHistoryKeyType) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageHistoryKeyType) ProtoMessage() {}

func (x *MessageHistoryKeyType) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_content_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageHistoryKeyType.ProtoReflect.Descriptor instead.
func (*MessageHistoryKeyType) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_content_proto_rawDescGZIP(), []int{5}
}

func (m *MessageHistoryKeyType) GetKey() isMessageHistoryKeyType_Key {
	if m != nil {
		return m.Key
	}
	return nil
}

func (x *MessageHistoryKeyType) GetChacha20Poly1305() []byte {
	if x, ok := x.GetKey().(*MessageHistoryKeyType_Chacha20Poly1305); ok {
		return x.Chacha20Poly1305
	}
	return nil
}

type isMessageHistoryKeyType_Key interface {
	isMessageHistoryKeyType_Key()
}

type MessageHistoryKeyType_Chacha20Poly1305 struct {
	Chacha20Poly1305 []byte `protobuf:"bytes,1,opt,name=chacha20_poly1305,json=chacha20Poly1305,proto3,oneof"`
}

func (*MessageHistoryKeyType_Chacha20Poly1305) isMessageHistoryKeyType_Key() {}

// Version 1 of the encrypted envelope
type PlaintextEnvelope_V1 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Expected to be EncodedContent
	Content []byte `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
	// A unique value that can be used to ensure that the same content can
	// produce different hashes. May be the sender timestamp.
	IdempotencyKey string `protobuf:"bytes,2,opt,name=idempotency_key,json=idempotencyKey,proto3" json:"idempotency_key,omitempty"`
}

func (x *PlaintextEnvelope_V1) Reset() {
	*x = PlaintextEnvelope_V1{}
	mi := &file_mls_message_contents_content_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PlaintextEnvelope_V1) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PlaintextEnvelope_V1) ProtoMessage() {}

func (x *PlaintextEnvelope_V1) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_content_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PlaintextEnvelope_V1.ProtoReflect.Descriptor instead.
func (*PlaintextEnvelope_V1) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_content_proto_rawDescGZIP(), []int{2, 0}
}

func (x *PlaintextEnvelope_V1) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

func (x *PlaintextEnvelope_V1) GetIdempotencyKey() string {
	if x != nil {
		return x.IdempotencyKey
	}
	return ""
}

// Version 2 of the encrypted envelope
type PlaintextEnvelope_V2 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A unique value that can be used to ensure that the same content can
	// produce different hashes. May be the sender timestamp.
	IdempotencyKey string `protobuf:"bytes,1,opt,name=idempotency_key,json=idempotencyKey,proto3" json:"idempotency_key,omitempty"`
	// Types that are assignable to MessageType:
	//
	//	*PlaintextEnvelope_V2_Content
	//	*PlaintextEnvelope_V2_Request
	//	*PlaintextEnvelope_V2_Reply
	MessageType isPlaintextEnvelope_V2_MessageType `protobuf_oneof:"message_type"`
}

func (x *PlaintextEnvelope_V2) Reset() {
	*x = PlaintextEnvelope_V2{}
	mi := &file_mls_message_contents_content_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PlaintextEnvelope_V2) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PlaintextEnvelope_V2) ProtoMessage() {}

func (x *PlaintextEnvelope_V2) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_content_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PlaintextEnvelope_V2.ProtoReflect.Descriptor instead.
func (*PlaintextEnvelope_V2) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_content_proto_rawDescGZIP(), []int{2, 1}
}

func (x *PlaintextEnvelope_V2) GetIdempotencyKey() string {
	if x != nil {
		return x.IdempotencyKey
	}
	return ""
}

func (m *PlaintextEnvelope_V2) GetMessageType() isPlaintextEnvelope_V2_MessageType {
	if m != nil {
		return m.MessageType
	}
	return nil
}

func (x *PlaintextEnvelope_V2) GetContent() []byte {
	if x, ok := x.GetMessageType().(*PlaintextEnvelope_V2_Content); ok {
		return x.Content
	}
	return nil
}

func (x *PlaintextEnvelope_V2) GetRequest() *MessageHistoryRequest {
	if x, ok := x.GetMessageType().(*PlaintextEnvelope_V2_Request); ok {
		return x.Request
	}
	return nil
}

func (x *PlaintextEnvelope_V2) GetReply() *MessageHistoryReply {
	if x, ok := x.GetMessageType().(*PlaintextEnvelope_V2_Reply); ok {
		return x.Reply
	}
	return nil
}

type isPlaintextEnvelope_V2_MessageType interface {
	isPlaintextEnvelope_V2_MessageType()
}

type PlaintextEnvelope_V2_Content struct {
	// Expected to be EncodedContent
	Content []byte `protobuf:"bytes,2,opt,name=content,proto3,oneof"`
}

type PlaintextEnvelope_V2_Request struct {
	// Initiator sends a request to receive message history
	Request *MessageHistoryRequest `protobuf:"bytes,3,opt,name=request,proto3,oneof"`
}

type PlaintextEnvelope_V2_Reply struct {
	// Some other authorized installation sends a reply
	Reply *MessageHistoryReply `protobuf:"bytes,4,opt,name=reply,proto3,oneof"`
}

func (*PlaintextEnvelope_V2_Content) isPlaintextEnvelope_V2_MessageType() {}

func (*PlaintextEnvelope_V2_Request) isPlaintextEnvelope_V2_MessageType() {}

func (*PlaintextEnvelope_V2_Reply) isPlaintextEnvelope_V2_MessageType() {}

var File_mls_message_contents_content_proto protoreflect.FileDescriptor

var file_mls_message_contents_content_proto_rawDesc = []byte{
	0x0a, 0x22, 0x6d, 0x6c, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x19, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x22,
	0x95, 0x01, 0x0a, 0x0d, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x49,
	0x64, 0x12, 0x21, 0x0a, 0x0c, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x5f, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69,
	0x74, 0x79, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x79, 0x70, 0x65, 0x49, 0x64, 0x12, 0x23, 0x0a,
	0x0d, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x6d, 0x61, 0x6a, 0x6f, 0x72, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x0c, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x4d, 0x61, 0x6a,
	0x6f, 0x72, 0x12, 0x23, 0x0a, 0x0d, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x6d, 0x69,
	0x6e, 0x6f, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0c, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x4d, 0x69, 0x6e, 0x6f, 0x72, 0x22, 0x8f, 0x03, 0x0a, 0x0e, 0x45, 0x6e, 0x63, 0x6f,
	0x64, 0x65, 0x64, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x3c, 0x0a, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e,
	0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65,
	0x49, 0x64, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x59, 0x0a, 0x0a, 0x70, 0x61, 0x72, 0x61,
	0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x39, 0x2e, 0x78,
	0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x65, 0x64,
	0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2e, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65,
	0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0a, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74,
	0x65, 0x72, 0x73, 0x12, 0x1f, 0x0a, 0x08, 0x66, 0x61, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x08, 0x66, 0x61, 0x6c, 0x6c, 0x62, 0x61, 0x63,
	0x6b, 0x88, 0x01, 0x01, 0x12, 0x4d, 0x0a, 0x0b, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x26, 0x2e, 0x78, 0x6d, 0x74, 0x70,
	0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x48, 0x01, 0x52, 0x0b, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x88, 0x01, 0x01, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x1a, 0x3d, 0x0a,
	0x0f, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x0b, 0x0a, 0x09,
	0x5f, 0x66, 0x61, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x42, 0x0e, 0x0a, 0x0c, 0x5f, 0x63, 0x6f,
	0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0xdf, 0x03, 0x0a, 0x11, 0x50, 0x6c,
	0x61, 0x69, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x45, 0x6e, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x12,
	0x41, 0x0a, 0x02, 0x76, 0x31, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x78, 0x6d,
	0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x50, 0x6c, 0x61, 0x69, 0x6e, 0x74, 0x65, 0x78,
	0x74, 0x45, 0x6e, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x2e, 0x56, 0x31, 0x48, 0x00, 0x52, 0x02,
	0x76, 0x31, 0x12, 0x41, 0x0a, 0x02, 0x76, 0x32, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2f,
	0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x50, 0x6c, 0x61, 0x69, 0x6e,
	0x74, 0x65, 0x78, 0x74, 0x45, 0x6e, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x2e, 0x56, 0x32, 0x48,
	0x00, 0x52, 0x02, 0x76, 0x32, 0x1a, 0x47, 0x0a, 0x02, 0x56, 0x31, 0x12, 0x18, 0x0a, 0x07, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x27, 0x0a, 0x0f, 0x69, 0x64, 0x65, 0x6d, 0x70, 0x6f, 0x74,
	0x65, 0x6e, 0x63, 0x79, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e,
	0x69, 0x64, 0x65, 0x6d, 0x70, 0x6f, 0x74, 0x65, 0x6e, 0x63, 0x79, 0x4b, 0x65, 0x79, 0x1a, 0xef,
	0x01, 0x0a, 0x02, 0x56, 0x32, 0x12, 0x27, 0x0a, 0x0f, 0x69, 0x64, 0x65, 0x6d, 0x70, 0x6f, 0x74,
	0x65, 0x6e, 0x63, 0x79, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e,
	0x69, 0x64, 0x65, 0x6d, 0x70, 0x6f, 0x74, 0x65, 0x6e, 0x63, 0x79, 0x4b, 0x65, 0x79, 0x12, 0x1a,
	0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x48,
	0x00, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x4c, 0x0a, 0x07, 0x72, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x30, 0x2e, 0x78, 0x6d,
	0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48,
	0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52,
	0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x46, 0x0a, 0x05, 0x72, 0x65, 0x70, 0x6c,
	0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d,
	0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x73, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x69, 0x73, 0x74, 0x6f,
	0x72, 0x79, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x48, 0x00, 0x52, 0x05, 0x72, 0x65, 0x70, 0x6c, 0x79,
	0x42, 0x0e, 0x0a, 0x0c, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65,
	0x42, 0x09, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0x51, 0x0a, 0x15, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x70, 0x69, 0x6e, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x69, 0x6e, 0x43, 0x6f, 0x64, 0x65, 0x22, 0x9f,
	0x01, 0x0a, 0x13, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72,
	0x79, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x57, 0x0a, 0x0e, 0x65, 0x6e, 0x63, 0x72, 0x79,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x30, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x4b, 0x65, 0x79, 0x54, 0x79, 0x70,
	0x65, 0x52, 0x0d, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x4b, 0x65, 0x79,
	0x22, 0x4d, 0x0a, 0x15, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x69, 0x73, 0x74, 0x6f,
	0x72, 0x79, 0x4b, 0x65, 0x79, 0x54, 0x79, 0x70, 0x65, 0x12, 0x2d, 0x0a, 0x11, 0x63, 0x68, 0x61,
	0x63, 0x68, 0x61, 0x32, 0x30, 0x5f, 0x70, 0x6f, 0x6c, 0x79, 0x31, 0x33, 0x30, 0x35, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x10, 0x63, 0x68, 0x61, 0x63, 0x68, 0x61, 0x32, 0x30,
	0x50, 0x6f, 0x6c, 0x79, 0x31, 0x33, 0x30, 0x35, 0x42, 0x05, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x2a,
	0x3c, 0x0a, 0x0b, 0x43, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x17,
	0x0a, 0x13, 0x43, 0x4f, 0x4d, 0x50, 0x52, 0x45, 0x53, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x44, 0x45,
	0x46, 0x4c, 0x41, 0x54, 0x45, 0x10, 0x00, 0x12, 0x14, 0x0a, 0x10, 0x43, 0x4f, 0x4d, 0x50, 0x52,
	0x45, 0x53, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x47, 0x5a, 0x49, 0x50, 0x10, 0x01, 0x42, 0xec, 0x01,
	0x0a, 0x1d, 0x63, 0x6f, 0x6d, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x42,
	0x0c, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a,
	0x3b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x78, 0x6d, 0x74, 0x70,
	0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2d, 0x6e, 0x6f, 0x64, 0x65, 0x2d, 0x67, 0x6f, 0x2f, 0x70, 0x6b,
	0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x6c, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xa2, 0x02, 0x03, 0x58,
	0x4d, 0x4d, 0xaa, 0x02, 0x18, 0x58, 0x6d, 0x74, 0x70, 0x2e, 0x4d, 0x6c, 0x73, 0x2e, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xca, 0x02, 0x18,
	0x58, 0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x6c, 0x73, 0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xe2, 0x02, 0x24, 0x58, 0x6d, 0x74, 0x70, 0x5c,
	0x4d, 0x6c, 0x73, 0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x73, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea,
	0x02, 0x1a, 0x58, 0x6d, 0x74, 0x70, 0x3a, 0x3a, 0x4d, 0x6c, 0x73, 0x3a, 0x3a, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mls_message_contents_content_proto_rawDescOnce sync.Once
	file_mls_message_contents_content_proto_rawDescData = file_mls_message_contents_content_proto_rawDesc
)

func file_mls_message_contents_content_proto_rawDescGZIP() []byte {
	file_mls_message_contents_content_proto_rawDescOnce.Do(func() {
		file_mls_message_contents_content_proto_rawDescData = protoimpl.X.CompressGZIP(file_mls_message_contents_content_proto_rawDescData)
	})
	return file_mls_message_contents_content_proto_rawDescData
}

var file_mls_message_contents_content_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_mls_message_contents_content_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_mls_message_contents_content_proto_goTypes = []any{
	(Compression)(0),              // 0: xmtp.mls.message_contents.Compression
	(*ContentTypeId)(nil),         // 1: xmtp.mls.message_contents.ContentTypeId
	(*EncodedContent)(nil),        // 2: xmtp.mls.message_contents.EncodedContent
	(*PlaintextEnvelope)(nil),     // 3: xmtp.mls.message_contents.PlaintextEnvelope
	(*MessageHistoryRequest)(nil), // 4: xmtp.mls.message_contents.MessageHistoryRequest
	(*MessageHistoryReply)(nil),   // 5: xmtp.mls.message_contents.MessageHistoryReply
	(*MessageHistoryKeyType)(nil), // 6: xmtp.mls.message_contents.MessageHistoryKeyType
	nil,                           // 7: xmtp.mls.message_contents.EncodedContent.ParametersEntry
	(*PlaintextEnvelope_V1)(nil),  // 8: xmtp.mls.message_contents.PlaintextEnvelope.V1
	(*PlaintextEnvelope_V2)(nil),  // 9: xmtp.mls.message_contents.PlaintextEnvelope.V2
}
var file_mls_message_contents_content_proto_depIdxs = []int32{
	1, // 0: xmtp.mls.message_contents.EncodedContent.type:type_name -> xmtp.mls.message_contents.ContentTypeId
	7, // 1: xmtp.mls.message_contents.EncodedContent.parameters:type_name -> xmtp.mls.message_contents.EncodedContent.ParametersEntry
	0, // 2: xmtp.mls.message_contents.EncodedContent.compression:type_name -> xmtp.mls.message_contents.Compression
	8, // 3: xmtp.mls.message_contents.PlaintextEnvelope.v1:type_name -> xmtp.mls.message_contents.PlaintextEnvelope.V1
	9, // 4: xmtp.mls.message_contents.PlaintextEnvelope.v2:type_name -> xmtp.mls.message_contents.PlaintextEnvelope.V2
	6, // 5: xmtp.mls.message_contents.MessageHistoryReply.encryption_key:type_name -> xmtp.mls.message_contents.MessageHistoryKeyType
	4, // 6: xmtp.mls.message_contents.PlaintextEnvelope.V2.request:type_name -> xmtp.mls.message_contents.MessageHistoryRequest
	5, // 7: xmtp.mls.message_contents.PlaintextEnvelope.V2.reply:type_name -> xmtp.mls.message_contents.MessageHistoryReply
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_mls_message_contents_content_proto_init() }
func file_mls_message_contents_content_proto_init() {
	if File_mls_message_contents_content_proto != nil {
		return
	}
	file_mls_message_contents_content_proto_msgTypes[1].OneofWrappers = []any{}
	file_mls_message_contents_content_proto_msgTypes[2].OneofWrappers = []any{
		(*PlaintextEnvelope_V1_)(nil),
		(*PlaintextEnvelope_V2_)(nil),
	}
	file_mls_message_contents_content_proto_msgTypes[5].OneofWrappers = []any{
		(*MessageHistoryKeyType_Chacha20Poly1305)(nil),
	}
	file_mls_message_contents_content_proto_msgTypes[8].OneofWrappers = []any{
		(*PlaintextEnvelope_V2_Content)(nil),
		(*PlaintextEnvelope_V2_Request)(nil),
		(*PlaintextEnvelope_V2_Reply)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_mls_message_contents_content_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mls_message_contents_content_proto_goTypes,
		DependencyIndexes: file_mls_message_contents_content_proto_depIdxs,
		EnumInfos:         file_mls_message_contents_content_proto_enumTypes,
		MessageInfos:      file_mls_message_contents_content_proto_msgTypes,
	}.Build()
	File_mls_message_contents_content_proto = out.File
	file_mls_message_contents_content_proto_rawDesc = nil
	file_mls_message_contents_content_proto_goTypes = nil
	file_mls_message_contents_content_proto_depIdxs = nil
}
