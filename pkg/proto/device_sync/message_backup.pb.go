// Definitions for backups

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: device_sync/message_backup.proto

package device_sync

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

// Group message kind
type GroupMessageKindSave int32

const (
	GroupMessageKindSave_GROUP_MESSAGE_KIND_SAVE_UNSPECIFIED       GroupMessageKindSave = 0
	GroupMessageKindSave_GROUP_MESSAGE_KIND_SAVE_APPLICATION       GroupMessageKindSave = 1
	GroupMessageKindSave_GROUP_MESSAGE_KIND_SAVE_MEMBERSHIP_CHANGE GroupMessageKindSave = 2
)

// Enum value maps for GroupMessageKindSave.
var (
	GroupMessageKindSave_name = map[int32]string{
		0: "GROUP_MESSAGE_KIND_SAVE_UNSPECIFIED",
		1: "GROUP_MESSAGE_KIND_SAVE_APPLICATION",
		2: "GROUP_MESSAGE_KIND_SAVE_MEMBERSHIP_CHANGE",
	}
	GroupMessageKindSave_value = map[string]int32{
		"GROUP_MESSAGE_KIND_SAVE_UNSPECIFIED":       0,
		"GROUP_MESSAGE_KIND_SAVE_APPLICATION":       1,
		"GROUP_MESSAGE_KIND_SAVE_MEMBERSHIP_CHANGE": 2,
	}
)

func (x GroupMessageKindSave) Enum() *GroupMessageKindSave {
	p := new(GroupMessageKindSave)
	*p = x
	return p
}

func (x GroupMessageKindSave) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (GroupMessageKindSave) Descriptor() protoreflect.EnumDescriptor {
	return file_device_sync_message_backup_proto_enumTypes[0].Descriptor()
}

func (GroupMessageKindSave) Type() protoreflect.EnumType {
	return &file_device_sync_message_backup_proto_enumTypes[0]
}

func (x GroupMessageKindSave) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use GroupMessageKindSave.Descriptor instead.
func (GroupMessageKindSave) EnumDescriptor() ([]byte, []int) {
	return file_device_sync_message_backup_proto_rawDescGZIP(), []int{0}
}

// Group message delivery status
type DeliveryStatusSave int32

const (
	DeliveryStatusSave_DELIVERY_STATUS_SAVE_UNSPECIFIED DeliveryStatusSave = 0
	DeliveryStatusSave_DELIVERY_STATUS_SAVE_UNPUBLISHED DeliveryStatusSave = 1
	DeliveryStatusSave_DELIVERY_STATUS_SAVE_PUBLISHED   DeliveryStatusSave = 2
	DeliveryStatusSave_DELIVERY_STATUS_SAVE_FAILED      DeliveryStatusSave = 3
)

// Enum value maps for DeliveryStatusSave.
var (
	DeliveryStatusSave_name = map[int32]string{
		0: "DELIVERY_STATUS_SAVE_UNSPECIFIED",
		1: "DELIVERY_STATUS_SAVE_UNPUBLISHED",
		2: "DELIVERY_STATUS_SAVE_PUBLISHED",
		3: "DELIVERY_STATUS_SAVE_FAILED",
	}
	DeliveryStatusSave_value = map[string]int32{
		"DELIVERY_STATUS_SAVE_UNSPECIFIED": 0,
		"DELIVERY_STATUS_SAVE_UNPUBLISHED": 1,
		"DELIVERY_STATUS_SAVE_PUBLISHED":   2,
		"DELIVERY_STATUS_SAVE_FAILED":      3,
	}
)

func (x DeliveryStatusSave) Enum() *DeliveryStatusSave {
	p := new(DeliveryStatusSave)
	*p = x
	return p
}

func (x DeliveryStatusSave) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DeliveryStatusSave) Descriptor() protoreflect.EnumDescriptor {
	return file_device_sync_message_backup_proto_enumTypes[1].Descriptor()
}

func (DeliveryStatusSave) Type() protoreflect.EnumType {
	return &file_device_sync_message_backup_proto_enumTypes[1]
}

func (x DeliveryStatusSave) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DeliveryStatusSave.Descriptor instead.
func (DeliveryStatusSave) EnumDescriptor() ([]byte, []int) {
	return file_device_sync_message_backup_proto_rawDescGZIP(), []int{1}
}

// Group message content type
type ContentTypeSave int32

