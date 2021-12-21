package diff

// Stat returnes the number of items added to and removed from the first to
// the second list
func Stat[K comparable](l, r []K) (int, int) {
	added, removed := List(l, r)
	return len(added), len(removed)
}

// List returns two lists, the first one contains the items that were added from left
// to right, the second one contains the items that were removed from left to right.
func List[K comparable](l, r []K) ([]K, []K) {
	ml := listToMap(l)
	mr := listToMap(r)

	var added []K
	for k := range mr {
		if _, found := ml[k]; !found {
			added = append(added, k)
		}
	}
	var removed []K
	for k := range ml {
		if _, found := mr[k]; !found {
			removed = append(removed, k)
		}
	}

	return added, removed
}

func listToMap[K comparable](l []K) map[K]struct{} {
	m := make(map[K]struct{}, len(l))
	for _, e := range l {
		m[e] = struct{}{}
	}
	return m
}
