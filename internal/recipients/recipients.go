package recipients

import (
	"bytes"
	"strings"

	"github.com/gopasspw/gopass/internal/set"
)

// Marshal all in memory Recipients line by line to []byte.
func Marshal(r []string) []byte {
	if len(r) == 0 {
		return []byte("\n")
	}

	out := bytes.Buffer{}
	for _, k := range set.Sorted(r) {
		_, _ = out.WriteString(k)
		_, _ = out.WriteString("\n")
	}

	return out.Bytes()
}

// Unmarshal Recipients line by line from a io.Reader. Handles Unix, Windows and Mac line endings.
func Unmarshal(buf []byte) []string {
	in := strings.ReplaceAll(string(buf), "\r", "\n")

	return set.Apply(set.SortedFiltered(strings.Split(in, "\n"), func(k string) bool {
		return k != ""
	}), func(k string) string {
		return strings.TrimSpace(k)
	})
}
