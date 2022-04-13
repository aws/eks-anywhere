package networkutils_test

import (
	"errors"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/networkutils/mocks"
)

var (
	validPorts   = []string{"443", "8080", "32000"}
	invalidPorts = []string{"", "443a", "abc", "0", "123456"}
)

func TestIsPortValidExpectValid(t *testing.T) {
	for _, port := range validPorts {
		if !networkutils.IsPortValid(port) {
			t.Fatalf("Expected port %s to be valid", port)
		}
	}
}

func TestIsPortValidExpectInvalid(t *testing.T) {
	for _, port := range invalidPorts {
		if networkutils.IsPortValid(port) {
			t.Fatalf("Expected port %s to be invalid", port)
		}
	}
}

func TestIsIPInUsePass(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	client := mocks.NewMockNetClient(ctrl)
	client.EXPECT().DialTimeout(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(5).
		Return(nil, errors.New("no connection"))

	res := networkutils.IsIPInUse(client, "10.10.10.10")
	g.Expect(res).To(gomega.BeFalse())
}

func TestIsIPInUseFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	conn := NewMockConn(ctrl)
	conn.EXPECT().Close().Return(nil)

	client := mocks.NewMockNetClient(ctrl)
	client.EXPECT().DialTimeout(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(conn, nil)

	res := networkutils.IsIPInUse(client, "10.10.10.10")
	g.Expect(res).To(gomega.BeTrue())
}

// MockConn is a mock of NetClient interface. It is hand written.
type MockConn struct {
	ctrl     *gomock.Controller
	recorder *MockConnMockRecorder
}

var _ net.Conn = &MockConn{}

// MockConnMockRecorder is the mock recorder for MockConn.
type MockConnMockRecorder struct {
	mock *MockConn
}

// NewMockConn creates a new mock instance.
func NewMockConn(ctrl *gomock.Controller) *MockConn {
	mock := &MockConn{ctrl: ctrl}
	mock.recorder = &MockConnMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConn) EXPECT() *MockConnMockRecorder {
	return m.recorder
}

// DialTimeout mocks base method.
func (m *MockConn) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

func (m *MockConn) Read(b []byte) (n int, err error)   { panic("unimplemented") }
func (m *MockConn) Write(b []byte) (n int, err error)  { panic("unimplemented") }
func (m *MockConn) LocalAddr() net.Addr                { panic("unimplemented") }
func (m *MockConn) RemoteAddr() net.Addr               { panic("unimplemented") }
func (m *MockConn) SetDeadline(t time.Time) error      { panic("unimplemented") }
func (m *MockConn) SetReadDeadline(t time.Time) error  { panic("unimplemented") }
func (m *MockConn) SetWriteDeadline(t time.Time) error { panic("unimplemented") }

func (mr *MockConnMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockConn)(nil).Close))
}
