// Message API

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: message_api/v1/message_api.proto

package message_apiv1

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
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

// Sort direction
type SortDirection int32

const (
	SortDirection_SORT_DIRECTION_UNSPECIFIED SortDirection = 0
	SortDirection_SORT_DIRECTION_ASCENDING   SortDirection = 1
	SortDirection_SORT_DIRECTION_DESCENDING  SortDirection = 2
)

// Enum value maps for SortDirection.
var (
	SortDirection_name = map[int32]string{
		0: "SORT_DIRECTION_UNSPECIFIED",
		1: "SORT_DIRECTION_ASCENDING",
		2: "SORT_DIRECTION_DESCENDING",
	}
	SortDirection_value = map[string]int32{
		"SORT_DIRECTION_UNSPECIFIED": 0,
		"SORT_DIRECTION_ASCENDING":   1,
		"SORT_DIRECTION_DESCENDING":  2,
	}
)

func (x SortDirection) Enum() *SortDirection {
	p := new(SortDirection)
	*p = x
	return p
}

func (x SortDirection) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SortDirection) Descriptor() protoreflect.EnumDescriptor {
	return file_message_api_v1_message_api_proto_enumTypes[0].Descriptor()
}

func (SortDirection) Type() protoreflect.EnumType {
	return &file_message_api_v1_message_api_proto_enumTypes[0]
}

func (x SortDirection) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SortDirection.Descriptor instead.
func (SortDirection) EnumDescriptor() ([]byte, []int) {
	return file_message_api_v1_message_api_proto_rawDescGZIP(), []int{0}
}

// This is based off of the go-waku Index type, but with the
// receiverTime and pubsubTopic removed for simplicity.
// Both removed fields are optional
type IndexCursor struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Digest        []byte                 `protobuf:"bytes,1,opt,name=digest,proto3" json:"digest,omitempty"`
	SenderTimeNs  uint64                 `protobuf:"varint,2,opt,name=sender_time_ns,json=senderTimeNs,proto3" json:"sender_time_ns,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *IndexCursor) Reset() {
	*x = IndexCursor{}
	mi := &file_message_api_v1_message_api_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *IndexCursor) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IndexCursor) ProtoMessage() {}

func (x *IndexCursor) ProtoReflect() protoreflect.Message {
	mi := &file_message_api_v1_message_api_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IndexCursor.ProtoReflect.Descriptor instead.
func (*IndexCursor) Descriptor() ([]byte, []int) {
	return file_message_api_v1_message_api_proto_rawDescGZIP(), []int{0}
}

func (x *IndexCursor) GetDigest() []byte {
	if x != nil {
		return x.Digest
	}
	return nil
}

func (x *IndexCursor) GetSenderTimeNs() uint64 {
	if x != nil {
		return x.SenderTimeNs
	}
	return 0
}

// Wrapper for potentially multiple types of cursor
type Cursor struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Making the cursor a one-of type, as I would like to change the way we
	// handle pagination to use a precomputed sort field.
	// This way we can handle both methods
	//
	// Types that are valid to be assigned to Cursor:
	//
	//	*Cursor_Index
	Cursor        isCursor_Cursor `protobuf_oneof:"cursor"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Cursor) Reset() {
	*x = Cursor{}
	mi := &file_message_api_v1_message_api_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Cursor) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Cursor) ProtoMessage() {}

