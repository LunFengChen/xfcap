package main

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	defaultHome  = "/data/local/tmp/xfcap"
	defaultRedir = 17892 // 本机监听；iptables REDIRECT 打到这里
)

type args struct {
	cmd       string
	packages  []string
	uids      []int
	pids      []int
	all       bool
	upstream  string
	redirPort int
	home      string
}

func parseArgs(argv []string) (*args, error) {
	a := &args{redirPort: defaultRedir, home: defaultHome}
	i := 0
	for i < len(argv) {
		s := argv[i]
		switch {
		case s == "--help":
			a.cmd = "help"
			return a, nil
		case s == "-p" || s == "--package":
			v, n, err := take(argv, i, "-p")
			if err != nil {
				return nil, err
			}
			a.packages = append(a.packages, v)
			i = n
		// -h 是电脑 mitm 地址，不是 --help。
		case s == "-h" || s == "--host":
			v, n, err := take(argv, i, "-h")
			if err != nil {
				return nil, err
			}
			a.upstream, i = v, n
		case s == "--uid":
			v, n, err := take(argv, i, "--uid")
			if err != nil {
				return nil, err
			}
			u, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("bad uid")
			}
			a.uids = append(a.uids, u)
			i = n
		case s == "--pid":
			v, n, err := take(argv, i, "--pid")
			if err != nil {
				return nil, err
			}
			p, err := strconv.Atoi(v)
			if err != nil || p <= 0 {
				return nil, fmt.Errorf("bad pid")
			}
			a.pids = append(a.pids, p)
			i = n
		case s == "--all":
			a.all = true
			i++
		case s == "--home":
			v, n, err := take(argv, i, "--home")
			if err != nil {
				return nil, err
			}
			a.home, i = v, n
		case strings.HasPrefix(s, "-"):
			return nil, fmt.Errorf("unknown flag %s", s)
		default:
			if a.cmd == "" {
				a.cmd = s
				i++
				continue
			}
			if a.cmd == "uid" {
				a.packages = append(a.packages, s)
				i++
				continue
			}
			return nil, fmt.Errorf("unknown arg %s", s)
		}
	}
	if a.cmd == "" {
		a.cmd = "help"
	}
	return a, nil
}

func (a *args) requireCapture() error {
	if !a.all && len(a.packages) == 0 && len(a.uids) == 0 && len(a.pids) == 0 {
		return fmt.Errorf("need -p PKG")
	}
	if a.upstream == "" {
		return fmt.Errorf("need -h HOST:PORT")
	}
	return nil
}

func take(argv []string, i int, flag string) (string, int, error) {
	if i+1 >= len(argv) {
		return "", i, fmt.Errorf("need %s value", flag)
	}
	return argv[i+1], i + 2, nil
}
