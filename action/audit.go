package action

import (
	"fmt"
	"os"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
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

	pwList := t.List(0)
	fmt.Printf("Checking %d secrets. This may take some time ...\n", len(pwList))
	bar := pb.StartNew(len(pwList))
	for _, secret := range pwList {
		bar.Increment()
		content, err := s.Store.GetFirstLine(secret)
		if err != nil {
			fmt.Println(color.RedString("Failed to retrieve secret '%s': %s", secret, err))
			continue
		}

		pw := string(content)
		if err = validator.Check(pw); err != nil {
			foundWeakPasswords = true
			fmt.Println(color.CyanString("Detected weak password for %s: %v\n", secret, err))
		}

		dupes[pw] = append(dupes[pw], secret)
	}
	bar.FinishPrint("Done")

	if !foundWeakPasswords {
		fmt.Println(color.GreenString("No weak passwords detected."))
	}

	foundDupes := false
	for _, dupe := range dupes {
		if len(dupe) > 1 {
			foundDupes = true
			fmt.Println(color.CyanString("Detected a shared password for %s\n", strings.Join(dupe, ", ")))
		}
	}
	if !foundDupes {
		fmt.Println(color.GreenString("No dupes found."))
	}

	if foundWeakPasswords || foundDupes {
		os.Exit(1)
	}

	return nil
}
