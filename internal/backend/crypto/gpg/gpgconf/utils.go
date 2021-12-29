package gpgconf

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// GPGOpts parses extra GPG options from the environment.
func GPGOpts() []string {
	for _, en := range []string{"GOPASS_GPG_OPTS", "PASSWORD_STORE_GPG_OPTS"} {
		if opts := os.Getenv(en); opts != "" {
			return strings.Fields(opts)
		}
	}
	return nil
}

// gpgConfigLoc returns the location of the GPG config file.
func gpgConfigLoc() string {
	if sv := os.Getenv("GNUPGHOME"); sv != "" {
		return filepath.Join(sv, "gpg.conf")
	}

	uhd, _ := os.UserHomeDir()
	return filepath.Join(uhd, ".gnupg", "gpg.conf")
}

func Config() (map[string]string, error) {
	fh, err := os.Open(gpgConfigLoc())
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	return parseGpgConfig(fh)
}

func parseGpgConfig(fh io.Reader) (map[string]string, error) {
	vals := make(map[string]string, 20)
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// ignore comments
		if strings.HasPrefix(line, "#") {
			continue
		}
		key, val, found := strings.Cut(line, " ")
		if !found {
			continue
		}
		vals[key] = strings.TrimSpace(val)
	}

	return vals, nil
}
