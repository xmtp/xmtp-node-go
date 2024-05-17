// Associations and signatures

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        (unknown)
// source: mls/message_contents/association.proto

package message_contents

import (
	message_contents "github.com/xmtp/xmtp-node-go/pkg/proto/message_contents"
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

// Allows for us to update the format of the association text without
// incrementing the entire proto
type AssociationTextVersion int32

const (
	AssociationTextVersion_ASSOCIATION_TEXT_VERSION_UNSPECIFIED AssociationTextVersion = 0
	AssociationTextVersion_ASSOCIATION_TEXT_VERSION_1           AssociationTextVersion = 1
)

// Enum value maps for AssociationTextVersion.
var (
	AssociationTextVersion_name = map[int32]string{
		0: "ASSOCIATION_TEXT_VERSION_UNSPECIFIED",
		1: "ASSOCIATION_TEXT_VERSION_1",
	}
	AssociationTextVersion_value = map[string]int32{
		"ASSOCIATION_TEXT_VERSION_UNSPECIFIED": 0,
		"ASSOCIATION_TEXT_VERSION_1":           1,
	}
)

func (x AssociationTextVersion) Enum() *AssociationTextVersion {
	p := new(AssociationTextVersion)
	*p = x
	return p
}

func (x AssociationTextVersion) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AssociationTextVersion) Descriptor() protoreflect.EnumDescriptor {
	return file_mls_message_contents_association_proto_enumTypes[0].Descriptor()
}

func (AssociationTextVersion) Type() protoreflect.EnumType {
	return &file_mls_message_contents_association_proto_enumTypes[0]
}

func (x AssociationTextVersion) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AssociationTextVersion.Descriptor instead.
func (AssociationTextVersion) EnumDescriptor() ([]byte, []int) {
	return file_mls_message_contents_association_proto_rawDescGZIP(), []int{0}
}

// Used for "Grant Messaging Access" associations
type GrantMessagingAccessAssociation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AssociationTextVersion AssociationTextVersion     `protobuf:"varint,1,opt,name=association_text_version,json=associationTextVersion,proto3,enum=xmtp.mls.message_contents.AssociationTextVersion" json:"association_text_version,omitempty"`
	Signature              *RecoverableEcdsaSignature `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"` // EIP-191 signature
	AccountAddress         string                     `protobuf:"bytes,3,opt,name=account_address,json=accountAddress,proto3" json:"account_address,omitempty"`
	CreatedNs              uint64                     `protobuf:"varint,4,opt,name=created_ns,json=createdNs,proto3" json:"created_ns,omitempty"`
}

func (x *GrantMessagingAccessAssociation) Reset() {
	*x = GrantMessagingAccessAssociation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_message_contents_association_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GrantMessagingAccessAssociation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GrantMessagingAccessAssociation) ProtoMessage() {}

func (x *GrantMessagingAccessAssociation) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_association_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GrantMessagingAccessAssociation.ProtoReflect.Descriptor instead.
func (*GrantMessagingAccessAssociation) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_association_proto_rawDescGZIP(), []int{0}
}

func (x *GrantMessagingAccessAssociation) GetAssociationTextVersion() AssociationTextVersion {
	if x != nil {
		return x.AssociationTextVersion
	}
	return AssociationTextVersion_ASSOCIATION_TEXT_VERSION_UNSPECIFIED
}

func (x *GrantMessagingAccessAssociation) GetSignature() *RecoverableEcdsaSignature {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *GrantMessagingAccessAssociation) GetAccountAddress() string {
	if x != nil {
		return x.AccountAddress
	}
	return ""
}

func (x *GrantMessagingAccessAssociation) GetCreatedNs() uint64 {
	if x != nil {
		return x.CreatedNs
	}
	return 0
}

// Used for "Revoke Messaging Access" associations
type RevokeMessagingAccessAssociation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AssociationTextVersion AssociationTextVersion     `protobuf:"varint,1,opt,name=association_text_version,json=associationTextVersion,proto3,enum=xmtp.mls.message_contents.AssociationTextVersion" json:"association_text_version,omitempty"`
	Signature              *RecoverableEcdsaSignature `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"` // EIP-191 signature
	AccountAddress         string                     `protobuf:"bytes,3,opt,name=account_address,json=accountAddress,proto3" json:"account_address,omitempty"`
	CreatedNs              uint64                     `protobuf:"varint,4,opt,name=created_ns,json=createdNs,proto3" json:"created_ns,omitempty"`
}

func (x *RevokeMessagingAccessAssociation) Reset() {
	*x = RevokeMessagingAccessAssociation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_message_contents_association_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RevokeMessagingAccessAssociation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RevokeMessagingAccessAssociation) ProtoMessage() {}

