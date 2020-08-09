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
	name = "summon-gopass"
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

	gc := &gc{
		gp: gp,
	}

	app := cli.NewApp()
	app.Name = name
	app.Version = version
	app.Usage = `Use "gopass-summon-provider" as provider for "summon"`
	app.Description = "" +
		"This command allows to use gopass as a secret provider for summon." +
		"To use it set the 'SUMMON_PROVIDER' variable to this executable or" +
		"copy or link it (as `gopass`) into the summon provider directory" +
		"'/usr/local/lib/summon/'. See 'summon' documentation for more details."
	app.Action = gc.Get

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
