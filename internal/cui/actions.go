package cui

import (
	"context"
	"errors"

	"github.com/urfave/cli/v2"
)

// Action is a action which can be selected
type Action struct {
	Name string
	Fn   func(context.Context, *cli.Context) error
}

// Actions is a list of actions
type Actions []Action

// Selection return the list of actions
func (ca Actions) Selection() []string {
	keys := make([]string, 0, len(ca))
	for _, a := range ca {
		keys = append(keys, a.Name)
	}
	return keys
}

// Run executes the selected action
func (ca Actions) Run(ctx context.Context, c *cli.Context, i int) error {
	if len(ca) < i || i >= len(ca) {
		return errors.New("action not found")
	}
	if ca[i].Fn == nil {
		return errors.New("action invalid")
	}
	return ca[i].Fn(ctx, c)
}
