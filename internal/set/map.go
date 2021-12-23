package set

func Map[K comparable](in []K) map[K]bool {
	m := make(map[K]bool, len(in))
	for _, i := range in {
		m[i] = true
	}
	return m
}