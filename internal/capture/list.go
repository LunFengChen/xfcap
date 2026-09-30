// Package capture：正在抓的 App 列表。多次 start -p 往这里追加。
package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Item struct {
	All  bool
	Pkg  string
	UID  int   // 主 UID，兼容旧表
	UIDs []int // 这个包会发 TCP 的全部 UID（:push / isolated / sharedUserId）
	Host string
}

func (c Item) Name() string {
	if c.All {
		return "全部 App"
	}
	if c.Pkg != "" {
		return c.Pkg
	}
	if uids := c.AllUIDs(); len(uids) == 1 {
		return fmt.Sprintf("uid %d", uids[0])
	}
	if c.UID > 0 {
		return fmt.Sprintf("uid %d", c.UID)
	}
	return "uid"
}

func (c Item) AllUIDs() []int {
	if c.All {
		return nil
	}
	seen := map[int]struct{}{}
	var out []int
	add := func(u int) {
		if u <= 0 {
			return
		}
		if _, ok := seen[u]; ok {
			return
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	for _, u := range c.UIDs {
		add(u)
	}
	add(c.UID)
	return out
}

func pathOf(home string) string {
	return filepath.Join(home, "run", "apps.txt")
}

func Load(home string) []Item {
	b, err := os.ReadFile(pathOf(home))
	if err != nil {
		return nil
	}
	var out []Item
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) < 3 {
			continue
		}
		uids := parseUIDs(f[1])
		c := Item{Pkg: f[0], Host: f[2], UIDs: uids}
		if len(uids) > 0 {
			c.UID = uids[0]
		}
		if c.Pkg == "*" {
			c.All = true
			c.Pkg = ""
			c.UID = 0
			c.UIDs = nil
		}
		out = append(out, c)
	}
	return out
}

func Save(home string, cs []Item) error {
	dir := filepath.Join(home, "run")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if len(cs) == 0 {
		_ = os.Remove(pathOf(home))
		return nil
	}
	var b strings.Builder
	for _, c := range cs {
		pkg := c.Pkg
		if c.All {
			pkg = "*"
		}
		fmt.Fprintf(&b, "%s\t%s\t%s\n", pkg, formatUIDs(c), c.Host)
	}
	return os.WriteFile(pathOf(home), []byte(b.String()), 0644)
}

func Add(exist, add []Item) []Item {
	out := append([]Item{}, exist...)
	for _, n := range add {
		replaced := false
		for i, e := range out {
			if same(e, n) {
				out[i] = n
				replaced = true
				break
			}
		}
		if !replaced {
			out = append(out, n)
		}
	}
	return out
}

func same(a, b Item) bool {
	if a.All && b.All {
		return true
	}
	if a.Pkg != "" && a.Pkg == b.Pkg {
		return true
	}
	if a.Pkg == "" && b.Pkg == "" && uidKey(a) == uidKey(b) && uidKey(a) != "" {
		return true
	}
	return false
}

func uidKey(c Item) string {
	return formatUIDs(c)
}

func Table(cs []Item) string {
	if len(cs) == 0 {
		return "没在抓\n"
	}
	rows := make([][]string, 0, len(cs))
	for _, c := range cs {
		uid := formatUIDs(c)
		if uid == "" {
			uid = "-"
		}
		if c.All {
			uid = "全部"
		}
		rows = append(rows, []string{c.Name(), uid, c.Host})
	}
	return formatTable([]string{"包名", "UID", "转到"}, rows)
}

func parseUIDs(s string) []int {
	var out []int
	seen := map[int]struct{}{}
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" || p == "0" {
			continue
		}
		u, err := strconv.Atoi(p)
		if err != nil || u <= 0 {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

func formatUIDs(c Item) string {
	us := c.AllUIDs()
	if len(us) == 0 {
		return "0"
	}
	ss := make([]string, len(us))
	for i, u := range us {
		ss[i] = strconv.Itoa(u)
	}
	return strings.Join(ss, ",")
}
