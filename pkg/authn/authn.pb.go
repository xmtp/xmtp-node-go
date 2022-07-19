// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.2
// source: authn.proto

package authn

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

// Signature represents a generalized public key signature,
// defined as a union to support cryptographic algorithm agility.
type Signature struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Union:
	//	*Signature_EcdsaCompact
	Union isSignature_Union `protobuf_oneof:"union"`
}

func (x *Signature) Reset() {
	*x = Signature{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authn_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Signature) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Signature) ProtoMessage() {}

func (x *Signature) ProtoReflect() protoreflect.Message {
	mi := &file_authn_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Signature.ProtoReflect.Descriptor instead.
func (*Signature) Descriptor() ([]byte, []int) {
	return file_authn_proto_rawDescGZIP(), []int{0}
}

func (m *Signature) GetUnion() isSignature_Union {
	if m != nil {
		return m.Union
	}
	return nil
}

func (x *Signature) GetEcdsaCompact() *Signature_ECDSACompact {
	if x, ok := x.GetUnion().(*Signature_EcdsaCompact); ok {
		return x.EcdsaCompact
	}
	return nil
}

type isSignature_Union interface {
	isSignature_Union()
}

type Signature_EcdsaCompact struct {
	EcdsaCompact *Signature_ECDSACompact `protobuf:"bytes,1,opt,name=ecdsa_compact,json=ecdsaCompact,proto3,oneof"`
}

func (*Signature_EcdsaCompact) isSignature_Union() {}

type Secp256K1Uncompresed struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// uncompressed point with prefix (0x04) [ P || X || Y ], 65 bytes
	Bytes []byte `protobuf:"bytes,1,opt,name=bytes,proto3" json:"bytes,omitempty"`
}

func (x *Secp256K1Uncompresed) Reset() {
	*x = Secp256K1Uncompresed{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authn_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Secp256K1Uncompresed) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Secp256K1Uncompresed) ProtoMessage() {}

