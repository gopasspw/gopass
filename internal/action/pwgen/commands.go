package pwgen

import (
	"github.com/urfave/cli/v2"
)

// GetCommands returns the pwgen subcommand
func GetCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:        "pwgen",
			Usage:       "Generate passwords",
			Description: "Print any number of password to the console.",
			Action:      Pwgen,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "no-numerals",
					Aliases: []string{"0"},
					Usage:   "Do not include numerals in the generated passwords.",
				},
				&cli.BoolFlag{
					Name:    "one-per-line",
					Aliases: []string{"1"},
					Usage:   "Print one password per line",
				},
				&cli.BoolFlag{
					Name:    "xkcd",
					Aliases: []string{"x"},
					Usage:   "Use multiple random english words combined to a password. By default, space is used as separator and all words are lowercase",
				},
				&cli.StringFlag{
					Name:    "xkcdsep",
					Aliases: []string{"xs"},
					Usage:   "Word separator for generated xkcd style password. If no separator is specified, the words are combined without spaces/separator and the first character of words is capitalised. This flag implies -xkcd",
					Value:   " ",
				},
				&cli.StringFlag{
					Name:    "xkcdlang",
					Aliases: []string{"xl"},
					Usage:   "Language to generate password from, currently de (german) and en (english, default) are supported",
					Value:   "en",
				},
			},
		},
	}
}
