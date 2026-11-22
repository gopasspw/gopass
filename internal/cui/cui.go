// Package cui provides a simple command line user interface
// for gopass. It is used to interact with the user.
package cui

import (
	"context"
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/termio"
)

// GetSelection show a navigable multiple-choice list to the user
// and returns the selected entry along with the action.
func GetSelection(ctx context.Context, prompt string, choices []string) (string, int) {
	if ctxutil.IsAlwaysYes(ctx) || !ctxutil.IsInteractive(ctx) {
		return "impossible", 0
	}

	_, errW := out.Writer(ctx)
	for i, c := range choices {
		fmt.Fprint(errW, color.GreenString("[%  d]", i))
		fmt.Fprintf(errW, " %s\n", c)
	}
	fmt.Fprintln(errW)
	var i int
	for {
		var err error
		i, err = termio.AskForInt(ctx, prompt, 0)
		if err == nil && i < len(choices) {
			break
		}
		if errors.Is(err, termio.ErrAborted) {
			return "aborted", 0
		}
		if err != nil {
			fmt.Fprintln(errW, err.Error())
		}
	}
	fmt.Fprintln(errW, i)

	return "default", i
}
