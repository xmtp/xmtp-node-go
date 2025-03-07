// Definitions for backups

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        (unknown)
// source: device_sync/group_backup.proto

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

// Group membership state
type GroupMembershipStateSave int32

const (
	GroupMembershipStateSave_GROUP_MEMBERSHIP_STATE_SAVE_UNSPECIFIED GroupMembershipStateSave = 0
	GroupMembershipStateSave_GROUP_MEMBERSHIP_STATE_SAVE_ALLOWED     GroupMembershipStateSave = 1
	GroupMembershipStateSave_GROUP_MEMBERSHIP_STATE_SAVE_REJECTED    GroupMembershipStateSave = 2
	GroupMembershipStateSave_GROUP_MEMBERSHIP_STATE_SAVE_PENDING     GroupMembershipStateSave = 3
)

// Enum value maps for GroupMembershipStateSave.
var (
	GroupMembershipStateSave_name = map[int32]string{
		0: "GROUP_MEMBERSHIP_STATE_SAVE_UNSPECIFIED",
		1: "GROUP_MEMBERSHIP_STATE_SAVE_ALLOWED",
		2: "GROUP_MEMBERSHIP_STATE_SAVE_REJECTED",
		3: "GROUP_MEMBERSHIP_STATE_SAVE_PENDING",
	}
	GroupMembershipStateSave_value = map[string]int32{
		"GROUP_MEMBERSHIP_STATE_SAVE_UNSPECIFIED": 0,
		"GROUP_MEMBERSHIP_STATE_SAVE_ALLOWED":     1,
		"GROUP_MEMBERSHIP_STATE_SAVE_REJECTED":    2,
		"GROUP_MEMBERSHIP_STATE_SAVE_PENDING":     3,
	}
)

func (x GroupMembershipStateSave) Enum() *GroupMembershipStateSave {
	p := new(GroupMembershipStateSave)
	*p = x
	return p
}

func (x GroupMembershipStateSave) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (GroupMembershipStateSave) Descriptor() protoreflect.EnumDescriptor {
	return file_device_sync_group_backup_proto_enumTypes[0].Descriptor()
}

func (GroupMembershipStateSave) Type() protoreflect.EnumType {
	return &file_device_sync_group_backup_proto_enumTypes[0]
}

func (x GroupMembershipStateSave) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use GroupMembershipStateSave.Descriptor instead.
func (GroupMembershipStateSave) EnumDescriptor() ([]byte, []int) {
	return file_device_sync_group_backup_proto_rawDescGZIP(), []int{0}
}

// Conversation type
type ConversationTypeSave int32

const (
	ConversationTypeSave_CONVERSATION_TYPE_SAVE_UNSPECIFIED ConversationTypeSave = 0
	ConversationTypeSave_CONVERSATION_TYPE_SAVE_GROUP       ConversationTypeSave = 1
	ConversationTypeSave_CONVERSATION_TYPE_SAVE_DM          ConversationTypeSave = 2
	ConversationTypeSave_CONVERSATION_TYPE_SAVE_SYNC        ConversationTypeSave = 3
)

// Enum value maps for ConversationTypeSave.
var (
	ConversationTypeSave_name = map[int32]string{
		0: "CONVERSATION_TYPE_SAVE_UNSPECIFIED",
		1: "CONVERSATION_TYPE_SAVE_GROUP",
		2: "CONVERSATION_TYPE_SAVE_DM",
		3: "CONVERSATION_TYPE_SAVE_SYNC",
	}
	ConversationTypeSave_value = map[string]int32{
		"CONVERSATION_TYPE_SAVE_UNSPECIFIED": 0,
		"CONVERSATION_TYPE_SAVE_GROUP":       1,
		"CONVERSATION_TYPE_SAVE_DM":          2,
		"CONVERSATION_TYPE_SAVE_SYNC":        3,
	}
)

func (x ConversationTypeSave) Enum() *ConversationTypeSave {
	p := new(ConversationTypeSave)
	*p = x
	return p
}

func (x ConversationTypeSave) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ConversationTypeSave) Descriptor() protoreflect.EnumDescriptor {
	return file_device_sync_group_backup_proto_enumTypes[1].Descriptor()
}

func (ConversationTypeSave) Type() protoreflect.EnumType {
	return &file_device_sync_group_backup_proto_enumTypes[1]
}

