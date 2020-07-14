package recipients

import (
	"bufio"
	"bytes"
	"sort"
	"strings"
)

// Marshal all in memory Recipients line by line to []byte.
func Marshal(r []string) []byte {
	if len(r) == 0 {
		return []byte("\n")
	}

	// deduplicate
	m := make(map[string]struct{}, len(r))
	for _, k := range r {
		m[k] = struct{}{}
	}
	// sort
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := bytes.Buffer{}
	for _, k := range keys {
		_, _ = out.WriteString(k)
		_, _ = out.WriteString("\n")
	}

	return out.Bytes()
}

// Unmarshal Recipients line by line from a io.Reader.
func Unmarshal(buf []byte) []string {
	m := make(map[string]struct{}, 5)
	scanner := bufio.NewScanner(bytes.NewReader(buf))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			// deduplicate
			m[line] = struct{}{}
		}
	}

	lst := make([]string, 0, len(m))
	for k := range m {
		lst = append(lst, k)
	}
	// sort
	sort.Strings(lst)

	return lst
}