func (x *Cursor) ProtoReflect() protoreflect.Message {
	mi := &file_message_api_v1_message_api_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Cursor.ProtoReflect.Descriptor instead.
func (*Cursor) Descriptor() ([]byte, []int) {
	return file_message_api_v1_message_api_proto_rawDescGZIP(), []int{1}
}

func (x *Cursor) GetCursor() isCursor_Cursor {
	if x != nil {
		return x.Cursor
	}
	return nil
}

func (x *Cursor) GetIndex() *IndexCursor {
	if x != nil {
		if x, ok := x.Cursor.(*Cursor_Index); ok {
			return x.Index
		}
	}
	return nil
}

type isCursor_Cursor interface {
	isCursor_Cursor()
}

type Cursor_Index struct {
	Index *IndexCursor `protobuf:"bytes,1,opt,name=index,proto3,oneof"`
}

func (*Cursor_Index) isCursor_Cursor() {}

// This is based off of the go-waku PagingInfo struct, but with the direction
// changed to our SortDirection enum format
type PagingInfo struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Note: this is a uint32, while go-waku's pageSize is a uint64
	Limit         uint32        `protobuf:"varint,1,opt,name=limit,proto3" json:"limit,omitempty"`
	Cursor        *Cursor       `protobuf:"bytes,2,opt,name=cursor,proto3" json:"cursor,omitempty"`
	Direction     SortDirection `protobuf:"varint,3,opt,name=direction,proto3,enum=xmtp.message_api.v1.SortDirection" json:"direction,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PagingInfo) Reset() {
	*x = PagingInfo{}
	mi := &file_message_api_v1_message_api_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PagingInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PagingInfo) ProtoMessage() {}

func (x *PagingInfo) ProtoReflect() protoreflect.Message {
	mi := &file_message_api_v1_message_api_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PagingInfo.ProtoReflect.Descriptor instead.
func (*PagingInfo) Descriptor() ([]byte, []int) {
	return file_message_api_v1_message_api_proto_rawDescGZIP(), []int{2}
}

func (x *PagingInfo) GetLimit() uint32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *PagingInfo) GetCursor() *Cursor {
	if x != nil {
		return x.Cursor
	}
	return nil
}

func (x *PagingInfo) GetDirection() SortDirection {
	if x != nil {
		return x.Direction
	}
	return SortDirection_SORT_DIRECTION_UNSPECIFIED
}

// Envelope encapsulates a message while in transit.
type Envelope struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The topic the message belongs to,
	// If the message includes the topic as well
	// it MUST be the same as the topic in the envelope.
	ContentTopic string `protobuf:"bytes,1,opt,name=content_topic,json=contentTopic,proto3" json:"content_topic,omitempty"`
	// Message creation timestamp
	// If the message includes the timestamp as well
	// it MUST be equivalent to the timestamp in the envelope.
	TimestampNs   uint64 `protobuf:"varint,2,opt,name=timestamp_ns,json=timestampNs,proto3" json:"timestamp_ns,omitempty"`
	Message       []byte `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Envelope) Reset() {
	*x = Envelope{}
	mi := &file_message_api_v1_message_api_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Envelope) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Envelope) ProtoMessage() {}

func (x *Envelope) ProtoReflect() protoreflect.Message {
	mi := &file_message_api_v1_message_api_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Envelope.ProtoReflect.Descriptor instead.
func (*Envelope) Descriptor() ([]byte, []int) {
	return file_message_api_v1_message_api_proto_rawDescGZIP(), []int{3}
}

func (x *Envelope) GetContentTopic() string {
	if x != nil {
		return x.ContentTopic
	}
	return ""
}

func (x *Envelope) GetTimestampNs() uint64 {
	if x != nil {
		return x.TimestampNs
	}
	return 0
}

func (x *Envelope) GetMessage() []byte {
	if x != nil {
		return x.Message
	}
	return nil
}

// Publish
type PublishRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Envelopes     []*Envelope            `protobuf:"bytes,1,rep,name=envelopes,proto3" json:"envelopes,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PublishRequest) Reset() {
	*x = PublishRequest{}
	mi := &file_message_api_v1_message_api_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PublishRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublishRequest) ProtoMessage() {}

func (x *PublishRequest) ProtoReflect() protoreflect.Message {
	mi := &file_message_api_v1_message_api_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublishRequest.ProtoReflect.Descriptor instead.
func (*PublishRequest) Descriptor() ([]byte, []int) {
	return file_message_api_v1_message_api_proto_rawDescGZIP(), []int{4}
}

func (x *PublishRequest) GetEnvelopes() []*Envelope {
	if x != nil {
		return x.Envelopes
	}
	return nil
}

// Empty message as a response for Publish
type PublishResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PublishResponse) Reset() {
	*x = PublishResponse{}
	mi := &file_message_api_v1_message_api_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PublishResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublishResponse) ProtoMessage() {}

func (x *PublishResponse) ProtoReflect() protoreflect.Message {
	mi := &file_message_api_v1_message_api_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublishResponse.ProtoReflect.Descriptor instead.
func (*PublishResponse) Descriptor() ([]byte, []int) {
	return file_message_api_v1_message_api_proto_rawDescGZIP(), []int{5}
}

// Subscribe
type SubscribeRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ContentTopics []string               `protobuf:"bytes,1,rep,name=content_topics,json=contentTopics,proto3" json:"content_topics,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SubscribeRequest) Reset() {
	*x = SubscribeRequest{}
	mi := &file_message_api_v1_message_api_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubscribeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscribeRequest) ProtoMessage() {}

