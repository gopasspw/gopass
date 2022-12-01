package set

// Map takes a slice of a given type and create a boolean map with keys
// of that type.
func Map[K comparable](in []K) map[K]bool {
	m := make(map[K]bool, len(in))
	for _, i := range in {
		m[i] = true
	}

	return m
}

// Apply applies the given function to every element of the slice.
func Apply[K comparable](in []K, f func(K) K) []K {
	out := make([]K, len(in))
	for i, v := range in {
		out[i] = f(v)
	}

	return out
}
