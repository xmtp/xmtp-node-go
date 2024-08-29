// Invitation is used by an initiator to invite participants
// into a new conversation. Invitation carries the chosen topic name
// and encryption scheme and key material to be used for message encryption.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: message_contents/invitation.proto

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

// Version of consent proof payload
type ConsentProofPayloadVersion int32

const (
	ConsentProofPayloadVersion_CONSENT_PROOF_PAYLOAD_VERSION_UNSPECIFIED ConsentProofPayloadVersion = 0
	ConsentProofPayloadVersion_CONSENT_PROOF_PAYLOAD_VERSION_1           ConsentProofPayloadVersion = 1
)

// Enum value maps for ConsentProofPayloadVersion.
var (
	ConsentProofPayloadVersion_name = map[int32]string{
		0: "CONSENT_PROOF_PAYLOAD_VERSION_UNSPECIFIED",
		1: "CONSENT_PROOF_PAYLOAD_VERSION_1",
	}
	ConsentProofPayloadVersion_value = map[string]int32{
		"CONSENT_PROOF_PAYLOAD_VERSION_UNSPECIFIED": 0,
		"CONSENT_PROOF_PAYLOAD_VERSION_1":           1,
	}
)

func (x ConsentProofPayloadVersion) Enum() *ConsentProofPayloadVersion {
	p := new(ConsentProofPayloadVersion)
	*p = x
	return p
}

func (x ConsentProofPayloadVersion) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ConsentProofPayloadVersion) Descriptor() protoreflect.EnumDescriptor {
	return file_message_contents_invitation_proto_enumTypes[0].Descriptor()
}

func (ConsentProofPayloadVersion) Type() protoreflect.EnumType {
	return &file_message_contents_invitation_proto_enumTypes[0]
}

func (x ConsentProofPayloadVersion) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ConsentProofPayloadVersion.Descriptor instead.
func (ConsentProofPayloadVersion) EnumDescriptor() ([]byte, []int) {
	return file_message_contents_invitation_proto_rawDescGZIP(), []int{0}
}

// Unsealed invitation V1
type InvitationV1 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// topic name chosen for this conversation.
	// It MUST be randomly generated bytes (length >= 32),
	// then base64 encoded without padding
	Topic string `protobuf:"bytes,1,opt,name=topic,proto3" json:"topic,omitempty"`
	// A context object defining metadata
	Context *InvitationV1_Context `protobuf:"bytes,2,opt,name=context,proto3" json:"context,omitempty"`
	// message encryption scheme and keys for this conversation.
	//
	// Types that are assignable to Encryption:
	//
	//	*InvitationV1_Aes256GcmHkdfSha256
	Encryption isInvitationV1_Encryption `protobuf_oneof:"encryption"`
	// The user's consent proof
	ConsentProof *ConsentProofPayload `protobuf:"bytes,4,opt,name=consent_proof,json=consentProof,proto3" json:"consent_proof,omitempty"`
}

func (x *InvitationV1) Reset() {
	*x = InvitationV1{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_contents_invitation_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InvitationV1) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InvitationV1) ProtoMessage() {}

func (x *InvitationV1) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_invitation_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InvitationV1.ProtoReflect.Descriptor instead.
func (*InvitationV1) Descriptor() ([]byte, []int) {
	return file_message_contents_invitation_proto_rawDescGZIP(), []int{0}
}

func (x *InvitationV1) GetTopic() string {
	if x != nil {
		return x.Topic
	}
	return ""
}

func (x *InvitationV1) GetContext() *InvitationV1_Context {
	if x != nil {
		return x.Context
	}
	return nil
}

func (m *InvitationV1) GetEncryption() isInvitationV1_Encryption {
	if m != nil {
		return m.Encryption
	}
	return nil
}

func (x *InvitationV1) GetAes256GcmHkdfSha256() *InvitationV1_Aes256GcmHkdfsha256 {
	if x, ok := x.GetEncryption().(*InvitationV1_Aes256GcmHkdfSha256); ok {
		return x.Aes256GcmHkdfSha256
	}
	return nil
}

func (x *InvitationV1) GetConsentProof() *ConsentProofPayload {
	if x != nil {
		return x.ConsentProof
	}
	return nil
}

type isInvitationV1_Encryption interface {
	isInvitationV1_Encryption()
}

