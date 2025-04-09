// Definitions for backups

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: device_sync/device_sync.proto

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

// Elements selected for backup
type BackupElementSelection int32

const (
	BackupElementSelection_BACKUP_ELEMENT_SELECTION_UNSPECIFIED BackupElementSelection = 0
	BackupElementSelection_BACKUP_ELEMENT_SELECTION_MESSAGES    BackupElementSelection = 1
	BackupElementSelection_BACKUP_ELEMENT_SELECTION_CONSENT     BackupElementSelection = 2
)

// Enum value maps for BackupElementSelection.
var (
	BackupElementSelection_name = map[int32]string{
		0: "BACKUP_ELEMENT_SELECTION_UNSPECIFIED",
		1: "BACKUP_ELEMENT_SELECTION_MESSAGES",
		2: "BACKUP_ELEMENT_SELECTION_CONSENT",
	}
	BackupElementSelection_value = map[string]int32{
		"BACKUP_ELEMENT_SELECTION_UNSPECIFIED": 0,
		"BACKUP_ELEMENT_SELECTION_MESSAGES":    1,
		"BACKUP_ELEMENT_SELECTION_CONSENT":     2,
	}
)

func (x BackupElementSelection) Enum() *BackupElementSelection {
	p := new(BackupElementSelection)
	*p = x
	return p
}

func (x BackupElementSelection) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (BackupElementSelection) Descriptor() protoreflect.EnumDescriptor {
	return file_device_sync_device_sync_proto_enumTypes[0].Descriptor()
}

func (BackupElementSelection) Type() protoreflect.EnumType {
	return &file_device_sync_device_sync_proto_enumTypes[0]
}

func (x BackupElementSelection) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use BackupElementSelection.Descriptor instead.
func (BackupElementSelection) EnumDescriptor() ([]byte, []int) {
	return file_device_sync_device_sync_proto_rawDescGZIP(), []int{0}
}

// Union type representing everything that can be serialied and saved in a backup archive.
type BackupElement struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Element:
	//
	//	*BackupElement_Metadata
	//	*BackupElement_Group
	//	*BackupElement_GroupMessage
	//	*BackupElement_Consent
	Element       isBackupElement_Element `protobuf_oneof:"element"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *BackupElement) Reset() {
	*x = BackupElement{}
	mi := &file_device_sync_device_sync_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BackupElement) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BackupElement) ProtoMessage() {}

func (x *BackupElement) ProtoReflect() protoreflect.Message {
	mi := &file_device_sync_device_sync_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BackupElement.ProtoReflect.Descriptor instead.
func (*BackupElement) Descriptor() ([]byte, []int) {
	return file_device_sync_device_sync_proto_rawDescGZIP(), []int{0}
}

func (x *BackupElement) GetElement() isBackupElement_Element {
	if x != nil {
		return x.Element
	}
	return nil
}

func (x *BackupElement) GetMetadata() *BackupMetadataSave {
	if x != nil {
		if x, ok := x.Element.(*BackupElement_Metadata); ok {
			return x.Metadata
		}
	}
	return nil
}

func (x *BackupElement) GetGroup() *GroupSave {
	if x != nil {
		if x, ok := x.Element.(*BackupElement_Group); ok {
			return x.Group
		}
	}
	return nil
}

func (x *BackupElement) GetGroupMessage() *GroupMessageSave {
	if x != nil {
		if x, ok := x.Element.(*BackupElement_GroupMessage); ok {
			return x.GroupMessage
		}
	}
	return nil
}

func (x *BackupElement) GetConsent() *ConsentSave {
	if x != nil {
		if x, ok := x.Element.(*BackupElement_Consent); ok {
			return x.Consent
		}
	}
	return nil
}

type isBackupElement_Element interface {
	isBackupElement_Element()
}

type BackupElement_Metadata struct {
	Metadata *BackupMetadataSave `protobuf:"bytes,1,opt,name=metadata,proto3,oneof"`
}

type BackupElement_Group struct {
	Group *GroupSave `protobuf:"bytes,2,opt,name=group,proto3,oneof"`
}

type BackupElement_GroupMessage struct {
	GroupMessage *GroupMessageSave `protobuf:"bytes,3,opt,name=group_message,json=groupMessage,proto3,oneof"`
}

type BackupElement_Consent struct {
	Consent *ConsentSave `protobuf:"bytes,4,opt,name=consent,proto3,oneof"`
}

func (*BackupElement_Metadata) isBackupElement_Element() {}

func (*BackupElement_Group) isBackupElement_Element() {}

func (*BackupElement_GroupMessage) isBackupElement_Element() {}

func (*BackupElement_Consent) isBackupElement_Element() {}

// Proto representation of backup metadata
// (Backup version is explicitly missing - it's stored as a header.)
type BackupMetadataSave struct {
	state         protoimpl.MessageState   `protogen:"open.v1"`
	Elements      []BackupElementSelection `protobuf:"varint,2,rep,packed,name=elements,proto3,enum=xmtp.device_sync.BackupElementSelection" json:"elements,omitempty"`
	ExportedAtNs  int64                    `protobuf:"varint,3,opt,name=exported_at_ns,json=exportedAtNs,proto3" json:"exported_at_ns,omitempty"`
	StartNs       *int64                   `protobuf:"varint,4,opt,name=start_ns,json=startNs,proto3,oneof" json:"start_ns,omitempty"`
	EndNs         *int64                   `protobuf:"varint,5,opt,name=end_ns,json=endNs,proto3,oneof" json:"end_ns,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *BackupMetadataSave) Reset() {
	*x = BackupMetadataSave{}
	mi := &file_device_sync_device_sync_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BackupMetadataSave) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BackupMetadataSave) ProtoMessage() {}

