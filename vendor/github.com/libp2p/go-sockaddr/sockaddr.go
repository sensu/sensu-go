package sockaddr

import (
	"unsafe"

	sockaddrnet "github.com/libp2p/go-sockaddr/net"
)

import "C"

// Socklen is a type for the length of a sockaddr.
type Socklen uint

// SockaddrToAny converts a Sockaddr into a RawSockaddrAny
// The implementation is platform dependent.
func SockaddrToAny(sa sockaddrnet.Sockaddr) (*sockaddrnet.RawSockaddrAny, Socklen, error) {
	return sockaddrToAny(sa)
}

// SockaddrToAny converts a RawSockaddrAny into a Sockaddr
// The implementation is platform dependent.
func AnyToSockaddr(rsa *sockaddrnet.RawSockaddrAny) (sockaddrnet.Sockaddr, error) {
	return anyToSockaddr(rsa)
}

// AnyToCAny casts a *RawSockaddrAny to a *C.struct_sockaddr_any
func AnyToCAny(a *sockaddrnet.RawSockaddrAny) *C.struct_sockaddr_any {
	return (*C.struct_sockaddr_any)(unsafe.Pointer(a))
}

// CAnyToAny casts a *C.struct_sockaddr_any to a *RawSockaddrAny
func CAnyToAny(a *C.struct_sockaddr_any) *sockaddrnet.RawSockaddrAny {
	return (*sockaddrnet.RawSockaddrAny)(unsafe.Pointer(a))
}
