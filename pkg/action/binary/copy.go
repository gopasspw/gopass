package binary

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/gopasspw/gopass/pkg/action"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/store/secret"
	"github.com/gopasspw/gopass/pkg/store/sub"

	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
)

// Copy copies either from the filesystem to the store or from the store
// to the filesystem
func Copy(ctx context.Context, c *cli.Context, store storer) error {
	from := c.Args().Get(0)
	to := c.Args().Get(1)

	// argument checking is in s.binaryCopy
	if err := binaryCopy(ctx, c, from, to, false, store); err != nil {
		return action.ExitError(ctx, action.ExitUnknown, err, "%s", err)
	}
	return nil
}

// Move works like Copy but will remove (shred/wipe) the source
// after a successful copy. Mostly useful for securely moving secrets into
// the store if they are no longer needed / wanted on disk afterwards
func Move(ctx context.Context, c *cli.Context, store storer) error {
	from := c.Args().Get(0)
	to := c.Args().Get(1)

	// argument checking is in s.binaryCopy
	if err := binaryCopy(ctx, c, from, to, true, store); err != nil {
		return action.ExitError(ctx, action.ExitUnknown, err, "%s", err)
	}
	return nil
}

// binaryCopy implements the control flow for copy and move. We support two
// workflows:
// 1. From the filesystem to the store
// 2. From the store to the filesystem
//
// Copying secrets in the store must be done through the regular copy command
func binaryCopy(ctx context.Context, c *cli.Context, from, to string, deleteSource bool, store storer) error {
	if from == "" || to == "" {
		op := "copy"
		if deleteSource {
			op = "move"
		}
		return errors.Errorf("Usage: %s %s from to", c.App.Name, op)
	}

	switch {
	case fsutil.IsFile(from) && fsutil.IsFile(to):
		// copying from on file to another file is not supported
		return errors.New("ambiquity detected. Only from or to can be a file")
	case store.Exists(ctx, from+Suffix) && store.Exists(ctx, to+Suffix):
		// copying from one secret to another secret is not supported
		return errors.New("ambiquity detected. Either from or to must be a file")
	case fsutil.IsFile(from) && !fsutil.IsFile(to):
		return binaryCopyFromFileToStore(ctx, from, to, deleteSource, store)
	case !fsutil.IsFile(from):
		return binaryCopyFromStoreToFile(ctx, from, to, deleteSource, store)
	default:
		return errors.Errorf("ambiquity detected. Unhandled case. Please report a bug")
	}
}

func binaryCopyFromFileToStore(ctx context.Context, from, to string, deleteSource bool, store storer) error {
	// if the source is a file the destination must no to avoid ambiquities
	// if necessary this can be resolved by using a absolute path for the file
	// and a relative one for the secret
	if !strings.HasSuffix(to, Suffix) {
		to += Suffix
	}

	// copy from FS to store
	buf, err := ioutil.ReadFile(from)
	if err != nil {
		return errors.Wrapf(err, "failed to read file from '%s'", from)
	}

	if err := store.Set(sub.WithReason(ctx, fmt.Sprintf("Copied data from %s to %s", from, to)), to, secret.New("", base64.StdEncoding.EncodeToString(buf))); err != nil {
		return errors.Wrapf(err, "failed to save buffer to store")
	}

	if !deleteSource {
		return nil
	}

	// it's important that we return if the validation fails, because
	// in that case we don't want to shred our (only) copy of this data!
	if err := binaryValidate(ctx, buf, to, store); err != nil {
		return errors.Wrapf(err, "failed to validate written data")
	}
	if err := fsutil.Shred(from, 8); err != nil {
		return errors.Wrapf(err, "failed to shred data")
	}
	return nil
}

func binaryCopyFromStoreToFile(ctx context.Context, from, to string, deleteSource bool, store storer) error {
	// if the source is no file we assume it's a secret and to is a filename
	// (which may already exist or not)
	if !strings.HasSuffix(from, Suffix) {
		from += Suffix
	}

	// copy from store to FS
	buf, err := binaryGet(ctx, from, store)
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
	if err := binaryValidate(ctx, buf, from, store); err != nil {
		return errors.Wrapf(err, "failed to validate the written data")
	}
	if err := store.Delete(ctx, from); err != nil {
		return errors.Wrapf(err, "failed to delete '%s' from the store", from)
	}
	return nil
}

func binaryValidate(ctx context.Context, buf []byte, name string, store storer) error {
	h := sha256.New()
	_, _ = h.Write(buf)
	fileSum := fmt.Sprintf("%x", h.Sum(nil))

	h.Reset()

	var err error
	buf, err = binaryGet(ctx, name, store)
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

func binaryGet(ctx context.Context, name string, store storer) ([]byte, error) {
	sec, err := store.Get(ctx, name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read '%s' from the store", name)
	}
	buf, err := base64.StdEncoding.DecodeString(sec.Body())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encode to base64")
	}
	return buf, nil
}
