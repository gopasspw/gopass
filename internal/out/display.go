package out

import (
	"fmt"
	"strings"
	"unicode"
)

// Untrusted marks a string from an untrusted origin, e.g. a third-party age
// agent answering on the agent socket. Rendering it strips terminal control
// characters (C0/C1, including ESC and OSC introducers) that fmt's %s would
// otherwise pass through verbatim. To mark the rendering of an existing
// Stringer as untrusted, wrap its String() result: Untrusted(x.String()).
type Untrusted string

// String implements fmt.Stringer by stripping control characters.
func (u Untrusted) String() string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}

		return r
	}, string(u))
}

// Truncated bounds the rendered display length of V, mirroring
// io.LimitedReader's shape: V renders first, then the output is cut at the
// last rune boundary at or before N bytes and suffixed with "...". N <= 0
// disables truncation; a nil V renders as the empty string.
type Truncated struct {
	V fmt.Stringer
	N int
}

// String implements fmt.Stringer by bounding the rendered length of V.
func (t Truncated) String() string {
	if t.V == nil {
		return ""
	}

	s := t.V.String()
	if t.N <= 0 || len(s) <= t.N {
		return s
	}

	var b strings.Builder
	for _, r := range s {
		if b.Len() >= t.N {
			break
		}
		b.WriteRune(r)
	}

	return b.String() + "..."
}
