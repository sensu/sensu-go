package sockaddrnet

import (
	"bytes"
	"net"
	"testing"
)

func assertIPEq(t *testing.T, ip net.IP, sa Sockaddr) {
	switch s := sa.(type) {
	case *SockaddrInet4:
		if !bytes.Equal(s.Addr[:], ip.To4()) {
			t.Error("IPs not equal")
		}
	case *SockaddrInet6:
		if !bytes.Equal(s.Addr[:], ip.To16()) {
			t.Error("IPs not equal")
		}
	default:
		t.Error("not a known sockaddr")
	}
}

func subtestIPSockaddr(t *testing.T, ip net.IP) {
	assertIPEq(t, ip, IPAndZoneToSockaddr(ip, ""))
}

func TestIPAndZoneToSockaddr(t *testing.T) {
	subtestIPSockaddr(t, net.ParseIP("127.0.0.1"))
	subtestIPSockaddr(t, net.IPv4zero)
	subtestIPSockaddr(t, net.IP(net.IPv4zero.To4()))
	subtestIPSockaddr(t, net.IPv6unspecified)
	assertIPEq(t, net.IPv4zero, IPAndZoneToSockaddr(nil, ""))
}
