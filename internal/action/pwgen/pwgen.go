// Package pwgen implements the subcommands to operate the stand alone password generator.
// The reason why it's not part of the action package is that we did try to split that
// but ran into issues and undid most of that work - except this package. If this bothers
// you feel free to propose a PR to move it back into the action package.
package pwgen

import (
	"strconv"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/pwgen"
	"github.com/gopasspw/gopass/pkg/pwgen/xkcdgen"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

// Pwgen handles the pwgen subcommand.
func Pwgen(c *cli.Context) error {
	pwLen := 12
	if lenStr := c.Args().Get(0); lenStr != "" {
		i, err := strconv.Atoi(lenStr)
		if err != nil {
			return exit.Error(exit.Usage, err, "Failed to convert password length arg: %s", err)
		}
		if i > 0 {
			pwLen = i
		}
	}

	pwNum := 10
	if numStr := c.Args().Get(1); numStr != "" {
		i, err := strconv.Atoi(numStr)
		if err != nil {
			return exit.Error(exit.Usage, err, "Failed to convert password number arg: %s", err)
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
	sep := config.String(c.Context, "pwgen.xkcd.sep")
	if c.IsSet("sep") {
		sep = c.String("sep")
	}
	lang := config.String(c.Context, "pwgen.xkcd.lang")
	if c.IsSet("lang") {
		lang = c.String("lang")
	}
	length := config.Int(c.Context, "pwgen.xkcd.len")
	if length < 1 {
		length = 4
	}
	for i := 0; i < num; i++ {
		s, err := xkcdgen.RandomLengthDelim(length, sep, lang)
		if err != nil {
			return err
		}
		out.Print(c.Context, s)
	}

	return nil
}

func pwGen(c *cli.Context, pwLen, pwNum int) error {
	ctx := c.Context

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
			ctx := out.WithNewline(ctx, false)
			out.Print(ctx, pwgen.GeneratePasswordCharset(pwLen, charset))
			out.Print(ctx, " ")
		}
		out.Print(ctx, "")
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
