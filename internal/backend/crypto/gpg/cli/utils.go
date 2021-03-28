package cli

import (
	"os"
	"path/filepath"
	"strings"
)

func splitPacket(in string) map[string]string {
	m := make(map[string]string, 3)
	p := strings.Split(in, ":")
	if len(p) < 3 {
		return m
	}
	p = strings.Split(strings.TrimSpace(p[2]), " ")
	for i := 0; i+1 < len(p); i += 2 {
		m[p[i]] = strings.Trim(p[i+1], ",")
	}
	return m
}

// gpgConfigLoc returns the location of the GPG config file
func gpgConfigLoc() string {
	if sv := os.Getenv("GNUPGHOME"); sv != "" {
		return filepath.Join(sv, "gpg.conf")
	}

	uhd, _ := os.UserHomeDir()
	return filepath.Join(uhd, ".gnupg", "gpg.conf")
}

// GPGOpts parses extra GPG options from the environment
func GPGOpts() []string {
	for _, en := range []string{"GOPASS_GPG_OPTS", "PASSWORD_STORE_GPG_OPTS"} {
		if opts := os.Getenv(en); opts != "" {
			return strings.Fields(opts)
		}
	}
	return nil
}
