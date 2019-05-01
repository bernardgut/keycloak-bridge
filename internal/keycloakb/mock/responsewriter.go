// Code generated by MockGen. DO NOT EDIT.
// Source: net/http (interfaces: ResponseWriter)

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	http "net/http"
	reflect "reflect"
)

// ResponseWriter is a mock of ResponseWriter interface
type ResponseWriter struct {
	ctrl     *gomock.Controller
	recorder *ResponseWriterMockRecorder
}

// ResponseWriterMockRecorder is the mock recorder for ResponseWriter
type ResponseWriterMockRecorder struct {
	mock *ResponseWriter
}

// NewResponseWriter creates a new mock instance
func NewResponseWriter(ctrl *gomock.Controller) *ResponseWriter {
	mock := &ResponseWriter{ctrl: ctrl}
	mock.recorder = &ResponseWriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *ResponseWriter) EXPECT() *ResponseWriterMockRecorder {
	return m.recorder
}

// Header mocks base method
func (m *ResponseWriter) Header() http.Header {
	ret := m.ctrl.Call(m, "Header")
	ret0, _ := ret[0].(http.Header)
	return ret0
}

// Header indicates an expected call of Header
func (mr *ResponseWriterMockRecorder) Header() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Header", reflect.TypeOf((*ResponseWriter)(nil).Header))
}

// Write mocks base method
func (m *ResponseWriter) Write(arg0 []byte) (int, error) {
	ret := m.ctrl.Call(m, "Write", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write
func (mr *ResponseWriterMockRecorder) Write(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*ResponseWriter)(nil).Write), arg0)
}

// WriteHeader mocks base method
func (m *ResponseWriter) WriteHeader(arg0 int) {
	m.ctrl.Call(m, "WriteHeader", arg0)
}

// WriteHeader indicates an expected call of WriteHeader
func (mr *ResponseWriterMockRecorder) WriteHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteHeader", reflect.TypeOf((*ResponseWriter)(nil).WriteHeader), arg0)
}