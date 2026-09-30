package redir

import "testing"

func TestPickSameUpstream(t *testing.T) {
	s := &Server{}
	u := Upstream{Kind: "http", Host: "1.2.3.4", Port: 9000}
	s.SetRoutes(map[int]Upstream{10175: u, 10141: u}, nil)
	got, ok := s.pick(0)
	if !ok || got.Addr() != "1.2.3.4:9000" {
		t.Fatalf("%v %v", got, ok)
	}
}

func TestPickDifferentUpstream(t *testing.T) {
	s := &Server{}
	s.SetRoutes(map[int]Upstream{
		10175: {Kind: "http", Host: "1.2.3.4", Port: 9000},
		10141: {Kind: "http", Host: "1.2.3.4", Port: 9001},
	}, nil)
	if _, ok := s.pick(0); ok {
		t.Fatal("want drop")
	}
}