func (x *SubscribeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_message_api_v1_message_api_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscribeRequest.ProtoReflect.Descriptor instead.
func (*SubscribeRequest) Descriptor() ([]byte, []int) {
	return file_message_api_v1_message_api_proto_rawDescGZIP(), []int{6}
}

func (x *SubscribeRequest) GetContentTopics() []string {
	if x != nil {
		return x.ContentTopics
	}
	return nil
}

// SubscribeAll
type SubscribeAllRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SubscribeAllRequest) Reset() {
	*x = SubscribeAllRequest{}
	mi := &file_message_api_v1_message_api_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubscribeAllRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscribeAllRequest) ProtoMessage() {}

func (x *SubscribeAllRequest) ProtoReflect() protoreflect.Message {
	mi := &file_message_api_v1_message_api_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscribeAllRequest.ProtoReflect.Descriptor instead.
func (*SubscribeAllRequest) Descriptor() ([]byte, []int) {
	return file_message_api_v1_message_api_proto_rawDescGZIP(), []int{7}
}

// Query
type QueryRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ContentTopics []string               `protobuf:"bytes,1,rep,name=content_topics,json=contentTopics,proto3" json:"content_topics,omitempty"`
	StartTimeNs   uint64                 `protobuf:"varint,2,opt,name=start_time_ns,json=startTimeNs,proto3" json:"start_time_ns,omitempty"`
	EndTimeNs     uint64                 `protobuf:"varint,3,opt,name=end_time_ns,json=endTimeNs,proto3" json:"end_time_ns,omitempty"`
	PagingInfo    *PagingInfo            `protobuf:"bytes,4,opt,name=paging_info,json=pagingInfo,proto3" json:"paging_info,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *QueryRequest) Reset() {
	*x = QueryRequest{}
	mi := &file_message_api_v1_message_api_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *QueryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryRequest) ProtoMessage() {}

func (x *QueryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_message_api_v1_message_api_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryRequest.ProtoReflect.Descriptor instead.
func (*QueryRequest) Descriptor() ([]byte, []int) {
	return file_message_api_v1_message_api_proto_rawDescGZIP(), []int{8}
}

func (x *QueryRequest) GetContentTopics() []string {
	if x != nil {
		return x.ContentTopics
	}
	return nil
}

func (x *QueryRequest) GetStartTimeNs() uint64 {
	if x != nil {
		return x.StartTimeNs
	}
	return 0
}

func (x *QueryRequest) GetEndTimeNs() uint64 {
	if x != nil {
		return x.EndTimeNs
	}
	return 0
}

func (x *QueryRequest) GetPagingInfo() *PagingInfo {
	if x != nil {
		return x.PagingInfo
	}
	return nil
}

// The response, containing envelopes, for a query
type QueryResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Envelopes     []*Envelope            `protobuf:"bytes,1,rep,name=envelopes,proto3" json:"envelopes,omitempty"`
	PagingInfo    *PagingInfo            `protobuf:"bytes,2,opt,name=paging_info,json=pagingInfo,proto3" json:"paging_info,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *QueryResponse) Reset() {
	*x = QueryResponse{}
	mi := &file_message_api_v1_message_api_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *QueryResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryResponse) ProtoMessage() {}

func (x *QueryResponse) ProtoReflect() protoreflect.Message {
	mi := &file_message_api_v1_message_api_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryResponse.ProtoReflect.Descriptor instead.
func (*QueryResponse) Descriptor() ([]byte, []int) {
	return file_message_api_v1_message_api_proto_rawDescGZIP(), []int{9}
}

func (x *QueryResponse) GetEnvelopes() []*Envelope {
	if x != nil {
		return x.Envelopes
	}
	return nil
}

func (x *QueryResponse) GetPagingInfo() *PagingInfo {
	if x != nil {
		return x.PagingInfo
	}
	return nil
}

// BatchQuery
type BatchQueryRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Requests      []*QueryRequest        `protobuf:"bytes,1,rep,name=requests,proto3" json:"requests,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *BatchQueryRequest) Reset() {
	*x = BatchQueryRequest{}
	mi := &file_message_api_v1_message_api_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BatchQueryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BatchQueryRequest) ProtoMessage() {}

func (x *BatchQueryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_message_api_v1_message_api_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BatchQueryRequest.ProtoReflect.Descriptor instead.
func (*BatchQueryRequest) Descriptor() ([]byte, []int) {
	return file_message_api_v1_message_api_proto_rawDescGZIP(), []int{10}
}

func (x *BatchQueryRequest) GetRequests() []*QueryRequest {
	if x != nil {
		return x.Requests
	}
	return nil
}

// Response containing a list of QueryResponse messages
type BatchQueryResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Responses     []*QueryResponse       `protobuf:"bytes,1,rep,name=responses,proto3" json:"responses,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *BatchQueryResponse) Reset() {
	*x = BatchQueryResponse{}
	mi := &file_message_api_v1_message_api_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BatchQueryResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BatchQueryResponse) ProtoMessage() {}

