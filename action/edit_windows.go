// +build windows

package action

import "os"

func getEditor() string {
	if ed := os.Getenv("EDITOR"); ed != "" {
		return ed
	}
	return "notepad.exe"
}
