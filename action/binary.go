package action

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/store/sub"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	// BinarySuffix is the suffix that is appended to binaries in the store
	BinarySuffix = ".b64"
)

// BinaryCat prints to or reads from STDIN/STDOUT
func (s *Action) BinaryCat(ctx context.Context, c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return s.exitError(ctx, ExitNoName, nil, "need a name")
	}
	if !strings.HasSuffix(name, BinarySuffix) {
		name += BinarySuffix
	}

	// handle pipe to stdin
	info, err := os.Stdin.Stat()
	if err != nil {
		return s.exitError(ctx, ExitIO, err, "failed to stat stdin: %s", err)
	}

	// if content is piped to stdin, read and save it
	if info.Mode()&os.ModeCharDevice == 0 {
		content := &bytes.Buffer{}

		if written, err := io.Copy(content, os.Stdin); err != nil {
			return s.exitError(ctx, ExitIO, err, "Failed to copy after %d bytes: %s", written, err)
		}

		return s.Store.Set(
			sub.WithReason(ctx, "Read secret from STDIN"),
			name,
			secret.New("", base64.StdEncoding.EncodeToString(content.Bytes())),
		)
	}

	buf, err := s.binaryGet(ctx, name)
	if err != nil {
		return s.exitError(ctx, ExitDecrypt, err, "failed to read secret: %s", err)
	}
	color.Yellow(string(buf))
	return nil
}

// BinarySum decodes binary content and computes the SHA256 checksum
func (s *Action) BinarySum(ctx context.Context, c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return s.exitError(ctx, ExitUsage, nil, "Usage: %s binary sha256 name", s.Name)
	}
	if !strings.HasSuffix(name, BinarySuffix) {
		name += BinarySuffix
	}

	buf, err := s.binaryGet(ctx, name)
	if err != nil {
		return s.exitError(ctx, ExitDecrypt, err, "failed to read secret: %s", err)
	}

	h := sha256.New()
	_, _ = h.Write(buf)
	fmt.Println(color.YellowString("%x", h.Sum(nil)))

	return nil
}

// BinaryCopy copies either from the filesystem to the store or from the store
// to the filesystem
func (s *Action) BinaryCopy(ctx context.Context, c *cli.Context) error {
	from := c.Args().Get(0)
	to := c.Args().Get(1)

	if err := s.binaryCopy(ctx, from, to, false); err != nil {
		return s.exitError(ctx, ExitUnknown, err, "%s", err)
	}
	return nil
}

// BinaryMove works like BinaryCopy but will remove (shred/wipe) the source
// after a successful copy. Mostly useful for securely moving secrets into
// the store if they are no longer needed / wanted on disk afterwards
func (s *Action) BinaryMove(ctx context.Context, c *cli.Context) error {
	from := c.Args().Get(0)
	to := c.Args().Get(1)

	if err := s.binaryCopy(ctx, from, to, true); err != nil {
		return s.exitError(ctx, ExitUnknown, err, "%s", err)
	}
	return nil
}

// binaryCopy implements the control flow for copy and move. We support two
// workflows:
// 1. From the filesystem to the store
// 2. From the store to the filesystem
//
// Copying secrets in the store must be done through the regular copy command
func (s *Action) binaryCopy(ctx context.Context, from, to string, deleteSource bool) error {
	if from == "" || to == "" {
		op := "copy"
		if deleteSource {
			op = "move"
		}
		return errors.Errorf("Usage: %s binary %s from to", s.Name, op)
	}
	switch {
	case fsutil.IsFile(from) && fsutil.IsFile(to):
		// copying from on file to another file is not supported
		return errors.New("ambiquity detected. Only from or to can be a file")
	case s.Store.Exists(from) && s.Store.Exists(to):
		// copying from one secret to another secret is not supported
		return errors.New("ambiquity detected. Either from or to must be a file")
	case fsutil.IsFile(from) && !fsutil.IsFile(to):
		// if the source is a file the destination must no to avoid ambiquities
		// if necessary this can be resolved by using a absolute path for the file
		// and a relative one for the secret
		if !strings.HasSuffix(to, BinarySuffix) {
			to += BinarySuffix
		}
		// copy from FS to store
		buf, err := ioutil.ReadFile(from)
		if err != nil {
			return errors.Wrapf(err, "failed to read file from '%s'", from)
		}
		if err := s.Store.Set(sub.WithReason(ctx, fmt.Sprintf("Copied data from %s to %s", from, to)), to, secret.New("", base64.StdEncoding.EncodeToString(buf))); err != nil {
			return errors.Wrapf(err, "failed to save buffer to store")
		}
		if deleteSource {
			// it's important that we return if the validation fails, because
			// in that case we don't want to shred our (only) copy of this data!
			if err := s.binaryValidate(ctx, buf, to); err != nil {
				return errors.Wrapf(err, "failed to validate written data")
			}
			if err := fsutil.Shred(from, 8); err != nil {
				return errors.Wrapf(err, "failed to shred data")
			}
		}
		return nil
	case !fsutil.IsFile(from):
		// if the source is no file we assume it's a secret and to is a filename
		// (which may already exist or not)
		if !strings.HasSuffix(from, BinarySuffix) {
			from += BinarySuffix
		}
		// copy from store to FS
		buf, err := s.binaryGet(ctx, from)
		if err != nil {
			return errors.Wrapf(err, "failed to read data from '%s'", from)
		}
		if err := ioutil.WriteFile(to, buf, 0600); err != nil {
			return errors.Wrapf(err, "failed to write data to '%s'", to)
		}
		if deleteSource {
			// as before: if validation of the written data fails, we MUST NOT
			// delete the (only) source
			if err := s.binaryValidate(ctx, buf, from); err != nil {
				return errors.Wrapf(err, "failed to validate the written data")
			}
			if err := s.Store.Delete(ctx, from); err != nil {
				return errors.Wrapf(err, "failed to delete '%s' from the store", from)
			}
		}
		return nil
	default:
		return errors.Errorf("ambiquity detected. Unhandled case. Please report a bug")
	}
}

func (s *Action) binaryValidate(ctx context.Context, buf []byte, name string) error {
	h := sha256.New()
	_, _ = h.Write(buf)
	fileSum := fmt.Sprintf("%x", h.Sum(nil))

	h.Reset()

	var err error
	buf, err = s.binaryGet(ctx, name)
	if err != nil {
		return errors.Wrapf(err, "failed to read '%s' from the store", name)
	}
	_, _ = h.Write(buf)
	storeSum := fmt.Sprintf("%x", h.Sum(nil))

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
	buf, err := base64.StdEncoding.DecodeString(sec.Body())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encode to base64")
	}
	return buf, nil
}
