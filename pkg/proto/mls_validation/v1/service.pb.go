// Message API

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: mls_validation/v1/service.proto

package mls_validationv1

import (
	identity "github.com/xmtp/xmtp-node-go/pkg/proto/identity"
	v1 "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	associations "github.com/xmtp/xmtp-node-go/pkg/proto/identity/associations"
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

// Contains a batch of serialized Key Packages
type ValidateInboxIdKeyPackagesRequest struct {
	state         protoimpl.MessageState                          `protogen:"open.v1"`
	KeyPackages   []*ValidateInboxIdKeyPackagesRequest_KeyPackage `protobuf:"bytes,1,rep,name=key_packages,json=keyPackages,proto3" json:"key_packages,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ValidateInboxIdKeyPackagesRequest) Reset() {
	*x = ValidateInboxIdKeyPackagesRequest{}
	mi := &file_mls_validation_v1_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ValidateInboxIdKeyPackagesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateInboxIdKeyPackagesRequest) ProtoMessage() {}

func (x *ValidateInboxIdKeyPackagesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateInboxIdKeyPackagesRequest.ProtoReflect.Descriptor instead.
func (*ValidateInboxIdKeyPackagesRequest) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{0}
}

func (x *ValidateInboxIdKeyPackagesRequest) GetKeyPackages() []*ValidateInboxIdKeyPackagesRequest_KeyPackage {
	if x != nil {
		return x.KeyPackages
	}
	return nil
}

// Validates a Inbox-ID Key Package Type
type ValidateInboxIdKeyPackagesResponse struct {
	state         protoimpl.MessageState                         `protogen:"open.v1"`
	Responses     []*ValidateInboxIdKeyPackagesResponse_Response `protobuf:"bytes,1,rep,name=responses,proto3" json:"responses,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ValidateInboxIdKeyPackagesResponse) Reset() {
	*x = ValidateInboxIdKeyPackagesResponse{}
	mi := &file_mls_validation_v1_service_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ValidateInboxIdKeyPackagesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateInboxIdKeyPackagesResponse) ProtoMessage() {}

func (x *ValidateInboxIdKeyPackagesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateInboxIdKeyPackagesResponse.ProtoReflect.Descriptor instead.
func (*ValidateInboxIdKeyPackagesResponse) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{1}
}

func (x *ValidateInboxIdKeyPackagesResponse) GetResponses() []*ValidateInboxIdKeyPackagesResponse_Response {
	if x != nil {
		return x.Responses
	}
	return nil
}

// Contains a batch of serialized Key Packages
type ValidateKeyPackagesRequest struct {
	state         protoimpl.MessageState                   `protogen:"open.v1"`
	KeyPackages   []*ValidateKeyPackagesRequest_KeyPackage `protobuf:"bytes,1,rep,name=key_packages,json=keyPackages,proto3" json:"key_packages,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ValidateKeyPackagesRequest) Reset() {
	*x = ValidateKeyPackagesRequest{}
	mi := &file_mls_validation_v1_service_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ValidateKeyPackagesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateKeyPackagesRequest) ProtoMessage() {}

func (x *ValidateKeyPackagesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateKeyPackagesRequest.ProtoReflect.Descriptor instead.
func (*ValidateKeyPackagesRequest) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{2}
}

func (x *ValidateKeyPackagesRequest) GetKeyPackages() []*ValidateKeyPackagesRequest_KeyPackage {
	if x != nil {
		return x.KeyPackages
	}
	return nil
}

// Response to ValidateKeyPackagesRequest
type ValidateKeyPackagesResponse struct {
	state         protoimpl.MessageState                            `protogen:"open.v1"`
	Responses     []*ValidateKeyPackagesResponse_ValidationResponse `protobuf:"bytes,1,rep,name=responses,proto3" json:"responses,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ValidateKeyPackagesResponse) Reset() {
	*x = ValidateKeyPackagesResponse{}
	mi := &file_mls_validation_v1_service_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ValidateKeyPackagesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateKeyPackagesResponse) ProtoMessage() {}

func (x *ValidateKeyPackagesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateKeyPackagesResponse.ProtoReflect.Descriptor instead.
func (*ValidateKeyPackagesResponse) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{3}
}

func (x *ValidateKeyPackagesResponse) GetResponses() []*ValidateKeyPackagesResponse_ValidationResponse {
	if x != nil {
		return x.Responses
	}
	return nil
}

// Contains a batch of serialized Group Messages
type ValidateGroupMessagesRequest struct {
	state         protoimpl.MessageState                       `protogen:"open.v1"`
	GroupMessages []*ValidateGroupMessagesRequest_GroupMessage `protobuf:"bytes,1,rep,name=group_messages,json=groupMessages,proto3" json:"group_messages,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ValidateGroupMessagesRequest) Reset() {
	*x = ValidateGroupMessagesRequest{}
	mi := &file_mls_validation_v1_service_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ValidateGroupMessagesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateGroupMessagesRequest) ProtoMessage() {}

func (x *ValidateGroupMessagesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateGroupMessagesRequest.ProtoReflect.Descriptor instead.
func (*ValidateGroupMessagesRequest) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{4}
}

func (x *ValidateGroupMessagesRequest) GetGroupMessages() []*ValidateGroupMessagesRequest_GroupMessage {
	if x != nil {
		return x.GroupMessages
	}
	return nil
}

// Response to ValidateGroupMessagesRequest
type ValidateGroupMessagesResponse struct {
	state         protoimpl.MessageState                              `protogen:"open.v1"`
	Responses     []*ValidateGroupMessagesResponse_ValidationResponse `protobuf:"bytes,1,rep,name=responses,proto3" json:"responses,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ValidateGroupMessagesResponse) Reset() {
	*x = ValidateGroupMessagesResponse{}
	mi := &file_mls_validation_v1_service_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ValidateGroupMessagesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateGroupMessagesResponse) ProtoMessage() {}

func (x *ValidateGroupMessagesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateGroupMessagesResponse.ProtoReflect.Descriptor instead.
func (*ValidateGroupMessagesResponse) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{5}
}

func (x *ValidateGroupMessagesResponse) GetResponses() []*ValidateGroupMessagesResponse_ValidationResponse {
	if x != nil {
		return x.Responses
	}
	return nil
}

// Request to get a final association state for identity updates
type GetAssociationStateRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// List of identity updates
	OldUpdates    []*associations.IdentityUpdate `protobuf:"bytes,1,rep,name=old_updates,json=oldUpdates,proto3" json:"old_updates,omitempty"`
	NewUpdates    []*associations.IdentityUpdate `protobuf:"bytes,2,rep,name=new_updates,json=newUpdates,proto3" json:"new_updates,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetAssociationStateRequest) Reset() {
	*x = GetAssociationStateRequest{}
	mi := &file_mls_validation_v1_service_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetAssociationStateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetAssociationStateRequest) ProtoMessage() {}

func (x *GetAssociationStateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetAssociationStateRequest.ProtoReflect.Descriptor instead.
func (*GetAssociationStateRequest) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{6}
}

func (x *GetAssociationStateRequest) GetOldUpdates() []*associations.IdentityUpdate {
	if x != nil {
		return x.OldUpdates
	}
	return nil
}

func (x *GetAssociationStateRequest) GetNewUpdates() []*associations.IdentityUpdate {
	if x != nil {
		return x.NewUpdates
	}
	return nil
}

// Response to GetAssociationStateRequest, containing the final association state
// for an InboxID
type GetAssociationStateResponse struct {
	state            protoimpl.MessageState             `protogen:"open.v1"`
	AssociationState *associations.AssociationState     `protobuf:"bytes,1,opt,name=association_state,json=associationState,proto3" json:"association_state,omitempty"`
	StateDiff        *associations.AssociationStateDiff `protobuf:"bytes,2,opt,name=state_diff,json=stateDiff,proto3" json:"state_diff,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *GetAssociationStateResponse) Reset() {
	*x = GetAssociationStateResponse{}
	mi := &file_mls_validation_v1_service_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetAssociationStateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetAssociationStateResponse) ProtoMessage() {}

