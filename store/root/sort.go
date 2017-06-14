package root

// byLen is a list of mount points (string) that can be sorted by name
type byLen []string

// Len return the number of mount points in the list
func (s byLen) Len() int { return len(s) }

// Less returns if a Mount point is shorter than another
func (s byLen) Less(i, j int) bool { return len(s[i]) > len(s[j]) }

// Swap Mount Point in the list of Mount Points.
func (s byLen) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
