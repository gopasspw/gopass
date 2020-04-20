package simple

import (
	"sort"
	"strings"
)

type PathSlice []string

func (p PathSlice) Len() int { return len(p) }
func (p PathSlice) Less(i, j int) bool {
	return strings.Replace(p[i], "\\", "/", -1) < strings.Replace(p[j], "\\", "/", -1)
}
func (p PathSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func sortPaths(paths []string) {
	sort.Sort(PathSlice(paths))
}
