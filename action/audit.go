package action

import (
	"fmt"
	"io"
	"os"

	"github.com/muesli/crunchy"
	"github.com/urfave/cli"
)

// Audit validates passwords against common flaws
func (s *Action) Audit(c *cli.Context) error {
	t, err := s.Store.Tree()
	if err != nil {
		return err
	}

	validator := crunchy.NewValidator()
	var out io.Writer
	out = os.Stdout

	foundWeakPasswords := false
	for _, secret := range t.List(0) {
		content, err := s.Store.Get(secret)
		if err != nil {
			return err
		}

		if err = validator.Check(string(content)); err != nil {
			foundWeakPasswords = true
			fmt.Fprintf(out, "Detected weak password for %s: %v\n", secret, err)
		}
	}

	if !foundWeakPasswords {
		fmt.Fprintln(out, "No weak passwords detected.")
	}

	return nil
}
