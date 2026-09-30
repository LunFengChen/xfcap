package redir

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type Upstream struct {
	Kind string
	Host string
	Port int
}

func (u Upstream) Addr() string {
	return net.JoinHostPort(u.Host, strconv.Itoa(u.Port))
}

func Parse(raw string) (Upstream, error) {
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return Upstream{}, fmt.Errorf("bad host")
	}
	kind := strings.ToLower(u.Scheme)
	switch kind {
	case "http", "https":
		kind = "http"
	case "socks5", "socks5h", "socks":
		kind = "socks5"
	default:
		return Upstream{}, fmt.Errorf("host must be IP:PORT or socks5://IP:PORT")
	}
	host := u.Hostname()
	if host == "" {
		return Upstream{}, fmt.Errorf("missing host")
	}
	p := u.Port()
	if p == "" {
		return Upstream{}, fmt.Errorf("missing port")
	}
	port, err := strconv.Atoi(p)
	if err != nil || port <= 0 || port > 65535 {
		return Upstream{}, fmt.Errorf("bad port")
	}
	if ip := net.ParseIP(host); ip != nil {
		host = ip.String()
	}
	return Upstream{Kind: kind, Host: host, Port: port}, nil
}
