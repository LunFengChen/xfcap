package app

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

// UIDs 是这个包会发 TCP 的全部 Linux UID：主应用、:push、isolatedProcess、sharedUserId。
// iptables owner 只认 UID，不认 PID；:push 多数和主进程同 UID，isolated 才是另一档。
func UIDs(pkg string) ([]int, error) {
	out, err := exec.Command("dumpsys", "package", pkg).Output()
	seen := map[int]struct{}{}
	if err == nil {
		for _, u := range parseUserIDs(string(out)) {
			seen[u] = struct{}{}
		}
	}
	if u, err := dataDirUID(pkg); err == nil {
		seen[u] = struct{}{}
	}
	for _, u := range runningUIDs(pkg) {
		seen[u] = struct{}{}
	}
	if len(seen) == 0 {
		return nil, fmt.Errorf("package not found: %s", pkg)
	}
	return sorted(seen), nil
}

// UID 只回主 userId。多 UID 用 UIDs。
func UID(pkg string) (int, error) {
	us, err := UIDs(pkg)
	if err != nil {
		return 0, err
	}
	return us[0], nil
}

func FromPID(pid int) (int, error) {
	if pid <= 0 {
		return 0, fmt.Errorf("bad pid")
	}
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0, fmt.Errorf("pid %d not found", pid)
	}
	for _, line := range strings.Split(string(b), "\n") {
		if !strings.HasPrefix(line, "Uid:") {
			continue
		}
		f := strings.Fields(line)
		if len(f) < 2 {
			break
		}
		u, err := strconv.Atoi(f[1])
		if err != nil || u <= 0 {
			return 0, fmt.Errorf("pid %d has no uid", pid)
		}
		return u, nil
	}
	return 0, fmt.Errorf("pid %d has no uid", pid)
}

func parseUserIDs(dumpsys string) []int {
	seen := map[int]struct{}{}
	for _, key := range []string{"userId=", "appId=", "IsolatedUid=", "IsolatedUid:", "uid="} {
		s := dumpsys
		for {
			i := strings.Index(s, key)
			if i < 0 {
				break
			}
			if i > 0 && isAlpha(s[i-1]) {
				s = s[i+1:]
				continue
			}
			s = strings.TrimLeft(s[i+len(key):], " \t")
			n, ok := leadingInt(s)
			if ok && appUID(n) {
				seen[n] = struct{}{}
			}
		}
	}
	return sorted(seen)
}

func isAlpha(c byte) bool {
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z'
}

func appUID(n int) bool {
	if n >= 10000 && n <= 19999 {
		return true
	}
	if n >= 99000 && n <= 99999 {
		return true
	}
	return false
}

func dataDirUID(pkg string) (int, error) {
	st, err := os.Stat("/data/user/0/" + pkg)
	if err != nil {
		return 0, err
	}
	sys, ok := st.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("stat uid")
	}
	return int(sys.Uid), nil
}

func runningUIDs(pkg string) []int {
	ents, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}
	seen := map[int]struct{}{}
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		b, err := os.ReadFile("/proc/" + e.Name() + "/cmdline")
		if err != nil || len(b) == 0 {
			continue
		}
		name := string(b)
		if i := strings.IndexByte(name, 0); i >= 0 {
			name = name[:i]
		}
		if name != pkg && !strings.HasPrefix(name, pkg+":") {
			continue
		}
		u, err := FromPID(pid)
		if err != nil || !appUID(u) {
			continue
		}
		seen[u] = struct{}{}
	}
	return sorted(seen)
}

func Merge(a, b []int) []int {
	seen := map[int]struct{}{}
	for _, u := range a {
		if u > 0 {
			seen[u] = struct{}{}
		}
	}
	for _, u := range b {
		if u > 0 {
			seen[u] = struct{}{}
		}
	}
	return sorted(seen)
}

func leadingInt(s string) (int, bool) {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 {
		return 0, false
	}
	n, err := strconv.Atoi(s[:i])
	return n, err == nil
}

func sorted(seen map[int]struct{}) []int {
	out := make([]int, 0, len(seen))
	for u := range seen {
		out = append(out, u)
	}
	sort.Ints(out)
	return out
}
