// +build !linux,!windows

package action

import "os"

func getEditor() string {
	if ed := os.Getenv("EDITOR"); ed != "" {
		return ed
	}
	// given, this is a very opinionated default, but this should be available
	// on virtualy any UNIX system and the user can still set EDITOR to get
	// his favorite one
	return "vi"
}
