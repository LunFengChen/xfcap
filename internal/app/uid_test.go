package app

import (
	"os"
	"reflect"
	"testing"
)

func TestParseUserIDs(t *testing.T) {
	got := parseUserIDs(`
    userId=10175
    versionName=11.4.0
      IsolatedUid: 99001
      IsolatedProcess:
        uid=99002
    sharedUser=android.uid.shared userId=10012
`)
	want := []int{10012, 10175, 99001, 99002}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if got := parseUserIDs("nothing"); len(got) != 0 {
		t.Fatalf("%v", got)
	}
}

func TestParseUserIDsSkipsSmall(t *testing.T) {
	got := parseUserIDs("    uid=0\n    uid=1000\n    userId=10141\n")
	if !reflect.DeepEqual(got, []int{10141}) {
		t.Fatalf("%v", got)
	}
}

func TestFromPID(t *testing.T) {
	if _, err := FromPID(0); err == nil {
		t.Fatal("want bad pid")
	}
	u, err := FromPID(os.Getpid())
	if err != nil || u <= 0 {
		t.Fatalf("self uid %d %v", u, err)
	}
}

func TestMerge(t *testing.T) {
	got := Merge([]int{10175, 0}, []int{99001, 10175})
	if !reflect.DeepEqual(got, []int{10175, 99001}) {
		t.Fatalf("%v", got)
	}
}
