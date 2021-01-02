package action

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/secrets"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/pkg/errors"

	"github.com/urfave/cli/v2"
)

var (
	binstdin = os.Stdin
)

// Cat prints to or reads from STDIN/STDOUT
func (s *Action) Cat(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()
	if name == "" {
		return ExitError(ExitNoName, nil, "Usage: %s cat <NAME>", c.App.Name)
	}

	// handle pipe to stdin
	info, err := binstdin.Stat()
	if err != nil {
		return ExitError(ExitIO, err, "failed to stat stdin: %s", err)
	}

	// if content is piped to stdin, read and save it
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
	ct := http.DetectContentType(in)

	debug.Log("Read %d bytes of %s from %s to %s", len(in), ct, src, dst)

	sec := secrets.NewKV()
	if err := sec.Set("Content-Type", ct); err != nil {
		debug.Log("Failed to set Content-Type %q: %q", ct, err)
	}
	if err := sec.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(src))); err != nil {
		debug.Log("Failed to set Content-Disposition: %q", err)
	}

	if strings.HasPrefix(ct, "text/") {
		sec.Write(in)
	} else {
		sec.Write([]byte(base64.StdEncoding.EncodeToString(in)))
		if err := sec.Set("Content-Transfer-Encoding", "Base64"); err != nil {
			debug.Log("Failed to set Content-Transfer-Encoding: %q", err)
		}
	}

	return sec
}

// BinaryCopy copies either from the filesystem to the store or from the store
// to the filesystem
func (s *Action) BinaryCopy(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	from := c.Args().Get(0)
	to := c.Args().Get(1)

	// argument checking is in s.binaryCopy
	if err := s.binaryCopy(ctx, c, from, to, false); err != nil {
		return ExitError(ExitUnknown, err, "%s", err)
	}
	return nil
}

// BinaryMove works like Copy but will remove (shred/wipe) the source
// after a successful copy. Mostly useful for securely moving secrets into
// the store if they are no longer needed / wanted on disk afterwards
func (s *Action) BinaryMove(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	from := c.Args().Get(0)
	to := c.Args().Get(1)

	// argument checking is in s.binaryCopy
	if err := s.binaryCopy(ctx, c, from, to, true); err != nil {
		return ExitError(ExitUnknown, err, "%s", err)
	}
	return nil
}

// binaryCopy implements the control flow for copy and move. We support two
// workflows:
// 1. From the filesystem to the store
// 2. From the store to the filesystem
//
// Copying secrets in the store must be done through the regular copy command
func (s *Action) binaryCopy(ctx context.Context, c *cli.Context, from, to string, deleteSource bool) error {
	if from == "" || to == "" {
		op := "copy"
		if deleteSource {
			op = "move"
		}
		return errors.Errorf("Usage: %s fs%s from to", c.App.Name, op)
	}

	switch {
	case fsutil.IsFile(from) && fsutil.IsFile(to):
		// copying from on file to another file is not supported
		return errors.New("ambiquity detected. Only from or to can be a file")
	case s.Store.Exists(ctx, from) && s.Store.Exists(ctx, to):
		// copying from one secret to another secret is not supported
		return errors.New("ambiquity detected. Either from or to must be a file")
	case fsutil.IsFile(from) && !fsutil.IsFile(to):
		return s.binaryCopyFromFileToStore(ctx, from, to, deleteSource)
	case !fsutil.IsFile(from):
		return s.binaryCopyFromStoreToFile(ctx, from, to, deleteSource)
	default:
		return errors.Errorf("ambiquity detected. Unhandled case. Please report a bug")
	}
}

func (s *Action) binaryCopyFromFileToStore(ctx context.Context, from, to string, deleteSource bool) error {
	// if the source is a file the destination must no to avoid ambiquities
	// if necessary this can be resolved by using a absolute path for the file
	// and a relative one for the secret

	// copy from FS to store
	buf, err := ioutil.ReadFile(from)
	if err != nil {
		return errors.Wrapf(err, "failed to read file from '%s'", from)
	}

	if err := s.Store.Set(
		ctxutil.WithCommitMessage(ctx, fmt.Sprintf("Copied data from %s to %s", from, to)), to, secFromBytes(to, from, buf)); err != nil {
		return errors.Wrapf(err, "failed to save buffer to store")
	}

	if !deleteSource {
		return nil
	}

	// it's important that we return if the validation fails, because
	// in that case we don't want to shred our (only) copy of this data!
	if err := s.binaryValidate(ctx, buf, to); err != nil {
		return errors.Wrapf(err, "failed to validate written data")
	}
	if err := fsutil.Shred(from, 8); err != nil {
		return errors.Wrapf(err, "failed to shred data")
	}
	return nil
}

func (s *Action) binaryCopyFromStoreToFile(ctx context.Context, from, to string, deleteSource bool) error {
	// if the source is no file we assume it's a secret and to is a filename
	// (which may already exist or not)

	// copy from store to FS
	buf, err := s.binaryGet(ctx, from)
	if err != nil {
		return errors.Wrapf(err, "failed to read data from '%s'", from)
	}
	if err := ioutil.WriteFile(to, buf, 0600); err != nil {
		return errors.Wrapf(err, "failed to write data to '%s'", to)
	}

	if !deleteSource {
		return nil
	}

	// as before: if validation of the written data fails, we MUST NOT
	// delete the (only) source
	if err := s.binaryValidate(ctx, buf, from); err != nil {
		return errors.Wrapf(err, "failed to validate the written data")
	}
	if err := s.Store.Delete(ctx, from); err != nil {
		return errors.Wrapf(err, "failed to delete '%s' from the store", from)
	}
	return nil
}

func (s *Action) binaryValidate(ctx context.Context, buf []byte, name string) error {
	h := sha256.New()
	_, _ = h.Write(buf)
	fileSum := fmt.Sprintf("%x", h.Sum(nil))
	h.Reset()

	debug.Log("in: %s - '%s'", fileSum, string(buf))

	var err error
	buf, err = s.binaryGet(ctx, name)
	if err != nil {
		return errors.Wrapf(err, "failed to read '%s' from the store", name)
	}
	_, _ = h.Write(buf)
	storeSum := fmt.Sprintf("%x", h.Sum(nil))

	debug.Log("store: %s - '%s'", storeSum, string(buf))

	if fileSum != storeSum {
		return errors.Errorf("Hashsum mismatch (file: %s, store: %s)", fileSum, storeSum)
	}
	return nil
}

func (s *Action) binaryGet(ctx context.Context, name string) ([]byte, error) {
	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read '%s' from the store", name)
	}

	if cte, _ := sec.Get("content-transfer-encoding"); cte != "Base64" {
		return []byte(sec.Body()), nil
	}

	buf, err := base64.StdEncoding.DecodeString(sec.Body())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encode to base64")
	}
	return buf, nil
}

// Sum decodes binary content and computes the SHA256 checksum
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
	out.Yellow(ctx, "%x", h.Sum(nil))

	return nil
}
