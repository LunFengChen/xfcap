package redir

import "testing"

func TestParse(t *testing.T) {
	u, err := Parse("http://172.18.22.226:9000")
	if err != nil {
		t.Fatal(err)
	}
	if u.Kind != "http" || u.Host != "172.18.22.226" || u.Port != 9000 {
		t.Fatalf("%+v", u)
	}
	u, err = Parse("socks5://10.0.0.1:1080")
	if err != nil {
		t.Fatal(err)
	}
	if u.Kind != "socks5" || u.Port != 1080 {
		t.Fatalf("%+v", u)
	}
	if _, err := Parse("ftp://x:1"); err == nil {
		t.Fatal("want error")
	}
}
