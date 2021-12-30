// Copyright 2021 The gopass Authors. All rights reserved.
// Use of this source code is governed by the MIT license,
// that can be found in the LICENSE file.

// Gopass implements the gopass command line tool.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	rdebug "runtime/debug"
	"runtime/pprof"
	"sort"
	"time"

	"github.com/blang/semver/v4"
	"github.com/fatih/color"
	ap "github.com/gopasspw/gopass/internal/action"
	"github.com/gopasspw/gopass/internal/action/pwgen"
	_ "github.com/gopasspw/gopass/internal/backend/crypto"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	_ "github.com/gopasspw/gopass/internal/backend/storage"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/queue"
	"github.com/gopasspw/gopass/internal/store/leaf"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/protect"
	"github.com/gopasspw/gopass/pkg/termio"
	colorable "github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"
)

const (
	name = "gopass"
)

var (
	// Version is the released version of gopass.
	version string
)

func main() {
	// important: execute the func now but the returned func only on defer!
	// Example: https://go.dev/play/p/8214zCX6hVq.
	defer writeCPUProfile()()

	if err := protect.Pledge("stdio rpath wpath cpath tty proc exec"); err != nil {
		panic(err)
	}
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

	cli.ErrWriter = errorWriter{
		out: colorable.NewColorableStderr(),
	}
	sv := getVersion()
	cli.VersionPrinter = makeVersionPrinter(os.Stdout, sv)

	// run the app
	q := queue.New(ctx)
	ctx = queue.WithQueue(ctx, q)
	ctx, app := setupApp(ctx, sv)
	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
	// process all pending queue items
	q.Close(ctx)
	writeMemProfile()
	// done
}

func setupApp(ctx context.Context, sv semver.Version) (context.Context, *cli.App) {
	// try to read config (if it exists)
	cfg := config.LoadWithFallback()

	// set config values
	ctx = initContext(ctx, cfg)

	// initialize action handlers
	action, err := ap.New(cfg, sv)
	if err != nil {
		out.Errorf(ctx, "failed to initialize gopass: %s", err)
		os.Exit(ap.ExitUnknown)
	}

	// set some action callbacks
	if !cfg.AutoImport {
		ctx = ctxutil.WithImportFunc(ctx, termio.AskForKeyImport)
	}
	ctx = leaf.WithFsckFunc(ctx, termio.AskForConfirmation)

	app := cli.NewApp()

	app.Name = name
	app.Version = sv.String()
	app.Usage = "The standard unix password manager - rewritten in Go"
	app.EnableBashCompletion = true
	app.BashComplete = func(c *cli.Context) {
		cli.DefaultAppComplete(c)
		action.Complete(c)
	}

	app.Flags = ap.ShowFlags()
	app.Action = func(c *cli.Context) error {
		if err := action.IsInitialized(c); err != nil {
			return err
		}

		if c.Args().Present() {
			return action.Show(c)
		}
		return action.REPL(c)
	}

	app.Commands = getCommands(action, app)
	return ctx, app
}

func getCommands(action *ap.Action, app *cli.App) []*cli.Command {
	cmds := []*cli.Command{
		{
			Name:  "completion",
			Usage: "Bash and ZSH completion",
			Description: "" +
				"Source the output of this command with bash or zsh to get auto completion",
			Subcommands: []*cli.Command{{
				Name:   "bash",
				Usage:  "Source for auto completion in bash",
				Action: action.CompletionBash,
			}, {
				Name:  "zsh",
				Usage: "Source for auto completion in zsh",
				Action: func(c *cli.Context) error {
					return action.CompletionZSH(app)
				},
			}, {
				Name:  "fish",
				Usage: "Source for auto completion in fish",
				Action: func(c *cli.Context) error {
					return action.CompletionFish(app)
				},
			}, {
				Name:  "openbsdksh",
				Usage: "Source for auto completion in OpenBSD's ksh",
				Action: func(c *cli.Context) error {
					return action.CompletionOpenBSDKsh(app)
				},
			}},
		},
	}
	cmds = append(cmds, action.GetCommands()...)
	cmds = append(cmds, pwgen.GetCommands()...)
	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name < cmds[j].Name })
	return cmds
}

