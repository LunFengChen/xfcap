package redir

import (
	"io"
	"net"
	"strings"
	"testing"
)

func TestHTTPConnect(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	errc := make(chan error, 1)
	go func() {
		buf := make([]byte, 256)
		n, err := b.Read(buf)
		if err != nil {
			errc <- err
			return
		}
		s := string(buf[:n])
		if !strings.Contains(s, "CONNECT 1.2.3.4:443 HTTP/1.1") || !strings.Contains(s, "Host: 1.2.3.4:443") {
			errc <- errString(s)
			return
		}
		_, _ = io.WriteString(b, "HTTP/1.1 200 Connection established\r\n\r\n")
		errc <- nil
	}()
	c, err := httpCONNECT(a, net.TCPAddr{IP: net.IPv4(1, 2, 3, 4), Port: 443})
	if err != nil {
		t.Fatal(err)
	}
	if c == nil {
		t.Fatal("nil conn")
	}
	if err := <-errc; err != nil {
		t.Fatal(err)
	}
}

func TestSOCKS5Connect(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	errc := make(chan error, 1)
	go func() {
		var hello [3]byte
		if _, err := io.ReadFull(b, hello[:]); err != nil {
			errc <- err
			return
		}
		if hello[0] != 5 || hello[1] != 1 || hello[2] != 0 {
			errc <- errString("bad hello")
			return
		}
		if _, err := b.Write([]byte{5, 0}); err != nil {
			errc <- err
			return
		}
		req := make([]byte, 10)
		if _, err := io.ReadFull(b, req); err != nil {
			errc <- err
			return
		}
		if req[0] != 5 || req[1] != 1 || req[3] != 1 || req[4] != 1 || req[7] != 4 || req[8] != 1 || req[9] != 187 {
			errc <- errString("bad req")
			return
		}
		if _, err := b.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0}); err != nil {
			errc <- err
			return
		}
		errc <- nil
	}()
	if err := socks5CONNECT(a, net.TCPAddr{IP: net.IPv4(1, 2, 3, 4), Port: 443}); err != nil {
		t.Fatal(err)
	}
	if err := <-errc; err != nil {
		t.Fatal(err)
	}
}

type errString string

func (e errString) Error() string { return string(e) }
