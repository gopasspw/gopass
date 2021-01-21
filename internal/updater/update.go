package updater

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"runtime"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/pkg/errors"

	"github.com/blang/semver/v4"
)

var (
	// UpdateMoveAfterQuit is exported for testing
	UpdateMoveAfterQuit = true
)

// Update will start th interactive update assistant
func Update(ctx context.Context, currentVersion semver.Version) error {
	if err := IsUpdateable(ctx); err != nil {
		out.Error(ctx, "Your gopass binary is externally managed. Can not update: %q", err)
		return err
	}

	dest, err := executable(ctx)
	if err != nil {
		return err
	}

	rel, err := FetchLatestRelease(ctx)
	if err != nil {
		return err
	}

	debug.Log("Current: %s - Latest: %s", currentVersion.String(), rel.Version.String())
	// binary is newer or equal to the latest release -> nothing to do
	if currentVersion.GTE(rel.Version) {
		out.Print(ctx, "gopass is up to date (%s)", currentVersion.String())
		return nil
	}

	debug.Log("downloading SHA256SUMS ...")
	_, sha256sums, err := downloadAsset(ctx, rel.Assets, "SHA256SUMS")
	if err != nil {
		return err
	}

	debug.Log("downloading SHA256SUMS.sig ...")
	_, sig, err := downloadAsset(ctx, rel.Assets, "SHA256SUMS.sig")
	if err != nil {
		return err
	}

	debug.Log("verifying GPG signature ...")
	ok, err := gpgVerify(sha256sums, sig)
	if err != nil {
		return errors.Wrapf(err, "signature verification failed")
	}
	if !ok {
		return fmt.Errorf("GPG signature verification for SHA256SUMS failed")
	}
	debug.Log("GPG signature OK!")

	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}

	suffix := fmt.Sprintf("%s-%s.%s", runtime.GOOS, runtime.GOARCH, ext)
	debug.Log("downloading tarball %q ...", suffix)
	dlFilename, buf, err := downloadAsset(ctx, rel.Assets, suffix)
	if err != nil {
		return err
	}

	debug.Log("finding hashsum entry for %q", dlFilename)
	wantHash, err := findHashForFile(sha256sums, dlFilename)
	if err != nil {
		return err
	}

	debug.Log("calculating hashsum of downloaded archive ...")
	gotHash := sha256.Sum256(buf)
	if !bytes.Equal(wantHash, gotHash[:]) {
		return fmt.Errorf("SHA256 hash mismatch, want %02x, got %02x", wantHash, gotHash)
	}
	debug.Log("hashsums match!")

	debug.Log("extracting binary from tarball ...")
	if err := extractFile(buf, dlFilename, dest); err != nil {
		return err
	}
	debug.Log("extracted %q to %q", dlFilename, dest)

	debug.Log("success!")
	return nil
}
