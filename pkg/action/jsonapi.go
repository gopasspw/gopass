package action

import (
	"context"

	"runtime"
	"strings"

	"github.com/gopasspw/gopass/pkg/jsonapi"
	"github.com/gopasspw/gopass/pkg/jsonapi/manifest"
	"github.com/gopasspw/gopass/pkg/termio"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
)

// JSONAPI reads a json message on stdin and responds on stdout
func (s *Action) JSONAPI(ctx context.Context, c *cli.Context) error {
	api := jsonapi.API{Store: s.Store, Reader: stdin, Writer: stdout, Version: s.version}
	if err := api.ReadAndRespond(ctx); err != nil {
		return api.RespondError(err)
	}
	return nil
}

func (s *Action) getBrowser(ctx context.Context, c *cli.Context) (string, error) {
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

func (s *Action) getGlobalInstall(ctx context.Context, c *cli.Context) (bool, error) {
	if !c.IsSet("global") {
		return termio.AskForBool(ctx, color.BlueString("Install for all users? (might require sudo gopass)"), false)
	}
	return c.Bool("global"), nil
}

func (s *Action) getLibPath(ctx context.Context, c *cli.Context, browser string, global bool) (string, error) {
	if !c.IsSet("libpath") && runtime.GOOS == "linux" && browser == "firefox" && global {
		return termio.AskForString(ctx, color.BlueString("What is your lib path?"), "/usr/lib")
	}
	return c.String("libpath"), nil
}

func (s *Action) getWrapperPath(ctx context.Context, c *cli.Context, defaultWrapperPath string, wrapperName string) (string, error) {
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
