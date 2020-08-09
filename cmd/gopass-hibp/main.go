package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/gopass/api"
	"github.com/urfave/cli/v2"
)

const (
	name = "gopass-hibp"
)

var (
	// Version is the released version of gopass
	version string
)

func main() {
	ctx := context.Background()

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(ctx)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	defer func() {
		signal.Stop(sigChan)
		cancel()
	}()
	go func() {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
		}
	}()

	gp, err := api.New(ctx)
	if err != nil {
		out.Red(ctx, "Failed to initialize gopass API: %s", err)
		os.Exit(1)
	}

	hibp := &hibp{
		gp: gp,
	}

	app := cli.NewApp()
	app.Name = name
	app.Version = version
	app.Usage = "haveibeenpwned.com leak checker for gopass"
	app.EnableBashCompletion = true
	app.Commands = []*cli.Command{
		{
			Name:  "api",
			Usage: "Detect leaked passwords using the HIBPv2 API",
			Description: "" +
				"This command will decrypt all secrets and check the passwords against the public " +
				"havibeenpwned.com v2 API.",
			Action: func(c *cli.Context) error {
				return hibp.CheckAPI(c.Context, c.Bool("force"))
			},
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "force",
					Aliases: []string{"f"},
					Usage:   "Force checking secrets against the public API",
				},
			},
		},
		{
			Name:  "dump",
			Usage: "Detect leaked passwords using the HIBP SHA-1 dumps",
			Description: "" +
				"This command will decrypt all secrets and check the passwords against the " +
				"havibeenpwned.com SHA-1 dumps (ordered by hash). " +
				"To use the dumps you need to download the dumps from https://haveibeenpwned.com/passwords first. Be sure to grab the one that says '(ordered by hash)'. " +
				"This is a very expensive operation, for advanced users. " +
				"Most users should probably use the API. " +
				"If you want to use the dumps you need to use 7z to extract the dump: 7z x pwned-passwords-ordered-2.0.txt.7z.",
			Action: func(c *cli.Context) error {
				return hibp.CheckDump(c.Context, c.Bool("force"), c.StringSlice("files"))
			},
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "force",
					Aliases: []string{"f"},
					Usage:   "Force checking secrets against the dumps",
				},
				&cli.StringSliceFlag{
					Name:  "files",
					Usage: "One or more HIBP v1/v2 dumps",
				},
			},
		},
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