const (
	ContentTypeSave_CONTENT_TYPE_SAVE_UNSPECIFIED             ContentTypeSave = 0
	ContentTypeSave_CONTENT_TYPE_SAVE_UNKNOWN                 ContentTypeSave = 1
	ContentTypeSave_CONTENT_TYPE_SAVE_TEXT                    ContentTypeSave = 2
	ContentTypeSave_CONTENT_TYPE_SAVE_GROUP_MEMBERSHIP_CHANGE ContentTypeSave = 3
	ContentTypeSave_CONTENT_TYPE_SAVE_GROUP_UPDATED           ContentTypeSave = 4
	ContentTypeSave_CONTENT_TYPE_SAVE_REACTION                ContentTypeSave = 5
	ContentTypeSave_CONTENT_TYPE_SAVE_READ_RECEIPT            ContentTypeSave = 6
	ContentTypeSave_CONTENT_TYPE_SAVE_REPLY                   ContentTypeSave = 7
	ContentTypeSave_CONTENT_TYPE_SAVE_ATTACHMENT              ContentTypeSave = 8
	ContentTypeSave_CONTENT_TYPE_SAVE_REMOTE_ATTACHMENT       ContentTypeSave = 9
	ContentTypeSave_CONTENT_TYPE_SAVE_TRANSACTION_REFERENCE   ContentTypeSave = 10
)

// Enum value maps for ContentTypeSave.
var (
	ContentTypeSave_name = map[int32]string{
		0:  "CONTENT_TYPE_SAVE_UNSPECIFIED",
		1:  "CONTENT_TYPE_SAVE_UNKNOWN",
		2:  "CONTENT_TYPE_SAVE_TEXT",
		3:  "CONTENT_TYPE_SAVE_GROUP_MEMBERSHIP_CHANGE",
		4:  "CONTENT_TYPE_SAVE_GROUP_UPDATED",
		5:  "CONTENT_TYPE_SAVE_REACTION",
		6:  "CONTENT_TYPE_SAVE_READ_RECEIPT",
		7:  "CONTENT_TYPE_SAVE_REPLY",
		8:  "CONTENT_TYPE_SAVE_ATTACHMENT",
		9:  "CONTENT_TYPE_SAVE_REMOTE_ATTACHMENT",
		10: "CONTENT_TYPE_SAVE_TRANSACTION_REFERENCE",
	}
	ContentTypeSave_value = map[string]int32{
		"CONTENT_TYPE_SAVE_UNSPECIFIED":             0,
		"CONTENT_TYPE_SAVE_UNKNOWN":                 1,
		"CONTENT_TYPE_SAVE_TEXT":                    2,
		"CONTENT_TYPE_SAVE_GROUP_MEMBERSHIP_CHANGE": 3,
		"CONTENT_TYPE_SAVE_GROUP_UPDATED":           4,
		"CONTENT_TYPE_SAVE_REACTION":                5,
		"CONTENT_TYPE_SAVE_READ_RECEIPT":            6,
		"CONTENT_TYPE_SAVE_REPLY":                   7,
		"CONTENT_TYPE_SAVE_ATTACHMENT":              8,
		"CONTENT_TYPE_SAVE_REMOTE_ATTACHMENT":       9,
		"CONTENT_TYPE_SAVE_TRANSACTION_REFERENCE":   10,
	}
)

func (x ContentTypeSave) Enum() *ContentTypeSave {
	p := new(ContentTypeSave)
	*p = x
	return p
}

func (x ContentTypeSave) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ContentTypeSave) Descriptor() protoreflect.EnumDescriptor {
	return file_device_sync_message_backup_proto_enumTypes[2].Descriptor()
}

func (ContentTypeSave) Type() protoreflect.EnumType {
	return &file_device_sync_message_backup_proto_enumTypes[2]
}

func (x ContentTypeSave) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ContentTypeSave.Descriptor instead.
func (ContentTypeSave) EnumDescriptor() ([]byte, []int) {
	return file_device_sync_message_backup_proto_rawDescGZIP(), []int{2}
}

