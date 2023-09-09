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
	"strings"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/urfave/cli/v2"
)

var binstdin = os.Stdin

// Cat prints to or reads from STDIN/STDOUT.
func (s *Action) Cat(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()
	if name == "" {
		return exit.Error(exit.NoName, nil, "Usage: %s cat <NAME>", c.App.Name)
	}

	// handle pipe to stdin.
	info, err := binstdin.Stat()
	if err != nil {
		return exit.Error(exit.IO, err, "failed to stat stdin: %s", err)
	}

	// if content is piped to stdin, read and save it.
	if info.Mode()&os.ModeCharDevice == 0 {
		debug.Log("Reading from STDIN ...")
		content := &bytes.Buffer{}

		if written, err := io.Copy(content, binstdin); err != nil {
			return exit.Error(exit.IO, err, "Failed to copy after %d bytes: %s", written, err)
		}

		sec, err := secFromBytes(name, "STDIN", content.Bytes())
		if err != nil {
			return exit.Error(exit.IO, err, "Failed to parse secret from STDIN: %v", err)
		}
		if err = s.Store.Set(
			ctxutil.WithCommitMessage(ctx, "Read secret from STDIN"),
			name,
			sec,
		); err != nil {
			return exit.Error(exit.Unknown, err, "Failed to write secret from STDIN: %v", err)
		}

		return nil
	}

	buf, err := s.binaryGet(ctx, name)
	if err != nil {
		return exit.Error(exit.Decrypt, err, "failed to read secret: %s", err)
	}
	debug.Log("read %d decoded bytes from secret %s", len(buf), name)

	fmt.Fprint(stdout, string(buf))

	return nil
}

func secFromBytes(dst, src string, in []byte) (gopass.Secret, error) {
	debug.Log("Read %d bytes from %s to %s", len(in), src, dst)

	sec := secrets.NewAKV()
	if err := sec.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(src))); err != nil {
		debug.Log("Failed to set Content-Disposition: %q", err)
	}
	if err := sec.Set("Content-Transfer-Encoding", "Base64"); err != nil {
		debug.Log("Failed to set Content-Transfer-Encoding: %q", err)
	}

	var written int
	encoder := base64.NewEncoder(base64.StdEncoding, sec)
	n, err := encoder.Write(in)
	if err != nil {
		debug.Log("Failed to write to base64 encoder: %v", err)

		return sec, err
	}
	written += n

	if err := encoder.Close(); err != nil {
		debug.Log("Failed to finalize base64 payload: %v", err)

		return sec, err
	}
	n, err = sec.Write([]byte("\n"))
	if err != nil {
		debug.Log("Failed to write to secret: %v", err)

		return sec, err
	}
	written += n

	debug.Log("Wrote %d bytes of Base64 encoded bytes to secret", written)

	return sec, nil
}

// BinaryCopy copies either from the filesystem to the store or from the store.
// to the filesystem.
func (s *Action) BinaryCopy(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	from := c.Args().Get(0)
	to := c.Args().Get(1)

	// argument checking is in s.binaryCopy.
	if err := s.binaryCopy(ctx, c, from, to, false); err != nil {
		return exit.Error(exit.Unknown, err, "%s", err)
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
		return exit.Error(exit.Unknown, err, "%s", err)
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

	sec, err := secFromBytes(to, from, buf)
	if err != nil {
		return fmt.Errorf("failed to parse secret from input: %w", err)
	}
	if err := s.Store.Set(
		ctxutil.WithCommitMessage(ctx, fmt.Sprintf("Copied data from %s to %s", from, to)), to, sec); err != nil {
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
	if err := os.WriteFile(to, buf, 0o600); err != nil {
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

func isBase64Encoded(sec gopass.Secret) bool {
	for _, k := range []string{
		"Content-Transfer-Encoding",
		"content-transfer-encoding",
	} {
		cte, _ := sec.Get(k)
		if strings.ToLower(cte) == "base64" {
			return true
		}
	}

	return false
}

func (s *Action) binaryGet(ctx context.Context, name string) ([]byte, error) {
	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q from the store: %w", name, err)
	}

	if !isBase64Encoded(sec) {
		debug.Log("handling non-base64 secret")

		// need to use sec.Bytes() otherwise the first line is missing.
		return sec.Bytes(), nil
	}

	debug.Log("decoding Base64 encoded secret")
	body := sec.Body()
	buf, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return nil, fmt.Errorf("failed to encode to base64: %w", err)
	}

	debug.Log("decoded %d Base64 chars into %d bytes", len(body), len(buf))
	if len(buf) < 1 {
		debug.Log("body:\n%v", body)
	}

	return buf, nil
}

// Sum decodes binary content and computes the SHA256 checksum.
func (s *Action) Sum(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()
	if name == "" {
		return exit.Error(exit.Usage, nil, "Usage: %s sha256 name", c.App.Name)
	}

	buf, err := s.binaryGet(ctx, name)
	if err != nil {
		return exit.Error(exit.Decrypt, err, "failed to read secret: %s", err)
	}

	h := sha256.New()
	_, _ = h.Write(buf)
	out.Printf(ctx, "%x", h.Sum(nil))

	return nil
}
