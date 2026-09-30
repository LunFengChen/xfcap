package redir

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

func dialUpstream(up Upstream, dst net.TCPAddr) (net.Conn, error) {
	c, err := net.DialTimeout("tcp", up.Addr(), 10*time.Second)
	if err != nil {
		return nil, err
	}
	_ = c.SetDeadline(time.Now().Add(15 * time.Second))
	var ready net.Conn
	switch up.Kind {
	case "socks5":
		err = socks5CONNECT(c, dst)
		ready = c
	default:
		ready, err = httpCONNECT(c, dst)
	}
	if err != nil {
		c.Close()
		return nil, err
	}
	_ = ready.SetDeadline(time.Time{})
	return ready, nil
}

func httpCONNECT(c net.Conn, dst net.TCPAddr) (net.Conn, error) {
	host := net.JoinHostPort(dst.IP.String(), strconv.Itoa(dst.Port))
	req := "CONNECT " + host + " HTTP/1.1\r\nHost: " + host + "\r\n\r\n"
	if _, err := io.WriteString(c, req); err != nil {
		return nil, err
	}
	br := bufio.NewReader(c)
	line, err := br.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if !strings.Contains(line, " 200") {
		return nil, fmt.Errorf("CONNECT %s", strings.TrimSpace(line))
	}
	for {
		h, err := br.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if h == "\r\n" || h == "\n" {
			break
		}
	}
	// 读完 HTTP 头后 bufio 可能已经吞了 TLS ClientHello，必须垫回去。
	if br.Buffered() == 0 {
		return c, nil
	}
	return &prefixConn{Conn: c, r: io.MultiReader(br, c)}, nil
}

type prefixConn struct {
	net.Conn
	r io.Reader
}

func (p *prefixConn) Read(b []byte) (int, error) { return p.r.Read(b) }

func socks5CONNECT(c net.Conn, dst net.TCPAddr) error {
	if _, err := c.Write([]byte{5, 1, 0}); err != nil {
		return err
	}
	var hello [2]byte
	if _, err := io.ReadFull(c, hello[:]); err != nil {
		return err
	}
	if hello[0] != 5 || hello[1] != 0 {
		return fmt.Errorf("socks5 hello")
	}
	ip4 := dst.IP.To4()
	if ip4 == nil {
		return fmt.Errorf("need ipv4 dest")
	}
	req := []byte{5, 1, 0, 1, ip4[0], ip4[1], ip4[2], ip4[3], byte(dst.Port >> 8), byte(dst.Port)}
	if _, err := c.Write(req); err != nil {
		return err
	}
	var hdr [4]byte
	if _, err := io.ReadFull(c, hdr[:]); err != nil {
		return err
	}
	if hdr[1] != 0 {
		return fmt.Errorf("socks5 status %d", hdr[1])
	}
	switch hdr[3] {
	case 1:
		_, err := io.ReadFull(c, make([]byte, 6))
		return err
	case 4:
		_, err := io.ReadFull(c, make([]byte, 18))
		return err
	case 3:
		var n [1]byte
		if _, err := io.ReadFull(c, n[:]); err != nil {
			return err
		}
		_, err := io.ReadFull(c, make([]byte, int(n[0])+2))
		return err
	default:
		return fmt.Errorf("socks5 atyp %d", hdr[3])
	}
}
