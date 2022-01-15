package set

// Filter filters all r's from the input list.
func Filter[K comparable](in []K, r ...K) []K {
	rs := Map(r)
	var out []K
	for _, i := range in {
		if !rs[i] {
			out = append(out, i)
		}
	}
	return out
}
