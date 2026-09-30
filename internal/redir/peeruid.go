package redir

import (
	"fmt"
	"net"
	"syscall"
)

func peerUID(c net.Conn) (int, error) {
	tc, ok := c.(*net.TCPConn)
	if !ok {
		return 0, fmt.Errorf("not tcp")
	}
	rc, err := tc.SyscallConn()
	if err != nil {
		return 0, err
	}
	var uid int
	var opErr error
	if err := rc.Control(func(fd uintptr) {
		uc, e := syscall.GetsockoptUcred(int(fd), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
		if e != nil {
			opErr = e
			return
		}
		uid = int(uc.Uid)
	}); err != nil {
		return 0, err
	}
	return uid, opErr
}
