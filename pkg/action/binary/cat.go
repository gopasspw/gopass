package binary

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"os"
	"strings"

	"github.com/justwatchcom/gopass/pkg/action"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/store/secret"
	"github.com/justwatchcom/gopass/pkg/store/sub"

	"github.com/urfave/cli"
)

// Cat prints to or reads from STDIN/STDOUT
func Cat(ctx context.Context, c *cli.Context, store storer) error {
	name := c.Args().First()
	if name == "" {
		return action.ExitError(ctx, action.ExitNoName, nil, "Usage: %s binary cat <NAME>", c.App.Name)
	}

	if !strings.HasSuffix(name, Suffix) {
		name += Suffix
	}

	// handle pipe to stdin
	info, err := os.Stdin.Stat()
	if err != nil {
		return action.ExitError(ctx, action.ExitIO, err, "failed to stat stdin: %s", err)
	}

	// if content is piped to stdin, read and save it
	if info.Mode()&os.ModeCharDevice == 0 {
		content := &bytes.Buffer{}

		if written, err := io.Copy(content, os.Stdin); err != nil {
			return action.ExitError(ctx, action.ExitIO, err, "Failed to copy after %d bytes: %s", written, err)
		}

		return store.Set(
			sub.WithReason(ctx, "Read secret from STDIN"),
			name,
			secret.New("", base64.StdEncoding.EncodeToString(content.Bytes())),
		)
	}

	buf, err := binaryGet(ctx, name, store)
	if err != nil {
		return action.ExitError(ctx, action.ExitDecrypt, err, "failed to read secret: %s", err)
	}

	out.Yellow(ctx, string(buf))
	return nil
}
