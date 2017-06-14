package root

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

func TestSortByLen(t *testing.T) {
	in := []string{
		"a",
		"bb",
		"ccc",
		"dddd",
	}
	out := []string{
		"dddd",
		"ccc",
		"bb",
		"a",
	}
	sort.Sort(byLen(in))
	for i, s := range in {
		if out[i] != s {
			t.Errorf("Mismatch at pos %d (%s - %s)", i, out[i], s)
		}
	}
}