func (x *GetAssociationStateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetAssociationStateResponse.ProtoReflect.Descriptor instead.
func (*GetAssociationStateResponse) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{7}
}

func (x *GetAssociationStateResponse) GetAssociationState() *associations.AssociationState {
	if x != nil {
		return x.AssociationState
	}
	return nil
}

func (x *GetAssociationStateResponse) GetStateDiff() *associations.AssociationStateDiff {
	if x != nil {
		return x.StateDiff
	}
	return nil
}

// Wrapper for each key package
type ValidateInboxIdKeyPackagesRequest_KeyPackage struct {
	state                        protoimpl.MessageState `protogen:"open.v1"`
	KeyPackageBytesTlsSerialized []byte                 `protobuf:"bytes,1,opt,name=key_package_bytes_tls_serialized,json=keyPackageBytesTlsSerialized,proto3" json:"key_package_bytes_tls_serialized,omitempty"`
	IsInboxIdCredential          bool                   `protobuf:"varint,2,opt,name=is_inbox_id_credential,json=isInboxIdCredential,proto3" json:"is_inbox_id_credential,omitempty"`
	unknownFields                protoimpl.UnknownFields
	sizeCache                    protoimpl.SizeCache
}

func (x *ValidateInboxIdKeyPackagesRequest_KeyPackage) Reset() {
	*x = ValidateInboxIdKeyPackagesRequest_KeyPackage{}
	mi := &file_mls_validation_v1_service_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ValidateInboxIdKeyPackagesRequest_KeyPackage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateInboxIdKeyPackagesRequest_KeyPackage) ProtoMessage() {}

