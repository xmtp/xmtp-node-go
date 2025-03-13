// reaction.proto
// This file defines the ReactionV2 message type and is associated with the following ContentTypeId:
//
// ContentTypeId {
//     authority_id: "xmtp.org",
//     type_id:      "reaction",
//     version_major: 2,
//     version_minor: 0,
// }
//

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        (unknown)
// source: mls/message_contents/content_types/reaction.proto

package content_types

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Action enum to represent reaction states
type ReactionAction int32

const (
	ReactionAction_REACTION_ACTION_UNSPECIFIED ReactionAction = 0
	ReactionAction_REACTION_ACTION_ADDED       ReactionAction = 1
	ReactionAction_REACTION_ACTION_REMOVED     ReactionAction = 2
)

// Enum value maps for ReactionAction.
var (
	ReactionAction_name = map[int32]string{
		0: "REACTION_ACTION_UNSPECIFIED",
		1: "REACTION_ACTION_ADDED",
		2: "REACTION_ACTION_REMOVED",
	}
	ReactionAction_value = map[string]int32{
		"REACTION_ACTION_UNSPECIFIED": 0,
		"REACTION_ACTION_ADDED":       1,
		"REACTION_ACTION_REMOVED":     2,
	}
)

func (x ReactionAction) Enum() *ReactionAction {
	p := new(ReactionAction)
	*p = x
	return p
}

func (x ReactionAction) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ReactionAction) Descriptor() protoreflect.EnumDescriptor {
	return file_mls_message_contents_content_types_reaction_proto_enumTypes[0].Descriptor()
}

func (ReactionAction) Type() protoreflect.EnumType {
	return &file_mls_message_contents_content_types_reaction_proto_enumTypes[0]
}

func (x ReactionAction) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ReactionAction.Descriptor instead.
func (ReactionAction) EnumDescriptor() ([]byte, []int) {
	return file_mls_message_contents_content_types_reaction_proto_rawDescGZIP(), []int{0}
}

// Schema enum to represent reaction content types
type ReactionSchema int32

const (
	ReactionSchema_REACTION_SCHEMA_UNSPECIFIED ReactionSchema = 0
	ReactionSchema_REACTION_SCHEMA_UNICODE     ReactionSchema = 1
	ReactionSchema_REACTION_SCHEMA_SHORTCODE   ReactionSchema = 2
	ReactionSchema_REACTION_SCHEMA_CUSTOM      ReactionSchema = 3
)

// Enum value maps for ReactionSchema.
var (
	ReactionSchema_name = map[int32]string{
		0: "REACTION_SCHEMA_UNSPECIFIED",
		1: "REACTION_SCHEMA_UNICODE",
		2: "REACTION_SCHEMA_SHORTCODE",
		3: "REACTION_SCHEMA_CUSTOM",
	}
	ReactionSchema_value = map[string]int32{
		"REACTION_SCHEMA_UNSPECIFIED": 0,
		"REACTION_SCHEMA_UNICODE":     1,
		"REACTION_SCHEMA_SHORTCODE":   2,
		"REACTION_SCHEMA_CUSTOM":      3,
	}
)

func (x ReactionSchema) Enum() *ReactionSchema {
	p := new(ReactionSchema)
	*p = x
	return p
}

func (x ReactionSchema) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ReactionSchema) Descriptor() protoreflect.EnumDescriptor {
	return file_mls_message_contents_content_types_reaction_proto_enumTypes[1].Descriptor()
}

func (ReactionSchema) Type() protoreflect.EnumType {
	return &file_mls_message_contents_content_types_reaction_proto_enumTypes[1]
}

func (x ReactionSchema) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ReactionSchema.Descriptor instead.
func (ReactionSchema) EnumDescriptor() ([]byte, []int) {
	return file_mls_message_contents_content_types_reaction_proto_rawDescGZIP(), []int{1}
}

