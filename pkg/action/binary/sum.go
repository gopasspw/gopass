package binary

import (
	"context"
	"crypto/sha256"
	"strings"

	"github.com/gopasspw/gopass/pkg/action"
	"github.com/gopasspw/gopass/pkg/out"

	"gopkg.in/urfave/cli.v1"
)

// Sum decodes binary content and computes the SHA256 checksum
func Sum(ctx context.Context, c *cli.Context, store storer) error {
	name := c.Args().First()
	if name == "" {
		return action.ExitError(ctx, action.ExitUsage, nil, "Usage: %s sha256 name", c.App.Name)
	}

	if !strings.HasSuffix(name, Suffix) {
		name += Suffix
	}

	buf, err := binaryGet(ctx, name, store)
	if err != nil {
		return action.ExitError(ctx, action.ExitDecrypt, err, "failed to read secret: %s", err)
	}

	h := sha256.New()
	_, _ = h.Write(buf)
	out.Yellow(ctx, "%x", h.Sum(nil))

	return nil
}
