//go:build windows

package action

import (
	"fmt"
	"io"
)

func clearScreen(w io.Writer, rl cleaner) error {
	return fmt.Errorf("not implemented on windows. see https://github.com/ergochat/readline/issues/36")
}