// Reaction message type
type ReactionV2 struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The message ID being reacted to
	Reference string `protobuf:"bytes,1,opt,name=reference,proto3" json:"reference,omitempty"`
	// The inbox ID of the user who sent the message being reacted to
	// Optional for group messages
	ReferenceInboxId string `protobuf:"bytes,2,opt,name=reference_inbox_id,json=referenceInboxId,proto3" json:"reference_inbox_id,omitempty"`
	// The action of the reaction (added or removed)
	Action ReactionAction `protobuf:"varint,3,opt,name=action,proto3,enum=xmtp.mls.message_contents.content_types.ReactionAction" json:"action,omitempty"`
	// The content of the reaction
	Content string `protobuf:"bytes,4,opt,name=content,proto3" json:"content,omitempty"`
	// The schema of the reaction content
	Schema        ReactionSchema `protobuf:"varint,5,opt,name=schema,proto3,enum=xmtp.mls.message_contents.content_types.ReactionSchema" json:"schema,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ReactionV2) Reset() {
	*x = ReactionV2{}
	mi := &file_mls_message_contents_content_types_reaction_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ReactionV2) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReactionV2) ProtoMessage() {}

func (x *ReactionV2) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_content_types_reaction_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReactionV2.ProtoReflect.Descriptor instead.
func (*ReactionV2) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_content_types_reaction_proto_rawDescGZIP(), []int{0}
}

func (x *ReactionV2) GetReference() string {
	if x != nil {
		return x.Reference
	}
	return ""
}

func (x *ReactionV2) GetReferenceInboxId() string {
	if x != nil {
		return x.ReferenceInboxId
	}
	return ""
}

func (x *ReactionV2) GetAction() ReactionAction {
	if x != nil {
		return x.Action
	}
	return ReactionAction_REACTION_ACTION_UNSPECIFIED
}

func (x *ReactionV2) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *ReactionV2) GetSchema() ReactionSchema {
	if x != nil {
		return x.Schema
	}
	return ReactionSchema_REACTION_SCHEMA_UNSPECIFIED
}

var File_mls_message_contents_content_types_reaction_proto protoreflect.FileDescriptor

var file_mls_message_contents_content_types_reaction_proto_rawDesc = string([]byte{
	0x0a, 0x31, 0x6d, 0x6c, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x2f, 0x72, 0x65, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x27, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x22, 0x94, 0x02, 0x0a,
	0x0a, 0x52, 0x65, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x56, 0x32, 0x12, 0x1c, 0x0a, 0x09, 0x72,
	0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x12, 0x2c, 0x0a, 0x12, 0x72, 0x65, 0x66,
	0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x5f, 0x69, 0x6e, 0x62, 0x6f, 0x78, 0x5f, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65,
	0x49, 0x6e, 0x62, 0x6f, 0x78, 0x49, 0x64, 0x12, 0x4f, 0x0a, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x37, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d,
	0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x73, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65,
	0x73, 0x2e, 0x52, 0x65, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x12, 0x4f, 0x0a, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x37, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x52, 0x65, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x06, 0x73, 0x63, 0x68,
	0x65, 0x6d, 0x61, 0x2a, 0x69, 0x0a, 0x0e, 0x52, 0x65, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x41,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1f, 0x0a, 0x1b, 0x52, 0x45, 0x41, 0x43, 0x54, 0x49, 0x4f,
	0x4e, 0x5f, 0x41, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49,
	0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x19, 0x0a, 0x15, 0x52, 0x45, 0x41, 0x43, 0x54, 0x49,
	0x4f, 0x4e, 0x5f, 0x41, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x41, 0x44, 0x44, 0x45, 0x44, 0x10,
	0x01, 0x12, 0x1b, 0x0a, 0x17, 0x52, 0x45, 0x41, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x41, 0x43,
	0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x52, 0x45, 0x4d, 0x4f, 0x56, 0x45, 0x44, 0x10, 0x02, 0x2a, 0x89,
	0x01, 0x0a, 0x0e, 0x52, 0x65, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x63, 0x68, 0x65, 0x6d,
	0x61, 0x12, 0x1f, 0x0a, 0x1b, 0x52, 0x45, 0x41, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x53, 0x43,
	0x48, 0x45, 0x4d, 0x41, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44,
	0x10, 0x00, 0x12, 0x1b, 0x0a, 0x17, 0x52, 0x45, 0x41, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x53,
	0x43, 0x48, 0x45, 0x4d, 0x41, 0x5f, 0x55, 0x4e, 0x49, 0x43, 0x4f, 0x44, 0x45, 0x10, 0x01, 0x12,
	0x1d, 0x0a, 0x19, 0x52, 0x45, 0x41, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x53, 0x43, 0x48, 0x45,
	0x4d, 0x41, 0x5f, 0x53, 0x48, 0x4f, 0x52, 0x54, 0x43, 0x4f, 0x44, 0x45, 0x10, 0x02, 0x12, 0x1a,
	0x0a, 0x16, 0x52, 0x45, 0x41, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x53, 0x43, 0x48, 0x45, 0x4d,
	0x41, 0x5f, 0x43, 0x55, 0x53, 0x54, 0x4f, 0x4d, 0x10, 0x03, 0x42, 0xbf, 0x02, 0x0a, 0x2b, 0x63,
	0x6f, 0x6d, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x42, 0x0d, 0x52, 0x65, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x49, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2f, 0x78, 0x6d, 0x74,
	0x70, 0x2d, 0x6e, 0x6f, 0x64, 0x65, 0x2d, 0x67, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x6c, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x73, 0xa2, 0x02, 0x04, 0x58, 0x4d, 0x4d, 0x43, 0xaa, 0x02, 0x25,
	0x58, 0x6d, 0x74, 0x70, 0x2e, 0x4d, 0x6c, 0x73, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x54, 0x79, 0x70, 0x65, 0x73, 0xca, 0x02, 0x25, 0x58, 0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x6c, 0x73,
	0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73,
	0x5c, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x73, 0xe2, 0x02, 0x31,
	0x58, 0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x6c, 0x73, 0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x5c, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x54, 0x79, 0x70, 0x65, 0x73, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0xea, 0x02, 0x28, 0x58, 0x6d, 0x74, 0x70, 0x3a, 0x3a, 0x4d, 0x6c, 0x73, 0x3a, 0x3a, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x3a, 0x3a,
	0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x73, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_mls_message_contents_content_types_reaction_proto_rawDescOnce sync.Once
	file_mls_message_contents_content_types_reaction_proto_rawDescData []byte
)

func file_mls_message_contents_content_types_reaction_proto_rawDescGZIP() []byte {
	file_mls_message_contents_content_types_reaction_proto_rawDescOnce.Do(func() {
		file_mls_message_contents_content_types_reaction_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_mls_message_contents_content_types_reaction_proto_rawDesc), len(file_mls_message_contents_content_types_reaction_proto_rawDesc)))
	})
	return file_mls_message_contents_content_types_reaction_proto_rawDescData
}

var file_mls_message_contents_content_types_reaction_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_mls_message_contents_content_types_reaction_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_mls_message_contents_content_types_reaction_proto_goTypes = []any{
	(ReactionAction)(0), // 0: xmtp.mls.message_contents.content_types.ReactionAction
	(ReactionSchema)(0), // 1: xmtp.mls.message_contents.content_types.ReactionSchema
	(*ReactionV2)(nil),  // 2: xmtp.mls.message_contents.content_types.ReactionV2
}
var file_mls_message_contents_content_types_reaction_proto_depIdxs = []int32{
	0, // 0: xmtp.mls.message_contents.content_types.ReactionV2.action:type_name -> xmtp.mls.message_contents.content_types.ReactionAction
	1, // 1: xmtp.mls.message_contents.content_types.ReactionV2.schema:type_name -> xmtp.mls.message_contents.content_types.ReactionSchema
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_mls_message_contents_content_types_reaction_proto_init() }
func file_mls_message_contents_content_types_reaction_proto_init() {
	if File_mls_message_contents_content_types_reaction_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_mls_message_contents_content_types_reaction_proto_rawDesc), len(file_mls_message_contents_content_types_reaction_proto_rawDesc)),
			NumEnums:      2,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mls_message_contents_content_types_reaction_proto_goTypes,
		DependencyIndexes: file_mls_message_contents_content_types_reaction_proto_depIdxs,
		EnumInfos:         file_mls_message_contents_content_types_reaction_proto_enumTypes,
		MessageInfos:      file_mls_message_contents_content_types_reaction_proto_msgTypes,
	}.Build()
	File_mls_message_contents_content_types_reaction_proto = out.File
	file_mls_message_contents_content_types_reaction_proto_goTypes = nil
	file_mls_message_contents_content_types_reaction_proto_depIdxs = nil
}
