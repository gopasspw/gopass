package password

import (
	"sort"
	"testing"
)

func TestMountPointSort(t *testing.T) {
	mps := []string{
		"sub1",
		"sub2/sub3",
		"sub2",
	}
	sort.Sort(byLen(mps))
	for i, v := range []string{
		"sub2/sub3",
		"sub1",
		"sub2",
	} {
		if mps[i] != v {
			t.Errorf("Mismatch at %d: %s vs. %s", i, v, mps[i])
		}
	}
}