func (x *BatchQueryResponse) ProtoReflect() protoreflect.Message {
	mi := &file_message_api_v1_message_api_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BatchQueryResponse.ProtoReflect.Descriptor instead.
func (*BatchQueryResponse) Descriptor() ([]byte, []int) {
	return file_message_api_v1_message_api_proto_rawDescGZIP(), []int{11}
}

func (x *BatchQueryResponse) GetResponses() []*QueryResponse {
	if x != nil {
		return x.Responses
	}
	return nil
}

var File_message_api_v1_message_api_proto protoreflect.FileDescriptor

const file_message_api_v1_message_api_proto_rawDesc = "" +
	"\n" +
	" message_api/v1/message_api.proto\x12\x13xmtp.message_api.v1\x1a\x1cgoogle/api/annotations.proto\x1a.protoc-gen-openapiv2/options/annotations.proto\"K\n" +
	"\vIndexCursor\x12\x16\n" +
	"\x06digest\x18\x01 \x01(\fR\x06digest\x12$\n" +
	"\x0esender_time_ns\x18\x02 \x01(\x04R\fsenderTimeNs\"L\n" +
	"\x06Cursor\x128\n" +
	"\x05index\x18\x01 \x01(\v2 .xmtp.message_api.v1.IndexCursorH\x00R\x05indexB\b\n" +
	"\x06cursor\"\x99\x01\n" +
	"\n" +
	"PagingInfo\x12\x14\n" +
	"\x05limit\x18\x01 \x01(\rR\x05limit\x123\n" +
	"\x06cursor\x18\x02 \x01(\v2\x1b.xmtp.message_api.v1.CursorR\x06cursor\x12@\n" +
	"\tdirection\x18\x03 \x01(\x0e2\".xmtp.message_api.v1.SortDirectionR\tdirection\"l\n" +
	"\bEnvelope\x12#\n" +
	"\rcontent_topic\x18\x01 \x01(\tR\fcontentTopic\x12!\n" +
	"\ftimestamp_ns\x18\x02 \x01(\x04R\vtimestampNs\x12\x18\n" +
	"\amessage\x18\x03 \x01(\fR\amessage\"M\n" +
	"\x0ePublishRequest\x12;\n" +
	"\tenvelopes\x18\x01 \x03(\v2\x1d.xmtp.message_api.v1.EnvelopeR\tenvelopes\"\x11\n" +
	"\x0fPublishResponse\"9\n" +
	"\x10SubscribeRequest\x12%\n" +
	"\x0econtent_topics\x18\x01 \x03(\tR\rcontentTopics\"\x15\n" +
	"\x13SubscribeAllRequest\"\xbb\x01\n" +
	"\fQueryRequest\x12%\n" +
	"\x0econtent_topics\x18\x01 \x03(\tR\rcontentTopics\x12\"\n" +
	"\rstart_time_ns\x18\x02 \x01(\x04R\vstartTimeNs\x12\x1e\n" +
	"\vend_time_ns\x18\x03 \x01(\x04R\tendTimeNs\x12@\n" +
	"\vpaging_info\x18\x04 \x01(\v2\x1f.xmtp.message_api.v1.PagingInfoR\n" +
	"pagingInfo\"\x8e\x01\n" +
	"\rQueryResponse\x12;\n" +
	"\tenvelopes\x18\x01 \x03(\v2\x1d.xmtp.message_api.v1.EnvelopeR\tenvelopes\x12@\n" +
	"\vpaging_info\x18\x02 \x01(\v2\x1f.xmtp.message_api.v1.PagingInfoR\n" +
	"pagingInfo\"R\n" +
	"\x11BatchQueryRequest\x12=\n" +
	"\brequests\x18\x01 \x03(\v2!.xmtp.message_api.v1.QueryRequestR\brequests\"V\n" +
	"\x12BatchQueryResponse\x12@\n" +
	"\tresponses\x18\x01 \x03(\v2\".xmtp.message_api.v1.QueryResponseR\tresponses*l\n" +
	"\rSortDirection\x12\x1e\n" +
	"\x1aSORT_DIRECTION_UNSPECIFIED\x10\x00\x12\x1c\n" +
	"\x18SORT_DIRECTION_ASCENDING\x10\x01\x12\x1d\n" +
	"\x19SORT_DIRECTION_DESCENDING\x10\x022\xc6\x05\n" +
	"\n" +
	"MessageApi\x12t\n" +
	"\aPublish\x12#.xmtp.message_api.v1.PublishRequest\x1a$.xmtp.message_api.v1.PublishResponse\"\x1e\x82\xd3\xe4\x93\x02\x18:\x01*\"\x13/message/v1/publish\x12u\n" +
	"\tSubscribe\x12%.xmtp.message_api.v1.SubscribeRequest\x1a\x1d.xmtp.message_api.v1.Envelope\" \x82\xd3\xe4\x93\x02\x1a:\x01*\"\x15/message/v1/subscribe0\x01\x12X\n" +
	"\n" +
	"Subscribe2\x12%.xmtp.message_api.v1.SubscribeRequest\x1a\x1d.xmtp.message_api.v1.Envelope\"\x00(\x010\x01\x12\x7f\n" +
	"\fSubscribeAll\x12(.xmtp.message_api.v1.SubscribeAllRequest\x1a\x1d.xmtp.message_api.v1.Envelope\"$\x82\xd3\xe4\x93\x02\x1e:\x01*\"\x19/message/v1/subscribe-all0\x01\x12l\n" +
	"\x05Query\x12!.xmtp.message_api.v1.QueryRequest\x1a\".xmtp.message_api.v1.QueryResponse\"\x1c\x82\xd3\xe4\x93\x02\x16:\x01*\"\x11/message/v1/query\x12\x81\x01\n" +
	"\n" +
	"BatchQuery\x12&.xmtp.message_api.v1.BatchQueryRequest\x1a'.xmtp.message_api.v1.BatchQueryResponse\"\"\x82\xd3\xe4\x93\x02\x1c:\x01*\"\x17/message/v1/batch-queryB\xef\x01\x92A\x13\x12\x11\n" +
	"\n" +
	"MessageApi2\x031.0\n" +
	"\x17com.xmtp.message_api.v1B\x0fMessageApiProtoP\x01ZCgithub.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1;message_apiv1\xa2\x02\x03XMX\xaa\x02\x12Xmtp.MessageApi.V1\xca\x02\x12Xmtp\\MessageApi\\V1\xe2\x02\x1eXmtp\\MessageApi\\V1\\GPBMetadata\xea\x02\x14Xmtp::MessageApi::V1b\x06proto3"

var (
	file_message_api_v1_message_api_proto_rawDescOnce sync.Once
	file_message_api_v1_message_api_proto_rawDescData []byte
)

func file_message_api_v1_message_api_proto_rawDescGZIP() []byte {
	file_message_api_v1_message_api_proto_rawDescOnce.Do(func() {
		file_message_api_v1_message_api_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_message_api_v1_message_api_proto_rawDesc), len(file_message_api_v1_message_api_proto_rawDesc)))
	})
	return file_message_api_v1_message_api_proto_rawDescData
}

