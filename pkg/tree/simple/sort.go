package simple

import (
	"sort"
	"strings"
)

type pathSlice []string

func (p pathSlice) Len() int { return len(p) }
func (p pathSlice) Less(i, j int) bool {
	return strings.Replace(p[i], "\\", "/", -1) < strings.Replace(p[j], "\\", "/", -1)
}
func (p pathSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func sortPaths(paths []string) {
	sort.Sort(pathSlice(paths))
}
