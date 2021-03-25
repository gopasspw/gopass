package diff

// List returnes the number of items added to and removed from the first to
// the second list
func List(l, r []string) (int, int) {
	ml := listToMap(l)
	mr := listToMap(r)

	var removed int
	for k := range ml {
		if _, found := mr[k]; !found {
			removed++
		}
	}
	var added int
	for k := range mr {
		if _, found := ml[k]; !found {
			added++
		}
	}

	return added, removed
}

func listToMap(l []string) map[string]struct{} {
	m := make(map[string]struct{}, len(l))
	for _, e := range l {
		m[e] = struct{}{}
	}
	return m
}
