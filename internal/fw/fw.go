// Package fw：OUTPUT 链按 UID REDIRECT TCP。不碰 tun、不劫持 DNS。
package fw

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	chainTCP   = "XFCAP_TCP"
	chainDNS   = "XFCAP_DNS"   // 旧版留下的，Cleanup 时拆掉
	chainInput = "XFCAP_INPUT" // 同上
	appUIDAll  = "10000-2147483647"
)

// 局域网不抓：电脑 mitm 在 172.16/12 里，转过去的包必须直出，否则打转。
var lanNets = []string{
	"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8",
	"169.254.0.0/16", "172.16.0.0/12", "192.168.0.0/16",
	"198.18.0.0/15", "224.0.0.0/4", "240.0.0.0/4",
}

type Spec struct {
	All   bool
	UIDs  []int
	Redir int
}

func Cleanup() {
	dropJump("nat", "OUTPUT", chainTCP)
	dropJump("nat", "OUTPUT", chainDNS)
	dropJump("filter", "INPUT", chainInput)
	flushChain("nat", chainTCP)
	flushChain("nat", chainDNS)
	flushChain("filter", chainInput)
}

func Install(s Spec) error {
	Cleanup()
	if err := ipt("-t", "nat", "-N", chainTCP); err != nil {
		return err
	}
	// uid 0 是 xfcap 自己连 mitm 的包，必须放行。
	if err := ipt("-t", "nat", "-A", chainTCP, "-m", "owner", "--uid-owner", "0", "-j", "RETURN"); err != nil {
		return err
	}
	for _, n := range lanNets {
		if err := ipt("-t", "nat", "-A", chainTCP, "-d", n, "-j", "RETURN"); err != nil {
			return err
		}
	}
	rp := strconv.Itoa(s.Redir)
	for _, uid := range ownerArgs(s) {
		if err := ipt("-t", "nat", "-A", chainTCP, "-m", "owner", "--uid-owner", uid, "-p", "tcp", "-j", "REDIRECT", "--to-ports", rp); err != nil {
			return err
		}
	}
	return ipt("-t", "nat", "-I", "OUTPUT", "1", "-j", chainTCP)
}

func HasOwnerMatch() bool {
	b, err := os.ReadFile("/proc/net/ip_tables_matches")
	if err != nil {
		return false
	}
	return strings.Contains(string(b), "owner")
}

func Installed() bool {
	return ipt("-t", "nat", "-C", "OUTPUT", "-j", chainTCP) == nil
}

func ownerArgs(s Spec) []string {
	if s.All {
		return []string{appUIDAll}
	}
	var out []string
	for _, u := range s.UIDs {
		if u > 0 {
			out = append(out, strconv.Itoa(u))
		}
	}
	return out
}

func ipt(args ...string) error {
	var last error
	for i := 0; i < 8; i++ {
		last = iptOnce(args...)
		if last == nil {
			return nil
		}
		msg := last.Error()
		if !strings.Contains(msg, "xtables lock") && !strings.Contains(msg, "temporarily unavailable") {
			return last
		}
		time.Sleep(time.Duration(i+1) * 80 * time.Millisecond)
	}
	return last
}

func iptOnce(args ...string) error {
	cmd := exec.Command("iptables", append([]string{"-w", "5"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if s := strings.TrimSpace(string(out)); s != "" {
			return &iptErr{s}
		}
		return err
	}
	return nil
}

type iptErr struct{ s string }

func (e *iptErr) Error() string { return e.s }

func dropJump(table, hook, chain string) {
	for ipt("-t", table, "-D", hook, "-j", chain) == nil {
	}
}

func flushChain(table, chain string) {
	_ = ipt("-t", table, "-F", chain)
	_ = ipt("-t", table, "-X", chain)
}
