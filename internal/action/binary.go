package action

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/urfave/cli/v2"
)

var (
	binstdin = os.Stdin
)

// Cat prints to or reads from STDIN/STDOUT.
func (s *Action) Cat(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()
	if name == "" {
		return ExitError(ExitNoName, nil, "Usage: %s cat <NAME>", c.App.Name)
	}

	// handle pipe to stdin.
	info, err := binstdin.Stat()
	if err != nil {
		return ExitError(ExitIO, err, "failed to stat stdin: %s", err)
	}

	// if content is piped to stdin, read and save it.
	if info.Mode()&os.ModeCharDevice == 0 {
		debug.Log("Reading from STDIN ...")
		content := &bytes.Buffer{}

		if written, err := io.Copy(content, binstdin); err != nil {
			return ExitError(ExitIO, err, "Failed to copy after %d bytes: %s", written, err)
		}

		return s.Store.Set(
			ctxutil.WithCommitMessage(ctx, "Read secret from STDIN"),
			name,
			secFromBytes(name, "STDIN", content.Bytes()),
		)
	}

	buf, err := s.binaryGet(ctx, name)
	if err != nil {
		return ExitError(ExitDecrypt, err, "failed to read secret: %s", err)
	}

	fmt.Fprint(stdout, string(buf))
	return nil
}

func secFromBytes(dst, src string, in []byte) gopass.Secret {
	debug.Log("Read %d bytes from %s to %s", len(in), src, dst)

	sec := secrets.NewKV()
	if err := sec.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(src))); err != nil {
		debug.Log("Failed to set Content-Disposition: %q", err)
	}

	sec.Write([]byte(base64.StdEncoding.EncodeToString(in)))
	if err := sec.Set("Content-Transfer-Encoding", "Base64"); err != nil {
		debug.Log("Failed to set Content-Transfer-Encoding: %q", err)
	}

	return sec
}

// BinaryCopy copies either from the filesystem to the store or from the store.
// to the filesystem.
func (s *Action) BinaryCopy(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	from := c.Args().Get(0)
	to := c.Args().Get(1)

	// argument checking is in s.binaryCopy.
	if err := s.binaryCopy(ctx, c, from, to, false); err != nil {
		return ExitError(ExitUnknown, err, "%s", err)
	}
	return nil
}

// BinaryMove works like Copy but will remove (shred/wipe) the source
// after a successful copy. Mostly useful for securely moving secrets into
// the store if they are no longer needed / wanted on disk afterwards.
func (s *Action) BinaryMove(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	from := c.Args().Get(0)
	to := c.Args().Get(1)

	// argument checking is in s.binaryCopy.
	if err := s.binaryCopy(ctx, c, from, to, true); err != nil {
		return ExitError(ExitUnknown, err, "%s", err)
	}
	return nil
}

// binaryCopy implements the control flow for copy and move. We support two
// workflows:.
// 1. From the filesystem to the store.
// 2. From the store to the filesystem.
//
// Copying secrets in the store must be done through the regular copy command.
func (s *Action) binaryCopy(ctx context.Context, c *cli.Context, from, to string, deleteSource bool) error {
	if from == "" || to == "" {
		op := "copy"
		if deleteSource {
			op = "move"
		}
		return fmt.Errorf("usage: %s fs%s from to", c.App.Name, op)
	}

	switch {
	case fsutil.IsFile(from) && fsutil.IsFile(to):
		// copying from on file to another file is not supported.
		return fmt.Errorf("ambiguity detected. Only from or to can be a file")
	case s.Store.Exists(ctx, from) && s.Store.Exists(ctx, to):
		// copying from one secret to another secret is not supported.
		return fmt.Errorf("ambiguity detected. Either from or to must be a file")
	case fsutil.IsFile(from) && !fsutil.IsFile(to):
		return s.binaryCopyFromFileToStore(ctx, from, to, deleteSource)
	case !fsutil.IsFile(from):
		return s.binaryCopyFromStoreToFile(ctx, from, to, deleteSource)
	default:
		return fmt.Errorf("ambiguity detected. Unhandled case. Please report a bug")
	}
}