func (x *Secp256K1Uncompresed) ProtoReflect() protoreflect.Message {
	mi := &file_authn_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Secp256K1Uncompresed.ProtoReflect.Descriptor instead.
func (*Secp256K1Uncompresed) Descriptor() ([]byte, []int) {
	return file_authn_proto_rawDescGZIP(), []int{1}
}

func (x *Secp256K1Uncompresed) GetBytes() []byte {
	if x != nil {
		return x.Bytes
	}
	return nil
}

// PublicKey represents a generalized public key,
// defined as a union to support cryptographic algorithm agility.
type PublicKey struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Timestamp uint64     `protobuf:"varint,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Signature *Signature `protobuf:"bytes,2,opt,name=signature,proto3,oneof" json:"signature,omitempty"`
	// Types that are assignable to Union:
	//	*PublicKey_Secp256K1Uncompressed
	Union isPublicKey_Union `protobuf_oneof:"union"`
}

func (x *PublicKey) Reset() {
	*x = PublicKey{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authn_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PublicKey) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublicKey) ProtoMessage() {}

func (x *PublicKey) ProtoReflect() protoreflect.Message {
	mi := &file_authn_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublicKey.ProtoReflect.Descriptor instead.
func (*PublicKey) Descriptor() ([]byte, []int) {
	return file_authn_proto_rawDescGZIP(), []int{2}
}

func (x *PublicKey) GetTimestamp() uint64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *PublicKey) GetSignature() *Signature {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (m *PublicKey) GetUnion() isPublicKey_Union {
	if m != nil {
		return m.Union
	}
	return nil
}

func (x *PublicKey) GetSecp256K1Uncompressed() *Secp256K1Uncompresed {
	if x, ok := x.GetUnion().(*PublicKey_Secp256K1Uncompressed); ok {
		return x.Secp256K1Uncompressed
	}
	return nil
}

type isPublicKey_Union interface {
	isPublicKey_Union()
}

type PublicKey_Secp256K1Uncompressed struct {
	Secp256K1Uncompressed *Secp256K1Uncompresed `protobuf:"bytes,3,opt,name=secp256k1_uncompressed,json=secp256k1Uncompressed,proto3,oneof"`
}

func (*PublicKey_Secp256K1Uncompressed) isPublicKey_Union() {}

type V1ClientAuthRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IdentityKeyBytes []byte     `protobuf:"bytes,1,opt,name=identity_key_bytes,json=identityKeyBytes,proto3" json:"identity_key_bytes,omitempty"`
	WalletSignature  *Signature `protobuf:"bytes,2,opt,name=wallet_signature,json=walletSignature,proto3" json:"wallet_signature,omitempty"`
	AuthDataBytes    []byte     `protobuf:"bytes,3,opt,name=auth_data_bytes,json=authDataBytes,proto3" json:"auth_data_bytes,omitempty"`
	AuthSignature    *Signature `protobuf:"bytes,4,opt,name=auth_signature,json=authSignature,proto3" json:"auth_signature,omitempty"`
}

func (x *V1ClientAuthRequest) Reset() {
	*x = V1ClientAuthRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authn_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *V1ClientAuthRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*V1ClientAuthRequest) ProtoMessage() {}

func (x *V1ClientAuthRequest) ProtoReflect() protoreflect.Message {
	mi := &file_authn_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use V1ClientAuthRequest.ProtoReflect.Descriptor instead.
func (*V1ClientAuthRequest) Descriptor() ([]byte, []int) {
	return file_authn_proto_rawDescGZIP(), []int{3}
}

func (x *V1ClientAuthRequest) GetIdentityKeyBytes() []byte {
	if x != nil {
		return x.IdentityKeyBytes
	}
	return nil
}

func (x *V1ClientAuthRequest) GetWalletSignature() *Signature {
	if x != nil {
		return x.WalletSignature
	}
	return nil
}

func (x *V1ClientAuthRequest) GetAuthDataBytes() []byte {
	if x != nil {
		return x.AuthDataBytes
	}
	return nil
}

func (x *V1ClientAuthRequest) GetAuthSignature() *Signature {
	if x != nil {
		return x.AuthSignature
	}
	return nil
}

type ClientAuthRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Version:
	//	*ClientAuthRequest_V1
	Version isClientAuthRequest_Version `protobuf_oneof:"version"`
}

func (x *ClientAuthRequest) Reset() {
	*x = ClientAuthRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authn_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientAuthRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientAuthRequest) ProtoMessage() {}

func (x *ClientAuthRequest) ProtoReflect() protoreflect.Message {
	mi := &file_authn_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientAuthRequest.ProtoReflect.Descriptor instead.
func (*ClientAuthRequest) Descriptor() ([]byte, []int) {
	return file_authn_proto_rawDescGZIP(), []int{4}
}

func (m *ClientAuthRequest) GetVersion() isClientAuthRequest_Version {
	if m != nil {
		return m.Version
	}
	return nil
}

func (x *ClientAuthRequest) GetV1() *V1ClientAuthRequest {
	if x, ok := x.GetVersion().(*ClientAuthRequest_V1); ok {
		return x.V1
	}
	return nil
}

type isClientAuthRequest_Version interface {
	isClientAuthRequest_Version()
}

type ClientAuthRequest_V1 struct {
	V1 *V1ClientAuthRequest `protobuf:"bytes,1,opt,name=v1,proto3,oneof"`
}

func (*ClientAuthRequest_V1) isClientAuthRequest_Version() {}

type V1ClientAuthResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AuthSuccessful bool   `protobuf:"varint,1,opt,name=auth_successful,json=authSuccessful,proto3" json:"auth_successful,omitempty"`
	ErrorStr       string `protobuf:"bytes,2,opt,name=error_str,json=errorStr,proto3" json:"error_str,omitempty"`
}

func (x *V1ClientAuthResponse) Reset() {
	*x = V1ClientAuthResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authn_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *V1ClientAuthResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*V1ClientAuthResponse) ProtoMessage() {}

func (x *V1ClientAuthResponse) ProtoReflect() protoreflect.Message {
	mi := &file_authn_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use V1ClientAuthResponse.ProtoReflect.Descriptor instead.
func (*V1ClientAuthResponse) Descriptor() ([]byte, []int) {
	return file_authn_proto_rawDescGZIP(), []int{5}
}

func (x *V1ClientAuthResponse) GetAuthSuccessful() bool {
	if x != nil {
		return x.AuthSuccessful
	}
	return false
}

func (x *V1ClientAuthResponse) GetErrorStr() string {
	if x != nil {
		return x.ErrorStr
	}
	return ""
}

type ClientAuthResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Version:
	//	*ClientAuthResponse_V1
	Version isClientAuthResponse_Version `protobuf_oneof:"version"`
}

func (x *ClientAuthResponse) Reset() {
	*x = ClientAuthResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authn_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientAuthResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientAuthResponse) ProtoMessage() {}

func (x *ClientAuthResponse) ProtoReflect() protoreflect.Message {
	mi := &file_authn_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientAuthResponse.ProtoReflect.Descriptor instead.
func (*ClientAuthResponse) Descriptor() ([]byte, []int) {
	return file_authn_proto_rawDescGZIP(), []int{6}
}

func (m *ClientAuthResponse) GetVersion() isClientAuthResponse_Version {
	if m != nil {
		return m.Version
	}
	return nil
}

func (x *ClientAuthResponse) GetV1() *V1ClientAuthResponse {
	if x, ok := x.GetVersion().(*ClientAuthResponse_V1); ok {
		return x.V1
	}
	return nil
}

type isClientAuthResponse_Version interface {
	isClientAuthResponse_Version()
}

type ClientAuthResponse_V1 struct {
	V1 *V1ClientAuthResponse `protobuf:"bytes,1,opt,name=v1,proto3,oneof"`
}

func (*ClientAuthResponse_V1) isClientAuthResponse_Version() {}

type AuthData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	WalletAddr string `protobuf:"bytes,1,opt,name=wallet_addr,json=walletAddr,proto3" json:"wallet_addr,omitempty"`
	PeerId     string `protobuf:"bytes,2,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	Timestamp  uint64 `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *AuthData) Reset() {
	*x = AuthData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authn_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthData) ProtoMessage() {}

