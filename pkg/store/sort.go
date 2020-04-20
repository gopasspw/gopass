package store

import (
	"os"
	"strings"
)

// ByPathLen sorts mount points by the number of level / path separators
type ByPathLen []string

var sep = string(os.PathSeparator)

func (s ByPathLen) Len() int { return len(s) }

func (s ByPathLen) Less(i, j int) bool {
	return strings.Count(s[i], sep) < strings.Count(s[j], sep)
}

func (s ByPathLen) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// ByLen is a list of mount points (string) that can be sorted by length
type ByLen []string

// Len return the number of mount points in the list
func (s ByLen) Len() int { return len(s) }

// Less returns if a Mount point is shorter than another
func (s ByLen) Less(i, j int) bool { return len(s[i]) > len(s[j]) }

// Swap Mount Point in the list of Mount Points.
func (s ByLen) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
