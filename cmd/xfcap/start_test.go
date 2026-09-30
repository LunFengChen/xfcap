package main

import (
	"testing"

	"github.com/LunFengChen/xfcap/internal/capture"
)

func TestSpecFromMultiUID(t *testing.T) {
	s := specFrom([]capture.Item{
		{Pkg: "com.zhihu.android", UID: 10175, UIDs: []int{10175, 99001}, Host: "1.2.3.4:9000"},
		{UID: 10175, UIDs: []int{10175}, Host: "1.2.3.4:9000"},
	}, 17892)
	if s.Redir != 17892 || s.All {
		t.Fatalf("%+v", s)
	}
	if len(s.UIDs) != 2 {
		t.Fatalf("uids %v", s.UIDs)
	}
}