func (s *Action) binaryCopyFromFileToStore(ctx context.Context, from, to string, deleteSource bool) error {
	// if the source is a file the destination must not to avoid ambiguities.
	// if necessary this can be resolved by using a absolute path for the file
	// and a relative one for the secret.

	// copy from FS to store.
	buf, err := os.ReadFile(from)
	if err != nil {
		return fmt.Errorf("failed to read file from %q: %w", from, err)
	}

	if err := s.Store.Set(
		ctxutil.WithCommitMessage(ctx, fmt.Sprintf("Copied data from %s to %s", from, to)), to, secFromBytes(to, from, buf)); err != nil {
		return fmt.Errorf("failed to save buffer to store: %w", err)
	}

	if !deleteSource {
		return nil
	}

	// it's important that we return if the validation fails, because
	// in that case we don't want to shred our (only) copy of this data!.
	if err := s.binaryValidate(ctx, buf, to); err != nil {
		return fmt.Errorf("failed to validate written data: %w", err)
	}
	if err := fsutil.Shred(from, 8); err != nil {
		return fmt.Errorf("failed to shred data: %w", err)
	}
	return nil
}

func (s *Action) binaryCopyFromStoreToFile(ctx context.Context, from, to string, deleteSource bool) error {
	// if the source is no file we assume it's a secret and to is a filename
	// (which may already exist or not).

	// copy from store to FS.
	buf, err := s.binaryGet(ctx, from)
	if err != nil {
		return fmt.Errorf("failed to read data from %q: %w", from, err)
	}
	if err := os.WriteFile(to, buf, 0600); err != nil {
		return fmt.Errorf("failed to write data to %q: %w", to, err)
	}

	if !deleteSource {
		return nil
	}

	// as before: if validation of the written data fails, we MUST NOT
	// delete the (only) source.
	if err := s.binaryValidate(ctx, buf, from); err != nil {
		return fmt.Errorf("failed to validate the written data: %w", err)
	}
	if err := s.Store.Delete(ctx, from); err != nil {
		return fmt.Errorf("failed to delete %q from the store: %w", from, err)
	}
	return nil
}

func (s *Action) binaryValidate(ctx context.Context, buf []byte, name string) error {
	h := sha256.New()
	_, _ = h.Write(buf)
	fileSum := fmt.Sprintf("%x", h.Sum(nil))
	h.Reset()

	debug.Log("in: %s - %q", fileSum, string(buf))

	var err error
	buf, err = s.binaryGet(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to read %q from the store: %w", name, err)
	}
	_, _ = h.Write(buf)
	storeSum := fmt.Sprintf("%x", h.Sum(nil))

	debug.Log("store: %s - %q", storeSum, string(buf))

	if fileSum != storeSum {
		return fmt.Errorf("hashsum mismatch (file: %s, store: %s)", fileSum, storeSum)
	}
	return nil
}

func (s *Action) binaryGet(ctx context.Context, name string) ([]byte, error) {
	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q from the store: %w", name, err)
	}

	if cte, _ := sec.Get("content-transfer-encoding"); cte != "Base64" {
		// need to use sec.Bytes() otherwise the first line is missing.
		return sec.Bytes(), nil
	}

	buf, err := base64.StdEncoding.DecodeString(sec.Body())
	if err != nil {
		return nil, fmt.Errorf("failed to encode to base64: %w", err)
	}
	return buf, nil
}

// Sum decodes binary content and computes the SHA256 checksum.
func (s *Action) Sum(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()
	if name == "" {
		return ExitError(ExitUsage, nil, "Usage: %s sha256 name", c.App.Name)
	}

	buf, err := s.binaryGet(ctx, name)
	if err != nil {
		return ExitError(ExitDecrypt, err, "failed to read secret: %s", err)
	}

	h := sha256.New()
	_, _ = h.Write(buf)
	out.Printf(ctx, "%x", h.Sum(nil))

	return nil
}
