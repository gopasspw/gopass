package binary

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gopasspw/gopass/internal/action"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/store/secret"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/urfave/cli/v2"
)

var (
	stdin  = os.Stdin
	stdout = os.Stdout
)

// Cat prints to or reads from STDIN/STDOUT
func Cat(c *cli.Context, store storer) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()
	if name == "" {
		return action.ExitError(action.ExitNoName, nil, "Usage: %s binary cat <NAME>", c.App.Name)
	}

	if !strings.HasSuffix(name, Suffix) {
		name += Suffix
	}

	// handle pipe to stdin
	info, err := stdin.Stat()
	if err != nil {
		return action.ExitError(action.ExitIO, err, "failed to stat stdin: %s", err)
	}

	// if content is piped to stdin, read and save it
	if info.Mode()&os.ModeCharDevice == 0 {
		debug.Log("Reading from STDIN ...")
		content := &bytes.Buffer{}

		if written, err := io.Copy(content, stdin); err != nil {
			return action.ExitError(action.ExitIO, err, "Failed to copy after %d bytes: %s", written, err)
		}

		debug.Log("Read %d bytes from STDIN to %s", content.Len(), name)
		return store.Set(
			ctxutil.WithCommitMessage(ctx, "Read secret from STDIN"),
			name,
			secret.New("", base64.StdEncoding.EncodeToString(content.Bytes())),
		)
	}

	buf, err := binaryGet(ctx, name, store)
	if err != nil {
		return action.ExitError(action.ExitDecrypt, err, "failed to read secret: %s", err)
	}

	fmt.Fprint(stdout, string(buf))
	return nil
}
