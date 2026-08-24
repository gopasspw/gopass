//go:build !windows

package termio

import (
	"os"

	"golang.org/x/term"
)

// makeRaw puts the terminal in raw mode so individual key presses can be
// read without line buffering. It returns the previous state to be restored
// with restore.
func makeRaw() (*termState, error) {
	st, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	return &termState{st: st}, nil
}

// restore restores the terminal to the given state.
func restore(state *termState) error {
	if state == nil {
		return nil
	}

	return term.Restore(int(os.Stdin.Fd()), state.st)
}

// termState wraps the terminal state for raw mode.
type termState struct {
	st *term.State
}
