package networkutils_test

import (
	"errors"
	"net"
	"reflect"
	"syscall"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

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
		Return(nil, errors.New("no connection"))

	res := networkutils.IsIPInUse(client, "10.10.10.10")
	g.Expect(res).To(gomega.BeFalse())
}

func TestIsIPInUseConnectionRefused(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	client := mocks.NewMockNetClient(ctrl)
	client.EXPECT().DialTimeout(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, syscall.ECONNREFUSED)

	res := networkutils.IsIPInUse(client, "10.10.10.10")
	g.Expect(res).To(gomega.BeTrue())
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

func TestIsPortInUsePass(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	client := mocks.NewMockNetClient(ctrl)
	client.EXPECT().DialTimeout("tcp", "10.10.10.10:80", 500*time.Millisecond).
		Return(nil, errors.New("no connection"))

	res := networkutils.IsPortInUse(client, "10.10.10.10", "80")
	g.Expect(res).To(gomega.BeFalse())
}

func TestIsPortInUseFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	conn := NewMockConn(ctrl)
	conn.EXPECT().Close().Return(nil)

	client := mocks.NewMockNetClient(ctrl)
	client.EXPECT().DialTimeout("tcp", "10.10.10.10:80", 500*time.Millisecond).
		Return(conn, nil)

	res := networkutils.IsPortInUse(client, "10.10.10.10", "80")
	g.Expect(res).To(gomega.BeTrue())
}

func TestGetLocalIP(t *testing.T) {
	_, err := networkutils.GetLocalIP()
	if err != nil {
		t.Fatalf("unable to get local IP: %v", err)
	}
}

func TestResolveLocalIPOverride(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		override    string
		wantErr     string
	}{
		{"IPv4", "192.0.2.11", "192.0.2.12", ""},
		{"IPv6", "2001:db8::11", "2001:db8::12", ""},
		{"mapped IPv4", "192.0.2.11", "::ffff:192.0.2.12", ""},
		{"IPv4 override for IPv6", "2001:db8::11", "192.0.2.12", "bootstrap IP 192.0.2.12 and Tinkerbell IP 2001:db8::11 must use the same IP family"},
		{"IPv6 override for IPv4", "192.0.2.11", "2001:db8::12", "bootstrap IP 2001:db8::12 and Tinkerbell IP 192.0.2.11 must use the same IP family"},
		{"invalid override", "2001:db8::11", "invalid", `invalid bootstrap IP "invalid"`},
		{"invalid destination", "invalid", "192.0.2.12", `invalid destination IP "invalid"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			// A nil client proves explicit overrides are validated without a route lookup.
			ip, err := networkutils.ResolveLocalIP(nil, tt.destination, tt.override)
			if tt.wantErr != "" {
				g.Expect(err).To(gomega.MatchError(tt.wantErr))
				g.Expect(ip).To(gomega.BeNil())
				return
			}
			g.Expect(err).To(gomega.Succeed())
			g.Expect(ip.Equal(net.ParseIP(tt.override))).To(gomega.BeTrue())
		})
	}
}

func TestResolveLocalIPRoute(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		network     string
		address     string
		source      net.Addr
		wantErr     string
	}{
		{"IPv4 route", "192.0.2.11", "udp4", "192.0.2.11:80", &net.UDPAddr{IP: net.ParseIP("192.0.2.12")}, ""},
		{"IPv6 route", "2001:db8::11", "udp6", "[2001:db8::11]:80", &net.UDPAddr{IP: net.ParseIP("2001:db8::12")}, ""},
		{"wrong source family", "2001:db8::11", "udp6", "[2001:db8::11]:80", &net.UDPAddr{IP: net.ParseIP("192.0.2.12")}, "bootstrap IP 192.0.2.12 and Tinkerbell IP 2001:db8::11 must use the same IP family"},
		{"missing source", "2001:db8::11", "udp6", "[2001:db8::11]:80", &net.UDPAddr{}, "failed to retrieve local ip: missing UDP source address"},
		{"unexpected address type", "2001:db8::11", "udp6", "[2001:db8::11]:80", &net.TCPAddr{IP: net.ParseIP("2001:db8::12")}, "failed to retrieve local ip: missing UDP source address"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			g := gomega.NewWithT(t)
			conn := NewMockConn(ctrl)
			conn.EXPECT().LocalAddr().Return(tt.source)
			conn.EXPECT().Close().Return(nil)
			client := mocks.NewMockNetClient(ctrl)
			client.EXPECT().DialTimeout(tt.network, tt.address, 5*time.Second).Return(conn, nil)
			ip, err := networkutils.ResolveLocalIP(client, tt.destination, "")
			if tt.wantErr != "" {
				g.Expect(err).To(gomega.MatchError(tt.wantErr))
				g.Expect(ip).To(gomega.BeNil())
				return
			}
			g.Expect(err).To(gomega.Succeed())
			g.Expect(ip.Equal(tt.source.(*net.UDPAddr).IP)).To(gomega.BeTrue())
		})
	}
}

func TestResolveLocalIPRouteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)
	routeError := errors.New("no route")
	client := mocks.NewMockNetClient(ctrl)
	client.EXPECT().DialTimeout("udp6", "[2001:db8::11]:80", 5*time.Second).Return(nil, routeError)
	ip, err := networkutils.ResolveLocalIP(client, "2001:db8::11", "")
	g.Expect(err).To(gomega.MatchError("failed to retrieve local ip: no route"))
	g.Expect(errors.Is(err, routeError)).To(gomega.BeTrue())
	g.Expect(ip).To(gomega.BeNil())
}

func TestInterfaceNameForIPLoopback(t *testing.T) {
	for _, address := range []string{"127.0.0.1", "::1"} {
		t.Run(address, func(t *testing.T) {
			g := gomega.NewWithT(t)
			name, err := networkutils.InterfaceNameForIP(net.ParseIP(address))
			g.Expect(err).To(gomega.Succeed())
			iface, err := net.InterfaceByName(name)
			g.Expect(err).To(gomega.Succeed())
			g.Expect(iface.Flags & net.FlagLoopback).NotTo(gomega.BeZero())
		})
	}
}

func TestInterfaceNameForIPNotLocal(t *testing.T) {
	g := gomega.NewWithT(t)
	name, err := networkutils.InterfaceNameForIP(net.ParseIP("2001:db8:ffff::1234"))
	g.Expect(err).To(gomega.MatchError("no local network interface owns IP 2001:db8:ffff::1234"))
	g.Expect(name).To(gomega.BeEmpty())
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

func (m *MockConn) Read(b []byte) (n int, err error)  { panic("unimplemented") }
func (m *MockConn) Write(b []byte) (n int, err error) { panic("unimplemented") }
func (m *MockConn) LocalAddr() net.Addr {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LocalAddr")
	ret0, _ := ret[0].(net.Addr)
	return ret0
}
func (m *MockConn) RemoteAddr() net.Addr               { panic("unimplemented") }
func (m *MockConn) SetDeadline(t time.Time) error      { panic("unimplemented") }
func (m *MockConn) SetReadDeadline(t time.Time) error  { panic("unimplemented") }
func (m *MockConn) SetWriteDeadline(t time.Time) error { panic("unimplemented") }

func (mr *MockConnMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockConn)(nil).Close))
}

func (mr *MockConnMockRecorder) LocalAddr() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LocalAddr", reflect.TypeOf((*MockConn)(nil).LocalAddr))
}