var file_message_api_v1_message_api_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_message_api_v1_message_api_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_message_api_v1_message_api_proto_goTypes = []any{
	(SortDirection)(0),          // 0: xmtp.message_api.v1.SortDirection
	(*IndexCursor)(nil),         // 1: xmtp.message_api.v1.IndexCursor
	(*Cursor)(nil),              // 2: xmtp.message_api.v1.Cursor
	(*PagingInfo)(nil),          // 3: xmtp.message_api.v1.PagingInfo
	(*Envelope)(nil),            // 4: xmtp.message_api.v1.Envelope
	(*PublishRequest)(nil),      // 5: xmtp.message_api.v1.PublishRequest
	(*PublishResponse)(nil),     // 6: xmtp.message_api.v1.PublishResponse
	(*SubscribeRequest)(nil),    // 7: xmtp.message_api.v1.SubscribeRequest
	(*SubscribeAllRequest)(nil), // 8: xmtp.message_api.v1.SubscribeAllRequest
	(*QueryRequest)(nil),        // 9: xmtp.message_api.v1.QueryRequest
	(*QueryResponse)(nil),       // 10: xmtp.message_api.v1.QueryResponse
	(*BatchQueryRequest)(nil),   // 11: xmtp.message_api.v1.BatchQueryRequest
	(*BatchQueryResponse)(nil),  // 12: xmtp.message_api.v1.BatchQueryResponse
}
var file_message_api_v1_message_api_proto_depIdxs = []int32{
	1,  // 0: xmtp.message_api.v1.Cursor.index:type_name -> xmtp.message_api.v1.IndexCursor
	2,  // 1: xmtp.message_api.v1.PagingInfo.cursor:type_name -> xmtp.message_api.v1.Cursor
	0,  // 2: xmtp.message_api.v1.PagingInfo.direction:type_name -> xmtp.message_api.v1.SortDirection
	4,  // 3: xmtp.message_api.v1.PublishRequest.envelopes:type_name -> xmtp.message_api.v1.Envelope
	3,  // 4: xmtp.message_api.v1.QueryRequest.paging_info:type_name -> xmtp.message_api.v1.PagingInfo
	4,  // 5: xmtp.message_api.v1.QueryResponse.envelopes:type_name -> xmtp.message_api.v1.Envelope
	3,  // 6: xmtp.message_api.v1.QueryResponse.paging_info:type_name -> xmtp.message_api.v1.PagingInfo
	9,  // 7: xmtp.message_api.v1.BatchQueryRequest.requests:type_name -> xmtp.message_api.v1.QueryRequest
	10, // 8: xmtp.message_api.v1.BatchQueryResponse.responses:type_name -> xmtp.message_api.v1.QueryResponse
	5,  // 9: xmtp.message_api.v1.MessageApi.Publish:input_type -> xmtp.message_api.v1.PublishRequest
	7,  // 10: xmtp.message_api.v1.MessageApi.Subscribe:input_type -> xmtp.message_api.v1.SubscribeRequest
	7,  // 11: xmtp.message_api.v1.MessageApi.Subscribe2:input_type -> xmtp.message_api.v1.SubscribeRequest
	8,  // 12: xmtp.message_api.v1.MessageApi.SubscribeAll:input_type -> xmtp.message_api.v1.SubscribeAllRequest
	9,  // 13: xmtp.message_api.v1.MessageApi.Query:input_type -> xmtp.message_api.v1.QueryRequest
	11, // 14: xmtp.message_api.v1.MessageApi.BatchQuery:input_type -> xmtp.message_api.v1.BatchQueryRequest
	6,  // 15: xmtp.message_api.v1.MessageApi.Publish:output_type -> xmtp.message_api.v1.PublishResponse
	4,  // 16: xmtp.message_api.v1.MessageApi.Subscribe:output_type -> xmtp.message_api.v1.Envelope
	4,  // 17: xmtp.message_api.v1.MessageApi.Subscribe2:output_type -> xmtp.message_api.v1.Envelope
	4,  // 18: xmtp.message_api.v1.MessageApi.SubscribeAll:output_type -> xmtp.message_api.v1.Envelope
	10, // 19: xmtp.message_api.v1.MessageApi.Query:output_type -> xmtp.message_api.v1.QueryResponse
	12, // 20: xmtp.message_api.v1.MessageApi.BatchQuery:output_type -> xmtp.message_api.v1.BatchQueryResponse
	15, // [15:21] is the sub-list for method output_type
	9,  // [9:15] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_message_api_v1_message_api_proto_init() }
func file_message_api_v1_message_api_proto_init() {
	if File_message_api_v1_message_api_proto != nil {
		return
	}
	file_message_api_v1_message_api_proto_msgTypes[1].OneofWrappers = []any{
		(*Cursor_Index)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_message_api_v1_message_api_proto_rawDesc), len(file_message_api_v1_message_api_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_message_api_v1_message_api_proto_goTypes,
		DependencyIndexes: file_message_api_v1_message_api_proto_depIdxs,
		EnumInfos:         file_message_api_v1_message_api_proto_enumTypes,
		MessageInfos:      file_message_api_v1_message_api_proto_msgTypes,
	}.Build()
	File_message_api_v1_message_api_proto = out.File
	file_message_api_v1_message_api_proto_goTypes = nil
	file_message_api_v1_message_api_proto_depIdxs = nil
}
