package cli

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// parseTS parses the passed string as an Epoch int and returns
// the time struct or the zero time struct
func parseTS(str string) time.Time {
	t := time.Time{}

	if sec, err := strconv.ParseInt(str, 10, 64); err == nil {
		t = time.Unix(sec, 0)
	}

	return t
}

// parseInt parses the passed string as an int and returns it
// or 0 on errors
func parseInt(str string) int {
	i := 0

	if iv, err := strconv.ParseInt(str, 10, 32); err == nil {
		i = int(iv)
	}

	return i
}

func splitPacket(in string) map[string]string {
	m := make(map[string]string, 3)
	p := strings.Split(in, ":")
	if len(p) < 3 {
		return m
	}
	p = strings.Split(strings.TrimSpace(p[2]), " ")
	for i := 0; i+1 < len(p); i += 2 {
		m[p[i]] = p[i+1]
	}
	return m
}

// see https://www.gnupg.org/documentation/manuals/gnupg/Invoking-GPG_002dAGENT.html
func tty() string {
	fd0 := "/proc/self/fd/0"
	dest, err := os.Readlink(fd0)
	if err != nil {
		return ""
	}
	return dest
}
