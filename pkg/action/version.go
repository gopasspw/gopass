package action

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/protect"
	"github.com/gopasspw/gopass/pkg/updater"

	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

// Version prints the gopass version
func (s *Action) Version(ctx context.Context, c *cli.Context) error {
	version := make(chan string, 1)
	go s.checkVersion(ctx, version)

	_ = s.Initialized(ctx, c)

	cli.VersionPrinter(c)

	cryptoVer := versionInfo(ctx, s.Store.Crypto(ctx, ""))
	rcsVer := versionInfo(ctx, s.Store.RCS(ctx, ""))
	storageVer := versionInfo(ctx, s.Store.Storage(ctx, ""))

	tpl := "%-10s - %10s - %10s - %10s\n"
	fmt.Fprintf(stdout, tpl, "<root>", cryptoVer, rcsVer, storageVer)

	// report all used crypto, sync and fs backends
	for _, mp := range s.Store.MountPoints() {
		cv := versionInfo(ctx, s.Store.Crypto(ctx, mp))
		rv := versionInfo(ctx, s.Store.RCS(ctx, mp))
		sv := versionInfo(ctx, s.Store.Storage(ctx, mp))

		if cv != cryptoVer || rv != rcsVer || sv != storageVer {
			fmt.Fprintf(stdout, tpl, mp, cv, rv, sv)
		}
	}

	fmt.Fprintf(stdout, "Available Crypto Backends: %s\n", strings.Join(backend.CryptoBackends(), ", "))
	fmt.Fprintf(stdout, "Available RCS Backends: %s\n", strings.Join(backend.RCSBackends(), ", "))
	fmt.Fprintf(stdout, "Available Storage Backends: %s\n", strings.Join(backend.StorageBackends(), ", "))

	select {
	case vi := <-version:
		if vi != "" {
			fmt.Fprintln(stdout, vi)
		}
	case <-time.After(2 * time.Second):
		out.Error(ctx, "Version check timed out")
	case <-ctx.Done():
		return ExitError(ctx, ExitAborted, nil, "user aborted")
	}

	return nil
}

type versioner interface {
	Name() string
	Version(context.Context) semver.Version
}

func versionInfo(ctx context.Context, v versioner) string {
	if v == nil {
		return "<none>"
	}
	return fmt.Sprintf("%s %s", v.Name(), v.Version(ctx))
}

func (s *Action) checkVersion(ctx context.Context, u chan string) {
	if disabled := os.Getenv("CHECKPOINT_DISABLE"); disabled != "" {
		u <- ""
		return
	}

	if strings.HasSuffix(s.version.String(), "+HEAD") || protect.ProtectEnabled {
		// chan not check version against HEAD or
		// against pledge(2)'d OpenBSD
		u <- ""
		return
	}

	r, err := updater.LatestRelease(ctx, len(s.version.Pre) > 0)
	if err != nil {
		u <- color.RedString("\nError checking latest version: %s", err)
		return
	}

	if s.version.LT(r.Version()) {
		notice := fmt.Sprintf("\nYour version (%s) of gopass is out of date!\nThe latest version is %s.\n", s.version, r.Version().String())
		notice += "You can update by downloading from www.justwatch.com/gopass"
		if err := updater.IsUpdateable(ctx); err == nil {
			notice += " by running 'gopass update'"
		}
		notice += " or via your package manager"
		u <- color.YellowString(notice)
	}
	u <- ""
}