func (x *ValidateInboxIdKeyPackagesRequest_KeyPackage) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateInboxIdKeyPackagesRequest_KeyPackage.ProtoReflect.Descriptor instead.
func (*ValidateInboxIdKeyPackagesRequest_KeyPackage) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{0, 0}
}

func (x *ValidateInboxIdKeyPackagesRequest_KeyPackage) GetKeyPackageBytesTlsSerialized() []byte {
	if x != nil {
		return x.KeyPackageBytesTlsSerialized
	}
	return nil
}

func (x *ValidateInboxIdKeyPackagesRequest_KeyPackage) GetIsInboxIdCredential() bool {
	if x != nil {
		return x.IsInboxIdCredential
	}
	return false
}

// one response corresponding to information about one key package
type ValidateInboxIdKeyPackagesResponse_Response struct {
	state                 protoimpl.MessageState  `protogen:"open.v1"`
	IsOk                  bool                    `protobuf:"varint,1,opt,name=is_ok,json=isOk,proto3" json:"is_ok,omitempty"`
	ErrorMessage          string                  `protobuf:"bytes,2,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	Credential            *identity.MlsCredential `protobuf:"bytes,3,opt,name=credential,proto3" json:"credential,omitempty"`
	InstallationPublicKey []byte                  `protobuf:"bytes,4,opt,name=installation_public_key,json=installationPublicKey,proto3" json:"installation_public_key,omitempty"`
	Expiration            uint64                  `protobuf:"varint,5,opt,name=expiration,proto3" json:"expiration,omitempty"`
	unknownFields         protoimpl.UnknownFields
	sizeCache             protoimpl.SizeCache
}

func (x *ValidateInboxIdKeyPackagesResponse_Response) Reset() {
	*x = ValidateInboxIdKeyPackagesResponse_Response{}
	mi := &file_mls_validation_v1_service_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ValidateInboxIdKeyPackagesResponse_Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateInboxIdKeyPackagesResponse_Response) ProtoMessage() {}

func (x *ValidateInboxIdKeyPackagesResponse_Response) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateInboxIdKeyPackagesResponse_Response.ProtoReflect.Descriptor instead.
func (*ValidateInboxIdKeyPackagesResponse_Response) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{1, 0}
}

func (x *ValidateInboxIdKeyPackagesResponse_Response) GetIsOk() bool {
	if x != nil {
		return x.IsOk
	}
	return false
}

func (x *ValidateInboxIdKeyPackagesResponse_Response) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

func (x *ValidateInboxIdKeyPackagesResponse_Response) GetCredential() *identity.MlsCredential {
	if x != nil {
		return x.Credential
	}
	return nil
}

func (x *ValidateInboxIdKeyPackagesResponse_Response) GetInstallationPublicKey() []byte {
	if x != nil {
		return x.InstallationPublicKey
	}
	return nil
}

func (x *ValidateInboxIdKeyPackagesResponse_Response) GetExpiration() uint64 {
	if x != nil {
		return x.Expiration
	}
	return 0
}

// Wrapper for each key package
type ValidateKeyPackagesRequest_KeyPackage struct {
	state                        protoimpl.MessageState `protogen:"open.v1"`
	KeyPackageBytesTlsSerialized []byte                 `protobuf:"bytes,1,opt,name=key_package_bytes_tls_serialized,json=keyPackageBytesTlsSerialized,proto3" json:"key_package_bytes_tls_serialized,omitempty"`
	IsInboxIdCredential          bool                   `protobuf:"varint,2,opt,name=is_inbox_id_credential,json=isInboxIdCredential,proto3" json:"is_inbox_id_credential,omitempty"`
	unknownFields                protoimpl.UnknownFields
	sizeCache                    protoimpl.SizeCache
}

func (x *ValidateKeyPackagesRequest_KeyPackage) Reset() {
	*x = ValidateKeyPackagesRequest_KeyPackage{}
	mi := &file_mls_validation_v1_service_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ValidateKeyPackagesRequest_KeyPackage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateKeyPackagesRequest_KeyPackage) ProtoMessage() {}

func (x *ValidateKeyPackagesRequest_KeyPackage) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateKeyPackagesRequest_KeyPackage.ProtoReflect.Descriptor instead.
func (*ValidateKeyPackagesRequest_KeyPackage) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{2, 0}
}

func (x *ValidateKeyPackagesRequest_KeyPackage) GetKeyPackageBytesTlsSerialized() []byte {
	if x != nil {
		return x.KeyPackageBytesTlsSerialized
	}
	return nil
}

func (x *ValidateKeyPackagesRequest_KeyPackage) GetIsInboxIdCredential() bool {
	if x != nil {
		return x.IsInboxIdCredential
	}
	return false
}

// An individual response to one key package
type ValidateKeyPackagesResponse_ValidationResponse struct {
	state                   protoimpl.MessageState `protogen:"open.v1"`
	IsOk                    bool                   `protobuf:"varint,1,opt,name=is_ok,json=isOk,proto3" json:"is_ok,omitempty"`
	ErrorMessage            string                 `protobuf:"bytes,2,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	InstallationId          []byte                 `protobuf:"bytes,3,opt,name=installation_id,json=installationId,proto3" json:"installation_id,omitempty"`
	AccountAddress          string                 `protobuf:"bytes,4,opt,name=account_address,json=accountAddress,proto3" json:"account_address,omitempty"`
	CredentialIdentityBytes []byte                 `protobuf:"bytes,5,opt,name=credential_identity_bytes,json=credentialIdentityBytes,proto3" json:"credential_identity_bytes,omitempty"`
	Expiration              uint64                 `protobuf:"varint,6,opt,name=expiration,proto3" json:"expiration,omitempty"`
	unknownFields           protoimpl.UnknownFields
	sizeCache               protoimpl.SizeCache
}

