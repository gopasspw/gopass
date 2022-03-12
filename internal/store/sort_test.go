package store

import (
	"sort"
	"testing"
)

func TestMountPointSort(t *testing.T) {
	t.Parallel()

	mps := []string{
		"sub2",
		"sub1isveryverylong",
		"sub2/sub3",
	}

	sort.Sort(ByPathLen(mps))

	for i, v := range []string{
		"sub2",
		"sub1isveryverylong",
		"sub2/sub3",
	} {
		t.Logf("[%d] %s - Want: %s", i, mps[i], v)

		if mps[i] != v {
			t.Errorf("Mismatch at %d: %s vs. %s", i, v, mps[i])
		}
	}
}

func TestMountPointReverseSort(t *testing.T) {
	t.Parallel()

	mps := []string{
		"sub2",
		"sub1isveryverylong",
		"sub2/sub3",
	}

	sort.Sort(sort.Reverse(ByPathLen(mps)))

	for i, v := range []string{
		"sub2/sub3",
		"sub2",
		"sub1isveryverylong",
	} {
		t.Logf("[%d] %s - Want: %s", i, mps[i], v)

		if mps[i] != v {
			t.Errorf("Mismatch at %d: %s vs. %s", i, v, mps[i])
		}
	}
}

func TestSortByLen(t *testing.T) {
	t.Parallel()

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

	sort.Sort(ByLen(in))

	for i, s := range in {
		if out[i] != s {
			t.Errorf("Mismatch at pos %d (%s - %s)", i, out[i], s)
		}
	}
}
