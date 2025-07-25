// Code generated by mockery v2.52.2. DO NOT EDIT.

package mocks

import (
	context "context"

	apiv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"

	metadata "google.golang.org/grpc/metadata"

	mock "github.com/stretchr/testify/mock"
)

// MockMlsApi_SubscribeGroupMessagesServer is an autogenerated mock type for the MlsApi_SubscribeGroupMessagesServer type
type MockMlsApi_SubscribeGroupMessagesServer struct {
	mock.Mock
}

type MockMlsApi_SubscribeGroupMessagesServer_Expecter struct {
	mock *mock.Mock
}

func (_m *MockMlsApi_SubscribeGroupMessagesServer) EXPECT() *MockMlsApi_SubscribeGroupMessagesServer_Expecter {
	return &MockMlsApi_SubscribeGroupMessagesServer_Expecter{mock: &_m.Mock}
}

// Context provides a mock function with no fields
func (_m *MockMlsApi_SubscribeGroupMessagesServer) Context() context.Context {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Context")
	}

	var r0 context.Context
	if rf, ok := ret.Get(0).(func() context.Context); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(context.Context)
		}
	}

	return r0
}

// MockMlsApi_SubscribeGroupMessagesServer_Context_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Context'
type MockMlsApi_SubscribeGroupMessagesServer_Context_Call struct {
	*mock.Call
}

// Context is a helper method to define mock.On call
func (_e *MockMlsApi_SubscribeGroupMessagesServer_Expecter) Context() *MockMlsApi_SubscribeGroupMessagesServer_Context_Call {
	return &MockMlsApi_SubscribeGroupMessagesServer_Context_Call{Call: _e.mock.On("Context")}
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_Context_Call) Run(run func()) *MockMlsApi_SubscribeGroupMessagesServer_Context_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_Context_Call) Return(_a0 context.Context) *MockMlsApi_SubscribeGroupMessagesServer_Context_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_Context_Call) RunAndReturn(run func() context.Context) *MockMlsApi_SubscribeGroupMessagesServer_Context_Call {
	_c.Call.Return(run)
	return _c
}

// RecvMsg provides a mock function with given fields: m
func (_m *MockMlsApi_SubscribeGroupMessagesServer) RecvMsg(m interface{}) error {
	ret := _m.Called(m)

	if len(ret) == 0 {
		panic("no return value specified for RecvMsg")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockMlsApi_SubscribeGroupMessagesServer_RecvMsg_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RecvMsg'
type MockMlsApi_SubscribeGroupMessagesServer_RecvMsg_Call struct {
	*mock.Call
}

// RecvMsg is a helper method to define mock.On call
//   - m interface{}
func (_e *MockMlsApi_SubscribeGroupMessagesServer_Expecter) RecvMsg(m interface{}) *MockMlsApi_SubscribeGroupMessagesServer_RecvMsg_Call {
	return &MockMlsApi_SubscribeGroupMessagesServer_RecvMsg_Call{Call: _e.mock.On("RecvMsg", m)}
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_RecvMsg_Call) Run(run func(m interface{})) *MockMlsApi_SubscribeGroupMessagesServer_RecvMsg_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_RecvMsg_Call) Return(_a0 error) *MockMlsApi_SubscribeGroupMessagesServer_RecvMsg_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_RecvMsg_Call) RunAndReturn(run func(interface{}) error) *MockMlsApi_SubscribeGroupMessagesServer_RecvMsg_Call {
	_c.Call.Return(run)
	return _c
}

// Send provides a mock function with given fields: _a0
func (_m *MockMlsApi_SubscribeGroupMessagesServer) Send(_a0 *apiv1.GroupMessage) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Send")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*apiv1.GroupMessage) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockMlsApi_SubscribeGroupMessagesServer_Send_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Send'
type MockMlsApi_SubscribeGroupMessagesServer_Send_Call struct {
	*mock.Call
}

// Send is a helper method to define mock.On call
//   - _a0 *apiv1.GroupMessage
func (_e *MockMlsApi_SubscribeGroupMessagesServer_Expecter) Send(_a0 interface{}) *MockMlsApi_SubscribeGroupMessagesServer_Send_Call {
	return &MockMlsApi_SubscribeGroupMessagesServer_Send_Call{Call: _e.mock.On("Send", _a0)}
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_Send_Call) Run(run func(_a0 *apiv1.GroupMessage)) *MockMlsApi_SubscribeGroupMessagesServer_Send_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*apiv1.GroupMessage))
	})
	return _c
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_Send_Call) Return(_a0 error) *MockMlsApi_SubscribeGroupMessagesServer_Send_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_Send_Call) RunAndReturn(run func(*apiv1.GroupMessage) error) *MockMlsApi_SubscribeGroupMessagesServer_Send_Call {
	_c.Call.Return(run)
	return _c
}

// SendHeader provides a mock function with given fields: _a0
func (_m *MockMlsApi_SubscribeGroupMessagesServer) SendHeader(_a0 metadata.MD) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for SendHeader")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(metadata.MD) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockMlsApi_SubscribeGroupMessagesServer_SendHeader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendHeader'
type MockMlsApi_SubscribeGroupMessagesServer_SendHeader_Call struct {
	*mock.Call
}

