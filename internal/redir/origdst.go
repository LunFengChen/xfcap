package redir

import (
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

// REDIRECT 之后 socket 的对端变成 127.0.0.1:17892，真正要去的地址在 SO_ORIGINAL_DST。
const soOriginalDst = 80

func originalDst(c net.Conn) (net.TCPAddr, error) {
	tc, ok := c.(*net.TCPConn)
	if !ok {
		return net.TCPAddr{}, fmt.Errorf("not tcp")
	}
	rc, err := tc.SyscallConn()
	if err != nil {
		return net.TCPAddr{}, err
	}
	var addr net.TCPAddr
	var opErr error
	if err := rc.Control(func(fd uintptr) {
		addr, opErr = origDstIPv4(fd)
	}); err != nil {
		return net.TCPAddr{}, err
	}
	return addr, opErr
}

func origDstIPv4(fd uintptr) (net.TCPAddr, error) {
	var sa syscall.RawSockaddrInet4
	n := uint32(unsafe.Sizeof(sa))
	_, _, e := syscall.Syscall6(syscall.SYS_GETSOCKOPT, fd, syscall.SOL_IP, soOriginalDst, uintptr(unsafe.Pointer(&sa)), uintptr(unsafe.Pointer(&n)), 0)
	if e != 0 {
		return net.TCPAddr{}, e
	}
	return sockaddrInet4(sa), nil
}

func sockaddrInet4(sa syscall.RawSockaddrInet4) net.TCPAddr {
	port := binary.BigEndian.Uint16((*[2]byte)(unsafe.Pointer(&sa.Port))[:])
	return net.TCPAddr{
		IP:   net.IPv4(sa.Addr[0], sa.Addr[1], sa.Addr[2], sa.Addr[3]),
		Port: int(port),
	}
}
