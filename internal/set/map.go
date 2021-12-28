package set

// Map takes a slice of a given type and create a boolean map with keys
// of that type
func Map[K comparable](in []K) map[K]bool {
	m := make(map[K]bool, len(in))
	for _, i := range in {
		m[i] = true
	}
	return m
}
