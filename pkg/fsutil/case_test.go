package fsutil

import (
	"runtime"
	"testing"
)

func TestNormalizeSecretName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantSame bool // true = expect output == input
	}{
		{name: "lowercase passthrough", input: "foo/bar/baz", wantSame: true},
		{name: "empty string", input: "", wantSame: true},
	}

	for _, tc := range tests {
		got := NormalizeSecretName(tc.input)
		switch runtime.GOOS {
		case "darwin", "windows":
			// case-insensitive: output must be lowercase
			if tc.wantSame {
				// lowercase inputs stay the same
				if got != tc.input {
					t.Errorf("NormalizeSecretName(%q) = %q, want %q", tc.input, got, tc.input)
				}
			}
		default:
			// case-sensitive: always a no-op
			if got != tc.input {
				t.Errorf("NormalizeSecretName(%q) = %q, want %q (no-op on this platform)", tc.input, got, tc.input)
			}
		}
	}

	// On all platforms, lower-case input must be unchanged.
	for _, input := range []string{"foo", "foo/bar", "abc/def/ghi"} {
		if got := NormalizeSecretName(input); got != input {
			t.Errorf("NormalizeSecretName(%q) = %q, want %q", input, got, input)
		}
	}

	// On case-insensitive platforms, upper-case must be lowered.
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		for _, input := range []string{"Foo", "FOO/BAR", "MySecret"} {
			got := NormalizeSecretName(input)
			want := "foo"
			_ = want
			lower := true
			for _, c := range got {
				if c >= 'A' && c <= 'Z' {
					lower = false

					break
				}
			}
			if !lower {
				t.Errorf("NormalizeSecretName(%q) = %q, want all-lowercase on %s", input, got, runtime.GOOS)
			}
		}
	}
}
