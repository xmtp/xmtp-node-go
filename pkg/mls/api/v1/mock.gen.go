// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1 (interfaces: MlsApi_SubscribeGroupMessagesServer,MlsApi_SubscribeWelcomeMessagesServer)

// Package api is a generated GoMock package.
package api

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	apiv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	metadata "google.golang.org/grpc/metadata"
)

// MockMlsApi_SubscribeGroupMessagesServer is a mock of MlsApi_SubscribeGroupMessagesServer interface.
type MockMlsApi_SubscribeGroupMessagesServer struct {
	ctrl     *gomock.Controller
	recorder *MockMlsApi_SubscribeGroupMessagesServerMockRecorder
}

// MockMlsApi_SubscribeGroupMessagesServerMockRecorder is the mock recorder for MockMlsApi_SubscribeGroupMessagesServer.
type MockMlsApi_SubscribeGroupMessagesServerMockRecorder struct {
	mock *MockMlsApi_SubscribeGroupMessagesServer
}

// NewMockMlsApi_SubscribeGroupMessagesServer creates a new mock instance.
func NewMockMlsApi_SubscribeGroupMessagesServer(ctrl *gomock.Controller) *MockMlsApi_SubscribeGroupMessagesServer {
	mock := &MockMlsApi_SubscribeGroupMessagesServer{ctrl: ctrl}
	mock.recorder = &MockMlsApi_SubscribeGroupMessagesServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMlsApi_SubscribeGroupMessagesServer) EXPECT() *MockMlsApi_SubscribeGroupMessagesServerMockRecorder {
	return m.recorder
}

// Context mocks base method.
func (m *MockMlsApi_SubscribeGroupMessagesServer) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context.
func (mr *MockMlsApi_SubscribeGroupMessagesServerMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockMlsApi_SubscribeGroupMessagesServer)(nil).Context))
}

// RecvMsg mocks base method.
func (m *MockMlsApi_SubscribeGroupMessagesServer) RecvMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg.
func (mr *MockMlsApi_SubscribeGroupMessagesServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockMlsApi_SubscribeGroupMessagesServer)(nil).RecvMsg), arg0)
}

// Send mocks base method.
func (m *MockMlsApi_SubscribeGroupMessagesServer) Send(arg0 *apiv1.GroupMessage) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockMlsApi_SubscribeGroupMessagesServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockMlsApi_SubscribeGroupMessagesServer)(nil).Send), arg0)
}

// SendHeader mocks base method.
func (m *MockMlsApi_SubscribeGroupMessagesServer) SendHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader.
func (mr *MockMlsApi_SubscribeGroupMessagesServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockMlsApi_SubscribeGroupMessagesServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method.
func (m *MockMlsApi_SubscribeGroupMessagesServer) SendMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg.
func (mr *MockMlsApi_SubscribeGroupMessagesServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockMlsApi_SubscribeGroupMessagesServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method.
func (m *MockMlsApi_SubscribeGroupMessagesServer) SetHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader.
func (mr *MockMlsApi_SubscribeGroupMessagesServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockMlsApi_SubscribeGroupMessagesServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method.
func (m *MockMlsApi_SubscribeGroupMessagesServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer.
func (mr *MockMlsApi_SubscribeGroupMessagesServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockMlsApi_SubscribeGroupMessagesServer)(nil).SetTrailer), arg0)
}

// MockMlsApi_SubscribeWelcomeMessagesServer is a mock of MlsApi_SubscribeWelcomeMessagesServer interface.
type MockMlsApi_SubscribeWelcomeMessagesServer struct {
	ctrl     *gomock.Controller
	recorder *MockMlsApi_SubscribeWelcomeMessagesServerMockRecorder
}

// MockMlsApi_SubscribeWelcomeMessagesServerMockRecorder is the mock recorder for MockMlsApi_SubscribeWelcomeMessagesServer.
type MockMlsApi_SubscribeWelcomeMessagesServerMockRecorder struct {
	mock *MockMlsApi_SubscribeWelcomeMessagesServer
}

// NewMockMlsApi_SubscribeWelcomeMessagesServer creates a new mock instance.
func NewMockMlsApi_SubscribeWelcomeMessagesServer(ctrl *gomock.Controller) *MockMlsApi_SubscribeWelcomeMessagesServer {
	mock := &MockMlsApi_SubscribeWelcomeMessagesServer{ctrl: ctrl}
	mock.recorder = &MockMlsApi_SubscribeWelcomeMessagesServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMlsApi_SubscribeWelcomeMessagesServer) EXPECT() *MockMlsApi_SubscribeWelcomeMessagesServerMockRecorder {
	return m.recorder
}

// Context mocks base method.
func (m *MockMlsApi_SubscribeWelcomeMessagesServer) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context.
func (mr *MockMlsApi_SubscribeWelcomeMessagesServerMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockMlsApi_SubscribeWelcomeMessagesServer)(nil).Context))
}

// RecvMsg mocks base method.
func (m *MockMlsApi_SubscribeWelcomeMessagesServer) RecvMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg.
func (mr *MockMlsApi_SubscribeWelcomeMessagesServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockMlsApi_SubscribeWelcomeMessagesServer)(nil).RecvMsg), arg0)
}

// Send mocks base method.
func (m *MockMlsApi_SubscribeWelcomeMessagesServer) Send(arg0 *apiv1.WelcomeMessage) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockMlsApi_SubscribeWelcomeMessagesServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockMlsApi_SubscribeWelcomeMessagesServer)(nil).Send), arg0)
}

// SendHeader mocks base method.
func (m *MockMlsApi_SubscribeWelcomeMessagesServer) SendHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader.
func (mr *MockMlsApi_SubscribeWelcomeMessagesServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockMlsApi_SubscribeWelcomeMessagesServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method.
func (m *MockMlsApi_SubscribeWelcomeMessagesServer) SendMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg.
func (mr *MockMlsApi_SubscribeWelcomeMessagesServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockMlsApi_SubscribeWelcomeMessagesServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method.
func (m *MockMlsApi_SubscribeWelcomeMessagesServer) SetHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader.
func (mr *MockMlsApi_SubscribeWelcomeMessagesServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockMlsApi_SubscribeWelcomeMessagesServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method.
func (m *MockMlsApi_SubscribeWelcomeMessagesServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer.
func (mr *MockMlsApi_SubscribeWelcomeMessagesServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockMlsApi_SubscribeWelcomeMessagesServer)(nil).SetTrailer), arg0)
}
