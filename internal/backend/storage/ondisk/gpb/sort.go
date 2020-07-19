package gpb

// ByRevision sorts to latest revision to the top, i.e. [0]
type ByRevision []*Revision

func (r ByRevision) Len() int      { return len(r) }
func (r ByRevision) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r ByRevision) Less(i, j int) bool {
	return r[i].Created.AsTime().Before(r[j].Created.AsTime())
}
