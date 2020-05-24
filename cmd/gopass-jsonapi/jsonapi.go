package main

import (
	"context"
	"os"

	"runtime"
	"strings"

	"github.com/blang/semver"
	"github.com/gopasspw/gopass/cmd/gopass-jsonapi/internal/jsonapi"
	"github.com/gopasspw/gopass/cmd/gopass-jsonapi/internal/jsonapi/manifest"
	"github.com/gopasspw/gopass/internal/termio"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var (
	stdin  = os.Stdin
	stdout = os.Stdout
)

type jsonapiCLI struct {
	gp gopass.Store
}

// listen reads a json message on stdin and responds on stdout
func (s *jsonapiCLI) listen(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	api := jsonapi.API{Store: s.gp, Reader: stdin, Writer: stdout, Version: semver.Version{}}
	if err := api.ReadAndRespond(ctx); err != nil {
		return api.RespondError(err)
	}
	return nil
}

func (s *jsonapiCLI) getBrowser(ctx context.Context, c *cli.Context) (string, error) {
	browser := c.String("browser")
	if browser != "" {
		return browser, nil
	}

	browser, err := termio.AskForString(ctx, color.BlueString("For which browser do you want to install gopass native messaging? [%s]", strings.Join(manifest.ValidBrowsers(), ",")), manifest.DefaultBrowser)
	if err != nil {
		return "", errors.Wrapf(err, "failed to ask for user input")
	}
	if !manifest.ValidBrowser(browser) {
		return "", errors.Errorf("%s not one of %s", browser, strings.Join(manifest.ValidBrowsers(), ","))
	}
	return browser, nil
}

func (s *jsonapiCLI) getGlobalInstall(ctx context.Context, c *cli.Context) (bool, error) {
	if !c.IsSet("global") {
		return termio.AskForBool(ctx, color.BlueString("Install for all users? (might require sudo gopass)"), false)
	}
	return c.Bool("global"), nil
}

func (s *jsonapiCLI) getLibPath(ctx context.Context, c *cli.Context, browser string, global bool) (string, error) {
	if !c.IsSet("libpath") && runtime.GOOS == "linux" && browser == "firefox" && global {
		return termio.AskForString(ctx, color.BlueString("What is your lib path?"), "/usr/lib")
	}
	return c.String("libpath"), nil
}

func (s *jsonapiCLI) getWrapperPath(ctx context.Context, c *cli.Context, defaultWrapperPath string, wrapperName string) (string, error) {
	path := c.String("path")
	if path != "" {
		return path, nil
	}
	path, err := termio.AskForString(ctx, color.BlueString("In which path should %s be installed?", wrapperName), defaultWrapperPath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to ask for user input")
	}
	return path, nil
}
