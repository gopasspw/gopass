// Package pwgen implements the subcommands to operate the stand alone password generator.
// The reason why it's not part of the action package is that we did try to split that
// but ran into issues and undid most of that work - except this package. If this bothers
// you feel free to propose a PR to move it back into the action package.
package pwgen

import (
	"context"
	"strconv"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/pwgen"
	"github.com/gopasspw/gopass/pkg/pwgen/xkcdgen"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

// Pwgen handles the pwgen subcommand.
func Pwgen(ctx context.Context, cmd *cli.Command) error {
	pwLen := 12
	if lenStr := cmd.Args().Get(0); lenStr != "" {
		i, err := strconv.Atoi(lenStr)
		if err != nil {
			return exit.Error(exit.Usage, err, "Failed to convert password length arg: %s", err)
		}
		if i > 0 {
			pwLen = i
		}
	}

	pwNum := 10
	if numStr := cmd.Args().Get(1); numStr != "" {
		i, err := strconv.Atoi(numStr)
		if err != nil {
			return exit.Error(exit.Usage, err, "Failed to convert password number arg: %s", err)
		}
		if i > 0 {
			pwNum = i
		}
	}

	if cmd.Bool("xkcd") || cmd.Bool("xkcd-capitalize") || cmd.Bool("xkcd-numbers") {
		return xkcdGen(ctx, cmd, pwLen, pwNum)
	}

	return pwGen(ctx, cmd, pwLen, pwNum)
}

func xkcdGen(ctx context.Context, cmd *cli.Command, length, num int) error {
	sep := config.String(ctx, "pwgen.xkcd-sep")
	if cmd.IsSet("xkcd-sep") {
		sep = cmd.String("xkcd-sep")
	}
	lang := config.String(ctx, "pwgen.xkcd-lang")
	if cmd.IsSet("xkcd-lang") {
		lang = cmd.String("xkcd-lang")
	}
	if length < 1 {
		length = config.Int(ctx, "pwgen.xkcd-len")
		if length < 1 {
			length = 4
		}
	}
	capitalize := config.Bool(ctx, "pwgen.xkcd-capitalize")
	if cmd.IsSet("xkcd-capitalize") {
		capitalize = cmd.Bool("xkcd-capitalize")
	}
	numbers := config.Bool(ctx, "pwgen.xkcd-numbers")
	if cmd.IsSet("xkcd-numbers") {
		numbers = cmd.Bool("xkcd-numbers")
	}

	for range num {
		s, err := xkcdgen.RandomLengthDelim(length, sep, lang, capitalize, numbers)
		if err != nil {
			return err
		}
		out.Print(ctx, s)
	}

	return nil
}

func pwGen(ctx context.Context, cmd *cli.Command, pwLen, pwNum int) error {
	perLine := numPerLine(pwLen)
	if cmd.Bool("one-per-line") {
		perLine = 1
	}

	charset := pwgen.CharAlphaNum

	switch {
	case cmd.Bool("no-numerals") && cmd.Bool("no-capitalize"):
		charset = pwgen.Lower
	case cmd.Bool("no-numerals"):
		charset = pwgen.CharAlpha
	case cmd.Bool("no-capitalize"):
		charset = pwgen.Digits + pwgen.Lower
	}

	if cmd.Bool("ambiguous") {
		charset = pwgen.Prune(charset, pwgen.Ambiq)
	}

	if cmd.Bool("symbols") {
		charset += pwgen.Syms
	}

	for range pwNum {
		for range perLine {
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
