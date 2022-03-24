package cui

import (
	"context"
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/termio"
)

// GetSelection show a navigateable multiple-choice list to the user
// and returns the selected entry along with the action.
func GetSelection(ctx context.Context, prompt string, choices []string) (string, int) {
	if ctxutil.IsAlwaysYes(ctx) || !ctxutil.IsInteractive(ctx) {
		return "impossible", 0
	}

	for i, c := range choices {
		fmt.Print(color.GreenString("[%  d]", i))
		fmt.Printf(" %s\n", c)
	}
	fmt.Println()
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
			fmt.Println(err.Error())
		}
	}
	fmt.Println(i)

	return "default", i
}