// Proto representation of a stored group message
type GroupMessageSave struct {
	state                 protoimpl.MessageState `protogen:"open.v1"`
	Id                    []byte                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	GroupId               []byte                 `protobuf:"bytes,2,opt,name=group_id,json=groupId,proto3" json:"group_id,omitempty"`
	DecryptedMessageBytes []byte                 `protobuf:"bytes,3,opt,name=decrypted_message_bytes,json=decryptedMessageBytes,proto3" json:"decrypted_message_bytes,omitempty"`
	SentAtNs              int64                  `protobuf:"varint,4,opt,name=sent_at_ns,json=sentAtNs,proto3" json:"sent_at_ns,omitempty"`
	Kind                  GroupMessageKindSave   `protobuf:"varint,5,opt,name=kind,proto3,enum=xmtp.device_sync.message_backup.GroupMessageKindSave" json:"kind,omitempty"`
	SenderInstallationId  []byte                 `protobuf:"bytes,6,opt,name=sender_installation_id,json=senderInstallationId,proto3" json:"sender_installation_id,omitempty"`
	SenderInboxId         string                 `protobuf:"bytes,7,opt,name=sender_inbox_id,json=senderInboxId,proto3" json:"sender_inbox_id,omitempty"`
	DeliveryStatus        DeliveryStatusSave     `protobuf:"varint,8,opt,name=delivery_status,json=deliveryStatus,proto3,enum=xmtp.device_sync.message_backup.DeliveryStatusSave" json:"delivery_status,omitempty"`
	ContentType           ContentTypeSave        `protobuf:"varint,9,opt,name=content_type,json=contentType,proto3,enum=xmtp.device_sync.message_backup.ContentTypeSave" json:"content_type,omitempty"`
	VersionMajor          int32                  `protobuf:"varint,10,opt,name=version_major,json=versionMajor,proto3" json:"version_major,omitempty"`
	VersionMinor          int32                  `protobuf:"varint,11,opt,name=version_minor,json=versionMinor,proto3" json:"version_minor,omitempty"`
	AuthorityId           string                 `protobuf:"bytes,12,opt,name=authority_id,json=authorityId,proto3" json:"authority_id,omitempty"`
	ReferenceId           []byte                 `protobuf:"bytes,13,opt,name=reference_id,json=referenceId,proto3,oneof" json:"reference_id,omitempty"`
	SequenceId            *int64                 `protobuf:"varint,14,opt,name=sequence_id,json=sequenceId,proto3,oneof" json:"sequence_id,omitempty"`
	OriginatorId          *int64                 `protobuf:"varint,15,opt,name=originator_id,json=originatorId,proto3,oneof" json:"originator_id,omitempty"`
	unknownFields         protoimpl.UnknownFields
	sizeCache             protoimpl.SizeCache
}

func (x *GroupMessageSave) Reset() {
	*x = GroupMessageSave{}
	mi := &file_device_sync_message_backup_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GroupMessageSave) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupMessageSave) ProtoMessage() {}

func (x *GroupMessageSave) ProtoReflect() protoreflect.Message {
	mi := &file_device_sync_message_backup_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupMessageSave.ProtoReflect.Descriptor instead.
func (*GroupMessageSave) Descriptor() ([]byte, []int) {
	return file_device_sync_message_backup_proto_rawDescGZIP(), []int{0}
}

func (x *GroupMessageSave) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *GroupMessageSave) GetGroupId() []byte {
	if x != nil {
		return x.GroupId
	}
	return nil
}

func (x *GroupMessageSave) GetDecryptedMessageBytes() []byte {
	if x != nil {
		return x.DecryptedMessageBytes
	}
	return nil
}

func (x *GroupMessageSave) GetSentAtNs() int64 {
	if x != nil {
		return x.SentAtNs
	}
	return 0
}

func (x *GroupMessageSave) GetKind() GroupMessageKindSave {
	if x != nil {
		return x.Kind
	}
	return GroupMessageKindSave_GROUP_MESSAGE_KIND_SAVE_UNSPECIFIED
}

func (x *GroupMessageSave) GetSenderInstallationId() []byte {
	if x != nil {
		return x.SenderInstallationId
	}
	return nil
}

func (x *GroupMessageSave) GetSenderInboxId() string {
	if x != nil {
		return x.SenderInboxId
	}
	return ""
}

func (x *GroupMessageSave) GetDeliveryStatus() DeliveryStatusSave {
	if x != nil {
		return x.DeliveryStatus
	}
	return DeliveryStatusSave_DELIVERY_STATUS_SAVE_UNSPECIFIED
}

