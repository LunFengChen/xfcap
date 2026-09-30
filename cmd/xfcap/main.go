// xfcap 跑在 Root 手机上：按 App UID 把 TCP REDIRECT 到本机，再转到电脑 mitm。
// 不建 tun。多次 start -p 是往列表里加 App，不是新开一份进程。
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/LunFengChen/xfcap/internal/app"
	"github.com/LunFengChen/xfcap/internal/capture"
	"github.com/LunFengChen/xfcap/internal/fw"
	"github.com/LunFengChen/xfcap/internal/redir"
)

func main() {
	a, err := parseArgs(os.Args[1:])
	if err != nil {
		fail(err.Error())
	}
	switch a.cmd {
	case "help":
		os.Stdout.WriteString("xfcap start -p PKG -h HOST:PORT\nxfcap start --pid PID -h HOST:PORT\nxfcap stop|status|doctor|uid\n")
	case "doctor":
		needRoot()
		doctor(a)
	case "uid":
		uidCmd(a)
	case "start":
		needRoot()
		start(a)
	case "stop":
		needRoot()
		stop(a)
	case "status":
		status(a)
	default:
		fail("unknown cmd")
	}
}

func needRoot() {
	if os.Geteuid() != 0 {
		fail("need root")
	}
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "xfcap:", msg)
	os.Exit(1)
}

func doctor(a *args) {
	iptv, _ := exec.Command("iptables", "--version").Output()
	own := "no"
	if fw.HasOwnerMatch() {
		own = "yes"
	}
	fmt.Printf("root uid=%d  iptables=%s  owner=%s  redir=%d\n", os.Geteuid(), strings.TrimSpace(string(iptv)), own, a.redirPort)
	if fw.Installed() {
		fmt.Println("XFCAP_TCP: yes")
	} else {
		fmt.Println("XFCAP_TCP: no")
	}
	if alivePID(a.home) > 0 {
		fmt.Printf("daemon: pid=%d\n", alivePID(a.home))
		os.Stdout.WriteString(capture.Table(capture.Load(a.home)))
	} else {
		fmt.Println("daemon: no")
	}
}

func uidCmd(a *args) {
	if len(a.packages) == 0 && len(a.pids) == 0 {
		fail("uid needs package or --pid")
	}
	for _, p := range a.packages {
		us, err := app.UIDs(p)
		if err != nil {
			fail(err.Error())
		}
		ss := make([]string, len(us))
		for i, u := range us {
			ss[i] = fmt.Sprintf("%d", u)
		}
		fmt.Printf("%s  uid=%s\n", p, strings.Join(ss, ","))
	}
	for _, pid := range a.pids {
		u, err := app.FromPID(pid)
		if err != nil {
			fail(err.Error())
		}
		fmt.Printf("pid %d  uid=%d\n", pid, u)
	}
}

func stop(a *args) {
	fw.Cleanup()
	(&redir.Server{}).Close()
	killDaemon(a.home)
	_ = capture.Save(a.home, nil)
	fmt.Println("已停止")
}

func status(a *args) {
	if alivePID(a.home) <= 0 {
		fmt.Println("没在抓")
		return
	}
	os.Stdout.WriteString(capture.Table(capture.Load(a.home)))
}