func (x *ValidateKeyPackagesResponse_ValidationResponse) Reset() {
	*x = ValidateKeyPackagesResponse_ValidationResponse{}
	mi := &file_mls_validation_v1_service_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ValidateKeyPackagesResponse_ValidationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateKeyPackagesResponse_ValidationResponse) ProtoMessage() {}

func (x *ValidateKeyPackagesResponse_ValidationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateKeyPackagesResponse_ValidationResponse.ProtoReflect.Descriptor instead.
func (*ValidateKeyPackagesResponse_ValidationResponse) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{3, 0}
}

func (x *ValidateKeyPackagesResponse_ValidationResponse) GetIsOk() bool {
	if x != nil {
		return x.IsOk
	}
	return false
}

func (x *ValidateKeyPackagesResponse_ValidationResponse) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

func (x *ValidateKeyPackagesResponse_ValidationResponse) GetInstallationId() []byte {
	if x != nil {
		return x.InstallationId
	}
	return nil
}

func (x *ValidateKeyPackagesResponse_ValidationResponse) GetAccountAddress() string {
	if x != nil {
		return x.AccountAddress
	}
	return ""
}

func (x *ValidateKeyPackagesResponse_ValidationResponse) GetCredentialIdentityBytes() []byte {
	if x != nil {
		return x.CredentialIdentityBytes
	}
	return nil
}

func (x *ValidateKeyPackagesResponse_ValidationResponse) GetExpiration() uint64 {
	if x != nil {
		return x.Expiration
	}
	return 0
}

// Wrapper for each message
type ValidateGroupMessagesRequest_GroupMessage struct {
	state                          protoimpl.MessageState `protogen:"open.v1"`
	GroupMessageBytesTlsSerialized []byte                 `protobuf:"bytes,1,opt,name=group_message_bytes_tls_serialized,json=groupMessageBytesTlsSerialized,proto3" json:"group_message_bytes_tls_serialized,omitempty"`
	unknownFields                  protoimpl.UnknownFields
	sizeCache                      protoimpl.SizeCache
}

