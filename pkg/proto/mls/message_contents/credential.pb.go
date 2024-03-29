// Credentials and revocations

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        (unknown)
// source: mls/message_contents/credential.proto

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

// A credential that can be used in MLS leaf nodes
type MlsCredential struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InstallationPublicKey []byte `protobuf:"bytes,1,opt,name=installation_public_key,json=installationPublicKey,proto3" json:"installation_public_key,omitempty"`
	// Types that are assignable to Association:
	//
	//	*MlsCredential_Eip_191
	Association isMlsCredential_Association `protobuf_oneof:"association"`
}

func (x *MlsCredential) Reset() {
	*x = MlsCredential{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_message_contents_credential_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MlsCredential) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MlsCredential) ProtoMessage() {}

func (x *MlsCredential) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_credential_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MlsCredential.ProtoReflect.Descriptor instead.
func (*MlsCredential) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_credential_proto_rawDescGZIP(), []int{0}
}

func (x *MlsCredential) GetInstallationPublicKey() []byte {
	if x != nil {
		return x.InstallationPublicKey
	}
	return nil
}

func (m *MlsCredential) GetAssociation() isMlsCredential_Association {
	if m != nil {
		return m.Association
	}
	return nil
}

func (x *MlsCredential) GetEip_191() *Eip191Association {
	if x, ok := x.GetAssociation().(*MlsCredential_Eip_191); ok {
		return x.Eip_191
	}
	return nil
}

type isMlsCredential_Association interface {
	isMlsCredential_Association()
}

type MlsCredential_Eip_191 struct {
	Eip_191 *Eip191Association `protobuf:"bytes,2,opt,name=eip_191,json=eip191,proto3,oneof"`
}

func (*MlsCredential_Eip_191) isMlsCredential_Association() {}

// A declaration and proof that a credential is no longer valid
type CredentialRevocation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InstallationPublicKey []byte `protobuf:"bytes,1,opt,name=installation_public_key,json=installationPublicKey,proto3" json:"installation_public_key,omitempty"`
	// Types that are assignable to Association:
	//
	//	*CredentialRevocation_Eip_191
	Association isCredentialRevocation_Association `protobuf_oneof:"association"`
}

func (x *CredentialRevocation) Reset() {
	*x = CredentialRevocation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_message_contents_credential_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CredentialRevocation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CredentialRevocation) ProtoMessage() {}

func (x *CredentialRevocation) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_credential_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CredentialRevocation.ProtoReflect.Descriptor instead.
func (*CredentialRevocation) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_credential_proto_rawDescGZIP(), []int{1}
}

func (x *CredentialRevocation) GetInstallationPublicKey() []byte {
	if x != nil {
		return x.InstallationPublicKey
	}
	return nil
}

func (m *CredentialRevocation) GetAssociation() isCredentialRevocation_Association {
	if m != nil {
		return m.Association
	}
	return nil
}

func (x *CredentialRevocation) GetEip_191() *Eip191Association {
	if x, ok := x.GetAssociation().(*CredentialRevocation_Eip_191); ok {
		return x.Eip_191
	}
	return nil
}

type isCredentialRevocation_Association interface {
	isCredentialRevocation_Association()
}

type CredentialRevocation_Eip_191 struct {
	Eip_191 *Eip191Association `protobuf:"bytes,2,opt,name=eip_191,json=eip191,proto3,oneof"`
}

func (*CredentialRevocation_Eip_191) isCredentialRevocation_Association() {}

var File_mls_message_contents_credential_proto protoreflect.FileDescriptor

