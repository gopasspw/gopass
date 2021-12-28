package gpgconf

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strings"
)

// Path returns the path to a GPG component
func Path(key string) (string, error) {
	buf := &bytes.Buffer{}
	cmd := exec.Command("gpgconf")
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	key = strings.TrimSpace(strings.ToLower(key))
	sc := bufio.NewScanner(buf)
	for sc.Scan() {
		p := strings.Split(strings.TrimSpace(sc.Text()), ":")
		if len(p) < 3 {
			continue
		}
		if key == p[0] {
			return p[2], nil
		}
	}

	return "", nil
}
