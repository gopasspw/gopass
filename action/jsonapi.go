package action

import (
	"context"

	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/utils/jsonapi"
	"github.com/justwatchcom/gopass/utils/jsonapi/manifest"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// JSONAPI reads a json message on stdin and responds on stdout
func (s *Action) JSONAPI(ctx context.Context, c *cli.Context) error {
	api := jsonapi.API{Store: s.Store, Reader: stdin, Writer: stdout}
	if err := api.ReadAndRespond(ctx); err != nil {
		return api.RespondError(err)
	}
	return nil
}

// SetupNativeMessaging sets up manifest for gopass as native messaging host
func (s *Action) SetupNativeMessaging(ctx context.Context, c *cli.Context) error {
	browser, err := s.getBrowser(ctx, c)
	if err != nil {
		return err
	}

	globalInstall, err := s.getGlobalInstall(ctx, c)
	if err != nil {
		return err
	}

	libpath, err := s.getLibPath(ctx, c, browser, globalInstall)
	if err != nil {
		return err
	}

	wrapperPath, err := s.getWrapperPath(ctx, c)
	if err != nil {
		return err
	}

	if err := manifest.PrintSummary(browser, wrapperPath, libpath, globalInstall); err != nil {
		return err
	}

	if c.Bool("print-only") {
		return nil
	}

	install, err := s.askForBool(ctx, color.BlueString("Install manifest and wrapper?"), true)
	if install && err == nil {
		return manifest.SetUp(browser, wrapperPath, libpath, globalInstall)
	}
	return err
}

func (s *Action) getBrowser(ctx context.Context, c *cli.Context) (string, error) {
	browser := c.String("browser")
	if browser != "" {
		return browser, nil
	}

	browser, err := s.askForString(ctx, color.BlueString("For which browser do you want to install gopass native messaging? [%s]", strings.Join(manifest.ValidBrowsers[:], ",")), manifest.DefaultBrowser)
	if err != nil {
		return "", errors.Wrapf(err, "failed to ask for user input")
	}
	if !stringInSlice(browser, manifest.ValidBrowsers) {
		return "", errors.Errorf("%s not one of %s", browser, strings.Join(manifest.ValidBrowsers[:], ","))
	}
	return browser, nil
}

func (s *Action) getGlobalInstall(ctx context.Context, c *cli.Context) (bool, error) {
	if !c.IsSet("global") {
		return s.askForBool(ctx, color.BlueString("Install for all users? (might require sudo gopass)"), false)
	}
	return c.Bool("global"), nil
}

func (s *Action) getLibPath(ctx context.Context, c *cli.Context, browser string, global bool) (string, error) {
	if !c.IsSet("libpath") && runtime.GOOS == "linux" && browser == "firefox" && global {
		return s.askForString(ctx, color.BlueString("What is your lib path?"), "/usr/lib")
	}
	return c.String("libpath"), nil
}

func (s *Action) getWrapperPath(ctx context.Context, c *cli.Context) (string, error) {
	path := c.String("path")
	if path != "" {
		return path, nil
	}
	path, err := s.askForString(ctx, color.BlueString("In which path should gopass_wrapper.sh be installed?"), config.Directory())
	if err != nil {
		return "", errors.Wrapf(err, "failed to ask for user input")
	}
	return path, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
