// +build linux

package action

import (
	"os"
	"os/exec"
)

func getEditor() string {
	if ed := os.Getenv("EDITOR"); ed != "" {
		return ed
	}
	if p, err := exec.LookPath("editor"); err == nil {
		return p
	}
	// if neither EDITOR is set nor "editor" available we'll just assume that vi
	// is installed. If this fails the user will have to set $EDITOR
	return "vi"
}