type InvitationV1_Aes256GcmHkdfSha256 struct {
	// Specify the encryption method to process the key material properly.
	Aes256GcmHkdfSha256 *InvitationV1_Aes256GcmHkdfsha256 `protobuf:"bytes,3,opt,name=aes256_gcm_hkdf_sha256,json=aes256GcmHkdfSha256,proto3,oneof"`
}

func (*InvitationV1_Aes256GcmHkdfSha256) isInvitationV1_Encryption() {}

// Sealed Invitation V1 Header
// Header carries information that is unencrypted, thus readable by the network
// it is however authenticated as associated data with the AEAD scheme used
// to encrypt the invitation body, thus providing tamper evidence.
type SealedInvitationHeaderV1 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sender    *SignedPublicKeyBundle `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	Recipient *SignedPublicKeyBundle `protobuf:"bytes,2,opt,name=recipient,proto3" json:"recipient,omitempty"`
	CreatedNs uint64                 `protobuf:"varint,3,opt,name=created_ns,json=createdNs,proto3" json:"created_ns,omitempty"`
}

func (x *SealedInvitationHeaderV1) Reset() {
	*x = SealedInvitationHeaderV1{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_contents_invitation_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SealedInvitationHeaderV1) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SealedInvitationHeaderV1) ProtoMessage() {}

func (x *SealedInvitationHeaderV1) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_invitation_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SealedInvitationHeaderV1.ProtoReflect.Descriptor instead.
func (*SealedInvitationHeaderV1) Descriptor() ([]byte, []int) {
	return file_message_contents_invitation_proto_rawDescGZIP(), []int{1}
}

func (x *SealedInvitationHeaderV1) GetSender() *SignedPublicKeyBundle {
	if x != nil {
		return x.Sender
	}
	return nil
}

func (x *SealedInvitationHeaderV1) GetRecipient() *SignedPublicKeyBundle {
	if x != nil {
		return x.Recipient
	}
	return nil
}

func (x *SealedInvitationHeaderV1) GetCreatedNs() uint64 {
	if x != nil {
		return x.CreatedNs
	}
	return 0
}

// Sealed Invitation V1
// Invitation encrypted with key material derived from the sender's and
// recipient's public key bundles using simplified X3DH where
// the sender's ephemeral key is replaced with sender's pre-key.
type SealedInvitationV1 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// encoded SealedInvitationHeaderV1 used as associated data for Ciphertext
	HeaderBytes []byte `protobuf:"bytes,1,opt,name=header_bytes,json=headerBytes,proto3" json:"header_bytes,omitempty"`
	// Ciphertext.payload MUST contain encrypted InvitationV1.
	Ciphertext *Ciphertext `protobuf:"bytes,2,opt,name=ciphertext,proto3" json:"ciphertext,omitempty"`
}

func (x *SealedInvitationV1) Reset() {
	*x = SealedInvitationV1{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_contents_invitation_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SealedInvitationV1) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SealedInvitationV1) ProtoMessage() {}

func (x *SealedInvitationV1) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_invitation_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SealedInvitationV1.ProtoReflect.Descriptor instead.
func (*SealedInvitationV1) Descriptor() ([]byte, []int) {
	return file_message_contents_invitation_proto_rawDescGZIP(), []int{2}
}

func (x *SealedInvitationV1) GetHeaderBytes() []byte {
	if x != nil {
		return x.HeaderBytes
	}
	return nil
}

func (x *SealedInvitationV1) GetCiphertext() *Ciphertext {
	if x != nil {
		return x.Ciphertext
	}
	return nil
}

// Versioned Sealed Invitation
type SealedInvitation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Version:
	//
	//	*SealedInvitation_V1
	Version isSealedInvitation_Version `protobuf_oneof:"version"`
}

func (x *SealedInvitation) Reset() {
	*x = SealedInvitation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_contents_invitation_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SealedInvitation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SealedInvitation) ProtoMessage() {}

func (x *SealedInvitation) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_invitation_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SealedInvitation.ProtoReflect.Descriptor instead.
func (*SealedInvitation) Descriptor() ([]byte, []int) {
	return file_message_contents_invitation_proto_rawDescGZIP(), []int{3}
}

func (m *SealedInvitation) GetVersion() isSealedInvitation_Version {
	if m != nil {
		return m.Version
	}
	return nil
}

func (x *SealedInvitation) GetV1() *SealedInvitationV1 {
	if x, ok := x.GetVersion().(*SealedInvitation_V1); ok {
		return x.V1
	}
	return nil
}

type isSealedInvitation_Version interface {
	isSealedInvitation_Version()
}

type SealedInvitation_V1 struct {
	V1 *SealedInvitationV1 `protobuf:"bytes,1,opt,name=v1,proto3,oneof"`
}

func (*SealedInvitation_V1) isSealedInvitation_Version() {}

// Payload for user's consent proof to be set in the invitation
// Signifying the conversation should be preapproved for the user on receipt
type ConsentProofPayload struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// the user's signature in hex format
	Signature string `protobuf:"bytes,1,opt,name=signature,proto3" json:"signature,omitempty"`
	// approximate time when the user signed
	Timestamp uint64 `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	// version of the payload
	PayloadVersion ConsentProofPayloadVersion `protobuf:"varint,3,opt,name=payload_version,json=payloadVersion,proto3,enum=xmtp.message_contents.ConsentProofPayloadVersion" json:"payload_version,omitempty"`
}