func (x *RevokeMessagingAccessAssociation) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_association_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RevokeMessagingAccessAssociation.ProtoReflect.Descriptor instead.
func (*RevokeMessagingAccessAssociation) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_association_proto_rawDescGZIP(), []int{1}
}

func (x *RevokeMessagingAccessAssociation) GetAssociationTextVersion() AssociationTextVersion {
	if x != nil {
		return x.AssociationTextVersion
	}
	return AssociationTextVersion_ASSOCIATION_TEXT_VERSION_UNSPECIFIED
}

func (x *RevokeMessagingAccessAssociation) GetSignature() *RecoverableEcdsaSignature {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *RevokeMessagingAccessAssociation) GetAccountAddress() string {
	if x != nil {
		return x.AccountAddress
	}
	return ""
}

func (x *RevokeMessagingAccessAssociation) GetCreatedNs() uint64 {
	if x != nil {
		return x.CreatedNs
	}
	return 0
}

// LegacyCreateIdentityAssociation is used when a v3 installation key
// is signed by a v2 identity key, which in turn is signed via a
// 'CreateIdentity' wallet signature
type LegacyCreateIdentityAssociation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Signs SHA-256 hash of installation key
	Signature *RecoverableEcdsaSignature `protobuf:"bytes,1,opt,name=signature,proto3" json:"signature,omitempty"`
	// created_ns is encoded inside serialized key, account_address is recoverable
	// from the SignedPublicKey signature
	SignedLegacyCreateIdentityKey *message_contents.SignedPublicKey `protobuf:"bytes,2,opt,name=signed_legacy_create_identity_key,json=signedLegacyCreateIdentityKey,proto3" json:"signed_legacy_create_identity_key,omitempty"`
}

func (x *LegacyCreateIdentityAssociation) Reset() {
	*x = LegacyCreateIdentityAssociation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_message_contents_association_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LegacyCreateIdentityAssociation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LegacyCreateIdentityAssociation) ProtoMessage() {}

func (x *LegacyCreateIdentityAssociation) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_association_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LegacyCreateIdentityAssociation.ProtoReflect.Descriptor instead.
func (*LegacyCreateIdentityAssociation) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_association_proto_rawDescGZIP(), []int{2}
}

func (x *LegacyCreateIdentityAssociation) GetSignature() *RecoverableEcdsaSignature {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *LegacyCreateIdentityAssociation) GetSignedLegacyCreateIdentityKey() *message_contents.SignedPublicKey {
	if x != nil {
		return x.SignedLegacyCreateIdentityKey
	}
	return nil
}

// RecoverableEcdsaSignature
type RecoverableEcdsaSignature struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 65-bytes [ R || S || V ], with recovery id as the last byte
	Bytes []byte `protobuf:"bytes,1,opt,name=bytes,proto3" json:"bytes,omitempty"`
}

func (x *RecoverableEcdsaSignature) Reset() {
	*x = RecoverableEcdsaSignature{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_message_contents_association_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecoverableEcdsaSignature) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecoverableEcdsaSignature) ProtoMessage() {}

func (x *RecoverableEcdsaSignature) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_association_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecoverableEcdsaSignature.ProtoReflect.Descriptor instead.
func (*RecoverableEcdsaSignature) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_association_proto_rawDescGZIP(), []int{3}
}

func (x *RecoverableEcdsaSignature) GetBytes() []byte {
	if x != nil {
		return x.Bytes
	}
	return nil
}

// EdDSA signature bytes matching RFC 8032
type EdDsaSignature struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Bytes []byte `protobuf:"bytes,1,opt,name=bytes,proto3" json:"bytes,omitempty"`
}

func (x *EdDsaSignature) Reset() {
	*x = EdDsaSignature{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_message_contents_association_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EdDsaSignature) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EdDsaSignature) ProtoMessage() {}

func (x *EdDsaSignature) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_association_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EdDsaSignature.ProtoReflect.Descriptor instead.
func (*EdDsaSignature) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_association_proto_rawDescGZIP(), []int{4}
}

func (x *EdDsaSignature) GetBytes() []byte {
	if x != nil {
		return x.Bytes
	}
	return nil
}

var File_mls_message_contents_association_proto protoreflect.FileDescriptor

