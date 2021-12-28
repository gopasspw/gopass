package set

import (
	"sort"

	"golang.org/x/exp/maps"
)

type ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

// Sorted returns a sorted set of the input.
func Sorted[K ordered](l []K) []K {
	return SortedFiltered(l, func(k K) bool {
		return true
	})
}

// SortedFiltered returns a sorted set of the input, filtered by the predicate.
func SortedFiltered[K ordered](l []K, want func(K) bool) []K {
	if len(l) == 0 {
		return l
	}

	// deduplicate
	m := make(map[K]struct{}, len(l))
	for _, k := range l {
		if !want(k) {
			continue
		}
		m[k] = struct{}{}
	}
	// sort
	keys := maps.Keys(m)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return keys
}
