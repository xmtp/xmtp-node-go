// Client authentication protocol

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        (unknown)
// source: message_api/v1/authn.proto

package message_apiv1

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

// Token is used by clients to prove to the nodes
// that they are serving a specific wallet.
type Token struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// identity key signed by a wallet
	IdentityKey *message_contents.PublicKey `protobuf:"bytes,1,opt,name=identity_key,json=identityKey,proto3" json:"identity_key,omitempty"`
	// encoded bytes of AuthData
	AuthDataBytes []byte `protobuf:"bytes,2,opt,name=auth_data_bytes,json=authDataBytes,proto3" json:"auth_data_bytes,omitempty"`
	// identity key signature of AuthData bytes
	AuthDataSignature *message_contents.Signature `protobuf:"bytes,3,opt,name=auth_data_signature,json=authDataSignature,proto3" json:"auth_data_signature,omitempty"`
}

func (x *Token) Reset() {
	*x = Token{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_api_v1_authn_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Token) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Token) ProtoMessage() {}

func (x *Token) ProtoReflect() protoreflect.Message {
	mi := &file_message_api_v1_authn_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Token.ProtoReflect.Descriptor instead.
func (*Token) Descriptor() ([]byte, []int) {
	return file_message_api_v1_authn_proto_rawDescGZIP(), []int{0}
}

func (x *Token) GetIdentityKey() *message_contents.PublicKey {
	if x != nil {
		return x.IdentityKey
	}
	return nil
}

func (x *Token) GetAuthDataBytes() []byte {
	if x != nil {
		return x.AuthDataBytes
	}
	return nil
}

func (x *Token) GetAuthDataSignature() *message_contents.Signature {
	if x != nil {
		return x.AuthDataSignature
	}
	return nil
}

// AuthData carries token parameters that are authenticated
// by the identity key signature.
// It is embedded in the Token structure as bytes
// so that the bytes don't need to be reconstructed
// to verify the token signature.
type AuthData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// address of the wallet
	WalletAddr string `protobuf:"bytes,1,opt,name=wallet_addr,json=walletAddr,proto3" json:"wallet_addr,omitempty"`
	// time when the token was generated/signed
	CreatedNs uint64 `protobuf:"varint,2,opt,name=created_ns,json=createdNs,proto3" json:"created_ns,omitempty"`
}

func (x *AuthData) Reset() {
	*x = AuthData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_api_v1_authn_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthData) ProtoMessage() {}

func (x *AuthData) ProtoReflect() protoreflect.Message {
	mi := &file_message_api_v1_authn_proto_msgTypes[1]
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
	return file_message_api_v1_authn_proto_rawDescGZIP(), []int{1}
}

func (x *AuthData) GetWalletAddr() string {
	if x != nil {
		return x.WalletAddr
	}
	return ""
}

func (x *AuthData) GetCreatedNs() uint64 {
	if x != nil {
		return x.CreatedNs
	}
	return 0
}

var File_message_api_v1_authn_proto protoreflect.FileDescriptor

var file_message_api_v1_authn_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31,
	0x2f, 0x61, 0x75, 0x74, 0x68, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x78, 0x6d,
	0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x76,
	0x31, 0x1a, 0x21, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x73, 0x2f, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc6, 0x01, 0x0a, 0x05, 0x54, 0x6f, 0x6b, 0x65, 0x6e,
	0x12, 0x43, 0x0a, 0x0c, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x5f, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x50,
	0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x52, 0x0b, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69,
	0x74, 0x79, 0x4b, 0x65, 0x79, 0x12, 0x26, 0x0a, 0x0f, 0x61, 0x75, 0x74, 0x68, 0x5f, 0x64, 0x61,
	0x74, 0x61, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0d,
	0x61, 0x75, 0x74, 0x68, 0x44, 0x61, 0x74, 0x61, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x50, 0x0a,
	0x13, 0x61, 0x75, 0x74, 0x68, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x5f, 0x73, 0x69, 0x67, 0x6e, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x78, 0x6d, 0x74,
	0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x73, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x52, 0x11, 0x61, 0x75,
	0x74, 0x68, 0x44, 0x61, 0x74, 0x61, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x22,
	0x4a, 0x0a, 0x08, 0x41, 0x75, 0x74, 0x68, 0x44, 0x61, 0x74, 0x61, 0x12, 0x1f, 0x0a, 0x0b, 0x77,
	0x61, 0x6c, 0x6c, 0x65, 0x74, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x41, 0x64, 0x64, 0x72, 0x12, 0x1d, 0x0a, 0x0a,
	0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x4e, 0x73, 0x42, 0xd4, 0x01, 0x0a, 0x17,
	0x63, 0x6f, 0x6d, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x5f, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x42, 0x0a, 0x41, 0x75, 0x74, 0x68, 0x6e, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x43, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2d, 0x6e, 0x6f, 0x64, 0x65,
	0x2d, 0x67, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x3b, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x5f, 0x61, 0x70, 0x69, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x58, 0x4d, 0x58,
	0xaa, 0x02, 0x12, 0x58, 0x6d, 0x74, 0x70, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x41,
	0x70, 0x69, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x12, 0x58, 0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x41, 0x70, 0x69, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x1e, 0x58, 0x6d, 0x74,
	0x70, 0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x41, 0x70, 0x69, 0x5c, 0x56, 0x31, 0x5c,
	0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x14, 0x58, 0x6d,
	0x74, 0x70, 0x3a, 0x3a, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x41, 0x70, 0x69, 0x3a, 0x3a,
	0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_message_api_v1_authn_proto_rawDescOnce sync.Once
	file_message_api_v1_authn_proto_rawDescData = file_message_api_v1_authn_proto_rawDesc
)

func file_message_api_v1_authn_proto_rawDescGZIP() []byte {
	file_message_api_v1_authn_proto_rawDescOnce.Do(func() {
		file_message_api_v1_authn_proto_rawDescData = protoimpl.X.CompressGZIP(file_message_api_v1_authn_proto_rawDescData)
	})
	return file_message_api_v1_authn_proto_rawDescData
}

var file_message_api_v1_authn_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_message_api_v1_authn_proto_goTypes = []interface{}{
	(*Token)(nil),                      // 0: xmtp.message_api.v1.Token
	(*AuthData)(nil),                   // 1: xmtp.message_api.v1.AuthData
	(*message_contents.PublicKey)(nil), // 2: xmtp.message_contents.PublicKey
	(*message_contents.Signature)(nil), // 3: xmtp.message_contents.Signature
}
var file_message_api_v1_authn_proto_depIdxs = []int32{
	2, // 0: xmtp.message_api.v1.Token.identity_key:type_name -> xmtp.message_contents.PublicKey
	3, // 1: xmtp.message_api.v1.Token.auth_data_signature:type_name -> xmtp.message_contents.Signature
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_message_api_v1_authn_proto_init() }
func file_message_api_v1_authn_proto_init() {
	if File_message_api_v1_authn_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_message_api_v1_authn_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Token); i {
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
		file_message_api_v1_authn_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
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
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_message_api_v1_authn_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_message_api_v1_authn_proto_goTypes,
		DependencyIndexes: file_message_api_v1_authn_proto_depIdxs,
		MessageInfos:      file_message_api_v1_authn_proto_msgTypes,
	}.Build()
	File_message_api_v1_authn_proto = out.File
	file_message_api_v1_authn_proto_rawDesc = nil
	file_message_api_v1_authn_proto_goTypes = nil
	file_message_api_v1_authn_proto_depIdxs = nil
}