func (x *ValidateGroupMessagesRequest_GroupMessage) Reset() {
	*x = ValidateGroupMessagesRequest_GroupMessage{}
	mi := &file_mls_validation_v1_service_proto_msgTypes[12]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ValidateGroupMessagesRequest_GroupMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateGroupMessagesRequest_GroupMessage) ProtoMessage() {}

func (x *ValidateGroupMessagesRequest_GroupMessage) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[12]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateGroupMessagesRequest_GroupMessage.ProtoReflect.Descriptor instead.
func (*ValidateGroupMessagesRequest_GroupMessage) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{4, 0}
}

func (x *ValidateGroupMessagesRequest_GroupMessage) GetGroupMessageBytesTlsSerialized() []byte {
	if x != nil {
		return x.GroupMessageBytesTlsSerialized
	}
	return nil
}

// An individual response to one message
type ValidateGroupMessagesResponse_ValidationResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	IsOk          bool                   `protobuf:"varint,1,opt,name=is_ok,json=isOk,proto3" json:"is_ok,omitempty"`
	ErrorMessage  string                 `protobuf:"bytes,2,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	GroupId       string                 `protobuf:"bytes,3,opt,name=group_id,json=groupId,proto3" json:"group_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ValidateGroupMessagesResponse_ValidationResponse) Reset() {
	*x = ValidateGroupMessagesResponse_ValidationResponse{}
	mi := &file_mls_validation_v1_service_proto_msgTypes[13]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ValidateGroupMessagesResponse_ValidationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateGroupMessagesResponse_ValidationResponse) ProtoMessage() {}

func (x *ValidateGroupMessagesResponse_ValidationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[13]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateGroupMessagesResponse_ValidationResponse.ProtoReflect.Descriptor instead.
func (*ValidateGroupMessagesResponse_ValidationResponse) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{5, 0}
}

func (x *ValidateGroupMessagesResponse_ValidationResponse) GetIsOk() bool {
	if x != nil {
		return x.IsOk
	}
	return false
}

func (x *ValidateGroupMessagesResponse_ValidationResponse) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

func (x *ValidateGroupMessagesResponse_ValidationResponse) GetGroupId() string {
	if x != nil {
		return x.GroupId
	}
	return ""
}

var File_mls_validation_v1_service_proto protoreflect.FileDescriptor