var file_mls_message_contents_association_proto_rawDesc = []byte{
	0x0a, 0x26, 0x6d, 0x6c, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x61, 0x73, 0x73, 0x6f, 0x63, 0x69, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x19, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d,
	0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x73, 0x1a, 0x21, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xaa, 0x02, 0x0a, 0x1f, 0x47, 0x72, 0x61, 0x6e, 0x74,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x41,
	0x73, 0x73, 0x6f, 0x63, 0x69, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x6b, 0x0a, 0x18, 0x61, 0x73,
	0x73, 0x6f, 0x63, 0x69, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x76,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x31, 0x2e, 0x78,
	0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x41, 0x73, 0x73, 0x6f, 0x63, 0x69, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x54, 0x65, 0x78, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52,
	0x16, 0x61, 0x73, 0x73, 0x6f, 0x63, 0x69, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x65, 0x78, 0x74,
	0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x52, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x34, 0x2e, 0x78, 0x6d, 0x74,
	0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x52, 0x65, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x61, 0x62,
	0x6c, 0x65, 0x45, 0x63, 0x64, 0x73, 0x61, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x27, 0x0a, 0x0f, 0x61,
	0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x41, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f,
	0x6e, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x64, 0x4e, 0x73, 0x22, 0xab, 0x02, 0x0a, 0x20, 0x52, 0x65, 0x76, 0x6f, 0x6b, 0x65, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x41, 0x73, 0x73,
	0x6f, 0x63, 0x69, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x6b, 0x0a, 0x18, 0x61, 0x73, 0x73, 0x6f,
	0x63, 0x69, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x31, 0x2e, 0x78, 0x6d, 0x74,
	0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x41, 0x73, 0x73, 0x6f, 0x63, 0x69, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x54, 0x65, 0x78, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x16, 0x61,
	0x73, 0x73, 0x6f, 0x63, 0x69, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x65, 0x78, 0x74, 0x56, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x52, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x34, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e,
	0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x73, 0x2e, 0x52, 0x65, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x61, 0x62, 0x6c, 0x65,
	0x45, 0x63, 0x64, 0x73, 0x61, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x52, 0x09,
	0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x27, 0x0a, 0x0f, 0x61, 0x63, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x6e, 0x73,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x4e,
	0x73, 0x22, 0xe7, 0x01, 0x0a, 0x1f, 0x4c, 0x65, 0x67, 0x61, 0x63, 0x79, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x41, 0x73, 0x73, 0x6f, 0x63, 0x69,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x52, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x34, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e,
	0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x73, 0x2e, 0x52, 0x65, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x61, 0x62, 0x6c, 0x65,
	0x45, 0x63, 0x64, 0x73, 0x61, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x52, 0x09,
	0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x70, 0x0a, 0x21, 0x73, 0x69, 0x67,
	0x6e, 0x65, 0x64, 0x5f, 0x6c, 0x65, 0x67, 0x61, 0x63, 0x79, 0x5f, 0x63, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x5f, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x53, 0x69, 0x67,
	0x6e, 0x65, 0x64, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x52, 0x1d, 0x73, 0x69,
	0x67, 0x6e, 0x65, 0x64, 0x4c, 0x65, 0x67, 0x61, 0x63, 0x79, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x4b, 0x65, 0x79, 0x22, 0x31, 0x0a, 0x19, 0x52,
	0x65, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x61, 0x62, 0x6c, 0x65, 0x45, 0x63, 0x64, 0x73, 0x61, 0x53,
	0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x62, 0x79, 0x74, 0x65,
	0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x62, 0x79, 0x74, 0x65, 0x73, 0x22, 0x26,
	0x0a, 0x0e, 0x45, 0x64, 0x44, 0x73, 0x61, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x12, 0x14, 0x0a, 0x05, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x05, 0x62, 0x79, 0x74, 0x65, 0x73, 0x2a, 0x62, 0x0a, 0x16, 0x41, 0x73, 0x73, 0x6f, 0x63, 0x69,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x65, 0x78, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x12, 0x28, 0x0a, 0x24, 0x41, 0x53, 0x53, 0x4f, 0x43, 0x49, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f,
	0x54, 0x45, 0x58, 0x54, 0x5f, 0x56, 0x45, 0x52, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x55, 0x4e, 0x53,
	0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x1e, 0x0a, 0x1a, 0x41, 0x53,
	0x53, 0x4f, 0x43, 0x49, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x45, 0x58, 0x54, 0x5f, 0x56,
	0x45, 0x52, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x31, 0x10, 0x01, 0x42, 0xf0, 0x01, 0x0a, 0x1d, 0x63,
	0x6f, 0x6d, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x42, 0x10, 0x41, 0x73,
	0x73, 0x6f, 0x63, 0x69, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01,
	0x5a, 0x3b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x78, 0x6d, 0x74,
	0x70, 0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2d, 0x6e, 0x6f, 0x64, 0x65, 0x2d, 0x67, 0x6f, 0x2f, 0x70,
	0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x6c, 0x73, 0x2f, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xa2, 0x02, 0x03,
	0x58, 0x4d, 0x4d, 0xaa, 0x02, 0x18, 0x58, 0x6d, 0x74, 0x70, 0x2e, 0x4d, 0x6c, 0x73, 0x2e, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xca, 0x02,
	0x18, 0x58, 0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x6c, 0x73, 0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xe2, 0x02, 0x24, 0x58, 0x6d, 0x74, 0x70,
	0x5c, 0x4d, 0x6c, 0x73, 0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x73, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0xea, 0x02, 0x1a, 0x58, 0x6d, 0x74, 0x70, 0x3a, 0x3a, 0x4d, 0x6c, 0x73, 0x3a, 0x3a, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mls_message_contents_association_proto_rawDescOnce sync.Once
	file_mls_message_contents_association_proto_rawDescData = file_mls_message_contents_association_proto_rawDesc
)

