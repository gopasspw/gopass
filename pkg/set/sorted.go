package set

import (
	"maps"
	"slices"

	"golang.org/x/exp/constraints"
)

// SortedKeys returns the sorted keys of the map.
// The keys are sorted in ascending order.
func SortedKeys[K constraints.Ordered, V any](m map[K]V) []K {
	// sort
	keys := maps.Keys(m)

	return slices.Sorted(keys)
}

// Sorted returns a sorted set of the input.
// Duplicates are removed.
func Sorted[K constraints.Ordered](l []K) []K {
	return SortedFiltered(l, func(k K) bool {
		return true
	})
}

// SortedFiltered returns a sorted set of the input, filtered by the predicate.
// Duplicates are removed.
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
	return SortedKeys(m)
}