func (x ConversationTypeSave) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ConversationTypeSave.Descriptor instead.
func (ConversationTypeSave) EnumDescriptor() ([]byte, []int) {
	return file_device_sync_group_backup_proto_rawDescGZIP(), []int{1}
}

// Proto representation of a stored group
type GroupSave struct {
	state                    protoimpl.MessageState   `protogen:"open.v1"`
	Id                       []byte                   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	CreatedAtNs              int64                    `protobuf:"varint,2,opt,name=created_at_ns,json=createdAtNs,proto3" json:"created_at_ns,omitempty"`
	MembershipState          GroupMembershipStateSave `protobuf:"varint,3,opt,name=membership_state,json=membershipState,proto3,enum=xmtp.device_sync.group_backup.GroupMembershipStateSave" json:"membership_state,omitempty"`
	InstallationsLastChecked int64                    `protobuf:"varint,4,opt,name=installations_last_checked,json=installationsLastChecked,proto3" json:"installations_last_checked,omitempty"`
	AddedByInboxId           string                   `protobuf:"bytes,5,opt,name=added_by_inbox_id,json=addedByInboxId,proto3" json:"added_by_inbox_id,omitempty"`
	WelcomeId                *int64                   `protobuf:"varint,6,opt,name=welcome_id,json=welcomeId,proto3,oneof" json:"welcome_id,omitempty"`
	RotatedAtNs              int64                    `protobuf:"varint,7,opt,name=rotated_at_ns,json=rotatedAtNs,proto3" json:"rotated_at_ns,omitempty"`
	ConversationType         ConversationTypeSave     `protobuf:"varint,8,opt,name=conversation_type,json=conversationType,proto3,enum=xmtp.device_sync.group_backup.ConversationTypeSave" json:"conversation_type,omitempty"`
	DmId                     *string                  `protobuf:"bytes,9,opt,name=dm_id,json=dmId,proto3,oneof" json:"dm_id,omitempty"`
	LastMessageNs            *int64                   `protobuf:"varint,10,opt,name=last_message_ns,json=lastMessageNs,proto3,oneof" json:"last_message_ns,omitempty"`
	MessageDisappearFromNs   *int64                   `protobuf:"varint,11,opt,name=message_disappear_from_ns,json=messageDisappearFromNs,proto3,oneof" json:"message_disappear_from_ns,omitempty"`
	MessageDisappearInNs     *int64                   `protobuf:"varint,12,opt,name=message_disappear_in_ns,json=messageDisappearInNs,proto3,oneof" json:"message_disappear_in_ns,omitempty"`
	unknownFields            protoimpl.UnknownFields
	sizeCache                protoimpl.SizeCache
}

func (x *GroupSave) Reset() {
	*x = GroupSave{}
	mi := &file_device_sync_group_backup_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GroupSave) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupSave) ProtoMessage() {}

func (x *GroupSave) ProtoReflect() protoreflect.Message {
	mi := &file_device_sync_group_backup_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupSave.ProtoReflect.Descriptor instead.
func (*GroupSave) Descriptor() ([]byte, []int) {
	return file_device_sync_group_backup_proto_rawDescGZIP(), []int{0}
}

func (x *GroupSave) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *GroupSave) GetCreatedAtNs() int64 {
	if x != nil {
		return x.CreatedAtNs
	}
	return 0
}

func (x *GroupSave) GetMembershipState() GroupMembershipStateSave {
	if x != nil {
		return x.MembershipState
	}
	return GroupMembershipStateSave_GROUP_MEMBERSHIP_STATE_SAVE_UNSPECIFIED
}

func (x *GroupSave) GetInstallationsLastChecked() int64 {
	if x != nil {
		return x.InstallationsLastChecked
	}
	return 0
}

func (x *GroupSave) GetAddedByInboxId() string {
	if x != nil {
		return x.AddedByInboxId
	}
	return ""
}

func (x *GroupSave) GetWelcomeId() int64 {
	if x != nil && x.WelcomeId != nil {
		return *x.WelcomeId
	}
	return 0
}

func (x *GroupSave) GetRotatedAtNs() int64 {
	if x != nil {
		return x.RotatedAtNs
	}
	return 0
}

func (x *GroupSave) GetConversationType() ConversationTypeSave {
	if x != nil {
		return x.ConversationType
	}
	return ConversationTypeSave_CONVERSATION_TYPE_SAVE_UNSPECIFIED
}

