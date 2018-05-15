package action

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"runtime"
	"strings"

	"github.com/justwatchcom/gopass/pkg/config"
	"github.com/justwatchcom/gopass/pkg/jsonapi"
	"github.com/justwatchcom/gopass/pkg/jsonapi/manifest"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/termio"

	"github.com/fatih/color"
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
		return ExitError(ctx, ExitIO, err, "failed to get browser: %s", err)
	}

	globalInstall, err := s.getGlobalInstall(ctx, c)
	if err != nil {
		return ExitError(ctx, ExitIO, err, "failed to get global flag: %s", err)
	}

	libPath, err := s.getLibPath(ctx, c, browser, globalInstall)
	if err != nil {
		return ExitError(ctx, ExitIO, err, "failed to get lib path: %s", err)
	}

	wrapperPath, err := s.getWrapperPath(ctx, c)
	if err != nil {
		return ExitError(ctx, ExitIO, err, "failed to get wrapper path: %s", err)
	}
	wrapperPath = filepath.Join(wrapperPath, manifest.WrapperName)

	manifestPath := c.String("manifest-path")
	if manifestPath == "" {
		p, err := manifest.Path(browser, libPath, globalInstall)
		if err != nil {
			return ExitError(ctx, ExitUnknown, err, "failed to get manifest path: %s", err)
		}
		manifestPath = p
	}

	wrap, mf, err := manifest.Render(browser, wrapperPath, c.String("gopass-path"), globalInstall)
	if err != nil {
		return ExitError(ctx, ExitUnknown, err, "failed to render manifest: %s", err)
	}

	if c.Bool("print") {
		out.Print(ctx, "Native Messaging Setup Preview:\nWrapper Script (%s):\n%s\n\nManifest File (%s):\n%s\n", wrapperPath, string(wrap), manifestPath, string(mf))
	}

	if install, err := termio.AskForBool(ctx, color.BlueString("Install manifest and wrapper?"), true); err != nil || !install {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(wrapperPath), 0755); err != nil {
		return ExitError(ctx, ExitIO, err, "failed to create wrapper path: %s", err)
	}
	if err := ioutil.WriteFile(wrapperPath, wrap, 0755); err != nil {
		return ExitError(ctx, ExitIO, err, "failed to write wrapper script: %s", err)
	}
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return ExitError(ctx, ExitIO, err, "failed to create manifest path: %s", err)
	}
	if err := ioutil.WriteFile(manifestPath, mf, 0644); err != nil {
		return ExitError(ctx, ExitIO, err, "failed to write manifest file: %s", err)
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

func (s *Action) getWrapperPath(ctx context.Context, c *cli.Context) (string, error) {
	path := c.String("path")
	if path != "" {
		return path, nil
	}
	path, err := termio.AskForString(ctx, color.BlueString("In which path should gopass_wrapper.sh be installed?"), config.Directory())
	if err != nil {
		return "", errors.Wrapf(err, "failed to ask for user input")
	}
	return path, nil
}