const file_mls_validation_v1_service_proto_rawDesc = "" +
	"\n" +
	"\x1fmls_validation/v1/service.proto\x12\x16xmtp.mls_validation.v1\x1a\x1eidentity/api/v1/identity.proto\x1a'identity/associations/association.proto\x1a\x19identity/credential.proto\"\x98\x02\n" +
	"!ValidateInboxIdKeyPackagesRequest\x12g\n" +
	"\fkey_packages\x18\x01 \x03(\v2D.xmtp.mls_validation.v1.ValidateInboxIdKeyPackagesRequest.KeyPackageR\vkeyPackages\x1a\x89\x01\n" +
	"\n" +
	"KeyPackage\x12F\n" +
	" key_package_bytes_tls_serialized\x18\x01 \x01(\fR\x1ckeyPackageBytesTlsSerialized\x123\n" +
	"\x16is_inbox_id_credential\x18\x02 \x01(\bR\x13isInboxIdCredential\"\xe4\x02\n" +
	"\"ValidateInboxIdKeyPackagesResponse\x12a\n" +
	"\tresponses\x18\x01 \x03(\v2C.xmtp.mls_validation.v1.ValidateInboxIdKeyPackagesResponse.ResponseR\tresponses\x1a\xda\x01\n" +
	"\bResponse\x12\x13\n" +
	"\x05is_ok\x18\x01 \x01(\bR\x04isOk\x12#\n" +
	"\rerror_message\x18\x02 \x01(\tR\ferrorMessage\x12<\n" +
	"\n" +
	"credential\x18\x03 \x01(\v2\x1c.xmtp.identity.MlsCredentialR\n" +
	"credential\x126\n" +
	"\x17installation_public_key\x18\x04 \x01(\fR\x15installationPublicKey\x12\x1e\n" +
	"\n" +
	"expiration\x18\x05 \x01(\x04R\n" +
	"expiration\"\x8a\x02\n" +
	"\x1aValidateKeyPackagesRequest\x12`\n" +
	"\fkey_packages\x18\x01 \x03(\v2=.xmtp.mls_validation.v1.ValidateKeyPackagesRequest.KeyPackageR\vkeyPackages\x1a\x89\x01\n" +
	"\n" +
	"KeyPackage\x12F\n" +
	" key_package_bytes_tls_serialized\x18\x01 \x01(\fR\x1ckeyPackageBytesTlsSerialized\x123\n" +
	"\x16is_inbox_id_credential\x18\x02 \x01(\bR\x13isInboxIdCredential\"\x82\x03\n" +
	"\x1bValidateKeyPackagesResponse\x12d\n" +
	"\tresponses\x18\x01 \x03(\v2F.xmtp.mls_validation.v1.ValidateKeyPackagesResponse.ValidationResponseR\tresponses\x1a\xfc\x01\n" +
	"\x12ValidationResponse\x12\x13\n" +
	"\x05is_ok\x18\x01 \x01(\bR\x04isOk\x12#\n" +
	"\rerror_message\x18\x02 \x01(\tR\ferrorMessage\x12'\n" +
	"\x0finstallation_id\x18\x03 \x01(\fR\x0einstallationId\x12'\n" +
	"\x0faccount_address\x18\x04 \x01(\tR\x0eaccountAddress\x12:\n" +
	"\x19credential_identity_bytes\x18\x05 \x01(\fR\x17credentialIdentityBytes\x12\x1e\n" +
	"\n" +
	"expiration\x18\x06 \x01(\x04R\n" +
	"expiration\"\xe4\x01\n" +
	"\x1cValidateGroupMessagesRequest\x12h\n" +
	"\x0egroup_messages\x18\x01 \x03(\v2A.xmtp.mls_validation.v1.ValidateGroupMessagesRequest.GroupMessageR\rgroupMessages\x1aZ\n" +
	"\fGroupMessage\x12J\n" +
	"\"group_message_bytes_tls_serialized\x18\x01 \x01(\fR\x1egroupMessageBytesTlsSerialized\"\xf2\x01\n" +
	"\x1dValidateGroupMessagesResponse\x12f\n" +
	"\tresponses\x18\x01 \x03(\v2H.xmtp.mls_validation.v1.ValidateGroupMessagesResponse.ValidationResponseR\tresponses\x1ai\n" +
	"\x12ValidationResponse\x12\x13\n" +
	"\x05is_ok\x18\x01 \x01(\bR\x04isOk\x12#\n" +
	"\rerror_message\x18\x02 \x01(\tR\ferrorMessage\x12\x19\n" +
	"\bgroup_id\x18\x03 \x01(\tR\agroupId\"\xb6\x01\n" +
	"\x1aGetAssociationStateRequest\x12K\n" +
	"\vold_updates\x18\x01 \x03(\v2*.xmtp.identity.associations.IdentityUpdateR\n" +
	"oldUpdates\x12K\n" +
	"\vnew_updates\x18\x02 \x03(\v2*.xmtp.identity.associations.IdentityUpdateR\n" +
	"newUpdates\"\xc9\x01\n" +
	"\x1bGetAssociationStateResponse\x12Y\n" +
	"\x11association_state\x18\x01 \x01(\v2,.xmtp.identity.associations.AssociationStateR\x10associationState\x12O\n" +
	"\n" +
	"state_diff\x18\x02 \x01(\v20.xmtp.identity.associations.AssociationStateDiffR\tstateDiff2\xdb\x04\n" +
	"\rValidationApi\x12\x86\x01\n" +
	"\x15ValidateGroupMessages\x124.xmtp.mls_validation.v1.ValidateGroupMessagesRequest\x1a5.xmtp.mls_validation.v1.ValidateGroupMessagesResponse\"\x00\x12\x80\x01\n" +
	"\x13GetAssociationState\x122.xmtp.mls_validation.v1.GetAssociationStateRequest\x1a3.xmtp.mls_validation.v1.GetAssociationStateResponse\"\x00\x12\x8e\x01\n" +
	"\x1aValidateInboxIdKeyPackages\x122.xmtp.mls_validation.v1.ValidateKeyPackagesRequest\x1a:.xmtp.mls_validation.v1.ValidateInboxIdKeyPackagesResponse\"\x00\x12\xac\x01\n" +
	"#VerifySmartContractWalletSignatures\x12@.xmtp.identity.api.v1.VerifySmartContractWalletSignaturesRequest\x1aA.xmtp.identity.api.v1.VerifySmartContractWalletSignaturesResponse\"\x00B\xeb\x01\n" +
	"\x1acom.xmtp.mls_validation.v1B\fServiceProtoP\x01ZIgithub.com/xmtp/xmtp-node-go/pkg/proto/mls_validation/v1;mls_validationv1\xa2\x02\x03XMX\xaa\x02\x15Xmtp.MlsValidation.V1\xca\x02\x15Xmtp\\MlsValidation\\V1\xe2\x02!Xmtp\\MlsValidation\\V1\\GPBMetadata\xea\x02\x17Xmtp::MlsValidation::V1b\x06proto3"

