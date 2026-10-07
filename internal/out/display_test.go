package out

import (
	"bytes"
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestUntrustedString(t *testing.T) {
	t.Parallel()

	var _ fmt.Stringer = Untrusted("") // compile-time contract

	// Stripping the ESC introducer and BEL terminator neutralizes the OSC
	// sequence: the leftover bytes render as inert text, not as a command.
	assert.Equal(t, "]0;pwnedgopass", Untrusted("\x1b]0;pwned\x07gopass").String())
	// C0 whitespace controls.
	assert.Empty(t, Untrusted("\n\r\t").String())
	// DEL and a C1 control (CSI, as valid UTF-8).
	assert.Empty(t, Untrusted("\x7f\u009b").String())
	// Printable ASCII and multi-byte runes pass through untouched.
	assert.Equal(t, "abc αβγ", Untrusted("abc αβγ").String())
	assert.Empty(t, Untrusted("").String())
}

func TestTruncatedString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    Truncated
		want string
	}{
		{"short string passes through", Truncated{V: Untrusted("abc"), N: 10}, "abc"},
		{"exact fit has no ellipsis", Truncated{V: Untrusted("abcde"), N: 5}, "abcde"},
		{"ascii cut gets ellipsis", Truncated{V: Untrusted("abcdefgh"), N: 5}, "abcde..."},
		{"cut honors rune boundary", Truncated{V: Untrusted("αβγδε"), N: 5}, "αβγ..."},
		{"zero N disables truncation", Truncated{V: Untrusted("abc"), N: 0}, "abc"},
		{"negative N disables truncation", Truncated{V: Untrusted("abc"), N: -1}, "abc"},
		{"nil V renders empty", Truncated{V: nil, N: 5}, ""},
		{"strips through an inner Untrusted", Truncated{V: Untrusted("\x1bab"), N: 10}, "ab"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.t.String()
			assert.Equal(t, tc.want, got)
			assert.True(t, utf8.ValidString(got))
		})
	}
}

func TestDisplayComposition(t *testing.T) {
	t.Parallel()

	// Strip first, then bound the visible content.
	got := Truncated{V: Untrusted("\x1bgopass/1.2.3-long..."), N: 8}.String()
	assert.Equal(t, "gopass/1...", got)
}

func TestDisplayThroughPrintf(t *testing.T) {
	ctx := config.NewContextInMemory()

	orig := Stdout
	defer func() {
		Stdout = orig
	}()
	buf := &bytes.Buffer{}
	Stdout = buf

	Printf(ctx, "%s", Untrusted("\x1b[31mpwned"))

	// Without the ESC introducer the sequence is inert text on the wire.
	assert.Equal(t, "[31mpwned\n", buf.String())
}
