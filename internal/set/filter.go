package set

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
