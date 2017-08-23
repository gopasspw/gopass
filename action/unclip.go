package action

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/atotto/clipboard"
	"github.com/urfave/cli"
)

// Unclip tries to erase the content of the clipboard
func (s *Action) Unclip(c *cli.Context) error {
	timeout := c.Int("timeout")
	checksum := os.Getenv("GOPASS_UNCLIP_CHECKSUM")

	time.Sleep(time.Second * time.Duration(timeout))

	cur, err := clipboard.ReadAll()
	if err != nil {
		return s.exitError(ExitIO, err, "failed to read clipboard: %s", err)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(cur)))

	if hash != checksum {
		return nil
	}
	if err := clipboard.WriteAll(""); err != nil {
		return s.exitError(ExitIO, err, "failed to write clipboard: %s", err)
	}

	return nil
}
