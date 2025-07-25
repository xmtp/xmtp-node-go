// Message API

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: identity/api/v1/identity.proto

package apiv1

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	associations "github.com/xmtp/xmtp-node-go/pkg/proto/identity/associations"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type VerifySmartContractWalletSignaturesRequest struct {
	state         protoimpl.MessageState                                `protogen:"open.v1"`
	Signatures    []*VerifySmartContractWalletSignatureRequestSignature `protobuf:"bytes,1,rep,name=signatures,proto3" json:"signatures,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *VerifySmartContractWalletSignaturesRequest) Reset() {
	*x = VerifySmartContractWalletSignaturesRequest{}
	mi := &file_identity_api_v1_identity_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *VerifySmartContractWalletSignaturesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VerifySmartContractWalletSignaturesRequest) ProtoMessage() {}

func (x *VerifySmartContractWalletSignaturesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_identity_api_v1_identity_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VerifySmartContractWalletSignaturesRequest.ProtoReflect.Descriptor instead.
func (*VerifySmartContractWalletSignaturesRequest) Descriptor() ([]byte, []int) {
	return file_identity_api_v1_identity_proto_rawDescGZIP(), []int{0}
}

func (x *VerifySmartContractWalletSignaturesRequest) GetSignatures() []*VerifySmartContractWalletSignatureRequestSignature {
	if x != nil {
		return x.Signatures
	}
	return nil
}

type VerifySmartContractWalletSignatureRequestSignature struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// CAIP-10 string
	// https://github.com/ChainAgnostic/CAIPs/blob/main/CAIPs/caip-10.md
	AccountId string `protobuf:"bytes,1,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty"`
	// Specify the block number to verify the signature against
	BlockNumber *uint64 `protobuf:"varint,2,opt,name=block_number,json=blockNumber,proto3,oneof" json:"block_number,omitempty"`
	// The signature bytes
	Signature     []byte `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
	Hash          []byte `protobuf:"bytes,4,opt,name=hash,proto3" json:"hash,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *VerifySmartContractWalletSignatureRequestSignature) Reset() {
	*x = VerifySmartContractWalletSignatureRequestSignature{}
	mi := &file_identity_api_v1_identity_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *VerifySmartContractWalletSignatureRequestSignature) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VerifySmartContractWalletSignatureRequestSignature) ProtoMessage() {}

func (x *VerifySmartContractWalletSignatureRequestSignature) ProtoReflect() protoreflect.Message {
	mi := &file_identity_api_v1_identity_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VerifySmartContractWalletSignatureRequestSignature.ProtoReflect.Descriptor instead.
func (*VerifySmartContractWalletSignatureRequestSignature) Descriptor() ([]byte, []int) {
	return file_identity_api_v1_identity_proto_rawDescGZIP(), []int{1}
}

func (x *VerifySmartContractWalletSignatureRequestSignature) GetAccountId() string {
	if x != nil {
		return x.AccountId
	}
	return ""
}

func (x *VerifySmartContractWalletSignatureRequestSignature) GetBlockNumber() uint64 {
	if x != nil && x.BlockNumber != nil {
		return *x.BlockNumber
	}
	return 0
}

func (x *VerifySmartContractWalletSignatureRequestSignature) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *VerifySmartContractWalletSignatureRequestSignature) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

type VerifySmartContractWalletSignaturesResponse struct {
	state         protoimpl.MessageState                                            `protogen:"open.v1"`
	Responses     []*VerifySmartContractWalletSignaturesResponse_ValidationResponse `protobuf:"bytes,1,rep,name=responses,proto3" json:"responses,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *VerifySmartContractWalletSignaturesResponse) Reset() {
	*x = VerifySmartContractWalletSignaturesResponse{}
	mi := &file_identity_api_v1_identity_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *VerifySmartContractWalletSignaturesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VerifySmartContractWalletSignaturesResponse) ProtoMessage() {}

func (x *VerifySmartContractWalletSignaturesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_identity_api_v1_identity_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VerifySmartContractWalletSignaturesResponse.ProtoReflect.Descriptor instead.
func (*VerifySmartContractWalletSignaturesResponse) Descriptor() ([]byte, []int) {
	return file_identity_api_v1_identity_proto_rawDescGZIP(), []int{2}
}

func (x *VerifySmartContractWalletSignaturesResponse) GetResponses() []*VerifySmartContractWalletSignaturesResponse_ValidationResponse {
	if x != nil {
		return x.Responses
	}
	return nil
}

// Publishes an identity update to the network
type PublishIdentityUpdateRequest struct {
	state          protoimpl.MessageState       `protogen:"open.v1"`
	IdentityUpdate *associations.IdentityUpdate `protobuf:"bytes,1,opt,name=identity_update,json=identityUpdate,proto3" json:"identity_update,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *PublishIdentityUpdateRequest) Reset() {
	*x = PublishIdentityUpdateRequest{}
	mi := &file_identity_api_v1_identity_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PublishIdentityUpdateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublishIdentityUpdateRequest) ProtoMessage() {}

func (x *PublishIdentityUpdateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_identity_api_v1_identity_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublishIdentityUpdateRequest.ProtoReflect.Descriptor instead.
func (*PublishIdentityUpdateRequest) Descriptor() ([]byte, []int) {
	return file_identity_api_v1_identity_proto_rawDescGZIP(), []int{3}
}

func (x *PublishIdentityUpdateRequest) GetIdentityUpdate() *associations.IdentityUpdate {
	if x != nil {
		return x.IdentityUpdate
	}
	return nil
}

// The response when an identity update is published
type PublishIdentityUpdateResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PublishIdentityUpdateResponse) Reset() {
	*x = PublishIdentityUpdateResponse{}
	mi := &file_identity_api_v1_identity_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PublishIdentityUpdateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublishIdentityUpdateResponse) ProtoMessage() {}

func (x *PublishIdentityUpdateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_identity_api_v1_identity_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublishIdentityUpdateResponse.ProtoReflect.Descriptor instead.
func (*PublishIdentityUpdateResponse) Descriptor() ([]byte, []int) {
	return file_identity_api_v1_identity_proto_rawDescGZIP(), []int{4}
}

// Get all updates for an identity since the specified time
type GetIdentityUpdatesRequest struct {
	state         protoimpl.MessageState               `protogen:"open.v1"`
	Requests      []*GetIdentityUpdatesRequest_Request `protobuf:"bytes,1,rep,name=requests,proto3" json:"requests,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetIdentityUpdatesRequest) Reset() {
	*x = GetIdentityUpdatesRequest{}
	mi := &file_identity_api_v1_identity_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetIdentityUpdatesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetIdentityUpdatesRequest) ProtoMessage() {}

func (x *GetIdentityUpdatesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_identity_api_v1_identity_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetIdentityUpdatesRequest.ProtoReflect.Descriptor instead.
func (*GetIdentityUpdatesRequest) Descriptor() ([]byte, []int) {
	return file_identity_api_v1_identity_proto_rawDescGZIP(), []int{5}
}

func (x *GetIdentityUpdatesRequest) GetRequests() []*GetIdentityUpdatesRequest_Request {
	if x != nil {
		return x.Requests
	}
	return nil
}

// Returns all log entries for the requested identities
type GetIdentityUpdatesResponse struct {
	state         protoimpl.MessageState                 `protogen:"open.v1"`
	Responses     []*GetIdentityUpdatesResponse_Response `protobuf:"bytes,1,rep,name=responses,proto3" json:"responses,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetIdentityUpdatesResponse) Reset() {
	*x = GetIdentityUpdatesResponse{}
	mi := &file_identity_api_v1_identity_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetIdentityUpdatesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetIdentityUpdatesResponse) ProtoMessage() {}

func (x *GetIdentityUpdatesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_identity_api_v1_identity_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetIdentityUpdatesResponse.ProtoReflect.Descriptor instead.
func (*GetIdentityUpdatesResponse) Descriptor() ([]byte, []int) {
	return file_identity_api_v1_identity_proto_rawDescGZIP(), []int{6}
}

func (x *GetIdentityUpdatesResponse) GetResponses() []*GetIdentityUpdatesResponse_Response {
	if x != nil {
		return x.Responses
	}
	return nil
}

// Request to retrieve the XIDs for the given addresses
type GetInboxIdsRequest struct {
	state         protoimpl.MessageState        `protogen:"open.v1"`
	Requests      []*GetInboxIdsRequest_Request `protobuf:"bytes,1,rep,name=requests,proto3" json:"requests,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetInboxIdsRequest) Reset() {
	*x = GetInboxIdsRequest{}
	mi := &file_identity_api_v1_identity_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetInboxIdsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetInboxIdsRequest) ProtoMessage() {}

func (x *GetInboxIdsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_identity_api_v1_identity_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetInboxIdsRequest.ProtoReflect.Descriptor instead.
func (*GetInboxIdsRequest) Descriptor() ([]byte, []int) {
	return file_identity_api_v1_identity_proto_rawDescGZIP(), []int{7}
}

func (x *GetInboxIdsRequest) GetRequests() []*GetInboxIdsRequest_Request {
	if x != nil {
		return x.Requests
	}
	return nil
}

// Response with the XIDs for the requested addresses
type GetInboxIdsResponse struct {
	state         protoimpl.MessageState          `protogen:"open.v1"`
	Responses     []*GetInboxIdsResponse_Response `protobuf:"bytes,1,rep,name=responses,proto3" json:"responses,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetInboxIdsResponse) Reset() {
	*x = GetInboxIdsResponse{}
	mi := &file_identity_api_v1_identity_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetInboxIdsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetInboxIdsResponse) ProtoMessage() {}

func (x *GetInboxIdsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_identity_api_v1_identity_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetInboxIdsResponse.ProtoReflect.Descriptor instead.
func (*GetInboxIdsResponse) Descriptor() ([]byte, []int) {
	return file_identity_api_v1_identity_proto_rawDescGZIP(), []int{8}
}

func (x *GetInboxIdsResponse) GetResponses() []*GetInboxIdsResponse_Response {
	if x != nil {
		return x.Responses
	}
	return nil
}

type VerifySmartContractWalletSignaturesResponse_ValidationResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	IsValid       bool                   `protobuf:"varint,1,opt,name=is_valid,json=isValid,proto3" json:"is_valid,omitempty"`
	BlockNumber   *uint64                `protobuf:"varint,2,opt,name=block_number,json=blockNumber,proto3,oneof" json:"block_number,omitempty"`
	Error         *string                `protobuf:"bytes,3,opt,name=error,proto3,oneof" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *VerifySmartContractWalletSignaturesResponse_ValidationResponse) Reset() {
	*x = VerifySmartContractWalletSignaturesResponse_ValidationResponse{}
	mi := &file_identity_api_v1_identity_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *VerifySmartContractWalletSignaturesResponse_ValidationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VerifySmartContractWalletSignaturesResponse_ValidationResponse) ProtoMessage() {}

func (x *VerifySmartContractWalletSignaturesResponse_ValidationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_identity_api_v1_identity_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VerifySmartContractWalletSignaturesResponse_ValidationResponse.ProtoReflect.Descriptor instead.
func (*VerifySmartContractWalletSignaturesResponse_ValidationResponse) Descriptor() ([]byte, []int) {
	return file_identity_api_v1_identity_proto_rawDescGZIP(), []int{2, 0}
}

func (x *VerifySmartContractWalletSignaturesResponse_ValidationResponse) GetIsValid() bool {
	if x != nil {
		return x.IsValid
	}
	return false
}

func (x *VerifySmartContractWalletSignaturesResponse_ValidationResponse) GetBlockNumber() uint64 {
	if x != nil && x.BlockNumber != nil {
		return *x.BlockNumber
	}
	return 0
}

func (x *VerifySmartContractWalletSignaturesResponse_ValidationResponse) GetError() string {
	if x != nil && x.Error != nil {
		return *x.Error
	}
	return ""
}

// Points to the last entry the client has received. The sequence_id should be
// set to 0 if the client has not received anything.
type GetIdentityUpdatesRequest_Request struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	InboxId       string                 `protobuf:"bytes,1,opt,name=inbox_id,json=inboxId,proto3" json:"inbox_id,omitempty"`
	SequenceId    uint64                 `protobuf:"varint,2,opt,name=sequence_id,json=sequenceId,proto3" json:"sequence_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetIdentityUpdatesRequest_Request) Reset() {
	*x = GetIdentityUpdatesRequest_Request{}
	mi := &file_identity_api_v1_identity_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetIdentityUpdatesRequest_Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetIdentityUpdatesRequest_Request) ProtoMessage() {}

func (x *GetIdentityUpdatesRequest_Request) ProtoReflect() protoreflect.Message {
	mi := &file_identity_api_v1_identity_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetIdentityUpdatesRequest_Request.ProtoReflect.Descriptor instead.
func (*GetIdentityUpdatesRequest_Request) Descriptor() ([]byte, []int) {
	return file_identity_api_v1_identity_proto_rawDescGZIP(), []int{5, 0}
}

func (x *GetIdentityUpdatesRequest_Request) GetInboxId() string {
	if x != nil {
		return x.InboxId
	}
	return ""
}

func (x *GetIdentityUpdatesRequest_Request) GetSequenceId() uint64 {
	if x != nil {
		return x.SequenceId
	}
	return 0
}

// A single entry in the XID log on the server.
type GetIdentityUpdatesResponse_IdentityUpdateLog struct {
	state             protoimpl.MessageState       `protogen:"open.v1"`
	SequenceId        uint64                       `protobuf:"varint,1,opt,name=sequence_id,json=sequenceId,proto3" json:"sequence_id,omitempty"`
	ServerTimestampNs uint64                       `protobuf:"varint,2,opt,name=server_timestamp_ns,json=serverTimestampNs,proto3" json:"server_timestamp_ns,omitempty"`
	Update            *associations.IdentityUpdate `protobuf:"bytes,3,opt,name=update,proto3" json:"update,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *GetIdentityUpdatesResponse_IdentityUpdateLog) Reset() {
	*x = GetIdentityUpdatesResponse_IdentityUpdateLog{}
	mi := &file_identity_api_v1_identity_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetIdentityUpdatesResponse_IdentityUpdateLog) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetIdentityUpdatesResponse_IdentityUpdateLog) ProtoMessage() {}

func (x *GetIdentityUpdatesResponse_IdentityUpdateLog) ProtoReflect() protoreflect.Message {
	mi := &file_identity_api_v1_identity_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetIdentityUpdatesResponse_IdentityUpdateLog.ProtoReflect.Descriptor instead.
func (*GetIdentityUpdatesResponse_IdentityUpdateLog) Descriptor() ([]byte, []int) {
	return file_identity_api_v1_identity_proto_rawDescGZIP(), []int{6, 0}
}

func (x *GetIdentityUpdatesResponse_IdentityUpdateLog) GetSequenceId() uint64 {
	if x != nil {
		return x.SequenceId
	}
	return 0
}

func (x *GetIdentityUpdatesResponse_IdentityUpdateLog) GetServerTimestampNs() uint64 {
	if x != nil {
		return x.ServerTimestampNs
	}
	return 0
}

func (x *GetIdentityUpdatesResponse_IdentityUpdateLog) GetUpdate() *associations.IdentityUpdate {
	if x != nil {
		return x.Update
	}
	return nil
}

// The update log for a single identity, starting after the last cursor
type GetIdentityUpdatesResponse_Response struct {
	state         protoimpl.MessageState                          `protogen:"open.v1"`
	InboxId       string                                          `protobuf:"bytes,1,opt,name=inbox_id,json=inboxId,proto3" json:"inbox_id,omitempty"`
	Updates       []*GetIdentityUpdatesResponse_IdentityUpdateLog `protobuf:"bytes,2,rep,name=updates,proto3" json:"updates,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetIdentityUpdatesResponse_Response) Reset() {
	*x = GetIdentityUpdatesResponse_Response{}
	mi := &file_identity_api_v1_identity_proto_msgTypes[12]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetIdentityUpdatesResponse_Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetIdentityUpdatesResponse_Response) ProtoMessage() {}

func (x *GetIdentityUpdatesResponse_Response) ProtoReflect() protoreflect.Message {
	mi := &file_identity_api_v1_identity_proto_msgTypes[12]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetIdentityUpdatesResponse_Response.ProtoReflect.Descriptor instead.
func (*GetIdentityUpdatesResponse_Response) Descriptor() ([]byte, []int) {
	return file_identity_api_v1_identity_proto_rawDescGZIP(), []int{6, 1}
}

func (x *GetIdentityUpdatesResponse_Response) GetInboxId() string {
	if x != nil {
		return x.InboxId
	}
	return ""
}

func (x *GetIdentityUpdatesResponse_Response) GetUpdates() []*GetIdentityUpdatesResponse_IdentityUpdateLog {
	if x != nil {
		return x.Updates
	}
	return nil
}

// A single request for a given address
type GetInboxIdsRequest_Request struct {
	state          protoimpl.MessageState      `protogen:"open.v1"`
	Identifier     string                      `protobuf:"bytes,1,opt,name=identifier,proto3" json:"identifier,omitempty"`
	IdentifierKind associations.IdentifierKind `protobuf:"varint,2,opt,name=identifier_kind,json=identifierKind,proto3,enum=xmtp.identity.associations.IdentifierKind" json:"identifier_kind,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *GetInboxIdsRequest_Request) Reset() {
	*x = GetInboxIdsRequest_Request{}
	mi := &file_identity_api_v1_identity_proto_msgTypes[13]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetInboxIdsRequest_Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetInboxIdsRequest_Request) ProtoMessage() {}

func (x *GetInboxIdsRequest_Request) ProtoReflect() protoreflect.Message {
	mi := &file_identity_api_v1_identity_proto_msgTypes[13]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetInboxIdsRequest_Request.ProtoReflect.Descriptor instead.
func (*GetInboxIdsRequest_Request) Descriptor() ([]byte, []int) {
	return file_identity_api_v1_identity_proto_rawDescGZIP(), []int{7, 0}
}

func (x *GetInboxIdsRequest_Request) GetIdentifier() string {
	if x != nil {
		return x.Identifier
	}
	return ""
}

func (x *GetInboxIdsRequest_Request) GetIdentifierKind() associations.IdentifierKind {
	if x != nil {
		return x.IdentifierKind
	}
	return associations.IdentifierKind(0)
}

// A single response for a given address
type GetInboxIdsResponse_Response struct {
	state          protoimpl.MessageState      `protogen:"open.v1"`
	Identifier     string                      `protobuf:"bytes,1,opt,name=identifier,proto3" json:"identifier,omitempty"`
	InboxId        *string                     `protobuf:"bytes,2,opt,name=inbox_id,json=inboxId,proto3,oneof" json:"inbox_id,omitempty"`
	IdentifierKind associations.IdentifierKind `protobuf:"varint,3,opt,name=identifier_kind,json=identifierKind,proto3,enum=xmtp.identity.associations.IdentifierKind" json:"identifier_kind,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *GetInboxIdsResponse_Response) Reset() {
	*x = GetInboxIdsResponse_Response{}
	mi := &file_identity_api_v1_identity_proto_msgTypes[14]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetInboxIdsResponse_Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetInboxIdsResponse_Response) ProtoMessage() {}

func (x *GetInboxIdsResponse_Response) ProtoReflect() protoreflect.Message {
	mi := &file_identity_api_v1_identity_proto_msgTypes[14]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetInboxIdsResponse_Response.ProtoReflect.Descriptor instead.
func (*GetInboxIdsResponse_Response) Descriptor() ([]byte, []int) {
	return file_identity_api_v1_identity_proto_rawDescGZIP(), []int{8, 0}
}

func (x *GetInboxIdsResponse_Response) GetIdentifier() string {
	if x != nil {
		return x.Identifier
	}
	return ""
}

func (x *GetInboxIdsResponse_Response) GetInboxId() string {
	if x != nil && x.InboxId != nil {
		return *x.InboxId
	}
	return ""
}

func (x *GetInboxIdsResponse_Response) GetIdentifierKind() associations.IdentifierKind {
	if x != nil {
		return x.IdentifierKind
	}
	return associations.IdentifierKind(0)
}

var File_identity_api_v1_identity_proto protoreflect.FileDescriptor

const file_identity_api_v1_identity_proto_rawDesc = "" +
	"\n" +
	"\x1eidentity/api/v1/identity.proto\x12\x14xmtp.identity.api.v1\x1a\x1cgoogle/api/annotations.proto\x1a'identity/associations/association.proto\x1a.protoc-gen-openapiv2/options/annotations.proto\"\x96\x01\n" +
	"*VerifySmartContractWalletSignaturesRequest\x12h\n" +
	"\n" +
	"signatures\x18\x01 \x03(\v2H.xmtp.identity.api.v1.VerifySmartContractWalletSignatureRequestSignatureR\n" +
	"signatures\"\xbe\x01\n" +
	"2VerifySmartContractWalletSignatureRequestSignature\x12\x1d\n" +
	"\n" +
	"account_id\x18\x01 \x01(\tR\taccountId\x12&\n" +
	"\fblock_number\x18\x02 \x01(\x04H\x00R\vblockNumber\x88\x01\x01\x12\x1c\n" +
	"\tsignature\x18\x03 \x01(\fR\tsignature\x12\x12\n" +
	"\x04hash\x18\x04 \x01(\fR\x04hashB\x0f\n" +
	"\r_block_number\"\xb1\x02\n" +
	"+VerifySmartContractWalletSignaturesResponse\x12r\n" +
	"\tresponses\x18\x01 \x03(\v2T.xmtp.identity.api.v1.VerifySmartContractWalletSignaturesResponse.ValidationResponseR\tresponses\x1a\x8d\x01\n" +
	"\x12ValidationResponse\x12\x19\n" +
	"\bis_valid\x18\x01 \x01(\bR\aisValid\x12&\n" +
	"\fblock_number\x18\x02 \x01(\x04H\x00R\vblockNumber\x88\x01\x01\x12\x19\n" +
	"\x05error\x18\x03 \x01(\tH\x01R\x05error\x88\x01\x01B\x0f\n" +
	"\r_block_numberB\b\n" +
	"\x06_error\"s\n" +
	"\x1cPublishIdentityUpdateRequest\x12S\n" +
	"\x0fidentity_update\x18\x01 \x01(\v2*.xmtp.identity.associations.IdentityUpdateR\x0eidentityUpdate\"\x1f\n" +
	"\x1dPublishIdentityUpdateResponse\"\xb7\x01\n" +
	"\x19GetIdentityUpdatesRequest\x12S\n" +
	"\brequests\x18\x01 \x03(\v27.xmtp.identity.api.v1.GetIdentityUpdatesRequest.RequestR\brequests\x1aE\n" +
	"\aRequest\x12\x19\n" +
	"\binbox_id\x18\x01 \x01(\tR\ainboxId\x12\x1f\n" +
	"\vsequence_id\x18\x02 \x01(\x04R\n" +
	"sequenceId\"\xa6\x03\n" +
	"\x1aGetIdentityUpdatesResponse\x12W\n" +
	"\tresponses\x18\x01 \x03(\v29.xmtp.identity.api.v1.GetIdentityUpdatesResponse.ResponseR\tresponses\x1a\xa8\x01\n" +
	"\x11IdentityUpdateLog\x12\x1f\n" +
	"\vsequence_id\x18\x01 \x01(\x04R\n" +
	"sequenceId\x12.\n" +
	"\x13server_timestamp_ns\x18\x02 \x01(\x04R\x11serverTimestampNs\x12B\n" +
	"\x06update\x18\x03 \x01(\v2*.xmtp.identity.associations.IdentityUpdateR\x06update\x1a\x83\x01\n" +
	"\bResponse\x12\x19\n" +
	"\binbox_id\x18\x01 \x01(\tR\ainboxId\x12\\\n" +
	"\aupdates\x18\x02 \x03(\v2B.xmtp.identity.api.v1.GetIdentityUpdatesResponse.IdentityUpdateLogR\aupdates\"\xe2\x01\n" +
	"\x12GetInboxIdsRequest\x12L\n" +
	"\brequests\x18\x01 \x03(\v20.xmtp.identity.api.v1.GetInboxIdsRequest.RequestR\brequests\x1a~\n" +
	"\aRequest\x12\x1e\n" +
	"\n" +
	"identifier\x18\x01 \x01(\tR\n" +
	"identifier\x12S\n" +
	"\x0fidentifier_kind\x18\x02 \x01(\x0e2*.xmtp.identity.associations.IdentifierKindR\x0eidentifierKind\"\x96\x02\n" +
	"\x13GetInboxIdsResponse\x12P\n" +
	"\tresponses\x18\x01 \x03(\v22.xmtp.identity.api.v1.GetInboxIdsResponse.ResponseR\tresponses\x1a\xac\x01\n" +
	"\bResponse\x12\x1e\n" +
	"\n" +
	"identifier\x18\x01 \x01(\tR\n" +
	"identifier\x12\x1e\n" +
	"\binbox_id\x18\x02 \x01(\tH\x00R\ainboxId\x88\x01\x01\x12S\n" +
	"\x0fidentifier_kind\x18\x03 \x01(\x0e2*.xmtp.identity.associations.IdentifierKindR\x0eidentifierKindB\v\n" +
	"\t_inbox_id2\xe3\x05\n" +
	"\vIdentityApi\x12\xb1\x01\n" +
	"\x15PublishIdentityUpdate\x122.xmtp.identity.api.v1.PublishIdentityUpdateRequest\x1a3.xmtp.identity.api.v1.PublishIdentityUpdateResponse\"/\x82\xd3\xe4\x93\x02):\x01*\"$/identity/v1/publish-identity-update\x12\xa5\x01\n" +
	"\x12GetIdentityUpdates\x12/.xmtp.identity.api.v1.GetIdentityUpdatesRequest\x1a0.xmtp.identity.api.v1.GetIdentityUpdatesResponse\",\x82\xd3\xe4\x93\x02&:\x01*\"!/identity/v1/get-identity-updates\x12\x89\x01\n" +
	"\vGetInboxIds\x12(.xmtp.identity.api.v1.GetInboxIdsRequest\x1a).xmtp.identity.api.v1.GetInboxIdsResponse\"%\x82\xd3\xe4\x93\x02\x1f:\x01*\"\x1a/identity/v1/get-inbox-ids\x12\xeb\x01\n" +
	"#VerifySmartContractWalletSignatures\x12@.xmtp.identity.api.v1.VerifySmartContractWalletSignaturesRequest\x1aA.xmtp.identity.api.v1.VerifySmartContractWalletSignaturesResponse\"?\x82\xd3\xe4\x93\x029:\x01*\"4/identity/v1/verify-smart-contract-wallet-signaturesB\xf1\x01\x92A\x14\x12\x12\n" +
	"\vIdentityApi2\x031.0\n" +
	"\x18com.xmtp.identity.api.v1B\rIdentityProtoP\x01Z<github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1;apiv1\xa2\x02\x03XIA\xaa\x02\x14Xmtp.Identity.Api.V1\xca\x02\x14Xmtp\\Identity\\Api\\V1\xe2\x02 Xmtp\\Identity\\Api\\V1\\GPBMetadata\xea\x02\x17Xmtp::Identity::Api::V1b\x06proto3"

var (
	file_identity_api_v1_identity_proto_rawDescOnce sync.Once
	file_identity_api_v1_identity_proto_rawDescData []byte
)

func file_identity_api_v1_identity_proto_rawDescGZIP() []byte {
	file_identity_api_v1_identity_proto_rawDescOnce.Do(func() {
		file_identity_api_v1_identity_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_identity_api_v1_identity_proto_rawDesc), len(file_identity_api_v1_identity_proto_rawDesc)))
	})
	return file_identity_api_v1_identity_proto_rawDescData
}

var file_identity_api_v1_identity_proto_msgTypes = make([]protoimpl.MessageInfo, 15)
var file_identity_api_v1_identity_proto_goTypes = []any{
	(*VerifySmartContractWalletSignaturesRequest)(nil),                     // 0: xmtp.identity.api.v1.VerifySmartContractWalletSignaturesRequest
	(*VerifySmartContractWalletSignatureRequestSignature)(nil),             // 1: xmtp.identity.api.v1.VerifySmartContractWalletSignatureRequestSignature
	(*VerifySmartContractWalletSignaturesResponse)(nil),                    // 2: xmtp.identity.api.v1.VerifySmartContractWalletSignaturesResponse
	(*PublishIdentityUpdateRequest)(nil),                                   // 3: xmtp.identity.api.v1.PublishIdentityUpdateRequest
	(*PublishIdentityUpdateResponse)(nil),                                  // 4: xmtp.identity.api.v1.PublishIdentityUpdateResponse
	(*GetIdentityUpdatesRequest)(nil),                                      // 5: xmtp.identity.api.v1.GetIdentityUpdatesRequest
	(*GetIdentityUpdatesResponse)(nil),                                     // 6: xmtp.identity.api.v1.GetIdentityUpdatesResponse
	(*GetInboxIdsRequest)(nil),                                             // 7: xmtp.identity.api.v1.GetInboxIdsRequest
	(*GetInboxIdsResponse)(nil),                                            // 8: xmtp.identity.api.v1.GetInboxIdsResponse
	(*VerifySmartContractWalletSignaturesResponse_ValidationResponse)(nil), // 9: xmtp.identity.api.v1.VerifySmartContractWalletSignaturesResponse.ValidationResponse
	(*GetIdentityUpdatesRequest_Request)(nil),                              // 10: xmtp.identity.api.v1.GetIdentityUpdatesRequest.Request
	(*GetIdentityUpdatesResponse_IdentityUpdateLog)(nil),                   // 11: xmtp.identity.api.v1.GetIdentityUpdatesResponse.IdentityUpdateLog
	(*GetIdentityUpdatesResponse_Response)(nil),                            // 12: xmtp.identity.api.v1.GetIdentityUpdatesResponse.Response
	(*GetInboxIdsRequest_Request)(nil),                                     // 13: xmtp.identity.api.v1.GetInboxIdsRequest.Request
	(*GetInboxIdsResponse_Response)(nil),                                   // 14: xmtp.identity.api.v1.GetInboxIdsResponse.Response
	(*associations.IdentityUpdate)(nil),                                    // 15: xmtp.identity.associations.IdentityUpdate
	(associations.IdentifierKind)(0),                                       // 16: xmtp.identity.associations.IdentifierKind
}
var file_identity_api_v1_identity_proto_depIdxs = []int32{
	1,  // 0: xmtp.identity.api.v1.VerifySmartContractWalletSignaturesRequest.signatures:type_name -> xmtp.identity.api.v1.VerifySmartContractWalletSignatureRequestSignature
	9,  // 1: xmtp.identity.api.v1.VerifySmartContractWalletSignaturesResponse.responses:type_name -> xmtp.identity.api.v1.VerifySmartContractWalletSignaturesResponse.ValidationResponse
	15, // 2: xmtp.identity.api.v1.PublishIdentityUpdateRequest.identity_update:type_name -> xmtp.identity.associations.IdentityUpdate
	10, // 3: xmtp.identity.api.v1.GetIdentityUpdatesRequest.requests:type_name -> xmtp.identity.api.v1.GetIdentityUpdatesRequest.Request
	12, // 4: xmtp.identity.api.v1.GetIdentityUpdatesResponse.responses:type_name -> xmtp.identity.api.v1.GetIdentityUpdatesResponse.Response
	13, // 5: xmtp.identity.api.v1.GetInboxIdsRequest.requests:type_name -> xmtp.identity.api.v1.GetInboxIdsRequest.Request
	14, // 6: xmtp.identity.api.v1.GetInboxIdsResponse.responses:type_name -> xmtp.identity.api.v1.GetInboxIdsResponse.Response
	15, // 7: xmtp.identity.api.v1.GetIdentityUpdatesResponse.IdentityUpdateLog.update:type_name -> xmtp.identity.associations.IdentityUpdate
	11, // 8: xmtp.identity.api.v1.GetIdentityUpdatesResponse.Response.updates:type_name -> xmtp.identity.api.v1.GetIdentityUpdatesResponse.IdentityUpdateLog
	16, // 9: xmtp.identity.api.v1.GetInboxIdsRequest.Request.identifier_kind:type_name -> xmtp.identity.associations.IdentifierKind
	16, // 10: xmtp.identity.api.v1.GetInboxIdsResponse.Response.identifier_kind:type_name -> xmtp.identity.associations.IdentifierKind
	3,  // 11: xmtp.identity.api.v1.IdentityApi.PublishIdentityUpdate:input_type -> xmtp.identity.api.v1.PublishIdentityUpdateRequest
	5,  // 12: xmtp.identity.api.v1.IdentityApi.GetIdentityUpdates:input_type -> xmtp.identity.api.v1.GetIdentityUpdatesRequest
	7,  // 13: xmtp.identity.api.v1.IdentityApi.GetInboxIds:input_type -> xmtp.identity.api.v1.GetInboxIdsRequest
	0,  // 14: xmtp.identity.api.v1.IdentityApi.VerifySmartContractWalletSignatures:input_type -> xmtp.identity.api.v1.VerifySmartContractWalletSignaturesRequest
	4,  // 15: xmtp.identity.api.v1.IdentityApi.PublishIdentityUpdate:output_type -> xmtp.identity.api.v1.PublishIdentityUpdateResponse
	6,  // 16: xmtp.identity.api.v1.IdentityApi.GetIdentityUpdates:output_type -> xmtp.identity.api.v1.GetIdentityUpdatesResponse
	8,  // 17: xmtp.identity.api.v1.IdentityApi.GetInboxIds:output_type -> xmtp.identity.api.v1.GetInboxIdsResponse
	2,  // 18: xmtp.identity.api.v1.IdentityApi.VerifySmartContractWalletSignatures:output_type -> xmtp.identity.api.v1.VerifySmartContractWalletSignaturesResponse
	15, // [15:19] is the sub-list for method output_type
	11, // [11:15] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_identity_api_v1_identity_proto_init() }
func file_identity_api_v1_identity_proto_init() {
	if File_identity_api_v1_identity_proto != nil {
		return
	}
	file_identity_api_v1_identity_proto_msgTypes[1].OneofWrappers = []any{}
	file_identity_api_v1_identity_proto_msgTypes[9].OneofWrappers = []any{}
	file_identity_api_v1_identity_proto_msgTypes[14].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_identity_api_v1_identity_proto_rawDesc), len(file_identity_api_v1_identity_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   15,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_identity_api_v1_identity_proto_goTypes,
		DependencyIndexes: file_identity_api_v1_identity_proto_depIdxs,
		MessageInfos:      file_identity_api_v1_identity_proto_msgTypes,
	}.Build()
	File_identity_api_v1_identity_proto = out.File
	file_identity_api_v1_identity_proto_goTypes = nil
	file_identity_api_v1_identity_proto_depIdxs = nil
}
