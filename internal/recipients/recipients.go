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

// Unmarshal Recipients line by line from a io.Reader.
func Unmarshal(buf []byte) []string {
	return set.SortedFiltered(strings.Split(string(buf), "\n"), func(k string) bool {
		return k != ""
	})
}
