// Package redir：听 127.0.0.1:redir，按连接 UID 选上游，CONNECT/SOCKS5 过去。
package redir

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
)

type Server struct {
	mu    sync.Mutex
	ln    net.Listener
	byUID map[int]Upstream
	all   *Upstream
}

func (s *Server) SetRoutes(byUID map[int]Upstream, all *Upstream) {
	s.mu.Lock()
	s.byUID = byUID
	s.all = all
	s.mu.Unlock()
}

func (s *Server) Listen(port int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ln != nil {
		return nil
	}
	// OUTPUT REDIRECT 对本机发出的包改写到 127.0.0.1:port，不要听 0.0.0.0。
	ln, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return err
	}
	s.ln = ln
	go s.serve(ln)
	return nil
}

func (s *Server) Close() {
	s.mu.Lock()
	ln := s.ln
	s.ln = nil
	s.byUID = nil
	s.all = nil
	s.mu.Unlock()
	if ln != nil {
		ln.Close()
	}
}

func (s *Server) pick(uid int) (Upstream, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u, ok := s.byUID[uid]; ok {
		return u, true
	}
	if s.all != nil {
		return *s.all, true
	}
	// SO_PEERCRED 偶尔拿不到 UID。所有路由转到同一台 mitm 时，用那条，别把包丢掉。
	var shared *Upstream
	for _, u := range s.byUID {
		cp := u
		if shared == nil {
			shared = &cp
			continue
		}
		if shared.Kind != u.Kind || shared.Addr() != u.Addr() {
			return Upstream{}, false
		}
	}
	if shared != nil {
		return *shared, true
	}
	return Upstream{}, false
}

func (s *Server) serve(ln net.Listener) {
	for {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		go s.handle(c)
	}
}

func (s *Server) handle(c net.Conn) {
	defer c.Close()
	// SO_PEERCRED 是发出这条 REDIRECT 连接的 App UID。
	uid, uidErr := peerUID(c)
	up, ok := s.pick(uid)
	if !ok {
		fmt.Fprintf(os.Stderr, "drop peer uid=%d err=%v\n", uid, uidErr)
		return
	}
	dst, err := originalDst(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "drop origdst uid=%d err=%v\n", uid, err)
		return
	}
	upConn, err := dialUpstream(up, dst)
	if err != nil {
		fmt.Fprintf(os.Stderr, "drop dial uid=%d dst=%s up=%s err=%v\n", uid, dst.String(), up.Addr(), err)
		return
	}
	defer upConn.Close()
	relay(c, upConn)
}

func relay(a, b net.Conn) {
	done := make(chan struct{}, 2)
	go func() {
		io.Copy(a, b)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(b, a)
		done <- struct{}{}
	}()
	<-done
	a.Close()
	b.Close()
	<-done
}
