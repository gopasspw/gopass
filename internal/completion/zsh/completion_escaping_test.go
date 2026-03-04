package zsh

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
			name:     "colon",
			input:    "cloudflare/api-tokens/DNS:Edit/nasvic.top",
			expected: "cloudflare/api-tokens/DNS\\:Edit/nasvic.top",
		},
		{
			name:     "brackets",
			input:    "passwords/hostname-00[1-2].mgmt",
			expected: "passwords/hostname-00\\[1-2\\].mgmt",
		},
		{
			name:     "multiple special chars",
			input:    "test:with[brackets]and:colons",
			expected: "test\\:with\\[brackets\\]and\\:colons",
		},
		{
			name:     "backslash",
			input:    "test\\path:with[special]chars",
			expected: "test\\\\path\\:with\\[special\\]chars",
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
