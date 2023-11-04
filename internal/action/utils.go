//go:build !windows

package action

import "io"

func clearScreen(w io.Writer, rl cleaner) error {
	_, err := w.Write([]byte("\033[H"))
	rl.Clean()
	rl.Refresh()

	return err
}
