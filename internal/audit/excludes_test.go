package audit

import (
	"testing"
)

func TestFilterExcludes(t *testing.T) {
	tests := []struct {
		name     string
		excludes string
		in       []string
		want     []string
	}{
		{
			name:     "no excludes",
			excludes: "",
			in:       []string{"secret1", "secret2"},
			want:     []string{"secret1", "secret2"},
		},
		{
			name:     "exclude one secret",
			excludes: "secret1",
			in:       []string{"secret1", "secret2"},
			want:     []string{"secret2"},
		},
		{
			name:     "exclude all secrets",
			excludes: "secret1\nsecret2",
			in:       []string{"secret1", "secret2"},
			want:     []string{},
		},
		{
			name:     "exclude with comment",
			excludes: "# this is a comment\nsecret1",
			in:       []string{"secret1", "secret2"},
			want:     []string{"secret2"},
		},
		{
			name:     "exclude with empty lines",
			excludes: "\nsecret1\n\n",
			in:       []string{"secret1", "secret2"},
			want:     []string{"secret2"},
		},
		{
			name:     "exclude with regex",
			excludes: "secret.*",
			in:       []string{"secret1", "secret2", "other"},
			want:     []string{"other"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterExcludes(tt.excludes, tt.in)
			if len(got) != len(tt.want) {
				t.Errorf("FilterExcludes() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("FilterExcludes() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
