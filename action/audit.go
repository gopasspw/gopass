package action

import (
	"fmt"
	"strings"

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
	dupes := make(map[string][]string)
	foundWeakPasswords := false

	for _, secret := range t.List(0) {
		content, err := s.Store.Get(secret)
		if err != nil {
			return err
		}

		pw := string(content)
		if err = validator.Check(pw); err != nil {
			foundWeakPasswords = true
			fmt.Printf("Detected weak password for %s: %v\n", secret, err)
		}

		dupes[pw] = append(dupes[pw], secret)
	}

	if !foundWeakPasswords {
		fmt.Println("No weak passwords detected.")
	}
	for _, dupe := range dupes {
		if len(dupe) > 1 {
			fmt.Printf("Detected a shared password for %s\n", strings.Join(dupe, ", "))
		}
	}

	return nil
}
