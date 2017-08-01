package action

import (
	"fmt"
	"io"
	"os"

	"github.com/taganaka/go-cracklib"
	"github.com/urfave/cli"
)

// Check validates password against cracklib
func (s *Action) Check(c *cli.Context) error {
	t, err := s.Store.Tree()
	if err != nil {
		return err
	}

	var out io.Writer
	out = os.Stdout

	for _, secret := range t.List(0) {
		content, err := s.Store.Get(secret)
		if err != nil {
			return err
		}

		if m, ok := cracklib.FascistCheck(string(content)); !ok {
			fmt.Fprintf(out, "Weak password for %s: %s\n", secret, m)
		}
	}

	return nil
}
