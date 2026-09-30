package main

import "testing"

func TestParseStart(t *testing.T) {
	a, err := parseArgs([]string{"start", "-p", "com.zhihu.android", "-h", "172.18.22.226:9000"})
	if err != nil {
		t.Fatal(err)
	}
	if a.cmd != "start" || a.packages[0] != "com.zhihu.android" || a.upstream != "172.18.22.226:9000" {
		t.Fatalf("%+v", a)
	}
	if err := a.requireCapture(); err != nil {
		t.Fatal(err)
	}
}

func TestParsePID(t *testing.T) {
	a, err := parseArgs([]string{"start", "--pid", "1234", "-h", "1.2.3.4:9000"})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.pids) != 1 || a.pids[0] != 1234 {
		t.Fatalf("%+v", a)
	}
	if err := a.requireCapture(); err != nil {
		t.Fatal(err)
	}
}

func TestParseMissingFlagValue(t *testing.T) {
	if _, err := parseArgs([]string{"start", "-p"}); err == nil {
		t.Fatal("want error")
	}
}