func (x *GroupSave) GetDmId() string {
	if x != nil && x.DmId != nil {
		return *x.DmId
	}
	return ""
}

func (x *GroupSave) GetLastMessageNs() int64 {
	if x != nil && x.LastMessageNs != nil {
		return *x.LastMessageNs
	}
	return 0
}

func (x *GroupSave) GetMessageDisappearFromNs() int64 {
	if x != nil && x.MessageDisappearFromNs != nil {
		return *x.MessageDisappearFromNs
	}
	return 0
}

func (x *GroupSave) GetMessageDisappearInNs() int64 {
	if x != nil && x.MessageDisappearInNs != nil {
		return *x.MessageDisappearInNs
	}
	return 0
}

var File_device_sync_group_backup_proto protoreflect.FileDescriptor

var file_device_sync_group_backup_proto_rawDesc = string([]byte{
	0x0a, 0x1e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x73, 0x79, 0x6e, 0x63, 0x2f, 0x67, 0x72,
	0x6f, 0x75, 0x70, 0x5f, 0x62, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x1d, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x73, 0x79,
	0x6e, 0x63, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x62, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x22,
	0xe0, 0x05, 0x0a, 0x09, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x53, 0x61, 0x76, 0x65, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64, 0x12, 0x22, 0x0a,
	0x0d, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x5f, 0x6e, 0x73, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x4e,
	0x73, 0x12, 0x62, 0x0a, 0x10, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x5f,
	0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x37, 0x2e, 0x78, 0x6d,
	0x74, 0x70, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x73, 0x79, 0x6e, 0x63, 0x2e, 0x67,
	0x72, 0x6f, 0x75, 0x70, 0x5f, 0x62, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x2e, 0x47, 0x72, 0x6f, 0x75,
	0x70, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x53, 0x74, 0x61, 0x74, 0x65,
	0x53, 0x61, 0x76, 0x65, 0x52, 0x0f, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x3c, 0x0a, 0x1a, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x5f, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x63, 0x68, 0x65, 0x63,
	0x6b, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x18, 0x69, 0x6e, 0x73, 0x74, 0x61,
	0x6c, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x4c, 0x61, 0x73, 0x74, 0x43, 0x68, 0x65, 0x63,
	0x6b, 0x65, 0x64, 0x12, 0x29, 0x0a, 0x11, 0x61, 0x64, 0x64, 0x65, 0x64, 0x5f, 0x62, 0x79, 0x5f,
	0x69, 0x6e, 0x62, 0x6f, 0x78, 0x5f, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e,
	0x61, 0x64, 0x64, 0x65, 0x64, 0x42, 0x79, 0x49, 0x6e, 0x62, 0x6f, 0x78, 0x49, 0x64, 0x12, 0x22,
	0x0a, 0x0a, 0x77, 0x65, 0x6c, 0x63, 0x6f, 0x6d, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x03, 0x48, 0x00, 0x52, 0x09, 0x77, 0x65, 0x6c, 0x63, 0x6f, 0x6d, 0x65, 0x49, 0x64, 0x88,
	0x01, 0x01, 0x12, 0x22, 0x0a, 0x0d, 0x72, 0x6f, 0x74, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74,
	0x5f, 0x6e, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x72, 0x6f, 0x74, 0x61, 0x74,
	0x65, 0x64, 0x41, 0x74, 0x4e, 0x73, 0x12, 0x60, 0x0a, 0x11, 0x63, 0x6f, 0x6e, 0x76, 0x65, 0x72,
	0x73, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x33, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f,
	0x73, 0x79, 0x6e, 0x63, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x62, 0x61, 0x63, 0x6b, 0x75,
	0x70, 0x2e, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79,
	0x70, 0x65, 0x53, 0x61, 0x76, 0x65, 0x52, 0x10, 0x63, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x05, 0x64, 0x6d, 0x5f, 0x69,
	0x64, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x48, 0x01, 0x52, 0x04, 0x64, 0x6d, 0x49, 0x64, 0x88,
	0x01, 0x01, 0x12, 0x2b, 0x0a, 0x0f, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x5f, 0x6e, 0x73, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x03, 0x48, 0x02, 0x52, 0x0d, 0x6c,
	0x61, 0x73, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4e, 0x73, 0x88, 0x01, 0x01, 0x12,
	0x3e, 0x0a, 0x19, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x64, 0x69, 0x73, 0x61, 0x70,
	0x70, 0x65, 0x61, 0x72, 0x5f, 0x66, 0x72, 0x6f, 0x6d, 0x5f, 0x6e, 0x73, 0x18, 0x0b, 0x20, 0x01,
	0x28, 0x03, 0x48, 0x03, 0x52, 0x16, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x44, 0x69, 0x73,
	0x61, 0x70, 0x70, 0x65, 0x61, 0x72, 0x46, 0x72, 0x6f, 0x6d, 0x4e, 0x73, 0x88, 0x01, 0x01, 0x12,
	0x3a, 0x0a, 0x17, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x64, 0x69, 0x73, 0x61, 0x70,
	0x70, 0x65, 0x61, 0x72, 0x5f, 0x69, 0x6e, 0x5f, 0x6e, 0x73, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x03,
	0x48, 0x04, 0x52, 0x14, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x44, 0x69, 0x73, 0x61, 0x70,
	0x70, 0x65, 0x61, 0x72, 0x49, 0x6e, 0x4e, 0x73, 0x88, 0x01, 0x01, 0x42, 0x0d, 0x0a, 0x0b, 0x5f,
	0x77, 0x65, 0x6c, 0x63, 0x6f, 0x6d, 0x65, 0x5f, 0x69, 0x64, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x64,
	0x6d, 0x5f, 0x69, 0x64, 0x42, 0x12, 0x0a, 0x10, 0x5f, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x6e, 0x73, 0x42, 0x1c, 0x0a, 0x1a, 0x5f, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x5f, 0x64, 0x69, 0x73, 0x61, 0x70, 0x70, 0x65, 0x61, 0x72, 0x5f, 0x66,
	0x72, 0x6f, 0x6d, 0x5f, 0x6e, 0x73, 0x42, 0x1a, 0x0a, 0x18, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x5f, 0x64, 0x69, 0x73, 0x61, 0x70, 0x70, 0x65, 0x61, 0x72, 0x5f, 0x69, 0x6e, 0x5f,
	0x6e, 0x73, 0x2a, 0xc3, 0x01, 0x0a, 0x18, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x6d, 0x62,
	0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x53, 0x74, 0x61, 0x74, 0x65, 0x53, 0x61, 0x76, 0x65, 0x12,
	0x2b, 0x0a, 0x27, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x5f, 0x4d, 0x45, 0x4d, 0x42, 0x45, 0x52, 0x53,
	0x48, 0x49, 0x50, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x53, 0x41, 0x56, 0x45, 0x5f, 0x55,
	0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x27, 0x0a, 0x23,
	0x47, 0x52, 0x4f, 0x55, 0x50, 0x5f, 0x4d, 0x45, 0x4d, 0x42, 0x45, 0x52, 0x53, 0x48, 0x49, 0x50,
	0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x53, 0x41, 0x56, 0x45, 0x5f, 0x41, 0x4c, 0x4c, 0x4f,
	0x57, 0x45, 0x44, 0x10, 0x01, 0x12, 0x28, 0x0a, 0x24, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x5f, 0x4d,
	0x45, 0x4d, 0x42, 0x45, 0x52, 0x53, 0x48, 0x49, 0x50, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f,
	0x53, 0x41, 0x56, 0x45, 0x5f, 0x52, 0x45, 0x4a, 0x45, 0x43, 0x54, 0x45, 0x44, 0x10, 0x02, 0x12,
	0x27, 0x0a, 0x23, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x5f, 0x4d, 0x45, 0x4d, 0x42, 0x45, 0x52, 0x53,
	0x48, 0x49, 0x50, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x53, 0x41, 0x56, 0x45, 0x5f, 0x50,
	0x45, 0x4e, 0x44, 0x49, 0x4e, 0x47, 0x10, 0x03, 0x2a, 0xa0, 0x01, 0x0a, 0x14, 0x43, 0x6f, 0x6e,
	0x76, 0x65, 0x72, 0x73, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x53, 0x61, 0x76,
	0x65, 0x12, 0x26, 0x0a, 0x22, 0x43, 0x4f, 0x4e, 0x56, 0x45, 0x52, 0x53, 0x41, 0x54, 0x49, 0x4f,
	0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x53, 0x41, 0x56, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50,
	0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x20, 0x0a, 0x1c, 0x43, 0x4f, 0x4e,
	0x56, 0x45, 0x52, 0x53, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x53,
	0x41, 0x56, 0x45, 0x5f, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x10, 0x01, 0x12, 0x1d, 0x0a, 0x19, 0x43,
	0x4f, 0x4e, 0x56, 0x45, 0x52, 0x53, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45,
	0x5f, 0x53, 0x41, 0x56, 0x45, 0x5f, 0x44, 0x4d, 0x10, 0x02, 0x12, 0x1f, 0x0a, 0x1b, 0x43, 0x4f,
	0x4e, 0x56, 0x45, 0x52, 0x53, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f,
	0x53, 0x41, 0x56, 0x45, 0x5f, 0x53, 0x59, 0x4e, 0x43, 0x10, 0x03, 0x42, 0xf7, 0x01, 0x0a, 0x21,
	0x63, 0x6f, 0x6d, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f,
	0x73, 0x79, 0x6e, 0x63, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x62, 0x61, 0x63, 0x6b, 0x75,
	0x70, 0x42, 0x10, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2d, 0x6e, 0x6f, 0x64, 0x65,
	0x2d, 0x67, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x64, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x5f, 0x73, 0x79, 0x6e, 0x63, 0xa2, 0x02, 0x03, 0x58, 0x44, 0x47, 0xaa,
	0x02, 0x1b, 0x58, 0x6d, 0x74, 0x70, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x53, 0x79, 0x6e,
	0x63, 0x2e, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0xca, 0x02, 0x1b,
	0x58, 0x6d, 0x74, 0x70, 0x5c, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x53, 0x79, 0x6e, 0x63, 0x5c,
	0x47, 0x72, 0x6f, 0x75, 0x70, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0xe2, 0x02, 0x27, 0x58, 0x6d,
	0x74, 0x70, 0x5c, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x53, 0x79, 0x6e, 0x63, 0x5c, 0x47, 0x72,
	0x6f, 0x75, 0x70, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x1d, 0x58, 0x6d, 0x74, 0x70, 0x3a, 0x3a, 0x44, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x53, 0x79, 0x6e, 0x63, 0x3a, 0x3a, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x42,
	0x61, 0x63, 0x6b, 0x75, 0x70, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_device_sync_group_backup_proto_rawDescOnce sync.Once
	file_device_sync_group_backup_proto_rawDescData []byte
)

func file_device_sync_group_backup_proto_rawDescGZIP() []byte {
	file_device_sync_group_backup_proto_rawDescOnce.Do(func() {
		file_device_sync_group_backup_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_device_sync_group_backup_proto_rawDesc), len(file_device_sync_group_backup_proto_rawDesc)))
	})
	return file_device_sync_group_backup_proto_rawDescData
}

