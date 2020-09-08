package elsearm

import "testing"

func TestZero(t *testing.T) {
	z := Zero()
	if z == nil || *z != 0 {
		t.Errorf("Zero must be zero ptr")
	}
}
