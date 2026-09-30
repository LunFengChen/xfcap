package fw

import "testing"

func TestOwnerArgs(t *testing.T) {
	got := ownerArgs(Spec{All: true, UIDs: []int{0, 10210}})
	if len(got) != 1 || got[0] != appUIDAll {
		t.Fatalf("%v", got)
	}
	got = ownerArgs(Spec{UIDs: []int{0, 10210}})
	if len(got) != 1 || got[0] != "10210" {
		t.Fatalf("%v", got)
	}
}
