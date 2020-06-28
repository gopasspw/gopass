package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/gopasspw/gopass/pkg/gopass/api"
	"github.com/urfave/cli/v2"
)

const (
	name = "summon-gopass"
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
		panic(err)
	}

	gc := &gc{
		gp: gp,
	}

	app := cli.NewApp()
	app.Name = name
	app.Version = "0.0.1"
	app.Usage = `Use "summon-gopass" as provider for "summon"`
	app.Description = "" +
		"This command allows use gopass as a secret provider for summon." +
		"To use it set the 'SUMMON_PROVIDER' variable to this executabel or" +
		"copy or link it into the summon provider direcotry '/usr/local/lib/summon/'."+
		"See 'summon' documentaion for more details."
	app.Action = gc.Get
	
	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
