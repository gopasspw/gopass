package action

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/updater"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/protect"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// Version prints the gopass version
func (s *Action) Version(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	version := make(chan string, 1)
	go s.checkVersion(ctx, version)

	_ = s.Initialized(c)

	cli.VersionPrinter(c)

	cryptoVer := versionInfo(ctx, s.Store.Crypto(ctx, ""))
	storageVer := versionInfo(ctx, s.Store.Storage(ctx, ""))

	tpl := "%-10s - %10s - %10s\n"
	fmt.Fprintf(stdout, tpl, "<root>", cryptoVer, storageVer)

	// report all used crypto, sync and fs backends
	for _, mp := range s.Store.MountPoints() {
		cv := versionInfo(ctx, s.Store.Crypto(ctx, mp))
		sv := versionInfo(ctx, s.Store.Storage(ctx, mp))

		if cv != cryptoVer || sv != storageVer {
			fmt.Fprintf(stdout, tpl, mp, cv, sv)
		}
	}

	fmt.Fprintf(stdout, "Available Crypto Backends: %s\n", strings.Join(backend.CryptoBackends(), ", "))
	fmt.Fprintf(stdout, "Available Storage Backends: %s\n", strings.Join(backend.StorageBackends(), ", "))

	select {
	case vi := <-version:
		if vi != "" {
			fmt.Fprintln(stdout, vi)
		}
	case <-time.After(2 * time.Second):
		out.Error(ctx, "Version check timed out")
	case <-ctx.Done():
		return ExitError(ExitAborted, nil, "user aborted")
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

	r, err := updater.LatestRelease(len(s.version.Pre) > 0)
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
