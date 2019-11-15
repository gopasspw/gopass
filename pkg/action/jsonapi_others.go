// +build !windows

package action

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/pkg/config"
	"github.com/gopasspw/gopass/pkg/jsonapi/manifest"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/termio"

	"github.com/fatih/color"
	"github.com/urfave/cli"
)

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

	wrapperPath, err := s.getWrapperPath(ctx, c, config.Directory(), manifest.WrapperName)
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

	if os.Getenv("GNUPGHOME") != "" {
		out.Yellow(ctx, "You seem to have GNUPGHOME set. If you intend to use the path in GNUPGHOME, you need to manually add:\n"+
			"\n  export GNUPGHOME=/path/to/gpg-home\n\n to the wrapper script")
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
