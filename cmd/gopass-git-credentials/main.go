package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/urfave/cli/v2"
)

const (
	name = "gopass-git-credentials"
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

	gp, err := gopass.New(ctx)
	if err != nil {
		panic(err)
	}

	gc := &gc{
		gp: gp,
	}

	app := cli.NewApp()
	app.Name = name
	app.Version = "0.0.1"
	app.Usage = `Use "!gopass-git-credentials $@" as git's credential.helper`
	app.Description = "" +
		"This command allows you to cache your git-credentials with gopass." +
		"Activate by using `git config --global credential.helper \"!gopass-git-credentials $@\"`"
	app.EnableBashCompletion = true
	app.Commands = []*cli.Command{
		{
			Name:   "get",
			Hidden: true,
			Action: gc.Get,
			Before: gc.Before,
		},
		{
			Name:   "store",
			Hidden: true,
			Action: gc.Store,
			Before: gc.Before,
		},
		{
			Name:   "erase",
			Hidden: true,
			Action: gc.Erase,
			Before: gc.Before,
		},
		{
			Name:        "configure",
			Description: "This command configures gopass-git-credential as git's credential.helper",
			Action:      gc.Configure,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "global",
					Usage: "Install for current user",
				},
				&cli.BoolFlag{
					Name:  "local",
					Usage: "Install for current repository only",
				},
				&cli.BoolFlag{
					Name:  "system",
					Usage: "Install for all users, requires superuser rights",
				},
			},
		},
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
