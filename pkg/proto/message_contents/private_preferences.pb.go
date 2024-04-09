// Private Key Storage
//
// Following definitions are not used in the protocol, instead they provide a
// way for encoding private keys for storage.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: message_contents/private_preferences.proto

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

// PrivatePreferencesAction is a message used to update the client's preference
// store.
type PrivatePreferencesAction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to MessageType:
	//
	//	*PrivatePreferencesAction_AllowAddress_
	//	*PrivatePreferencesAction_DenyAddress_
	//	*PrivatePreferencesAction_AllowGroup_
	//	*PrivatePreferencesAction_DenyGroup_
	MessageType isPrivatePreferencesAction_MessageType `protobuf_oneof:"message_type"`
}

func (x *PrivatePreferencesAction) Reset() {
	*x = PrivatePreferencesAction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_contents_private_preferences_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PrivatePreferencesAction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PrivatePreferencesAction) ProtoMessage() {}

func (x *PrivatePreferencesAction) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_private_preferences_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PrivatePreferencesAction.ProtoReflect.Descriptor instead.
func (*PrivatePreferencesAction) Descriptor() ([]byte, []int) {
	return file_message_contents_private_preferences_proto_rawDescGZIP(), []int{0}
}

func (m *PrivatePreferencesAction) GetMessageType() isPrivatePreferencesAction_MessageType {
	if m != nil {
		return m.MessageType
	}
	return nil
}

func (x *PrivatePreferencesAction) GetAllowAddress() *PrivatePreferencesAction_AllowAddress {
	if x, ok := x.GetMessageType().(*PrivatePreferencesAction_AllowAddress_); ok {
		return x.AllowAddress
	}
	return nil
}

func (x *PrivatePreferencesAction) GetDenyAddress() *PrivatePreferencesAction_DenyAddress {
	if x, ok := x.GetMessageType().(*PrivatePreferencesAction_DenyAddress_); ok {
		return x.DenyAddress
	}
	return nil
}

func (x *PrivatePreferencesAction) GetAllowGroup() *PrivatePreferencesAction_AllowGroup {
	if x, ok := x.GetMessageType().(*PrivatePreferencesAction_AllowGroup_); ok {
		return x.AllowGroup
	}
	return nil
}

func (x *PrivatePreferencesAction) GetDenyGroup() *PrivatePreferencesAction_DenyGroup {
	if x, ok := x.GetMessageType().(*PrivatePreferencesAction_DenyGroup_); ok {
		return x.DenyGroup
	}
	return nil
}

type isPrivatePreferencesAction_MessageType interface {
	isPrivatePreferencesAction_MessageType()
}

type PrivatePreferencesAction_AllowAddress_ struct {
	AllowAddress *PrivatePreferencesAction_AllowAddress `protobuf:"bytes,1,opt,name=allow_address,json=allowAddress,proto3,oneof"`
}

type PrivatePreferencesAction_DenyAddress_ struct {
	DenyAddress *PrivatePreferencesAction_DenyAddress `protobuf:"bytes,2,opt,name=deny_address,json=denyAddress,proto3,oneof"`
}

type PrivatePreferencesAction_AllowGroup_ struct {
	AllowGroup *PrivatePreferencesAction_AllowGroup `protobuf:"bytes,3,opt,name=allow_group,json=allowGroup,proto3,oneof"`
}

type PrivatePreferencesAction_DenyGroup_ struct {
	DenyGroup *PrivatePreferencesAction_DenyGroup `protobuf:"bytes,4,opt,name=deny_group,json=denyGroup,proto3,oneof"`
}

func (*PrivatePreferencesAction_AllowAddress_) isPrivatePreferencesAction_MessageType() {}

func (*PrivatePreferencesAction_DenyAddress_) isPrivatePreferencesAction_MessageType() {}

func (*PrivatePreferencesAction_AllowGroup_) isPrivatePreferencesAction_MessageType() {}

func (*PrivatePreferencesAction_DenyGroup_) isPrivatePreferencesAction_MessageType() {}

// The payload that goes over the wire
type PrivatePreferencesPayload struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Version:
	//
	//	*PrivatePreferencesPayload_V1
	Version isPrivatePreferencesPayload_Version `protobuf_oneof:"version"`
}

func (x *PrivatePreferencesPayload) Reset() {
	*x = PrivatePreferencesPayload{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_contents_private_preferences_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PrivatePreferencesPayload) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PrivatePreferencesPayload) ProtoMessage() {}

func (x *PrivatePreferencesPayload) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_private_preferences_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PrivatePreferencesPayload.ProtoReflect.Descriptor instead.
func (*PrivatePreferencesPayload) Descriptor() ([]byte, []int) {
	return file_message_contents_private_preferences_proto_rawDescGZIP(), []int{1}
}