var (
	file_mls_validation_v1_service_proto_rawDescOnce sync.Once
	file_mls_validation_v1_service_proto_rawDescData []byte
)

func file_mls_validation_v1_service_proto_rawDescGZIP() []byte {
	file_mls_validation_v1_service_proto_rawDescOnce.Do(func() {
		file_mls_validation_v1_service_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_mls_validation_v1_service_proto_rawDesc), len(file_mls_validation_v1_service_proto_rawDesc)))
	})
	return file_mls_validation_v1_service_proto_rawDescData
}

var file_mls_validation_v1_service_proto_msgTypes = make([]protoimpl.MessageInfo, 14)
var file_mls_validation_v1_service_proto_goTypes = []any{
	(*ValidateInboxIdKeyPackagesRequest)(nil),                // 0: xmtp.mls_validation.v1.ValidateInboxIdKeyPackagesRequest
	(*ValidateInboxIdKeyPackagesResponse)(nil),               // 1: xmtp.mls_validation.v1.ValidateInboxIdKeyPackagesResponse
	(*ValidateKeyPackagesRequest)(nil),                       // 2: xmtp.mls_validation.v1.ValidateKeyPackagesRequest
	(*ValidateKeyPackagesResponse)(nil),                      // 3: xmtp.mls_validation.v1.ValidateKeyPackagesResponse
	(*ValidateGroupMessagesRequest)(nil),                     // 4: xmtp.mls_validation.v1.ValidateGroupMessagesRequest
	(*ValidateGroupMessagesResponse)(nil),                    // 5: xmtp.mls_validation.v1.ValidateGroupMessagesResponse
	(*GetAssociationStateRequest)(nil),                       // 6: xmtp.mls_validation.v1.GetAssociationStateRequest
	(*GetAssociationStateResponse)(nil),                      // 7: xmtp.mls_validation.v1.GetAssociationStateResponse
	(*ValidateInboxIdKeyPackagesRequest_KeyPackage)(nil),     // 8: xmtp.mls_validation.v1.ValidateInboxIdKeyPackagesRequest.KeyPackage
	(*ValidateInboxIdKeyPackagesResponse_Response)(nil),      // 9: xmtp.mls_validation.v1.ValidateInboxIdKeyPackagesResponse.Response
	(*ValidateKeyPackagesRequest_KeyPackage)(nil),            // 10: xmtp.mls_validation.v1.ValidateKeyPackagesRequest.KeyPackage
	(*ValidateKeyPackagesResponse_ValidationResponse)(nil),   // 11: xmtp.mls_validation.v1.ValidateKeyPackagesResponse.ValidationResponse
	(*ValidateGroupMessagesRequest_GroupMessage)(nil),        // 12: xmtp.mls_validation.v1.ValidateGroupMessagesRequest.GroupMessage
	(*ValidateGroupMessagesResponse_ValidationResponse)(nil), // 13: xmtp.mls_validation.v1.ValidateGroupMessagesResponse.ValidationResponse
	(*associations.IdentityUpdate)(nil),                      // 14: xmtp.identity.associations.IdentityUpdate
	(*associations.AssociationState)(nil),                    // 15: xmtp.identity.associations.AssociationState
	(*associations.AssociationStateDiff)(nil),                // 16: xmtp.identity.associations.AssociationStateDiff
	(*identity.MlsCredential)(nil),                           // 17: xmtp.identity.MlsCredential
	(*v1.VerifySmartContractWalletSignaturesRequest)(nil),    // 18: xmtp.identity.api.v1.VerifySmartContractWalletSignaturesRequest
	(*v1.VerifySmartContractWalletSignaturesResponse)(nil),   // 19: xmtp.identity.api.v1.VerifySmartContractWalletSignaturesResponse
}
var file_mls_validation_v1_service_proto_depIdxs = []int32{
	8,  // 0: xmtp.mls_validation.v1.ValidateInboxIdKeyPackagesRequest.key_packages:type_name -> xmtp.mls_validation.v1.ValidateInboxIdKeyPackagesRequest.KeyPackage
	9,  // 1: xmtp.mls_validation.v1.ValidateInboxIdKeyPackagesResponse.responses:type_name -> xmtp.mls_validation.v1.ValidateInboxIdKeyPackagesResponse.Response
	10, // 2: xmtp.mls_validation.v1.ValidateKeyPackagesRequest.key_packages:type_name -> xmtp.mls_validation.v1.ValidateKeyPackagesRequest.KeyPackage
	11, // 3: xmtp.mls_validation.v1.ValidateKeyPackagesResponse.responses:type_name -> xmtp.mls_validation.v1.ValidateKeyPackagesResponse.ValidationResponse
	12, // 4: xmtp.mls_validation.v1.ValidateGroupMessagesRequest.group_messages:type_name -> xmtp.mls_validation.v1.ValidateGroupMessagesRequest.GroupMessage
	13, // 5: xmtp.mls_validation.v1.ValidateGroupMessagesResponse.responses:type_name -> xmtp.mls_validation.v1.ValidateGroupMessagesResponse.ValidationResponse
	14, // 6: xmtp.mls_validation.v1.GetAssociationStateRequest.old_updates:type_name -> xmtp.identity.associations.IdentityUpdate
	14, // 7: xmtp.mls_validation.v1.GetAssociationStateRequest.new_updates:type_name -> xmtp.identity.associations.IdentityUpdate
	15, // 8: xmtp.mls_validation.v1.GetAssociationStateResponse.association_state:type_name -> xmtp.identity.associations.AssociationState
	16, // 9: xmtp.mls_validation.v1.GetAssociationStateResponse.state_diff:type_name -> xmtp.identity.associations.AssociationStateDiff
	17, // 10: xmtp.mls_validation.v1.ValidateInboxIdKeyPackagesResponse.Response.credential:type_name -> xmtp.identity.MlsCredential
	4,  // 11: xmtp.mls_validation.v1.ValidationApi.ValidateGroupMessages:input_type -> xmtp.mls_validation.v1.ValidateGroupMessagesRequest
	6,  // 12: xmtp.mls_validation.v1.ValidationApi.GetAssociationState:input_type -> xmtp.mls_validation.v1.GetAssociationStateRequest
	2,  // 13: xmtp.mls_validation.v1.ValidationApi.ValidateInboxIdKeyPackages:input_type -> xmtp.mls_validation.v1.ValidateKeyPackagesRequest
	18, // 14: xmtp.mls_validation.v1.ValidationApi.VerifySmartContractWalletSignatures:input_type -> xmtp.identity.api.v1.VerifySmartContractWalletSignaturesRequest
	5,  // 15: xmtp.mls_validation.v1.ValidationApi.ValidateGroupMessages:output_type -> xmtp.mls_validation.v1.ValidateGroupMessagesResponse
	7,  // 16: xmtp.mls_validation.v1.ValidationApi.GetAssociationState:output_type -> xmtp.mls_validation.v1.GetAssociationStateResponse
	1,  // 17: xmtp.mls_validation.v1.ValidationApi.ValidateInboxIdKeyPackages:output_type -> xmtp.mls_validation.v1.ValidateInboxIdKeyPackagesResponse
	19, // 18: xmtp.mls_validation.v1.ValidationApi.VerifySmartContractWalletSignatures:output_type -> xmtp.identity.api.v1.VerifySmartContractWalletSignaturesResponse
	15, // [15:19] is the sub-list for method output_type
	11, // [11:15] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_mls_validation_v1_service_proto_init() }
func file_mls_validation_v1_service_proto_init() {
	if File_mls_validation_v1_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_mls_validation_v1_service_proto_rawDesc), len(file_mls_validation_v1_service_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   14,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_mls_validation_v1_service_proto_goTypes,
		DependencyIndexes: file_mls_validation_v1_service_proto_depIdxs,
		MessageInfos:      file_mls_validation_v1_service_proto_msgTypes,
	}.Build()
	File_mls_validation_v1_service_proto = out.File
	file_mls_validation_v1_service_proto_goTypes = nil
	file_mls_validation_v1_service_proto_depIdxs = nil
}
