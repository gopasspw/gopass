package fish

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapePasswordName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single quote",
			input:    "password'with'quotes",
			expected: "password'\\''with'\\''quotes",
		},
		{
			name:     "backslash",
			input:    "path\\to\\password",
			expected: "path\\\\to\\\\password",
		},
		{
			name:     "backslash and quote",
			input:    "test\\with'quote",
			expected: "test\\\\with'\\''quote",
		},
		{
			name:     "colon",
			input:    "cloudflare/api-tokens/DNS:Edit/nasvic.top",
			expected: "cloudflare/api-tokens/DNS:Edit/nasvic.top",
		},
		{
			name:     "no special chars",
			input:    "simple/password",
			expected: "simple/password",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := escapePasswordName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
