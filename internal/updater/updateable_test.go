package updater

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUpdateable(t *testing.T) {
	originalExecutable := executable
	defer func() { executable = originalExecutable }()

	tests := []struct {
		name           string
		mockExecutable func(ctx context.Context) (string, error)
		envVars        map[string]string
		expectedError  string
	}{
		{
			name: "test binary",
			mockExecutable: func(ctx context.Context) (string, error) {
				return "/path/to/binary.test", nil
			},
			expectedError: "",
		},
		{
			name: "force update",
			mockExecutable: func(ctx context.Context) (string, error) {
				return "/path/to/binary", nil
			},
			envVars: map[string]string{
				"GOPASS_FORCE_UPDATE": "1",
			},
			expectedError: "",
		},
		{
			name: "binary in GOPATH",
			mockExecutable: func(ctx context.Context) (string, error) {
				return "/path/to/gopath/bin/binary", nil
			},
			envVars: map[string]string{
				"GOPATH": "/path/to/gopath",
			},
			expectedError: "use go get -u to update binary in GOPATH",
		},
		{
			name: "not a regular file",
			mockExecutable: func(ctx context.Context) (string, error) {
				return "/path/to/directory", nil
			},
			expectedError: "not a regular file",
		},
		{
			name: "cannot write to file",
			mockExecutable: func(ctx context.Context) (string, error) {
				return "/path/to/binary", nil
			},
			expectedError: "can not write \"/path/to/binary\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executable = tt.mockExecutable

			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			err := IsUpdateable(context.Background())
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}
