package cui

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/termio"
)

// GetSelection show a navigateable multiple-choice list to the user
// and returns the selected entry along with the action
func GetSelection(ctx context.Context, prompt string, choices []string) (string, int) {
	if ctxutil.IsAlwaysYes(ctx) || !ctxutil.IsInteractive(ctx) {
		return "impossible", 0
	}
	for i, c := range choices {
		fmt.Printf("[%  d] %s\n", i, c)
	}
	fmt.Println()
	var i int
	for {
		var err error
		i, err = termio.AskForInt(ctx, prompt, 0)
		if err == nil {
			break
		}
		if err == termio.ErrAborted {
			return "aborted", 0
		}
		fmt.Println(err.Error())
	}
	fmt.Println(i)
	return "default", i
}