func parseBuildInfo() (string, string, string) {
	bi, ok := rdebug.ReadBuildInfo()
	if !ok {
		return "HEAD", "", ""
	}
	var (
		commit string
		date   string
		dirty  string
	)
	for _, v := range bi.Settings {
		switch v.Key {
		case "gitrevision":
			commit = v.Value[len(v.Value)-8:]
		case "gitcommittime":
			if bt, err := time.Parse("2006-01-02T15:04:05Z", date); err == nil {
				date = bt.Format("2006-01-02 15:04:05")
			}
		case "gituncommitted":
			if v.Value == "true" {
				dirty = " (dirty)"
			}
		}
	}
	return commit, date, dirty
}

func makeVersionPrinter(out io.Writer, sv semver.Version) func(c *cli.Context) {
	return func(c *cli.Context) {
		commit, buildtime, dirty := parseBuildInfo()
		buildInfo := ""
		if commit != "" {
			buildInfo = commit + dirty
		}
		if buildtime != "" {
			if buildInfo != "" {
				buildInfo += " "
			}
			buildInfo += buildtime
		}
		if buildInfo != "" {
			buildInfo = "(" + buildInfo + ") "
		}
		fmt.Fprintf(
			out,
			"%s %s %s%s %s %s\n",
			name,
			sv.String(),
			buildInfo,
			runtime.Version(),
			runtime.GOOS,
			runtime.GOARCH,
		)
	}
}

type errorWriter struct {
	out io.Writer
}

func (e errorWriter) Write(p []byte) (int, error) {
	return e.out.Write([]byte("\n" + color.RedString("Error: %s", p)))
}

func initContext(ctx context.Context, cfg *config.Config) context.Context {
	// initialize from config, may be overridden by env vars
	ctx = cfg.WithContext(ctx)

	// always trust
	ctx = gpg.WithAlwaysTrust(ctx, true)

	// check recipients conflicts with always trust, make sure it's not enabled
	// when always trust is
	if gpg.IsAlwaysTrust(ctx) {
		ctx = leaf.WithCheckRecipients(ctx, false)
	}

	// only emit color codes when stdout is a terminal
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		color.NoColor = true
		ctx = ctxutil.WithTerminal(ctx, false)
		ctx = ctxutil.WithInteractive(ctx, false)
	}

	// reading from stdin?
	if info, err := os.Stdin.Stat(); err == nil && info.Mode()&os.ModeCharDevice == 0 {
		ctx = ctxutil.WithInteractive(ctx, false)
		ctx = ctxutil.WithStdin(ctx, true)
	}

	// disable colored output on windows since cmd.exe doesn't support ANSI color
	// codes. Other terminal may do, but until we can figure that out better
	// disable this for all terms on this platform
	if runtime.GOOS == "windows" {
		color.NoColor = true
	}

	return ctx
}

func writeCPUProfile() func() {
	cp := os.Getenv("GOPASS_CPU_PROFILE")
	if cp == "" {
		return func() {}
	}

	f, err := os.Create(cp)
	if err != nil {
		log.Fatalf("could not create CPU profile at %s: %s", cp, err)
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatalf("could not start CPU profile: %s", err)
	}

	return func() {
		pprof.StopCPUProfile()
		f.Close()
		debug.Log("wrote CPU profile to %s", cp)
	}
}

func writeMemProfile() {
	mp := os.Getenv("GOPASS_MEM_PROFILE")
	if mp == "" {
		return
	}
	f, err := os.Create(mp)
	if err != nil {
		log.Fatalf("could not write mem profile to %s: %s", mp, err)
	}
	defer f.Close()

	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatalf("could not write heap profile: %s", err)
	}
	debug.Log("wrote heap profile to %s", mp)
}
