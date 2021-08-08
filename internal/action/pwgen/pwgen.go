package pwgen

import (
	"fmt"
	"strconv"

	"github.com/gopasspw/gopass/internal/action"
	"github.com/gopasspw/gopass/pkg/pwgen"
	"github.com/gopasspw/gopass/pkg/pwgen/xkcdgen"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

// Pwgen handles the pwgen subcommand
func Pwgen(c *cli.Context) error {
	pwLen := 12
	if lenStr := c.Args().Get(0); lenStr != "" {
		i, err := strconv.Atoi(lenStr)
		if err != nil {
			return action.ExitError(action.ExitUsage, err, "Failed to convert password length arg: %s", err)
		}
		if i > 0 {
			pwLen = i
		}
	}

	pwNum := 10
	if numStr := c.Args().Get(1); numStr != "" {
		i, err := strconv.Atoi(numStr)
		if err != nil {
			return action.ExitError(action.ExitUsage, err, "Failed to convert password number arg: %s", err)
		}
		if i > 0 {
			pwNum = i
		}
	}

	if c.Bool("xkcd") {
		return xkcdGen(c, pwNum)
	}

	return pwGen(c, pwLen, pwNum)
}

func xkcdGen(c *cli.Context, num int) error {
	for i := 0; i < num; i++ {
		s, err := xkcdgen.RandomLengthDelim(4, c.String("sep"), c.String("lang"))
		if err != nil {
			return err
		}
		fmt.Println(s)
	}
	return nil
}

func pwGen(c *cli.Context, pwLen, pwNum int) error {
	perLine := numPerLine(pwLen)
	if c.Bool("one-per-line") {
		perLine = 1
	}
	charset := pwgen.CharAlphaNum
	switch {
	case c.Bool("no-numerals") && c.Bool("no-capitalize"):
		charset = pwgen.Lower
	case c.Bool("no-numerals"):
		charset = pwgen.CharAlpha
	case c.Bool("no-capitalize"):
		charset = pwgen.Digits + pwgen.Lower
	}
	if c.Bool("ambiguous") {
		charset = pwgen.Prune(charset, pwgen.Ambiq)
	}
	if c.Bool("symbols") {
		charset += pwgen.Syms
	}
	for i := 0; i < pwNum; i++ {
		for j := 0; j < perLine; j++ {
			fmt.Print(pwgen.GeneratePasswordCharset(pwLen, charset))
			fmt.Print(" ")
		}
		fmt.Println()
	}

	return nil
}

func numPerLine(pwLen int) int {
	cols, _, err := term.GetSize(0)
	if err != nil {
		return 1
	}

	return cols / (pwLen + 1)
}