func file_mls_message_contents_association_proto_rawDescGZIP() []byte {
	file_mls_message_contents_association_proto_rawDescOnce.Do(func() {
		file_mls_message_contents_association_proto_rawDescData = protoimpl.X.CompressGZIP(file_mls_message_contents_association_proto_rawDescData)
	})
	return file_mls_message_contents_association_proto_rawDescData
}

var file_mls_message_contents_association_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_mls_message_contents_association_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_mls_message_contents_association_proto_goTypes = []interface{}{
	(AssociationTextVersion)(0),              // 0: xmtp.mls.message_contents.AssociationTextVersion
	(*GrantMessagingAccessAssociation)(nil),  // 1: xmtp.mls.message_contents.GrantMessagingAccessAssociation
	(*RevokeMessagingAccessAssociation)(nil), // 2: xmtp.mls.message_contents.RevokeMessagingAccessAssociation
	(*LegacyCreateIdentityAssociation)(nil),  // 3: xmtp.mls.message_contents.LegacyCreateIdentityAssociation
	(*RecoverableEcdsaSignature)(nil),        // 4: xmtp.mls.message_contents.RecoverableEcdsaSignature
	(*EdDsaSignature)(nil),                   // 5: xmtp.mls.message_contents.EdDsaSignature
	(*message_contents.SignedPublicKey)(nil), // 6: xmtp.message_contents.SignedPublicKey
}
var file_mls_message_contents_association_proto_depIdxs = []int32{
	0, // 0: xmtp.mls.message_contents.GrantMessagingAccessAssociation.association_text_version:type_name -> xmtp.mls.message_contents.AssociationTextVersion
	4, // 1: xmtp.mls.message_contents.GrantMessagingAccessAssociation.signature:type_name -> xmtp.mls.message_contents.RecoverableEcdsaSignature
	0, // 2: xmtp.mls.message_contents.RevokeMessagingAccessAssociation.association_text_version:type_name -> xmtp.mls.message_contents.AssociationTextVersion
	4, // 3: xmtp.mls.message_contents.RevokeMessagingAccessAssociation.signature:type_name -> xmtp.mls.message_contents.RecoverableEcdsaSignature
	4, // 4: xmtp.mls.message_contents.LegacyCreateIdentityAssociation.signature:type_name -> xmtp.mls.message_contents.RecoverableEcdsaSignature
	6, // 5: xmtp.mls.message_contents.LegacyCreateIdentityAssociation.signed_legacy_create_identity_key:type_name -> xmtp.message_contents.SignedPublicKey
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_mls_message_contents_association_proto_init() }
func file_mls_message_contents_association_proto_init() {
	if File_mls_message_contents_association_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mls_message_contents_association_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GrantMessagingAccessAssociation); i {
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
		file_mls_message_contents_association_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RevokeMessagingAccessAssociation); i {
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
		file_mls_message_contents_association_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LegacyCreateIdentityAssociation); i {
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
		file_mls_message_contents_association_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecoverableEcdsaSignature); i {
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
		file_mls_message_contents_association_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EdDsaSignature); i {
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
			RawDescriptor: file_mls_message_contents_association_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mls_message_contents_association_proto_goTypes,
		DependencyIndexes: file_mls_message_contents_association_proto_depIdxs,
		EnumInfos:         file_mls_message_contents_association_proto_enumTypes,
		MessageInfos:      file_mls_message_contents_association_proto_msgTypes,
	}.Build()
	File_mls_message_contents_association_proto = out.File
	file_mls_message_contents_association_proto_rawDesc = nil
	file_mls_message_contents_association_proto_goTypes = nil
	file_mls_message_contents_association_proto_depIdxs = nil
}