func (m *PrivatePreferencesPayload) GetVersion() isPrivatePreferencesPayload_Version {
	if m != nil {
		return m.Version
	}
	return nil
}

func (x *PrivatePreferencesPayload) GetV1() *Ciphertext {
	if x, ok := x.GetVersion().(*PrivatePreferencesPayload_V1); ok {
		return x.V1
	}
	return nil
}

type isPrivatePreferencesPayload_Version interface {
	isPrivatePreferencesPayload_Version()
}

type PrivatePreferencesPayload_V1 struct {
	V1 *Ciphertext `protobuf:"bytes,1,opt,name=v1,proto3,oneof"`
}

func (*PrivatePreferencesPayload_V1) isPrivatePreferencesPayload_Version() {}

// Allow 1:1 direct message (DM) access
type PrivatePreferencesAction_AllowAddress struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Add the given wallet addresses to the allow list
	WalletAddresses []string `protobuf:"bytes,1,rep,name=wallet_addresses,json=walletAddresses,proto3" json:"wallet_addresses,omitempty"`
}

func (x *PrivatePreferencesAction_AllowAddress) Reset() {
	*x = PrivatePreferencesAction_AllowAddress{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_contents_private_preferences_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PrivatePreferencesAction_AllowAddress) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PrivatePreferencesAction_AllowAddress) ProtoMessage() {}

func (x *PrivatePreferencesAction_AllowAddress) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_private_preferences_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PrivatePreferencesAction_AllowAddress.ProtoReflect.Descriptor instead.
func (*PrivatePreferencesAction_AllowAddress) Descriptor() ([]byte, []int) {
	return file_message_contents_private_preferences_proto_rawDescGZIP(), []int{0, 0}
}

func (x *PrivatePreferencesAction_AllowAddress) GetWalletAddresses() []string {
	if x != nil {
		return x.WalletAddresses
	}
	return nil
}

// Deny (block) 1:1 direct message (DM) access
type PrivatePreferencesAction_DenyAddress struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Add the given wallet addresses to the deny list
	WalletAddresses []string `protobuf:"bytes,1,rep,name=wallet_addresses,json=walletAddresses,proto3" json:"wallet_addresses,omitempty"`
}

func (x *PrivatePreferencesAction_DenyAddress) Reset() {
	*x = PrivatePreferencesAction_DenyAddress{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_contents_private_preferences_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PrivatePreferencesAction_DenyAddress) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PrivatePreferencesAction_DenyAddress) ProtoMessage() {}

func (x *PrivatePreferencesAction_DenyAddress) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_private_preferences_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PrivatePreferencesAction_DenyAddress.ProtoReflect.Descriptor instead.
func (*PrivatePreferencesAction_DenyAddress) Descriptor() ([]byte, []int) {
	return file_message_contents_private_preferences_proto_rawDescGZIP(), []int{0, 1}
}

func (x *PrivatePreferencesAction_DenyAddress) GetWalletAddresses() []string {
	if x != nil {
		return x.WalletAddresses
	}
	return nil
}

// Allow Group access
type PrivatePreferencesAction_AllowGroup struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Add the given group_ids to the allow list
	GroupIds [][]byte `protobuf:"bytes,1,rep,name=group_ids,json=groupIds,proto3" json:"group_ids,omitempty"`
}

func (x *PrivatePreferencesAction_AllowGroup) Reset() {
	*x = PrivatePreferencesAction_AllowGroup{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_contents_private_preferences_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PrivatePreferencesAction_AllowGroup) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PrivatePreferencesAction_AllowGroup) ProtoMessage() {}

func (x *PrivatePreferencesAction_AllowGroup) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_private_preferences_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PrivatePreferencesAction_AllowGroup.ProtoReflect.Descriptor instead.
func (*PrivatePreferencesAction_AllowGroup) Descriptor() ([]byte, []int) {
	return file_message_contents_private_preferences_proto_rawDescGZIP(), []int{0, 2}
}

func (x *PrivatePreferencesAction_AllowGroup) GetGroupIds() [][]byte {
	if x != nil {
		return x.GroupIds
	}
	return nil
}

// Deny (deny) Group access
type PrivatePreferencesAction_DenyGroup struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Add the given group_ids to the deny list
	GroupIds [][]byte `protobuf:"bytes,1,rep,name=group_ids,json=groupIds,proto3" json:"group_ids,omitempty"`
}

func (x *PrivatePreferencesAction_DenyGroup) Reset() {
	*x = PrivatePreferencesAction_DenyGroup{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_contents_private_preferences_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PrivatePreferencesAction_DenyGroup) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PrivatePreferencesAction_DenyGroup) ProtoMessage() {}

