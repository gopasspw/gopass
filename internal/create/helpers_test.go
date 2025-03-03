package create

import (
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestFmtfn(t *testing.T) {
	tests := []struct {
		d        int
		n        string
		t        string
		expected string
	}{
		{0, "1", "test", color.GreenString("[1]") + " test                               "},
		{2, "2", "example", "  " + color.GreenString("[2]") + " example                            "},
		{4, "3", "sample", "    " + color.GreenString("[3]") + " sample                             "},
	}

	for _, tt := range tests {
		t.Run(tt.n, func(t *testing.T) {
			result := fmtfn(tt.d, tt.n, tt.t)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractHostname(t *testing.T) {
	t.Parallel()

	for in, out := range map[string]string{
		"":                                     "",
		"http://www.example.org/":              "www.example.org",
		"++#+++#jhlkadsrezu 33 553q ++++##$§&": "jhlkadsrezu_33_553q",
		"www.example.org/?foo=bar#abc":         "www.example.org",
		"a test":                               "a_test",
		"http://example.com":                   "example.com",
		"https://sub.example.com":              "sub.example.com",
		"ftp://example.com":                    "example.com",
		"example.com":                          "example.com",
		"invalid-url":                          "invalid-url",
	} {
		assert.Equal(t, out, extractHostname(in))
	}
}
