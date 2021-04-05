package cli

import (
	"bufio"
	"io"
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

func gpgConfig() (map[string]string, error) {
	fh, err := os.Open(gpgConfigLoc())
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	return parseGpgConfig(fh)
}

func parseGpgConfig(fh io.Reader) (map[string]string, error) {
	val := make(map[string]string, 20)
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// ignore comments
		if strings.HasPrefix(line, "#") {
			continue
		}
		p := strings.SplitN(line, " ", 2)
		if len(p) < 1 {
			continue
		}
		val[p[0]] = ""
		if len(p) < 2 {
			continue
		}
		val[p[0]] = strings.TrimSpace(p[1])
	}

	return val, nil
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
