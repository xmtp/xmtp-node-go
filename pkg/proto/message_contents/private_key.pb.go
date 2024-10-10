// Private Key Storage
//
// Following definitions are not used in the protocol, instead
// they provide a way for encoding private keys for storage.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        (unknown)
// source: message_contents/private_key.proto

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

// PrivateKey generalized to support different key types
type SignedPrivateKey struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// time the key was created
	CreatedNs uint64 `protobuf:"varint,1,opt,name=created_ns,json=createdNs,proto3" json:"created_ns,omitempty"`
	// private key
	//
	// Types that are assignable to Union:
	//
	//	*SignedPrivateKey_Secp256K1_
	Union isSignedPrivateKey_Union `protobuf_oneof:"union"`
	// public key for this private key
	PublicKey *SignedPublicKey `protobuf:"bytes,3,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
}

func (x *SignedPrivateKey) Reset() {
	*x = SignedPrivateKey{}
	mi := &file_message_contents_private_key_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SignedPrivateKey) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignedPrivateKey) ProtoMessage() {}

func (x *SignedPrivateKey) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_private_key_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignedPrivateKey.ProtoReflect.Descriptor instead.
func (*SignedPrivateKey) Descriptor() ([]byte, []int) {
	return file_message_contents_private_key_proto_rawDescGZIP(), []int{0}
}

func (x *SignedPrivateKey) GetCreatedNs() uint64 {
	if x != nil {
		return x.CreatedNs
	}
	return 0
}

func (m *SignedPrivateKey) GetUnion() isSignedPrivateKey_Union {
	if m != nil {
		return m.Union
	}
	return nil
}

func (x *SignedPrivateKey) GetSecp256K1() *SignedPrivateKey_Secp256K1 {
	if x, ok := x.GetUnion().(*SignedPrivateKey_Secp256K1_); ok {
		return x.Secp256K1
	}
	return nil
}

func (x *SignedPrivateKey) GetPublicKey() *SignedPublicKey {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

type isSignedPrivateKey_Union interface {
	isSignedPrivateKey_Union()
}

type SignedPrivateKey_Secp256K1_ struct {
	Secp256K1 *SignedPrivateKey_Secp256K1 `protobuf:"bytes,2,opt,name=secp256k1,proto3,oneof"`
}

func (*SignedPrivateKey_Secp256K1_) isSignedPrivateKey_Union() {}

// PrivateKeyBundle wraps the identityKey and the preKeys,
// enforces usage of signed keys.
type PrivateKeyBundleV2 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IdentityKey *SignedPrivateKey `protobuf:"bytes,1,opt,name=identity_key,json=identityKey,proto3" json:"identity_key,omitempty"`
	// all the known pre-keys, newer keys first,
	PreKeys []*SignedPrivateKey `protobuf:"bytes,2,rep,name=pre_keys,json=preKeys,proto3" json:"pre_keys,omitempty"`
}

func (x *PrivateKeyBundleV2) Reset() {
	*x = PrivateKeyBundleV2{}
	mi := &file_message_contents_private_key_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PrivateKeyBundleV2) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PrivateKeyBundleV2) ProtoMessage() {}

func (x *PrivateKeyBundleV2) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_private_key_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PrivateKeyBundleV2.ProtoReflect.Descriptor instead.
func (*PrivateKeyBundleV2) Descriptor() ([]byte, []int) {
	return file_message_contents_private_key_proto_rawDescGZIP(), []int{1}
}

func (x *PrivateKeyBundleV2) GetIdentityKey() *SignedPrivateKey {
	if x != nil {
		return x.IdentityKey
	}
	return nil
}

func (x *PrivateKeyBundleV2) GetPreKeys() []*SignedPrivateKey {
	if x != nil {
		return x.PreKeys
	}
	return nil
}

// LEGACY: PrivateKey generalized to support different key types
type PrivateKey struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// time the key was created
	Timestamp uint64 `protobuf:"varint,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	// private key
	//
	// Types that are assignable to Union:
	//
	//	*PrivateKey_Secp256K1_
	Union isPrivateKey_Union `protobuf_oneof:"union"`
	// public key for this private key
	PublicKey *PublicKey `protobuf:"bytes,3,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
}

func (x *PrivateKey) Reset() {
	*x = PrivateKey{}
	mi := &file_message_contents_private_key_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PrivateKey) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PrivateKey) ProtoMessage() {}

func (x *PrivateKey) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_private_key_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PrivateKey.ProtoReflect.Descriptor instead.
func (*PrivateKey) Descriptor() ([]byte, []int) {
	return file_message_contents_private_key_proto_rawDescGZIP(), []int{2}
}

func (x *PrivateKey) GetTimestamp() uint64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (m *PrivateKey) GetUnion() isPrivateKey_Union {
	if m != nil {
		return m.Union
	}
	return nil
}

func (x *PrivateKey) GetSecp256K1() *PrivateKey_Secp256K1 {
	if x, ok := x.GetUnion().(*PrivateKey_Secp256K1_); ok {
		return x.Secp256K1
	}
	return nil
}

func (x *PrivateKey) GetPublicKey() *PublicKey {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

type isPrivateKey_Union interface {
	isPrivateKey_Union()
}

type PrivateKey_Secp256K1_ struct {
	Secp256K1 *PrivateKey_Secp256K1 `protobuf:"bytes,2,opt,name=secp256k1,proto3,oneof"`
}

func (*PrivateKey_Secp256K1_) isPrivateKey_Union() {}

// LEGACY: PrivateKeyBundleV1 wraps the identityKey and the preKeys
type PrivateKeyBundleV1 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IdentityKey *PrivateKey `protobuf:"bytes,1,opt,name=identity_key,json=identityKey,proto3" json:"identity_key,omitempty"`
	// all the known pre-keys, newer keys first,
	PreKeys []*PrivateKey `protobuf:"bytes,2,rep,name=pre_keys,json=preKeys,proto3" json:"pre_keys,omitempty"`
}

func (x *PrivateKeyBundleV1) Reset() {
	*x = PrivateKeyBundleV1{}
	mi := &file_message_contents_private_key_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PrivateKeyBundleV1) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PrivateKeyBundleV1) ProtoMessage() {}

func (x *PrivateKeyBundleV1) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_private_key_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PrivateKeyBundleV1.ProtoReflect.Descriptor instead.
func (*PrivateKeyBundleV1) Descriptor() ([]byte, []int) {
	return file_message_contents_private_key_proto_rawDescGZIP(), []int{3}
}

func (x *PrivateKeyBundleV1) GetIdentityKey() *PrivateKey {
	if x != nil {
		return x.IdentityKey
	}
	return nil
}

func (x *PrivateKeyBundleV1) GetPreKeys() []*PrivateKey {
	if x != nil {
		return x.PreKeys
	}
	return nil
}

// Versioned PrivateKeyBundle
type PrivateKeyBundle struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Version:
	//
	//	*PrivateKeyBundle_V1
	//	*PrivateKeyBundle_V2
	Version isPrivateKeyBundle_Version `protobuf_oneof:"version"`
}

func (x *PrivateKeyBundle) Reset() {
	*x = PrivateKeyBundle{}
	mi := &file_message_contents_private_key_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PrivateKeyBundle) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PrivateKeyBundle) ProtoMessage() {}

func (x *PrivateKeyBundle) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_private_key_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PrivateKeyBundle.ProtoReflect.Descriptor instead.
func (*PrivateKeyBundle) Descriptor() ([]byte, []int) {
	return file_message_contents_private_key_proto_rawDescGZIP(), []int{4}
}

func (m *PrivateKeyBundle) GetVersion() isPrivateKeyBundle_Version {
	if m != nil {
		return m.Version
	}
	return nil
}

func (x *PrivateKeyBundle) GetV1() *PrivateKeyBundleV1 {
	if x, ok := x.GetVersion().(*PrivateKeyBundle_V1); ok {
		return x.V1
	}
	return nil
}

func (x *PrivateKeyBundle) GetV2() *PrivateKeyBundleV2 {
	if x, ok := x.GetVersion().(*PrivateKeyBundle_V2); ok {
		return x.V2
	}
	return nil
}

type isPrivateKeyBundle_Version interface {
	isPrivateKeyBundle_Version()
}

type PrivateKeyBundle_V1 struct {
	V1 *PrivateKeyBundleV1 `protobuf:"bytes,1,opt,name=v1,proto3,oneof"`
}

type PrivateKeyBundle_V2 struct {
	V2 *PrivateKeyBundleV2 `protobuf:"bytes,2,opt,name=v2,proto3,oneof"`
}

func (*PrivateKeyBundle_V1) isPrivateKeyBundle_Version() {}

func (*PrivateKeyBundle_V2) isPrivateKeyBundle_Version() {}

// PrivateKeyBundle encrypted with key material generated by
// signing a randomly generated "pre-key" with the user's wallet,
// i.e. EIP-191 signature of a "storage signature" message with
// the pre-key embedded in it.
// (see xmtp-js::PrivateKeyBundle.toEncryptedBytes for details)
type EncryptedPrivateKeyBundleV1 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// randomly generated pre-key
	WalletPreKey []byte `protobuf:"bytes,1,opt,name=wallet_pre_key,json=walletPreKey,proto3" json:"wallet_pre_key,omitempty"` // 32 bytes
	// MUST contain encrypted PrivateKeyBundle
	Ciphertext *Ciphertext `protobuf:"bytes,2,opt,name=ciphertext,proto3" json:"ciphertext,omitempty"`
}

func (x *EncryptedPrivateKeyBundleV1) Reset() {
	*x = EncryptedPrivateKeyBundleV1{}
	mi := &file_message_contents_private_key_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EncryptedPrivateKeyBundleV1) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EncryptedPrivateKeyBundleV1) ProtoMessage() {}

func (x *EncryptedPrivateKeyBundleV1) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_private_key_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EncryptedPrivateKeyBundleV1.ProtoReflect.Descriptor instead.
func (*EncryptedPrivateKeyBundleV1) Descriptor() ([]byte, []int) {
	return file_message_contents_private_key_proto_rawDescGZIP(), []int{5}
}

func (x *EncryptedPrivateKeyBundleV1) GetWalletPreKey() []byte {
	if x != nil {
		return x.WalletPreKey
	}
	return nil
}

func (x *EncryptedPrivateKeyBundleV1) GetCiphertext() *Ciphertext {
	if x != nil {
		return x.Ciphertext
	}
	return nil
}

// Versioned encrypted PrivateKeyBundle
type EncryptedPrivateKeyBundle struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Version:
	//
	//	*EncryptedPrivateKeyBundle_V1
	Version isEncryptedPrivateKeyBundle_Version `protobuf_oneof:"version"`
}

func (x *EncryptedPrivateKeyBundle) Reset() {
	*x = EncryptedPrivateKeyBundle{}
	mi := &file_message_contents_private_key_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EncryptedPrivateKeyBundle) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EncryptedPrivateKeyBundle) ProtoMessage() {}

func (x *EncryptedPrivateKeyBundle) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_private_key_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EncryptedPrivateKeyBundle.ProtoReflect.Descriptor instead.
func (*EncryptedPrivateKeyBundle) Descriptor() ([]byte, []int) {
	return file_message_contents_private_key_proto_rawDescGZIP(), []int{6}
}

func (m *EncryptedPrivateKeyBundle) GetVersion() isEncryptedPrivateKeyBundle_Version {
	if m != nil {
		return m.Version
	}
	return nil
}

func (x *EncryptedPrivateKeyBundle) GetV1() *EncryptedPrivateKeyBundleV1 {
	if x, ok := x.GetVersion().(*EncryptedPrivateKeyBundle_V1); ok {
		return x.V1
	}
	return nil
}

type isEncryptedPrivateKeyBundle_Version interface {
	isEncryptedPrivateKeyBundle_Version()
}

type EncryptedPrivateKeyBundle_V1 struct {
	V1 *EncryptedPrivateKeyBundleV1 `protobuf:"bytes,1,opt,name=v1,proto3,oneof"`
}

func (*EncryptedPrivateKeyBundle_V1) isEncryptedPrivateKeyBundle_Version() {}

// EC: SECP256k1
type SignedPrivateKey_Secp256K1 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Bytes []byte `protobuf:"bytes,1,opt,name=bytes,proto3" json:"bytes,omitempty"` // D big-endian, 32 bytes
}

func (x *SignedPrivateKey_Secp256K1) Reset() {
	*x = SignedPrivateKey_Secp256K1{}
	mi := &file_message_contents_private_key_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SignedPrivateKey_Secp256K1) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignedPrivateKey_Secp256K1) ProtoMessage() {}

func (x *SignedPrivateKey_Secp256K1) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_private_key_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignedPrivateKey_Secp256K1.ProtoReflect.Descriptor instead.
func (*SignedPrivateKey_Secp256K1) Descriptor() ([]byte, []int) {
	return file_message_contents_private_key_proto_rawDescGZIP(), []int{0, 0}
}

func (x *SignedPrivateKey_Secp256K1) GetBytes() []byte {
	if x != nil {
		return x.Bytes
	}
	return nil
}

// EC: SECP256k1
type PrivateKey_Secp256K1 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Bytes []byte `protobuf:"bytes,1,opt,name=bytes,proto3" json:"bytes,omitempty"` // D big-endian, 32 bytes
}

func (x *PrivateKey_Secp256K1) Reset() {
	*x = PrivateKey_Secp256K1{}
	mi := &file_message_contents_private_key_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PrivateKey_Secp256K1) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PrivateKey_Secp256K1) ProtoMessage() {}

func (x *PrivateKey_Secp256K1) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_private_key_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PrivateKey_Secp256K1.ProtoReflect.Descriptor instead.
func (*PrivateKey_Secp256K1) Descriptor() ([]byte, []int) {
	return file_message_contents_private_key_proto_rawDescGZIP(), []int{2, 0}
}

func (x *PrivateKey_Secp256K1) GetBytes() []byte {
	if x != nil {
		return x.Bytes
	}
	return nil
}

var File_message_contents_private_key_proto protoreflect.FileDescriptor

var file_message_contents_private_key_proto_rawDesc = []byte{
	0x0a, 0x22, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x73, 0x2f, 0x70, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x1a, 0x21, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x63, 0x69,
	0x70, 0x68, 0x65, 0x72, 0x74, 0x65, 0x78, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x21,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73,
	0x2f, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0xf7, 0x01, 0x0a, 0x10, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x50, 0x72, 0x69, 0x76,
	0x61, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x64, 0x5f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x64, 0x4e, 0x73, 0x12, 0x51, 0x0a, 0x09, 0x73, 0x65, 0x63, 0x70, 0x32, 0x35, 0x36,
	0x6b, 0x31, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x31, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73,
	0x2e, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x50, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x4b, 0x65,
	0x79, 0x2e, 0x53, 0x65, 0x63, 0x70, 0x32, 0x35, 0x36, 0x6b, 0x31, 0x48, 0x00, 0x52, 0x09, 0x73,
	0x65, 0x63, 0x70, 0x32, 0x35, 0x36, 0x6b, 0x31, 0x12, 0x45, 0x0a, 0x0a, 0x70, 0x75, 0x62, 0x6c,
	0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x78,
	0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x73, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x50, 0x75, 0x62, 0x6c, 0x69,
	0x63, 0x4b, 0x65, 0x79, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x1a,
	0x21, 0x0a, 0x09, 0x53, 0x65, 0x63, 0x70, 0x32, 0x35, 0x36, 0x6b, 0x31, 0x12, 0x14, 0x0a, 0x05,
	0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x62, 0x79, 0x74,
	0x65, 0x73, 0x42, 0x07, 0x0a, 0x05, 0x75, 0x6e, 0x69, 0x6f, 0x6e, 0x22, 0xa4, 0x01, 0x0a, 0x12,
	0x50, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x42, 0x75, 0x6e, 0x64, 0x6c, 0x65,
	0x56, 0x32, 0x12, 0x4a, 0x0a, 0x0c, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x5f, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73,
	0x2e, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x50, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x4b, 0x65,
	0x79, 0x52, 0x0b, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x4b, 0x65, 0x79, 0x12, 0x42,
	0x0a, 0x08, 0x70, 0x72, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x27, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x50,
	0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x52, 0x07, 0x70, 0x72, 0x65, 0x4b, 0x65,
	0x79, 0x73, 0x22, 0xe4, 0x01, 0x0a, 0x0a, 0x50, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x4b, 0x65,
	0x79, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12,
	0x4b, 0x0a, 0x09, 0x73, 0x65, 0x63, 0x70, 0x32, 0x35, 0x36, 0x6b, 0x31, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x50, 0x72, 0x69, 0x76, 0x61,
	0x74, 0x65, 0x4b, 0x65, 0x79, 0x2e, 0x53, 0x65, 0x63, 0x70, 0x32, 0x35, 0x36, 0x6b, 0x31, 0x48,
	0x00, 0x52, 0x09, 0x73, 0x65, 0x63, 0x70, 0x32, 0x35, 0x36, 0x6b, 0x31, 0x12, 0x3f, 0x0a, 0x0a,
	0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x20, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b,
	0x65, 0x79, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x1a, 0x21, 0x0a,
	0x09, 0x53, 0x65, 0x63, 0x70, 0x32, 0x35, 0x36, 0x6b, 0x31, 0x12, 0x14, 0x0a, 0x05, 0x62, 0x79,
	0x74, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x62, 0x79, 0x74, 0x65, 0x73,
	0x42, 0x07, 0x0a, 0x05, 0x75, 0x6e, 0x69, 0x6f, 0x6e, 0x22, 0x98, 0x01, 0x0a, 0x12, 0x50, 0x72,
	0x69, 0x76, 0x61, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x42, 0x75, 0x6e, 0x64, 0x6c, 0x65, 0x56, 0x31,
	0x12, 0x44, 0x0a, 0x0c, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x5f, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x50,
	0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x52, 0x0b, 0x69, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x4b, 0x65, 0x79, 0x12, 0x3c, 0x0a, 0x08, 0x70, 0x72, 0x65, 0x5f, 0x6b, 0x65,
	0x79, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73,
	0x2e, 0x50, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x52, 0x07, 0x70, 0x72, 0x65,
	0x4b, 0x65, 0x79, 0x73, 0x22, 0x9d, 0x01, 0x0a, 0x10, 0x50, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65,
	0x4b, 0x65, 0x79, 0x42, 0x75, 0x6e, 0x64, 0x6c, 0x65, 0x12, 0x3b, 0x0a, 0x02, 0x76, 0x31, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x50, 0x72,
	0x69, 0x76, 0x61, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x42, 0x75, 0x6e, 0x64, 0x6c, 0x65, 0x56, 0x31,
	0x48, 0x00, 0x52, 0x02, 0x76, 0x31, 0x12, 0x3b, 0x0a, 0x02, 0x76, 0x32, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x29, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x50, 0x72, 0x69, 0x76, 0x61,
	0x74, 0x65, 0x4b, 0x65, 0x79, 0x42, 0x75, 0x6e, 0x64, 0x6c, 0x65, 0x56, 0x32, 0x48, 0x00, 0x52,
	0x02, 0x76, 0x32, 0x42, 0x09, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x4a, 0x04,
	0x08, 0x03, 0x10, 0x04, 0x22, 0x86, 0x01, 0x0a, 0x1b, 0x45, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74,
	0x65, 0x64, 0x50, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x42, 0x75, 0x6e, 0x64,
	0x6c, 0x65, 0x56, 0x31, 0x12, 0x24, 0x0a, 0x0e, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x5f, 0x70,
	0x72, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x77, 0x61,
	0x6c, 0x6c, 0x65, 0x74, 0x50, 0x72, 0x65, 0x4b, 0x65, 0x79, 0x12, 0x41, 0x0a, 0x0a, 0x63, 0x69,
	0x70, 0x68, 0x65, 0x72, 0x74, 0x65, 0x78, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21,
	0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x43, 0x69, 0x70, 0x68, 0x65, 0x72, 0x74, 0x65, 0x78,
	0x74, 0x52, 0x0a, 0x63, 0x69, 0x70, 0x68, 0x65, 0x72, 0x74, 0x65, 0x78, 0x74, 0x22, 0x6c, 0x0a,
	0x19, 0x45, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x65, 0x64, 0x50, 0x72, 0x69, 0x76, 0x61, 0x74,
	0x65, 0x4b, 0x65, 0x79, 0x42, 0x75, 0x6e, 0x64, 0x6c, 0x65, 0x12, 0x44, 0x0a, 0x02, 0x76, 0x31,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x32, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x45,
	0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x65, 0x64, 0x50, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x4b,
	0x65, 0x79, 0x42, 0x75, 0x6e, 0x64, 0x6c, 0x65, 0x56, 0x31, 0x48, 0x00, 0x52, 0x02, 0x76, 0x31,
	0x42, 0x09, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x42, 0xd6, 0x01, 0x0a, 0x19,
	0x63, 0x6f, 0x6d, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x42, 0x0f, 0x50, 0x72, 0x69, 0x76, 0x61,
	0x74, 0x65, 0x4b, 0x65, 0x79, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x37, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2f, 0x78, 0x6d,
	0x74, 0x70, 0x2d, 0x6e, 0x6f, 0x64, 0x65, 0x2d, 0x67, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x73, 0xa2, 0x02, 0x03, 0x58, 0x4d, 0x58, 0xaa, 0x02, 0x14, 0x58, 0x6d,
	0x74, 0x70, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x73, 0xca, 0x02, 0x14, 0x58, 0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xe2, 0x02, 0x20, 0x58, 0x6d, 0x74, 0x70,
	0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73,
	0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x15, 0x58,
	0x6d, 0x74, 0x70, 0x3a, 0x3a, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_message_contents_private_key_proto_rawDescOnce sync.Once
	file_message_contents_private_key_proto_rawDescData = file_message_contents_private_key_proto_rawDesc
)

func file_message_contents_private_key_proto_rawDescGZIP() []byte {
	file_message_contents_private_key_proto_rawDescOnce.Do(func() {
		file_message_contents_private_key_proto_rawDescData = protoimpl.X.CompressGZIP(file_message_contents_private_key_proto_rawDescData)
	})
	return file_message_contents_private_key_proto_rawDescData
}

var file_message_contents_private_key_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_message_contents_private_key_proto_goTypes = []any{
	(*SignedPrivateKey)(nil),            // 0: xmtp.message_contents.SignedPrivateKey
	(*PrivateKeyBundleV2)(nil),          // 1: xmtp.message_contents.PrivateKeyBundleV2
	(*PrivateKey)(nil),                  // 2: xmtp.message_contents.PrivateKey
	(*PrivateKeyBundleV1)(nil),          // 3: xmtp.message_contents.PrivateKeyBundleV1
	(*PrivateKeyBundle)(nil),            // 4: xmtp.message_contents.PrivateKeyBundle
	(*EncryptedPrivateKeyBundleV1)(nil), // 5: xmtp.message_contents.EncryptedPrivateKeyBundleV1
	(*EncryptedPrivateKeyBundle)(nil),   // 6: xmtp.message_contents.EncryptedPrivateKeyBundle
	(*SignedPrivateKey_Secp256K1)(nil),  // 7: xmtp.message_contents.SignedPrivateKey.Secp256k1
	(*PrivateKey_Secp256K1)(nil),        // 8: xmtp.message_contents.PrivateKey.Secp256k1
	(*SignedPublicKey)(nil),             // 9: xmtp.message_contents.SignedPublicKey
	(*PublicKey)(nil),                   // 10: xmtp.message_contents.PublicKey
	(*Ciphertext)(nil),                  // 11: xmtp.message_contents.Ciphertext
}
var file_message_contents_private_key_proto_depIdxs = []int32{
	7,  // 0: xmtp.message_contents.SignedPrivateKey.secp256k1:type_name -> xmtp.message_contents.SignedPrivateKey.Secp256k1
	9,  // 1: xmtp.message_contents.SignedPrivateKey.public_key:type_name -> xmtp.message_contents.SignedPublicKey
	0,  // 2: xmtp.message_contents.PrivateKeyBundleV2.identity_key:type_name -> xmtp.message_contents.SignedPrivateKey
	0,  // 3: xmtp.message_contents.PrivateKeyBundleV2.pre_keys:type_name -> xmtp.message_contents.SignedPrivateKey
	8,  // 4: xmtp.message_contents.PrivateKey.secp256k1:type_name -> xmtp.message_contents.PrivateKey.Secp256k1
	10, // 5: xmtp.message_contents.PrivateKey.public_key:type_name -> xmtp.message_contents.PublicKey
	2,  // 6: xmtp.message_contents.PrivateKeyBundleV1.identity_key:type_name -> xmtp.message_contents.PrivateKey
	2,  // 7: xmtp.message_contents.PrivateKeyBundleV1.pre_keys:type_name -> xmtp.message_contents.PrivateKey
	3,  // 8: xmtp.message_contents.PrivateKeyBundle.v1:type_name -> xmtp.message_contents.PrivateKeyBundleV1
	1,  // 9: xmtp.message_contents.PrivateKeyBundle.v2:type_name -> xmtp.message_contents.PrivateKeyBundleV2
	11, // 10: xmtp.message_contents.EncryptedPrivateKeyBundleV1.ciphertext:type_name -> xmtp.message_contents.Ciphertext
	5,  // 11: xmtp.message_contents.EncryptedPrivateKeyBundle.v1:type_name -> xmtp.message_contents.EncryptedPrivateKeyBundleV1
	12, // [12:12] is the sub-list for method output_type
	12, // [12:12] is the sub-list for method input_type
	12, // [12:12] is the sub-list for extension type_name
	12, // [12:12] is the sub-list for extension extendee
	0,  // [0:12] is the sub-list for field type_name
}

func init() { file_message_contents_private_key_proto_init() }
func file_message_contents_private_key_proto_init() {
	if File_message_contents_private_key_proto != nil {
		return
	}
	file_message_contents_ciphertext_proto_init()
	file_message_contents_public_key_proto_init()
	file_message_contents_private_key_proto_msgTypes[0].OneofWrappers = []any{
		(*SignedPrivateKey_Secp256K1_)(nil),
	}
	file_message_contents_private_key_proto_msgTypes[2].OneofWrappers = []any{
		(*PrivateKey_Secp256K1_)(nil),
	}
	file_message_contents_private_key_proto_msgTypes[4].OneofWrappers = []any{
		(*PrivateKeyBundle_V1)(nil),
		(*PrivateKeyBundle_V2)(nil),
	}
	file_message_contents_private_key_proto_msgTypes[6].OneofWrappers = []any{
		(*EncryptedPrivateKeyBundle_V1)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_message_contents_private_key_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_message_contents_private_key_proto_goTypes,
		DependencyIndexes: file_message_contents_private_key_proto_depIdxs,
		MessageInfos:      file_message_contents_private_key_proto_msgTypes,
	}.Build()
	File_message_contents_private_key_proto = out.File
	file_message_contents_private_key_proto_rawDesc = nil
	file_message_contents_private_key_proto_goTypes = nil
	file_message_contents_private_key_proto_depIdxs = nil
}