func (x *ConsentProofPayload) Reset() {
	*x = ConsentProofPayload{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_contents_invitation_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConsentProofPayload) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConsentProofPayload) ProtoMessage() {}

func (x *ConsentProofPayload) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_invitation_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConsentProofPayload.ProtoReflect.Descriptor instead.
func (*ConsentProofPayload) Descriptor() ([]byte, []int) {
	return file_message_contents_invitation_proto_rawDescGZIP(), []int{4}
}

func (x *ConsentProofPayload) GetSignature() string {
	if x != nil {
		return x.Signature
	}
	return ""
}

func (x *ConsentProofPayload) GetTimestamp() uint64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *ConsentProofPayload) GetPayloadVersion() ConsentProofPayloadVersion {
	if x != nil {
		return x.PayloadVersion
	}
	return ConsentProofPayloadVersion_CONSENT_PROOF_PAYLOAD_VERSION_UNSPECIFIED
}

// Supported encryption schemes
// AES256-GCM-HKDF-SHA256
type InvitationV1_Aes256GcmHkdfsha256 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	KeyMaterial []byte `protobuf:"bytes,1,opt,name=key_material,json=keyMaterial,proto3" json:"key_material,omitempty"` // randomly generated key material (32 bytes)
}

func (x *InvitationV1_Aes256GcmHkdfsha256) Reset() {
	*x = InvitationV1_Aes256GcmHkdfsha256{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_contents_invitation_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InvitationV1_Aes256GcmHkdfsha256) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InvitationV1_Aes256GcmHkdfsha256) ProtoMessage() {}

func (x *InvitationV1_Aes256GcmHkdfsha256) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_invitation_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InvitationV1_Aes256GcmHkdfsha256.ProtoReflect.Descriptor instead.
func (*InvitationV1_Aes256GcmHkdfsha256) Descriptor() ([]byte, []int) {
	return file_message_contents_invitation_proto_rawDescGZIP(), []int{0, 0}
}

func (x *InvitationV1_Aes256GcmHkdfsha256) GetKeyMaterial() []byte {
	if x != nil {
		return x.KeyMaterial
	}
	return nil
}

