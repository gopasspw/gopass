package age

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindRecipients(t *testing.T) {
	ctx := context.Background()
	a := &Age{
		ghCache: &mockGHCache{},
	}

	tests := []struct {
		name     string
		search   []string
		expected []string
	}{
		{
			name:     "github key",
			search:   []string{"github:username"},
			expected: []string{"ssh-rsa AAAAB3Nza..."},
		},
		{
			name:     "ssh key",
			search:   []string{"ssh-rsa AAAAB3Nza..."},
			expected: []string{"ssh-rsa AAAAB3Nza..."},
		},
		{
			name:     "age key",
			search:   []string{"age1qxy2z..."},
			expected: []string{"age1qxy2z..."},
		},
		{
			name:     "unknown key",
			search:   []string{"unknown:key"},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipients, err := a.FindRecipients(ctx, tt.search...)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, recipients)
		})
	}
}

func TestParseRecipients(t *testing.T) {
	ctx := context.Background()
	a := &Age{
		ghCache: &mockGHCache{},
	}

	tests := []struct {
		name     string
		input    []string
		expected int
	}{
		{
			name:     "valid age key",
			input:    []string{"age1qxy2z..."},
			expected: 1,
		},
		{
			name:     "valid ssh key",
			input:    []string{"ssh-rsa AAAAB3Nza..."},
			expected: 1,
		},
		{
			name:     "github key",
			input:    []string{"github:username"},
			expected: 1,
		},
		{
			name:     "unknown key",
			input:    []string{"unknown:key"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipients, err := a.parseRecipients(ctx, tt.input)
			assert.NoError(t, err)
			assert.Len(t, recipients, tt.expected)
		})
	}
}

type mockGHCache struct{}

func (m *mockGHCache) ListKeys(ctx context.Context, user string) ([]string, error) {
	return []string{"ssh-rsa AAAAB3Nza..."}, nil
}