func (x *PrivatePreferencesAction_DenyGroup) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_private_preferences_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PrivatePreferencesAction_DenyGroup.ProtoReflect.Descriptor instead.
func (*PrivatePreferencesAction_DenyGroup) Descriptor() ([]byte, []int) {
	return file_message_contents_private_preferences_proto_rawDescGZIP(), []int{0, 3}
}

func (x *PrivatePreferencesAction_DenyGroup) GetGroupIds() [][]byte {
	if x != nil {
		return x.GroupIds
	}
	return nil
}

var File_message_contents_private_preferences_proto protoreflect.FileDescriptor

var file_message_contents_private_preferences_proto_rawDesc = []byte{
	0x0a, 0x2a, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x73, 0x2f, 0x70, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x65,
	0x72, 0x65, 0x6e, 0x63, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x78, 0x6d,
	0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x73, 0x1a, 0x21, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x63, 0x69, 0x70, 0x68, 0x65, 0x72, 0x74, 0x65, 0x78, 0x74,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xf6, 0x04, 0x0a, 0x18, 0x50, 0x72, 0x69, 0x76, 0x61,
	0x74, 0x65, 0x50, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x73, 0x41, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x63, 0x0a, 0x0d, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x5f, 0x61, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x3c, 0x2e, 0x78, 0x6d, 0x74,
	0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x73, 0x2e, 0x50, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x50, 0x72, 0x65, 0x66, 0x65, 0x72,
	0x65, 0x6e, 0x63, 0x65, 0x73, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x41, 0x6c, 0x6c, 0x6f,
	0x77, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x48, 0x00, 0x52, 0x0c, 0x61, 0x6c, 0x6c, 0x6f,
	0x77, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x60, 0x0a, 0x0c, 0x64, 0x65, 0x6e, 0x79,
	0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x3b,
	0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x50, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x50, 0x72,
	0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x73, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e,
	0x44, 0x65, 0x6e, 0x79, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x48, 0x00, 0x52, 0x0b, 0x64,
	0x65, 0x6e, 0x79, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x5d, 0x0a, 0x0b, 0x61, 0x6c,
	0x6c, 0x6f, 0x77, 0x5f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x3a, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x50, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x50,
	0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x73, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x2e, 0x41, 0x6c, 0x6c, 0x6f, 0x77, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x48, 0x00, 0x52, 0x0a, 0x61,
	0x6c, 0x6c, 0x6f, 0x77, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x12, 0x5a, 0x0a, 0x0a, 0x64, 0x65, 0x6e,
	0x79, 0x5f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x39, 0x2e,
	0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x50, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x50, 0x72, 0x65,
	0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x73, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x44,
	0x65, 0x6e, 0x79, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x48, 0x00, 0x52, 0x09, 0x64, 0x65, 0x6e, 0x79,
	0x47, 0x72, 0x6f, 0x75, 0x70, 0x1a, 0x39, 0x0a, 0x0c, 0x41, 0x6c, 0x6c, 0x6f, 0x77, 0x41, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x29, 0x0a, 0x10, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x5f,
	0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x0f, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73,
	0x1a, 0x38, 0x0a, 0x0b, 0x44, 0x65, 0x6e, 0x79, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12,
	0x29, 0x0a, 0x10, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0f, 0x77, 0x61, 0x6c, 0x6c, 0x65,
	0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x1a, 0x29, 0x0a, 0x0a, 0x41, 0x6c,
	0x6c, 0x6f, 0x77, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x12, 0x1b, 0x0a, 0x09, 0x67, 0x72, 0x6f, 0x75,
	0x70, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x08, 0x67, 0x72, 0x6f,
	0x75, 0x70, 0x49, 0x64, 0x73, 0x1a, 0x28, 0x0a, 0x09, 0x44, 0x65, 0x6e, 0x79, 0x47, 0x72, 0x6f,
	0x75, 0x70, 0x12, 0x1b, 0x0a, 0x09, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x69, 0x64, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x08, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x73, 0x42,
	0x0e, 0x0a, 0x0c, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x22,
	0x5b, 0x0a, 0x19, 0x50, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x50, 0x72, 0x65, 0x66, 0x65, 0x72,
	0x65, 0x6e, 0x63, 0x65, 0x73, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x33, 0x0a, 0x02,
	0x76, 0x31, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73,
	0x2e, 0x43, 0x69, 0x70, 0x68, 0x65, 0x72, 0x74, 0x65, 0x78, 0x74, 0x48, 0x00, 0x52, 0x02, 0x76,
	0x31, 0x42, 0x09, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x42, 0xde, 0x01, 0x0a,
	0x19, 0x63, 0x6f, 0x6d, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x42, 0x17, 0x50, 0x72, 0x69, 0x76,
	0x61, 0x74, 0x65, 0x50, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x73, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2d, 0x6e, 0x6f, 0x64, 0x65,
	0x2d, 0x67, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xa2, 0x02,
	0x03, 0x58, 0x4d, 0x58, 0xaa, 0x02, 0x14, 0x58, 0x6d, 0x74, 0x70, 0x2e, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xca, 0x02, 0x14, 0x58, 0x6d,
	0x74, 0x70, 0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x73, 0xe2, 0x02, 0x20, 0x58, 0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x15, 0x58, 0x6d, 0x74, 0x70, 0x3a, 0x3a, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_message_contents_private_preferences_proto_rawDescOnce sync.Once
	file_message_contents_private_preferences_proto_rawDescData = file_message_contents_private_preferences_proto_rawDesc
)

