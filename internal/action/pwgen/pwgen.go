package pwgen

import (
	"fmt"
	"strconv"

	"github.com/gopasspw/gopass/internal/action"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/pwgen"
	"github.com/gopasspw/gopass/pkg/pwgen/xkcdgen"
	"github.com/gopasspw/gopass/pkg/termutil"
	"github.com/urfave/cli/v2"
)

// Pwgen handles the pwgen subcommand
func Pwgen(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)

	pwLen := 12
	if lenStr := c.Args().Get(0); lenStr != "" {
		i, err := strconv.Atoi(lenStr)
		if err != nil {
			return action.ExitError(ctx, action.ExitUsage, err, "Failed to convert password length arg: %s", err)
		}
		if i > 0 {
			pwLen = i
		}
	}

	pwNum := 10
	if numStr := c.Args().Get(1); numStr != "" {
		i, err := strconv.Atoi(numStr)
		if err != nil {
			return action.ExitError(ctx, action.ExitUsage, err, "Failed to convert password number arg: %s", err)
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
		s, err := xkcdgen.RandomLengthDelim(4, c.String("xkcdsep"), c.String("xkcdlang"))
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
	if c.Bool("no-numerals") {
		charset = pwgen.CharAlpha
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
	_, cols := termutil.GetTermsize()
	if cols < 1 {
		return 1
	}

	return cols / (pwLen + 1)
}