func (x *AuthData) ProtoReflect() protoreflect.Message {
	mi := &file_authn_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthData.ProtoReflect.Descriptor instead.
func (*AuthData) Descriptor() ([]byte, []int) {
	return file_authn_proto_rawDescGZIP(), []int{7}
}

func (x *AuthData) GetWalletAddr() string {
	if x != nil {
		return x.WalletAddr
	}
	return ""
}

func (x *AuthData) GetPeerId() string {
	if x != nil {
		return x.PeerId
	}
	return ""
}

func (x *AuthData) GetTimestamp() uint64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

type Signature_ECDSACompact struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Bytes    []byte `protobuf:"bytes,1,opt,name=bytes,proto3" json:"bytes,omitempty"`        // compact representation [ R || S ], 64 bytes
	Recovery uint32 `protobuf:"varint,2,opt,name=recovery,proto3" json:"recovery,omitempty"` // recovery bit
}

func (x *Signature_ECDSACompact) Reset() {
	*x = Signature_ECDSACompact{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authn_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Signature_ECDSACompact) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Signature_ECDSACompact) ProtoMessage() {}

func (x *Signature_ECDSACompact) ProtoReflect() protoreflect.Message {
	mi := &file_authn_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Signature_ECDSACompact.ProtoReflect.Descriptor instead.
func (*Signature_ECDSACompact) Descriptor() ([]byte, []int) {
	return file_authn_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Signature_ECDSACompact) GetBytes() []byte {
	if x != nil {
		return x.Bytes
	}
	return nil
}

func (x *Signature_ECDSACompact) GetRecovery() uint32 {
	if x != nil {
		return x.Recovery
	}
	return 0
}

var File_authn_proto protoreflect.FileDescriptor

var file_authn_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x61, 0x75, 0x74, 0x68, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70,
	0x62, 0x22, 0x99, 0x01, 0x0a, 0x09, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12,
	0x41, 0x0a, 0x0d, 0x65, 0x63, 0x64, 0x73, 0x61, 0x5f, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x63, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x69, 0x67, 0x6e,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x2e, 0x45, 0x43, 0x44, 0x53, 0x41, 0x43, 0x6f, 0x6d, 0x70, 0x61,
	0x63, 0x74, 0x48, 0x00, 0x52, 0x0c, 0x65, 0x63, 0x64, 0x73, 0x61, 0x43, 0x6f, 0x6d, 0x70, 0x61,
	0x63, 0x74, 0x1a, 0x40, 0x0a, 0x0c, 0x45, 0x43, 0x44, 0x53, 0x41, 0x43, 0x6f, 0x6d, 0x70, 0x61,
	0x63, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x05, 0x62, 0x79, 0x74, 0x65, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x63, 0x6f,
	0x76, 0x65, 0x72, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x08, 0x72, 0x65, 0x63, 0x6f,
	0x76, 0x65, 0x72, 0x79, 0x42, 0x07, 0x0a, 0x05, 0x75, 0x6e, 0x69, 0x6f, 0x6e, 0x22, 0x2c, 0x0a,
	0x14, 0x53, 0x65, 0x63, 0x70, 0x32, 0x35, 0x36, 0x6b, 0x31, 0x55, 0x6e, 0x63, 0x6f, 0x6d, 0x70,
	0x72, 0x65, 0x73, 0x65, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x62, 0x79, 0x74, 0x65, 0x73, 0x22, 0xc5, 0x01, 0x0a, 0x09,
	0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x74, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x30, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x70, 0x62, 0x2e,
	0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x48, 0x01, 0x52, 0x09, 0x73, 0x69, 0x67,
	0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x88, 0x01, 0x01, 0x12, 0x51, 0x0a, 0x16, 0x73, 0x65, 0x63,
	0x70, 0x32, 0x35, 0x36, 0x6b, 0x31, 0x5f, 0x75, 0x6e, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73,
	0x73, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x70, 0x62, 0x2e, 0x53,
	0x65, 0x63, 0x70, 0x32, 0x35, 0x36, 0x6b, 0x31, 0x55, 0x6e, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65,
	0x73, 0x65, 0x64, 0x48, 0x00, 0x52, 0x15, 0x73, 0x65, 0x63, 0x70, 0x32, 0x35, 0x36, 0x6b, 0x31,
	0x55, 0x6e, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x65, 0x64, 0x42, 0x07, 0x0a, 0x05,
	0x75, 0x6e, 0x69, 0x6f, 0x6e, 0x42, 0x0c, 0x0a, 0x0a, 0x5f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74,
	0x75, 0x72, 0x65, 0x22, 0xdb, 0x01, 0x0a, 0x13, 0x56, 0x31, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x41, 0x75, 0x74, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2c, 0x0a, 0x12, 0x69,
	0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x5f, 0x6b, 0x65, 0x79, 0x5f, 0x62, 0x79, 0x74, 0x65,
	0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x10, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74,
	0x79, 0x4b, 0x65, 0x79, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x38, 0x0a, 0x10, 0x77, 0x61, 0x6c,
	0x6c, 0x65, 0x74, 0x5f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x52, 0x0f, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74,
	0x75, 0x72, 0x65, 0x12, 0x26, 0x0a, 0x0f, 0x61, 0x75, 0x74, 0x68, 0x5f, 0x64, 0x61, 0x74, 0x61,
	0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0d, 0x61, 0x75,
	0x74, 0x68, 0x44, 0x61, 0x74, 0x61, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x34, 0x0a, 0x0e, 0x61,
	0x75, 0x74, 0x68, 0x5f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x52, 0x0d, 0x61, 0x75, 0x74, 0x68, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x22, 0x49, 0x0a, 0x11, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x41, 0x75, 0x74, 0x68, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x29, 0x0a, 0x02, 0x76, 0x31, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x17, 0x2e, 0x70, 0x62, 0x2e, 0x56, 0x31, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x41, 0x75, 0x74, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x02, 0x76,
	0x31, 0x42, 0x09, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x5c, 0x0a, 0x14,
	0x56, 0x31, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x41, 0x75, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x27, 0x0a, 0x0f, 0x61, 0x75, 0x74, 0x68, 0x5f, 0x73, 0x75, 0x63,
	0x63, 0x65, 0x73, 0x73, 0x66, 0x75, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0e, 0x61,
	0x75, 0x74, 0x68, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x66, 0x75, 0x6c, 0x12, 0x1b, 0x0a,
	0x09, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x73, 0x74, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x53, 0x74, 0x72, 0x22, 0x4b, 0x0a, 0x12, 0x43, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x41, 0x75, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x2a, 0x0a, 0x02, 0x76, 0x31, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x70,
	0x62, 0x2e, 0x56, 0x31, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x41, 0x75, 0x74, 0x68, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x48, 0x00, 0x52, 0x02, 0x76, 0x31, 0x42, 0x09, 0x0a, 0x07,
	0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x62, 0x0a, 0x08, 0x41, 0x75, 0x74, 0x68, 0x44,
	0x61, 0x74, 0x61, 0x12, 0x1f, 0x0a, 0x0b, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x5f, 0x61, 0x64,
	0x64, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74,
	0x41, 0x64, 0x64, 0x72, 0x12, 0x17, 0x0a, 0x07, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x65, 0x65, 0x72, 0x49, 0x64, 0x12, 0x1c, 0x0a,
	0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x09, 0x5a, 0x07, 0x2e,
	0x3b, 0x61, 0x75, 0x74, 0x68, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_authn_proto_rawDescOnce sync.Once
	file_authn_proto_rawDescData = file_authn_proto_rawDesc
)

