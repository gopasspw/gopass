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
		_, _ = out.WriteString(strings.TrimSpace(k))
		_, _ = out.WriteString("\n")
	}

	return out.Bytes()
}

// Unmarshal Recipients line by line from a io.Reader. Handles Unix, Windows and Mac line endings.
// Note: Does not preserve comments!
func Unmarshal(buf []byte) []string {
	in := strings.ReplaceAll(string(buf), "\r", "\n")

	return set.Apply(set.SortedFiltered(strings.Split(in, "\n"), func(k string) bool {
		return k != "" && !strings.HasPrefix(k, "#")
	}), func(k string) string {
		out := strings.TrimSpace(k)
		if strings.Contains(out, " #") {
			out = out[:strings.Index(k, " #")]
		}

		return out
	})
}
