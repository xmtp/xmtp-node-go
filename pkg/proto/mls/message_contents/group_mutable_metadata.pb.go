// Group mutable metadata

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: mls/message_contents/group_mutable_metadata.proto

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

// Message for group mutable metadata
type GroupMutableMetadataV1 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GroupName string `protobuf:"bytes,1,opt,name=group_name,json=groupName,proto3" json:"group_name,omitempty"`
}

func (x *GroupMutableMetadataV1) Reset() {
	*x = GroupMutableMetadataV1{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_message_contents_group_mutable_metadata_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupMutableMetadataV1) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupMutableMetadataV1) ProtoMessage() {}

func (x *GroupMutableMetadataV1) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_group_mutable_metadata_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupMutableMetadataV1.ProtoReflect.Descriptor instead.
func (*GroupMutableMetadataV1) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_group_mutable_metadata_proto_rawDescGZIP(), []int{0}
}

func (x *GroupMutableMetadataV1) GetGroupName() string {
	if x != nil {
		return x.GroupName
	}
	return ""
}

var File_mls_message_contents_group_mutable_metadata_proto protoreflect.FileDescriptor

var file_mls_message_contents_group_mutable_metadata_proto_rawDesc = []byte{
	0x0a, 0x31, 0x6d, 0x6c, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x6d, 0x75, 0x74,
	0x61, 0x62, 0x6c, 0x65, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x19, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x22, 0x37,
	0x0a, 0x16, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x75, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x4d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x56, 0x31, 0x12, 0x1d, 0x0a, 0x0a, 0x67, 0x72, 0x6f, 0x75,
	0x70, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x67, 0x72,
	0x6f, 0x75, 0x70, 0x4e, 0x61, 0x6d, 0x65, 0x42, 0xf9, 0x01, 0x0a, 0x1d, 0x63, 0x6f, 0x6d, 0x2e,
	0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x42, 0x19, 0x47, 0x72, 0x6f, 0x75, 0x70,
	0x4d, 0x75, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2d, 0x6e, 0x6f, 0x64,
	0x65, 0x2d, 0x67, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d,
	0x6c, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x73, 0xa2, 0x02, 0x03, 0x58, 0x4d, 0x4d, 0xaa, 0x02, 0x18, 0x58, 0x6d, 0x74, 0x70,
	0x2e, 0x4d, 0x6c, 0x73, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x73, 0xca, 0x02, 0x18, 0x58, 0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x6c, 0x73, 0x5c,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xe2,
	0x02, 0x24, 0x58, 0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x6c, 0x73, 0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x1a, 0x58, 0x6d, 0x74, 0x70, 0x3a, 0x3a, 0x4d,
	0x6c, 0x73, 0x3a, 0x3a, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mls_message_contents_group_mutable_metadata_proto_rawDescOnce sync.Once
	file_mls_message_contents_group_mutable_metadata_proto_rawDescData = file_mls_message_contents_group_mutable_metadata_proto_rawDesc
)

func file_mls_message_contents_group_mutable_metadata_proto_rawDescGZIP() []byte {
	file_mls_message_contents_group_mutable_metadata_proto_rawDescOnce.Do(func() {
		file_mls_message_contents_group_mutable_metadata_proto_rawDescData = protoimpl.X.CompressGZIP(file_mls_message_contents_group_mutable_metadata_proto_rawDescData)
	})
	return file_mls_message_contents_group_mutable_metadata_proto_rawDescData
}

var file_mls_message_contents_group_mutable_metadata_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_mls_message_contents_group_mutable_metadata_proto_goTypes = []interface{}{
	(*GroupMutableMetadataV1)(nil), // 0: xmtp.mls.message_contents.GroupMutableMetadataV1
}
var file_mls_message_contents_group_mutable_metadata_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_mls_message_contents_group_mutable_metadata_proto_init() }
func file_mls_message_contents_group_mutable_metadata_proto_init() {
	if File_mls_message_contents_group_mutable_metadata_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mls_message_contents_group_mutable_metadata_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupMutableMetadataV1); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_mls_message_contents_group_mutable_metadata_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mls_message_contents_group_mutable_metadata_proto_goTypes,
		DependencyIndexes: file_mls_message_contents_group_mutable_metadata_proto_depIdxs,
		MessageInfos:      file_mls_message_contents_group_mutable_metadata_proto_msgTypes,
	}.Build()
	File_mls_message_contents_group_mutable_metadata_proto = out.File
	file_mls_message_contents_group_mutable_metadata_proto_rawDesc = nil
	file_mls_message_contents_group_mutable_metadata_proto_goTypes = nil
	file_mls_message_contents_group_mutable_metadata_proto_depIdxs = nil
}