func (x *GroupMessageSave) GetContentType() ContentTypeSave {
	if x != nil {
		return x.ContentType
	}
	return ContentTypeSave_CONTENT_TYPE_SAVE_UNSPECIFIED
}

func (x *GroupMessageSave) GetVersionMajor() int32 {
	if x != nil {
		return x.VersionMajor
	}
	return 0
}

func (x *GroupMessageSave) GetVersionMinor() int32 {
	if x != nil {
		return x.VersionMinor
	}
	return 0
}

func (x *GroupMessageSave) GetAuthorityId() string {
	if x != nil {
		return x.AuthorityId
	}
	return ""
}

func (x *GroupMessageSave) GetReferenceId() []byte {
	if x != nil {
		return x.ReferenceId
	}
	return nil
}

func (x *GroupMessageSave) GetSequenceId() int64 {
	if x != nil && x.SequenceId != nil {
		return *x.SequenceId
	}
	return 0
}

func (x *GroupMessageSave) GetOriginatorId() int64 {
	if x != nil && x.OriginatorId != nil {
		return *x.OriginatorId
	}
	return 0
}

var File_device_sync_message_backup_proto protoreflect.FileDescriptor

const file_device_sync_message_backup_proto_rawDesc = "" +
	"\n" +
	" device_sync/message_backup.proto\x12\x1fxmtp.device_sync.message_backup\"\x87\x06\n" +
	"\x10GroupMessageSave\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\fR\x02id\x12\x19\n" +
	"\bgroup_id\x18\x02 \x01(\fR\agroupId\x126\n" +
	"\x17decrypted_message_bytes\x18\x03 \x01(\fR\x15decryptedMessageBytes\x12\x1c\n" +
	"\n" +
	"sent_at_ns\x18\x04 \x01(\x03R\bsentAtNs\x12I\n" +
	"\x04kind\x18\x05 \x01(\x0e25.xmtp.device_sync.message_backup.GroupMessageKindSaveR\x04kind\x124\n" +
	"\x16sender_installation_id\x18\x06 \x01(\fR\x14senderInstallationId\x12&\n" +
	"\x0fsender_inbox_id\x18\a \x01(\tR\rsenderInboxId\x12\\\n" +
	"\x0fdelivery_status\x18\b \x01(\x0e23.xmtp.device_sync.message_backup.DeliveryStatusSaveR\x0edeliveryStatus\x12S\n" +
	"\fcontent_type\x18\t \x01(\x0e20.xmtp.device_sync.message_backup.ContentTypeSaveR\vcontentType\x12#\n" +
	"\rversion_major\x18\n" +
	" \x01(\x05R\fversionMajor\x12#\n" +
	"\rversion_minor\x18\v \x01(\x05R\fversionMinor\x12!\n" +
	"\fauthority_id\x18\f \x01(\tR\vauthorityId\x12&\n" +
	"\freference_id\x18\r \x01(\fH\x00R\vreferenceId\x88\x01\x01\x12$\n" +
	"\vsequence_id\x18\x0e \x01(\x03H\x01R\n" +
	"sequenceId\x88\x01\x01\x12(\n" +
	"\roriginator_id\x18\x0f \x01(\x03H\x02R\foriginatorId\x88\x01\x01B\x0f\n" +
	"\r_reference_idB\x0e\n" +
	"\f_sequence_idB\x10\n" +
	"\x0e_originator_id*\x97\x01\n" +
	"\x14GroupMessageKindSave\x12'\n" +
	"#GROUP_MESSAGE_KIND_SAVE_UNSPECIFIED\x10\x00\x12'\n" +
	"#GROUP_MESSAGE_KIND_SAVE_APPLICATION\x10\x01\x12-\n" +
	")GROUP_MESSAGE_KIND_SAVE_MEMBERSHIP_CHANGE\x10\x02*\xa5\x01\n" +
	"\x12DeliveryStatusSave\x12$\n" +
	" DELIVERY_STATUS_SAVE_UNSPECIFIED\x10\x00\x12$\n" +
	" DELIVERY_STATUS_SAVE_UNPUBLISHED\x10\x01\x12\"\n" +
	"\x1eDELIVERY_STATUS_SAVE_PUBLISHED\x10\x02\x12\x1f\n" +
	"\x1bDELIVERY_STATUS_SAVE_FAILED\x10\x03*\x9c\x03\n" +
	"\x0fContentTypeSave\x12!\n" +
	"\x1dCONTENT_TYPE_SAVE_UNSPECIFIED\x10\x00\x12\x1d\n" +
	"\x19CONTENT_TYPE_SAVE_UNKNOWN\x10\x01\x12\x1a\n" +
	"\x16CONTENT_TYPE_SAVE_TEXT\x10\x02\x12-\n" +
	")CONTENT_TYPE_SAVE_GROUP_MEMBERSHIP_CHANGE\x10\x03\x12#\n" +
	"\x1fCONTENT_TYPE_SAVE_GROUP_UPDATED\x10\x04\x12\x1e\n" +
	"\x1aCONTENT_TYPE_SAVE_REACTION\x10\x05\x12\"\n" +
	"\x1eCONTENT_TYPE_SAVE_READ_RECEIPT\x10\x06\x12\x1b\n" +
	"\x17CONTENT_TYPE_SAVE_REPLY\x10\a\x12 \n" +
	"\x1cCONTENT_TYPE_SAVE_ATTACHMENT\x10\b\x12'\n" +
	"#CONTENT_TYPE_SAVE_REMOTE_ATTACHMENT\x10\t\x12+\n" +
	"'CONTENT_TYPE_SAVE_TRANSACTION_REFERENCE\x10\n" +
	"B\x83\x02\n" +
	"#com.xmtp.device_sync.message_backupB\x12MessageBackupProtoP\x01Z2github.com/xmtp/xmtp-node-go/pkg/proto/device_sync\xa2\x02\x03XDM\xaa\x02\x1dXmtp.DeviceSync.MessageBackup\xca\x02\x1dXmtp\\DeviceSync\\MessageBackup\xe2\x02)Xmtp\\DeviceSync\\MessageBackup\\GPBMetadata\xea\x02\x1fXmtp::DeviceSync::MessageBackupb\x06proto3"