var file_device_sync_group_backup_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_device_sync_group_backup_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_device_sync_group_backup_proto_goTypes = []any{
	(GroupMembershipStateSave)(0), // 0: xmtp.device_sync.group_backup.GroupMembershipStateSave
	(ConversationTypeSave)(0),     // 1: xmtp.device_sync.group_backup.ConversationTypeSave
	(*GroupSave)(nil),             // 2: xmtp.device_sync.group_backup.GroupSave
}
var file_device_sync_group_backup_proto_depIdxs = []int32{
	0, // 0: xmtp.device_sync.group_backup.GroupSave.membership_state:type_name -> xmtp.device_sync.group_backup.GroupMembershipStateSave
	1, // 1: xmtp.device_sync.group_backup.GroupSave.conversation_type:type_name -> xmtp.device_sync.group_backup.ConversationTypeSave
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_device_sync_group_backup_proto_init() }
func file_device_sync_group_backup_proto_init() {
	if File_device_sync_group_backup_proto != nil {
		return
	}
	file_device_sync_group_backup_proto_msgTypes[0].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_device_sync_group_backup_proto_rawDesc), len(file_device_sync_group_backup_proto_rawDesc)),
			NumEnums:      2,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_device_sync_group_backup_proto_goTypes,
		DependencyIndexes: file_device_sync_group_backup_proto_depIdxs,
		EnumInfos:         file_device_sync_group_backup_proto_enumTypes,
		MessageInfos:      file_device_sync_group_backup_proto_msgTypes,
	}.Build()
	File_device_sync_group_backup_proto = out.File
	file_device_sync_group_backup_proto_goTypes = nil
	file_device_sync_group_backup_proto_depIdxs = nil
}
