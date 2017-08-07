package action

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/muesli/crunchy"
	"github.com/muesli/goprogressbar"
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

	bar := &goprogressbar.ProgressBar{
		Text:    "Secrets checked",
		Total:   int64(len(pwList)),
		Current: 0,
		Width:   120,
	}
	for _, secret := range pwList {
		bar.Current++
		bar.Text = fmt.Sprintf("%d of %d secrets checked", bar.Current, bar.Total)
		bar.LazyPrint()

		content, err := s.Store.GetFirstLine(secret)
		if err != nil {
			bar.Clear()
			fmt.Println(color.RedString("Failed to retrieve secret '%s': %s", secret, err))
			continue
		}

		pw := string(content)
		if err = validator.Check(pw); err != nil {
			foundWeakPasswords = true
			bar.Clear()
			fmt.Println(color.CyanString("Detected weak secret for '%s': %v", secret, err))
		}

		dupes[pw] = append(dupes[pw], secret)
	}

	if !foundWeakPasswords {
		fmt.Println(color.GreenString("No weak secrets detected."))
	} else {
		bar.Clear()
	}

	foundDupes := false
	for _, dupe := range dupes {
		if len(dupe) > 1 {
			foundDupes = true
			fmt.Println(color.CyanString("Detected a shared secret for %s", strings.Join(dupe, ", ")))
		}
	}
	if !foundDupes {
		fmt.Println(color.GreenString("No shared secrets found."))
	}

	if foundWeakPasswords || foundDupes {
		os.Exit(1)
	}

	return nil
}