// The context type
type InvitationV1_Context struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Expected to be a URI (ie xmtp.org/convo1)
	ConversationId string `protobuf:"bytes,1,opt,name=conversation_id,json=conversationId,proto3" json:"conversation_id,omitempty"`
	// Key value map of additional metadata that would be exposed to
	// application developers and could be used for filtering
	Metadata map[string]string `protobuf:"bytes,2,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *InvitationV1_Context) Reset() {
	*x = InvitationV1_Context{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_contents_invitation_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InvitationV1_Context) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InvitationV1_Context) ProtoMessage() {}

func (x *InvitationV1_Context) ProtoReflect() protoreflect.Message {
	mi := &file_message_contents_invitation_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InvitationV1_Context.ProtoReflect.Descriptor instead.
func (*InvitationV1_Context) Descriptor() ([]byte, []int) {
	return file_message_contents_invitation_proto_rawDescGZIP(), []int{0, 1}
}

func (x *InvitationV1_Context) GetConversationId() string {
	if x != nil {
		return x.ConversationId
	}
	return ""
}

func (x *InvitationV1_Context) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

var File_message_contents_invitation_proto protoreflect.FileDescriptor

var file_message_contents_invitation_proto_rawDesc = []byte{
	0x0a, 0x21, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x73, 0x2f, 0x69, 0x6e, 0x76, 0x69, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x15, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x1a, 0x21, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x63, 0x69, 0x70,
	0x68, 0x65, 0x72, 0x74, 0x65, 0x78, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x21, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2f,
	0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xbd, 0x04, 0x0a, 0x0c, 0x49, 0x6e, 0x76, 0x69, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x56,
	0x31, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x12, 0x45, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x78, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73,
	0x2e, 0x49, 0x6e, 0x76, 0x69, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x56, 0x31, 0x2e, 0x43, 0x6f,
	0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x12, 0x6e,
	0x0a, 0x16, 0x61, 0x65, 0x73, 0x32, 0x35, 0x36, 0x5f, 0x67, 0x63, 0x6d, 0x5f, 0x68, 0x6b, 0x64,
	0x66, 0x5f, 0x73, 0x68, 0x61, 0x32, 0x35, 0x36, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x37,
	0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x49, 0x6e, 0x76, 0x69, 0x74, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x56, 0x31, 0x2e, 0x41, 0x65, 0x73, 0x32, 0x35, 0x36, 0x67, 0x63, 0x6d, 0x48, 0x6b, 0x64,
	0x66, 0x73, 0x68, 0x61, 0x32, 0x35, 0x36, 0x48, 0x00, 0x52, 0x13, 0x61, 0x65, 0x73, 0x32, 0x35,
	0x36, 0x47, 0x63, 0x6d, 0x48, 0x6b, 0x64, 0x66, 0x53, 0x68, 0x61, 0x32, 0x35, 0x36, 0x12, 0x4f,
	0x0a, 0x0d, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x6f, 0x66, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x43, 0x6f,
	0x6e, 0x73, 0x65, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61,
	0x64, 0x52, 0x0c, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x1a,
	0x38, 0x0a, 0x13, 0x41, 0x65, 0x73, 0x32, 0x35, 0x36, 0x67, 0x63, 0x6d, 0x48, 0x6b, 0x64, 0x66,
	0x73, 0x68, 0x61, 0x32, 0x35, 0x36, 0x12, 0x21, 0x0a, 0x0c, 0x6b, 0x65, 0x79, 0x5f, 0x6d, 0x61,
	0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x6b, 0x65,
	0x79, 0x4d, 0x61, 0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x1a, 0xc6, 0x01, 0x0a, 0x07, 0x43, 0x6f,
	0x6e, 0x74, 0x65, 0x78, 0x74, 0x12, 0x27, 0x0a, 0x0f, 0x63, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e,
	0x63, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x55,
	0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x39, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x49, 0x6e, 0x76, 0x69, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x56, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x2e, 0x4d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x6d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x1a, 0x3b, 0x0a, 0x0d, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x42, 0x0c, 0x0a, 0x0a, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x22, 0xcb, 0x01, 0x0a, 0x18, 0x53, 0x65, 0x61, 0x6c, 0x65, 0x64, 0x49, 0x6e, 0x76, 0x69, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x56, 0x31, 0x12, 0x44, 0x0a,
	0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2c, 0x2e,
	0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x50, 0x75, 0x62, 0x6c,
	0x69, 0x63, 0x4b, 0x65, 0x79, 0x42, 0x75, 0x6e, 0x64, 0x6c, 0x65, 0x52, 0x06, 0x73, 0x65, 0x6e,
	0x64, 0x65, 0x72, 0x12, 0x4a, 0x0a, 0x09, 0x72, 0x65, 0x63, 0x69, 0x70, 0x69, 0x65, 0x6e, 0x74,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x53,
	0x69, 0x67, 0x6e, 0x65, 0x64, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x42, 0x75,
	0x6e, 0x64, 0x6c, 0x65, 0x52, 0x09, 0x72, 0x65, 0x63, 0x69, 0x70, 0x69, 0x65, 0x6e, 0x74, 0x12,
	0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x6e, 0x73, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x4e, 0x73, 0x22, 0x7a,
	0x0a, 0x12, 0x53, 0x65, 0x61, 0x6c, 0x65, 0x64, 0x49, 0x6e, 0x76, 0x69, 0x74, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x56, 0x31, 0x12, 0x21, 0x0a, 0x0c, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x5f, 0x62,
	0x79, 0x74, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x68, 0x65, 0x61, 0x64,
	0x65, 0x72, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x41, 0x0a, 0x0a, 0x63, 0x69, 0x70, 0x68, 0x65,
	0x72, 0x74, 0x65, 0x78, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x78, 0x6d,
	0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x73, 0x2e, 0x43, 0x69, 0x70, 0x68, 0x65, 0x72, 0x74, 0x65, 0x78, 0x74, 0x52, 0x0a,
	0x63, 0x69, 0x70, 0x68, 0x65, 0x72, 0x74, 0x65, 0x78, 0x74, 0x22, 0x60, 0x0a, 0x10, 0x53, 0x65,
	0x61, 0x6c, 0x65, 0x64, 0x49, 0x6e, 0x76, 0x69, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x3b,
	0x0a, 0x02, 0x76, 0x31, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x78, 0x6d, 0x74,
	0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x73, 0x2e, 0x53, 0x65, 0x61, 0x6c, 0x65, 0x64, 0x49, 0x6e, 0x76, 0x69, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x56, 0x31, 0x48, 0x00, 0x52, 0x02, 0x76, 0x31, 0x42, 0x09, 0x0a, 0x07, 0x76,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x4a, 0x04, 0x08, 0x02, 0x10, 0x03, 0x22, 0xad, 0x01, 0x0a,
	0x13, 0x43, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x50, 0x61, 0x79,
	0x6c, 0x6f, 0x61, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x12, 0x5a, 0x0a, 0x0f, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x5f, 0x76, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x31, 0x2e, 0x78, 0x6d, 0x74, 0x70,
	0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x50, 0x61,
	0x79, 0x6c, 0x6f, 0x61, 0x64, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0e, 0x70, 0x61,
	0x79, 0x6c, 0x6f, 0x61, 0x64, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x2a, 0x70, 0x0a, 0x1a,
	0x43, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x50, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x2d, 0x0a, 0x29, 0x43, 0x4f,
	0x4e, 0x53, 0x45, 0x4e, 0x54, 0x5f, 0x50, 0x52, 0x4f, 0x4f, 0x46, 0x5f, 0x50, 0x41, 0x59, 0x4c,
	0x4f, 0x41, 0x44, 0x5f, 0x56, 0x45, 0x52, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x55, 0x4e, 0x53, 0x50,
	0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x23, 0x0a, 0x1f, 0x43, 0x4f, 0x4e,
	0x53, 0x45, 0x4e, 0x54, 0x5f, 0x50, 0x52, 0x4f, 0x4f, 0x46, 0x5f, 0x50, 0x41, 0x59, 0x4c, 0x4f,
	0x41, 0x44, 0x5f, 0x56, 0x45, 0x52, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x31, 0x10, 0x01, 0x42, 0xd6,
	0x01, 0x0a, 0x19, 0x63, 0x6f, 0x6d, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x42, 0x0f, 0x49, 0x6e,
	0x76, 0x69, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a,
	0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x78, 0x6d, 0x74, 0x70,
	0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2d, 0x6e, 0x6f, 0x64, 0x65, 0x2d, 0x67, 0x6f, 0x2f, 0x70, 0x6b,
	0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xa2, 0x02, 0x03, 0x58, 0x4d, 0x58, 0xaa, 0x02,
	0x14, 0x58, 0x6d, 0x74, 0x70, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x73, 0xca, 0x02, 0x14, 0x58, 0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xe2, 0x02, 0x20, 0x58,
	0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x73, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea,
	0x02, 0x15, 0x58, 0x6d, 0x74, 0x70, 0x3a, 0x3a, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_message_contents_invitation_proto_rawDescOnce sync.Once
	file_message_contents_invitation_proto_rawDescData = file_message_contents_invitation_proto_rawDesc
)

func file_message_contents_invitation_proto_rawDescGZIP() []byte {
	file_message_contents_invitation_proto_rawDescOnce.Do(func() {
		file_message_contents_invitation_proto_rawDescData = protoimpl.X.CompressGZIP(file_message_contents_invitation_proto_rawDescData)
	})
	return file_message_contents_invitation_proto_rawDescData
}

var file_message_contents_invitation_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_message_contents_invitation_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_message_contents_invitation_proto_goTypes = []any{
	(ConsentProofPayloadVersion)(0),          // 0: xmtp.message_contents.ConsentProofPayloadVersion
	(*InvitationV1)(nil),                     // 1: xmtp.message_contents.InvitationV1
	(*SealedInvitationHeaderV1)(nil),         // 2: xmtp.message_contents.SealedInvitationHeaderV1
	(*SealedInvitationV1)(nil),               // 3: xmtp.message_contents.SealedInvitationV1
	(*SealedInvitation)(nil),                 // 4: xmtp.message_contents.SealedInvitation
	(*ConsentProofPayload)(nil),              // 5: xmtp.message_contents.ConsentProofPayload
	(*InvitationV1_Aes256GcmHkdfsha256)(nil), // 6: xmtp.message_contents.InvitationV1.Aes256gcmHkdfsha256
	(*InvitationV1_Context)(nil),             // 7: xmtp.message_contents.InvitationV1.Context
	nil,                                      // 8: xmtp.message_contents.InvitationV1.Context.MetadataEntry
	(*SignedPublicKeyBundle)(nil),            // 9: xmtp.message_contents.SignedPublicKeyBundle
	(*Ciphertext)(nil),                       // 10: xmtp.message_contents.Ciphertext
}
var file_message_contents_invitation_proto_depIdxs = []int32{
	7,  // 0: xmtp.message_contents.InvitationV1.context:type_name -> xmtp.message_contents.InvitationV1.Context
	6,  // 1: xmtp.message_contents.InvitationV1.aes256_gcm_hkdf_sha256:type_name -> xmtp.message_contents.InvitationV1.Aes256gcmHkdfsha256
	5,  // 2: xmtp.message_contents.InvitationV1.consent_proof:type_name -> xmtp.message_contents.ConsentProofPayload
	9,  // 3: xmtp.message_contents.SealedInvitationHeaderV1.sender:type_name -> xmtp.message_contents.SignedPublicKeyBundle
	9,  // 4: xmtp.message_contents.SealedInvitationHeaderV1.recipient:type_name -> xmtp.message_contents.SignedPublicKeyBundle
	10, // 5: xmtp.message_contents.SealedInvitationV1.ciphertext:type_name -> xmtp.message_contents.Ciphertext
	3,  // 6: xmtp.message_contents.SealedInvitation.v1:type_name -> xmtp.message_contents.SealedInvitationV1
	0,  // 7: xmtp.message_contents.ConsentProofPayload.payload_version:type_name -> xmtp.message_contents.ConsentProofPayloadVersion
	8,  // 8: xmtp.message_contents.InvitationV1.Context.metadata:type_name -> xmtp.message_contents.InvitationV1.Context.MetadataEntry
	9,  // [9:9] is the sub-list for method output_type
	9,  // [9:9] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_message_contents_invitation_proto_init() }
func file_message_contents_invitation_proto_init() {
	if File_message_contents_invitation_proto != nil {
		return
	}
	file_message_contents_ciphertext_proto_init()
	file_message_contents_public_key_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_message_contents_invitation_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*InvitationV1); i {
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
		file_message_contents_invitation_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*SealedInvitationHeaderV1); i {
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
		file_message_contents_invitation_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*SealedInvitationV1); i {
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
		file_message_contents_invitation_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*SealedInvitation); i {
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
		file_message_contents_invitation_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*ConsentProofPayload); i {
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
		file_message_contents_invitation_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*InvitationV1_Aes256GcmHkdfsha256); i {
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
		file_message_contents_invitation_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*InvitationV1_Context); i {
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
	file_message_contents_invitation_proto_msgTypes[0].OneofWrappers = []any{
		(*InvitationV1_Aes256GcmHkdfSha256)(nil),
	}
	file_message_contents_invitation_proto_msgTypes[3].OneofWrappers = []any{
		(*SealedInvitation_V1)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_message_contents_invitation_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_message_contents_invitation_proto_goTypes,
		DependencyIndexes: file_message_contents_invitation_proto_depIdxs,
		EnumInfos:         file_message_contents_invitation_proto_enumTypes,
		MessageInfos:      file_message_contents_invitation_proto_msgTypes,
	}.Build()
	File_message_contents_invitation_proto = out.File
	file_message_contents_invitation_proto_rawDesc = nil
	file_message_contents_invitation_proto_goTypes = nil
	file_message_contents_invitation_proto_depIdxs = nil
}