var (
	file_device_sync_message_backup_proto_rawDescOnce sync.Once
	file_device_sync_message_backup_proto_rawDescData []byte
)

func file_device_sync_message_backup_proto_rawDescGZIP() []byte {
	file_device_sync_message_backup_proto_rawDescOnce.Do(func() {
		file_device_sync_message_backup_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_device_sync_message_backup_proto_rawDesc), len(file_device_sync_message_backup_proto_rawDesc)))
	})
	return file_device_sync_message_backup_proto_rawDescData
}

var file_device_sync_message_backup_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_device_sync_message_backup_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_device_sync_message_backup_proto_goTypes = []any{
	(GroupMessageKindSave)(0), // 0: xmtp.device_sync.message_backup.GroupMessageKindSave
	(DeliveryStatusSave)(0),   // 1: xmtp.device_sync.message_backup.DeliveryStatusSave
	(ContentTypeSave)(0),      // 2: xmtp.device_sync.message_backup.ContentTypeSave
	(*GroupMessageSave)(nil),  // 3: xmtp.device_sync.message_backup.GroupMessageSave
}
var file_device_sync_message_backup_proto_depIdxs = []int32{
	0, // 0: xmtp.device_sync.message_backup.GroupMessageSave.kind:type_name -> xmtp.device_sync.message_backup.GroupMessageKindSave
	1, // 1: xmtp.device_sync.message_backup.GroupMessageSave.delivery_status:type_name -> xmtp.device_sync.message_backup.DeliveryStatusSave
	2, // 2: xmtp.device_sync.message_backup.GroupMessageSave.content_type:type_name -> xmtp.device_sync.message_backup.ContentTypeSave
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_device_sync_message_backup_proto_init() }
func file_device_sync_message_backup_proto_init() {
	if File_device_sync_message_backup_proto != nil {
		return
	}
	file_device_sync_message_backup_proto_msgTypes[0].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_device_sync_message_backup_proto_rawDesc), len(file_device_sync_message_backup_proto_rawDesc)),
			NumEnums:      3,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_device_sync_message_backup_proto_goTypes,
		DependencyIndexes: file_device_sync_message_backup_proto_depIdxs,
		EnumInfos:         file_device_sync_message_backup_proto_enumTypes,
		MessageInfos:      file_device_sync_message_backup_proto_msgTypes,
	}.Build()
	File_device_sync_message_backup_proto = out.File
	file_device_sync_message_backup_proto_goTypes = nil
	file_device_sync_message_backup_proto_depIdxs = nil
}
