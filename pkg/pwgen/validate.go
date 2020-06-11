package pwgen

import (
	"strings"
)

// containsAllClasses validates that the password contains at least one
// character from each given character class. Can also contain other classes.
func containsAllClasses(pw string, classes ...string) bool {
CLASSES:
	for _, class := range classes {
		for _, ch := range class {
			if strings.Contains(pw, string(ch)) {
				continue CLASSES
			}
		}
		return false
	}
	return true
}

// containsOnlyClasses validates that the password only contains characters
// from the given classes. Must not satisfy all classes.
func containsOnlyClasses(pw string, classes ...string) bool {
	for _, c := range pw {
		for _, class := range classes {
			if !strings.Contains(class, string(c)) {
				return false
			}
		}
	}
	return true
}

func containsMaxConsecutive(pw string, n int) bool {
	last := ""
	repCnt := 1
	for _, r := range pw {
		if last == string(r) {
			repCnt++
			if repCnt >= n {
				return false
			}
		} else {
			repCnt = 1
		}
		last = string(r)
	}
	return true
}
