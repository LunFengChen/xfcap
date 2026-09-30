package capture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddAndTable(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "xfcap")
	if err := os.MkdirAll(filepath.Join(home, "run"), 0755); err != nil {
		t.Fatal(err)
	}
	cs := Add(nil, []Item{item("com.zhihu.android", "172.18.22.226:9000", 10175, 99001)})
	cs = Add(cs, []Item{item("com.android.chrome", "172.18.22.226:9000", 10120)})
	if err := Save(home, cs); err != nil {
		t.Fatal(err)
	}
	got := Load(home)
	if len(got) != 2 {
		t.Fatalf("%+v", got)
	}
	if got[0].UID != 10175 || len(got[0].AllUIDs()) != 2 {
		t.Fatalf("%+v", got[0])
	}
	s := Table(got)
	if !contains(s, "com.zhihu.android") || !contains(s, "com.android.chrome") || !contains(s, "10175,99001") {
		t.Fatal(s)
	}
}

func item(pkg, host string, uids ...int) Item {
	c := Item{Pkg: pkg, Host: host, UIDs: uids}
	if len(uids) > 0 {
		c.UID = uids[0]
	}
	return c
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || stringIndex(s, sub) >= 0)
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