func (x *BackupMetadataSave) ProtoReflect() protoreflect.Message {
	mi := &file_device_sync_device_sync_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BackupMetadataSave.ProtoReflect.Descriptor instead.
func (*BackupMetadataSave) Descriptor() ([]byte, []int) {
	return file_device_sync_device_sync_proto_rawDescGZIP(), []int{1}
}

func (x *BackupMetadataSave) GetElements() []BackupElementSelection {
	if x != nil {
		return x.Elements
	}
	return nil
}

func (x *BackupMetadataSave) GetExportedAtNs() int64 {
	if x != nil {
		return x.ExportedAtNs
	}
	return 0
}

func (x *BackupMetadataSave) GetStartNs() int64 {
	if x != nil && x.StartNs != nil {
		return *x.StartNs
	}
	return 0
}

func (x *BackupMetadataSave) GetEndNs() int64 {
	if x != nil && x.EndNs != nil {
		return *x.EndNs
	}
	return 0
}

var File_device_sync_device_sync_proto protoreflect.FileDescriptor

const file_device_sync_device_sync_proto_rawDesc = "" +
	"\n" +
	"\x1ddevice_sync/device_sync.proto\x12\x10xmtp.device_sync\x1a device_sync/consent_backup.proto\x1a\x1edevice_sync/group_backup.proto\x1a device_sync/message_backup.proto\"\xc4\x02\n" +
	"\rBackupElement\x12B\n" +
	"\bmetadata\x18\x01 \x01(\v2$.xmtp.device_sync.BackupMetadataSaveH\x00R\bmetadata\x12@\n" +
	"\x05group\x18\x02 \x01(\v2(.xmtp.device_sync.group_backup.GroupSaveH\x00R\x05group\x12X\n" +
	"\rgroup_message\x18\x03 \x01(\v21.xmtp.device_sync.message_backup.GroupMessageSaveH\x00R\fgroupMessage\x12H\n" +
	"\aconsent\x18\x04 \x01(\v2,.xmtp.device_sync.consent_backup.ConsentSaveH\x00R\aconsentB\t\n" +
	"\aelement\"\xd4\x01\n" +
	"\x12BackupMetadataSave\x12D\n" +
	"\belements\x18\x02 \x03(\x0e2(.xmtp.device_sync.BackupElementSelectionR\belements\x12$\n" +
	"\x0eexported_at_ns\x18\x03 \x01(\x03R\fexportedAtNs\x12\x1e\n" +
	"\bstart_ns\x18\x04 \x01(\x03H\x00R\astartNs\x88\x01\x01\x12\x1a\n" +
	"\x06end_ns\x18\x05 \x01(\x03H\x01R\x05endNs\x88\x01\x01B\v\n" +
	"\t_start_nsB\t\n" +
	"\a_end_ns*\x8f\x01\n" +
	"\x16BackupElementSelection\x12(\n" +
	"$BACKUP_ELEMENT_SELECTION_UNSPECIFIED\x10\x00\x12%\n" +
	"!BACKUP_ELEMENT_SELECTION_MESSAGES\x10\x01\x12$\n" +
	" BACKUP_ELEMENT_SELECTION_CONSENT\x10\x02B\xb8\x01\n" +
	"\x14com.xmtp.device_syncB\x0fDeviceSyncProtoP\x01Z2github.com/xmtp/xmtp-node-go/pkg/proto/device_sync\xa2\x02\x03XDX\xaa\x02\x0fXmtp.DeviceSync\xca\x02\x0fXmtp\\DeviceSync\xe2\x02\x1bXmtp\\DeviceSync\\GPBMetadata\xea\x02\x10Xmtp::DeviceSyncb\x06proto3"

var (
	file_device_sync_device_sync_proto_rawDescOnce sync.Once
	file_device_sync_device_sync_proto_rawDescData []byte
)

func file_device_sync_device_sync_proto_rawDescGZIP() []byte {
	file_device_sync_device_sync_proto_rawDescOnce.Do(func() {
		file_device_sync_device_sync_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_device_sync_device_sync_proto_rawDesc), len(file_device_sync_device_sync_proto_rawDesc)))
	})
	return file_device_sync_device_sync_proto_rawDescData
}

