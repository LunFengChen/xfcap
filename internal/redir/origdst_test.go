package redir

import (
	"encoding/binary"
	"net"
	"syscall"
	"testing"
	"unsafe"
)

func TestSockaddrInet4(t *testing.T) {
	var sa syscall.RawSockaddrInet4
	sa.Addr = [4]byte{93, 184, 216, 34}
	binary.BigEndian.PutUint16((*[2]byte)(unsafe.Pointer(&sa.Port))[:], 443)
	got := sockaddrInet4(sa)
	if !got.IP.Equal(net.IPv4(93, 184, 216, 34)) || got.Port != 443 {
		t.Fatalf("%v", got)
	}
}