// SendHeader is a helper method to define mock.On call
//   - _a0 metadata.MD
func (_e *MockMlsApi_SubscribeGroupMessagesServer_Expecter) SendHeader(_a0 interface{}) *MockMlsApi_SubscribeGroupMessagesServer_SendHeader_Call {
	return &MockMlsApi_SubscribeGroupMessagesServer_SendHeader_Call{Call: _e.mock.On("SendHeader", _a0)}
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_SendHeader_Call) Run(run func(_a0 metadata.MD)) *MockMlsApi_SubscribeGroupMessagesServer_SendHeader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(metadata.MD))
	})
	return _c
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_SendHeader_Call) Return(_a0 error) *MockMlsApi_SubscribeGroupMessagesServer_SendHeader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_SendHeader_Call) RunAndReturn(run func(metadata.MD) error) *MockMlsApi_SubscribeGroupMessagesServer_SendHeader_Call {
	_c.Call.Return(run)
	return _c
}

// SendMsg provides a mock function with given fields: m
func (_m *MockMlsApi_SubscribeGroupMessagesServer) SendMsg(m interface{}) error {
	ret := _m.Called(m)

	if len(ret) == 0 {
		panic("no return value specified for SendMsg")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockMlsApi_SubscribeGroupMessagesServer_SendMsg_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendMsg'
type MockMlsApi_SubscribeGroupMessagesServer_SendMsg_Call struct {
	*mock.Call
}

// SendMsg is a helper method to define mock.On call
//   - m interface{}
func (_e *MockMlsApi_SubscribeGroupMessagesServer_Expecter) SendMsg(m interface{}) *MockMlsApi_SubscribeGroupMessagesServer_SendMsg_Call {
	return &MockMlsApi_SubscribeGroupMessagesServer_SendMsg_Call{Call: _e.mock.On("SendMsg", m)}
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_SendMsg_Call) Run(run func(m interface{})) *MockMlsApi_SubscribeGroupMessagesServer_SendMsg_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_SendMsg_Call) Return(_a0 error) *MockMlsApi_SubscribeGroupMessagesServer_SendMsg_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_SendMsg_Call) RunAndReturn(run func(interface{}) error) *MockMlsApi_SubscribeGroupMessagesServer_SendMsg_Call {
	_c.Call.Return(run)
	return _c
}

// SetHeader provides a mock function with given fields: _a0
func (_m *MockMlsApi_SubscribeGroupMessagesServer) SetHeader(_a0 metadata.MD) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for SetHeader")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(metadata.MD) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockMlsApi_SubscribeGroupMessagesServer_SetHeader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetHeader'
type MockMlsApi_SubscribeGroupMessagesServer_SetHeader_Call struct {
	*mock.Call
}

// SetHeader is a helper method to define mock.On call
//   - _a0 metadata.MD
func (_e *MockMlsApi_SubscribeGroupMessagesServer_Expecter) SetHeader(_a0 interface{}) *MockMlsApi_SubscribeGroupMessagesServer_SetHeader_Call {
	return &MockMlsApi_SubscribeGroupMessagesServer_SetHeader_Call{Call: _e.mock.On("SetHeader", _a0)}
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_SetHeader_Call) Run(run func(_a0 metadata.MD)) *MockMlsApi_SubscribeGroupMessagesServer_SetHeader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(metadata.MD))
	})
	return _c
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_SetHeader_Call) Return(_a0 error) *MockMlsApi_SubscribeGroupMessagesServer_SetHeader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_SetHeader_Call) RunAndReturn(run func(metadata.MD) error) *MockMlsApi_SubscribeGroupMessagesServer_SetHeader_Call {
	_c.Call.Return(run)
	return _c
}

// SetTrailer provides a mock function with given fields: _a0
func (_m *MockMlsApi_SubscribeGroupMessagesServer) SetTrailer(_a0 metadata.MD) {
	_m.Called(_a0)
}

// MockMlsApi_SubscribeGroupMessagesServer_SetTrailer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetTrailer'
type MockMlsApi_SubscribeGroupMessagesServer_SetTrailer_Call struct {
	*mock.Call
}

// SetTrailer is a helper method to define mock.On call
//   - _a0 metadata.MD
func (_e *MockMlsApi_SubscribeGroupMessagesServer_Expecter) SetTrailer(_a0 interface{}) *MockMlsApi_SubscribeGroupMessagesServer_SetTrailer_Call {
	return &MockMlsApi_SubscribeGroupMessagesServer_SetTrailer_Call{Call: _e.mock.On("SetTrailer", _a0)}
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_SetTrailer_Call) Run(run func(_a0 metadata.MD)) *MockMlsApi_SubscribeGroupMessagesServer_SetTrailer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(metadata.MD))
	})
	return _c
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_SetTrailer_Call) Return() *MockMlsApi_SubscribeGroupMessagesServer_SetTrailer_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockMlsApi_SubscribeGroupMessagesServer_SetTrailer_Call) RunAndReturn(run func(metadata.MD)) *MockMlsApi_SubscribeGroupMessagesServer_SetTrailer_Call {
	_c.Run(run)
	return _c
}

// NewMockMlsApi_SubscribeGroupMessagesServer creates a new instance of MockMlsApi_SubscribeGroupMessagesServer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockMlsApi_SubscribeGroupMessagesServer(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMlsApi_SubscribeGroupMessagesServer {
	mock := &MockMlsApi_SubscribeGroupMessagesServer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