var file_mls_message_contents_credential_proto_rawDesc = []byte{
	0x0a, 0x25, 0x6d, 0x6c, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x63, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61,
	0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x19, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c,
	0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x73, 0x1a, 0x26, 0x6d, 0x6c, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x61, 0x73, 0x73, 0x6f, 0x63, 0x69, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x9f, 0x01, 0x0a, 0x0d, 0x4d,
	0x6c, 0x73, 0x43, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x12, 0x36, 0x0a, 0x17,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x70, 0x75, 0x62,
	0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x15, 0x69,
	0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x75, 0x62, 0x6c, 0x69,
	0x63, 0x4b, 0x65, 0x79, 0x12, 0x47, 0x0a, 0x07, 0x65, 0x69, 0x70, 0x5f, 0x31, 0x39, 0x31, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73,
	0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x73, 0x2e, 0x45, 0x69, 0x70, 0x31, 0x39, 0x31, 0x41, 0x73, 0x73, 0x6f, 0x63, 0x69, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x06, 0x65, 0x69, 0x70, 0x31, 0x39, 0x31, 0x42, 0x0d, 0x0a,
	0x0b, 0x61, 0x73, 0x73, 0x6f, 0x63, 0x69, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xa6, 0x01, 0x0a,
	0x14, 0x43, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x52, 0x65, 0x76, 0x6f, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x36, 0x0a, 0x17, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x15, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x12, 0x47, 0x0a,
	0x07, 0x65, 0x69, 0x70, 0x5f, 0x31, 0x39, 0x31, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2c,
	0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x45, 0x69, 0x70, 0x31, 0x39,
	0x31, 0x41, 0x73, 0x73, 0x6f, 0x63, 0x69, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x06,
	0x65, 0x69, 0x70, 0x31, 0x39, 0x31, 0x42, 0x0d, 0x0a, 0x0b, 0x61, 0x73, 0x73, 0x6f, 0x63, 0x69,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0xef, 0x01, 0x0a, 0x1d, 0x63, 0x6f, 0x6d, 0x2e, 0x78, 0x6d,
	0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x42, 0x0f, 0x43, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x61, 0x6c, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3b, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2f, 0x78, 0x6d, 0x74, 0x70,
	0x2d, 0x6e, 0x6f, 0x64, 0x65, 0x2d, 0x67, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x6d, 0x6c, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xa2, 0x02, 0x03, 0x58, 0x4d, 0x4d, 0xaa, 0x02, 0x18,
	0x58, 0x6d, 0x74, 0x70, 0x2e, 0x4d, 0x6c, 0x73, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xca, 0x02, 0x18, 0x58, 0x6d, 0x74, 0x70, 0x5c,
	0x4d, 0x6c, 0x73, 0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x73, 0xe2, 0x02, 0x24, 0x58, 0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x6c, 0x73, 0x5c, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x5c, 0x47,
	0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x1a, 0x58, 0x6d, 0x74,
	0x70, 0x3a, 0x3a, 0x4d, 0x6c, 0x73, 0x3a, 0x3a, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mls_message_contents_credential_proto_rawDescOnce sync.Once
	file_mls_message_contents_credential_proto_rawDescData = file_mls_message_contents_credential_proto_rawDesc
)

func file_mls_message_contents_credential_proto_rawDescGZIP() []byte {
	file_mls_message_contents_credential_proto_rawDescOnce.Do(func() {
		file_mls_message_contents_credential_proto_rawDescData = protoimpl.X.CompressGZIP(file_mls_message_contents_credential_proto_rawDescData)
	})
	return file_mls_message_contents_credential_proto_rawDescData
}

var file_mls_message_contents_credential_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_mls_message_contents_credential_proto_goTypes = []interface{}{
	(*MlsCredential)(nil),        // 0: xmtp.mls.message_contents.MlsCredential
	(*CredentialRevocation)(nil), // 1: xmtp.mls.message_contents.CredentialRevocation
	(*Eip191Association)(nil),    // 2: xmtp.mls.message_contents.Eip191Association
}
var file_mls_message_contents_credential_proto_depIdxs = []int32{
	2, // 0: xmtp.mls.message_contents.MlsCredential.eip_191:type_name -> xmtp.mls.message_contents.Eip191Association
	2, // 1: xmtp.mls.message_contents.CredentialRevocation.eip_191:type_name -> xmtp.mls.message_contents.Eip191Association
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_mls_message_contents_credential_proto_init() }
func file_mls_message_contents_credential_proto_init() {
	if File_mls_message_contents_credential_proto != nil {
		return
	}
	file_mls_message_contents_association_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_mls_message_contents_credential_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MlsCredential); i {
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
		file_mls_message_contents_credential_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CredentialRevocation); i {
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
	file_mls_message_contents_credential_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*MlsCredential_Eip_191)(nil),
	}
	file_mls_message_contents_credential_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*CredentialRevocation_Eip_191)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_mls_message_contents_credential_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mls_message_contents_credential_proto_goTypes,
		DependencyIndexes: file_mls_message_contents_credential_proto_depIdxs,
		MessageInfos:      file_mls_message_contents_credential_proto_msgTypes,
	}.Build()
	File_mls_message_contents_credential_proto = out.File
	file_mls_message_contents_credential_proto_rawDesc = nil
	file_mls_message_contents_credential_proto_goTypes = nil
	file_mls_message_contents_credential_proto_depIdxs = nil
}