func file_authn_proto_rawDescGZIP() []byte {
	file_authn_proto_rawDescOnce.Do(func() {
		file_authn_proto_rawDescData = protoimpl.X.CompressGZIP(file_authn_proto_rawDescData)
	})
	return file_authn_proto_rawDescData
}

var file_authn_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_authn_proto_goTypes = []interface{}{
	(*Signature)(nil),              // 0: pb.Signature
	(*Secp256K1Uncompresed)(nil),   // 1: pb.Secp256k1Uncompresed
	(*PublicKey)(nil),              // 2: pb.PublicKey
	(*V1ClientAuthRequest)(nil),    // 3: pb.V1ClientAuthRequest
	(*ClientAuthRequest)(nil),      // 4: pb.ClientAuthRequest
	(*V1ClientAuthResponse)(nil),   // 5: pb.V1ClientAuthResponse
	(*ClientAuthResponse)(nil),     // 6: pb.ClientAuthResponse
	(*AuthData)(nil),               // 7: pb.AuthData
	(*Signature_ECDSACompact)(nil), // 8: pb.Signature.ECDSACompact
}
var file_authn_proto_depIdxs = []int32{
	8, // 0: pb.Signature.ecdsa_compact:type_name -> pb.Signature.ECDSACompact
	0, // 1: pb.PublicKey.signature:type_name -> pb.Signature
	1, // 2: pb.PublicKey.secp256k1_uncompressed:type_name -> pb.Secp256k1Uncompresed
	0, // 3: pb.V1ClientAuthRequest.wallet_signature:type_name -> pb.Signature
	0, // 4: pb.V1ClientAuthRequest.auth_signature:type_name -> pb.Signature
	3, // 5: pb.ClientAuthRequest.v1:type_name -> pb.V1ClientAuthRequest
	5, // 6: pb.ClientAuthResponse.v1:type_name -> pb.V1ClientAuthResponse
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_authn_proto_init() }
func file_authn_proto_init() {
	if File_authn_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_authn_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Signature); i {
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
		file_authn_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Secp256K1Uncompresed); i {
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
		file_authn_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PublicKey); i {
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
		file_authn_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*V1ClientAuthRequest); i {
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
		file_authn_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClientAuthRequest); i {
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
		file_authn_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*V1ClientAuthResponse); i {
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
		file_authn_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClientAuthResponse); i {
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
		file_authn_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthData); i {
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
		file_authn_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Signature_ECDSACompact); i {
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
	file_authn_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Signature_EcdsaCompact)(nil),
	}
	file_authn_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*PublicKey_Secp256K1Uncompressed)(nil),
	}
	file_authn_proto_msgTypes[4].OneofWrappers = []interface{}{
		(*ClientAuthRequest_V1)(nil),
	}
	file_authn_proto_msgTypes[6].OneofWrappers = []interface{}{
		(*ClientAuthResponse_V1)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_authn_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_authn_proto_goTypes,
		DependencyIndexes: file_authn_proto_depIdxs,
		MessageInfos:      file_authn_proto_msgTypes,
	}.Build()
	File_authn_proto = out.File
	file_authn_proto_rawDesc = nil
	file_authn_proto_goTypes = nil
	file_authn_proto_depIdxs = nil
}
