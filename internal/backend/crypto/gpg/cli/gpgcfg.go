package cli

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
)

func gpgConfigLoc() string {
	if sv := os.Getenv("GNUPGHOME"); sv != "" {
		return filepath.Join(sv, "gpg.conf")
	}

	uhd, _ := os.UserHomeDir()
	return filepath.Join(uhd, ".gnupg", "gpg.conf")
}

func fileContains(path, needle string) bool {
	fh, err := os.Open(path)
	if err != nil {
		debug.Log("failed to open %q for reading: %s", path, err)
		return false
	}
	defer fh.Close()

	s := bufio.NewScanner(fh)
	for s.Scan() {
		if strings.Contains(s.Text(), needle) {
			return true
		}
	}
	return false
}
