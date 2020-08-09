package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/gopass/api"
	"github.com/urfave/cli/v2"
)

const (
	name = "gopass-jsonapi"
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

	ja := &jsonapiCLI{
		gp: gp,
	}

	app := cli.NewApp()
	app.Name = name
	app.Version = version
	app.Usage = "Setup and run gopass-jsonapi as native messaging hosts, e.g. for browser plugins"
	app.EnableBashCompletion = true
	app.Action = func(c *cli.Context) error {
		if strings.HasSuffix(os.Args[0], "native_host") || strings.HasSuffix(os.Args[0], "native_host.exe") {
			return ja.listen(c)
		}
		return cli.ShowAppHelp(c)
	}
	app.Commands = []*cli.Command{
		{
			Name:        "listen",
			Usage:       "Listen and respond to messages via stdin/stdout",
			Description: "Gopass-jsonapi is started in listen mode from browser plugins using a wrapper specified in native messaging host manifests",
			Hidden:      true,
			Action: func(c *cli.Context) error {
				return ja.listen(c)
			},
		},
		{
			Name:        "configure",
			Usage:       "Setup gopass-jsonapi native messaging manifest for selected browser",
			Description: "To access gopass from browser plugins, a native app manifest must be installed at the correct location",
			Action: func(c *cli.Context) error {
				return ja.setup(c)
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "browser",
					Usage: "One of 'chrome' and 'firefox'",
				},
				&cli.StringFlag{
					Name:  "path",
					Usage: "Path to install 'gopass_wrapper.sh' to",
				},
				&cli.StringFlag{
					Name:  "manifest-path",
					Usage: "Path to install 'com.justwatch.gopass.json' to",
				},
				&cli.BoolFlag{
					Name:  "global",
					Usage: "Install for all users, requires superuser rights",
				},
				&cli.StringFlag{
					Name:  "libpath",
					Usage: "Library path for global installation on linux. Default is /usr/lib",
				},
				&cli.StringFlag{
					Name:  "gopass-path",
					Usage: "Path to gopass binary. Default is auto detected",
				},
				&cli.BoolFlag{
					Name:  "print",
					Usage: "Print installation summary before creating any files",
					Value: true,
				},
			},
		},
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