func file_message_contents_private_preferences_proto_rawDescGZIP() []byte {
	file_message_contents_private_preferences_proto_rawDescOnce.Do(func() {
		file_message_contents_private_preferences_proto_rawDescData = protoimpl.X.CompressGZIP(file_message_contents_private_preferences_proto_rawDescData)
	})
	return file_message_contents_private_preferences_proto_rawDescData
}

var file_message_contents_private_preferences_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_message_contents_private_preferences_proto_goTypes = []interface{}{
	(*PrivatePreferencesAction)(nil),              // 0: xmtp.message_contents.PrivatePreferencesAction
	(*PrivatePreferencesPayload)(nil),             // 1: xmtp.message_contents.PrivatePreferencesPayload
	(*PrivatePreferencesAction_AllowAddress)(nil), // 2: xmtp.message_contents.PrivatePreferencesAction.AllowAddress
	(*PrivatePreferencesAction_DenyAddress)(nil),  // 3: xmtp.message_contents.PrivatePreferencesAction.DenyAddress
	(*PrivatePreferencesAction_AllowGroup)(nil),   // 4: xmtp.message_contents.PrivatePreferencesAction.AllowGroup
	(*PrivatePreferencesAction_DenyGroup)(nil),    // 5: xmtp.message_contents.PrivatePreferencesAction.DenyGroup
	(*Ciphertext)(nil),                            // 6: xmtp.message_contents.Ciphertext
}
var file_message_contents_private_preferences_proto_depIdxs = []int32{
	2, // 0: xmtp.message_contents.PrivatePreferencesAction.allow_address:type_name -> xmtp.message_contents.PrivatePreferencesAction.AllowAddress
	3, // 1: xmtp.message_contents.PrivatePreferencesAction.deny_address:type_name -> xmtp.message_contents.PrivatePreferencesAction.DenyAddress
	4, // 2: xmtp.message_contents.PrivatePreferencesAction.allow_group:type_name -> xmtp.message_contents.PrivatePreferencesAction.AllowGroup
	5, // 3: xmtp.message_contents.PrivatePreferencesAction.deny_group:type_name -> xmtp.message_contents.PrivatePreferencesAction.DenyGroup
	6, // 4: xmtp.message_contents.PrivatePreferencesPayload.v1:type_name -> xmtp.message_contents.Ciphertext
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_message_contents_private_preferences_proto_init() }
func file_message_contents_private_preferences_proto_init() {
	if File_message_contents_private_preferences_proto != nil {
		return
	}
	file_message_contents_ciphertext_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_message_contents_private_preferences_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PrivatePreferencesAction); i {
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
		file_message_contents_private_preferences_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PrivatePreferencesPayload); i {
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
		file_message_contents_private_preferences_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PrivatePreferencesAction_AllowAddress); i {
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
		file_message_contents_private_preferences_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PrivatePreferencesAction_DenyAddress); i {
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
		file_message_contents_private_preferences_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PrivatePreferencesAction_AllowGroup); i {
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
		file_message_contents_private_preferences_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PrivatePreferencesAction_DenyGroup); i {
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
	file_message_contents_private_preferences_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*PrivatePreferencesAction_AllowAddress_)(nil),
		(*PrivatePreferencesAction_DenyAddress_)(nil),
		(*PrivatePreferencesAction_AllowGroup_)(nil),
		(*PrivatePreferencesAction_DenyGroup_)(nil),
	}
	file_message_contents_private_preferences_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*PrivatePreferencesPayload_V1)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_message_contents_private_preferences_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_message_contents_private_preferences_proto_goTypes,
		DependencyIndexes: file_message_contents_private_preferences_proto_depIdxs,
		MessageInfos:      file_message_contents_private_preferences_proto_msgTypes,
	}.Build()
	File_message_contents_private_preferences_proto = out.File
	file_message_contents_private_preferences_proto_rawDesc = nil
	file_message_contents_private_preferences_proto_goTypes = nil
	file_message_contents_private_preferences_proto_depIdxs = nil
}