var file_device_sync_device_sync_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_device_sync_device_sync_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_device_sync_device_sync_proto_goTypes = []any{
	(BackupElementSelection)(0), // 0: xmtp.device_sync.BackupElementSelection
	(*BackupElement)(nil),       // 1: xmtp.device_sync.BackupElement
	(*BackupMetadataSave)(nil),  // 2: xmtp.device_sync.BackupMetadataSave
	(*GroupSave)(nil),           // 3: xmtp.device_sync.group_backup.GroupSave
	(*GroupMessageSave)(nil),    // 4: xmtp.device_sync.message_backup.GroupMessageSave
	(*ConsentSave)(nil),         // 5: xmtp.device_sync.consent_backup.ConsentSave
}
var file_device_sync_device_sync_proto_depIdxs = []int32{
	2, // 0: xmtp.device_sync.BackupElement.metadata:type_name -> xmtp.device_sync.BackupMetadataSave
	3, // 1: xmtp.device_sync.BackupElement.group:type_name -> xmtp.device_sync.group_backup.GroupSave
	4, // 2: xmtp.device_sync.BackupElement.group_message:type_name -> xmtp.device_sync.message_backup.GroupMessageSave
	5, // 3: xmtp.device_sync.BackupElement.consent:type_name -> xmtp.device_sync.consent_backup.ConsentSave
	0, // 4: xmtp.device_sync.BackupMetadataSave.elements:type_name -> xmtp.device_sync.BackupElementSelection
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_device_sync_device_sync_proto_init() }
func file_device_sync_device_sync_proto_init() {
	if File_device_sync_device_sync_proto != nil {
		return
	}
	file_device_sync_consent_backup_proto_init()
	file_device_sync_group_backup_proto_init()
	file_device_sync_message_backup_proto_init()
	file_device_sync_device_sync_proto_msgTypes[0].OneofWrappers = []any{
		(*BackupElement_Metadata)(nil),
		(*BackupElement_Group)(nil),
		(*BackupElement_GroupMessage)(nil),
		(*BackupElement_Consent)(nil),
	}
	file_device_sync_device_sync_proto_msgTypes[1].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_device_sync_device_sync_proto_rawDesc), len(file_device_sync_device_sync_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_device_sync_device_sync_proto_goTypes,
		DependencyIndexes: file_device_sync_device_sync_proto_depIdxs,
		EnumInfos:         file_device_sync_device_sync_proto_enumTypes,
		MessageInfos:      file_device_sync_device_sync_proto_msgTypes,
	}.Build()
	File_device_sync_device_sync_proto = out.File
	file_device_sync_device_sync_proto_goTypes = nil
	file_device_sync_device_sync_proto_depIdxs = nil
}
