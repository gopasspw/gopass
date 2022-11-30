package set

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// SortedKeys returns the sorted keys of the map.
func SortedKeys[K constraints.Ordered, V any](m map[K]V) []K {
	// sort
	keys := maps.Keys(m)
	slices.Sort(keys)

	return keys
}

// Sorted returns a sorted set of the input.
func Sorted[K constraints.Ordered](l []K) []K {
	return SortedFiltered(l, func(k K) bool {
		return true
	})
}

// SortedFiltered returns a sorted set of the input, filtered by the predicate.
func SortedFiltered[K constraints.Ordered](l []K, want func(K) bool) []K {
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
	slices.Sort(keys)

	return keys
}
