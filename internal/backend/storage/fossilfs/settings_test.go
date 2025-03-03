package fossilfs

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFossil_fixConfig(t *testing.T) {
	f := &Fossil{}
	ctx := context.Background()

	err := f.fixConfig(ctx)
	require.NoError(t, err)
}

func TestFossil_InitConfig(t *testing.T) {
	f := &Fossil{}
	ctx := context.Background()

	err := f.InitConfig(ctx, "", "")
	require.NoError(t, err)
}

func TestFossil_ConfigSet(t *testing.T) {
	f := &Fossil{}
	ctx := context.Background()

	err := f.ConfigSet(ctx, "test-key", "test-value")
	require.NoError(t, err)
}

func TestFossil_ConfigGet(t *testing.T) {
	f := &Fossil{}
	ctx := context.Background()

	// Mock the command execution
	execCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, arg...)
		cmd := exec.CommandContext(ctx, os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
	defer func() { execCommand = exec.CommandContext }()

	value, err := f.ConfigGet(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, "test-value", value)
}

func TestFossil_ConfigList(t *testing.T) {
	f := &Fossil{}
	ctx := context.Background()

	// Mock the command execution
	execCommand := func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, arg...)
		cmd := exec.CommandContext(ctx, os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
	defer func() { execCommand = exec.CommandContext }()

	configs, err := f.ConfigList(ctx)
	require.NoError(t, err)
	assert.NotNil(t, configs)
	assert.Equal(t, "test-value", configs["test-key"])
}

// TestHelperProcess is a helper function to mock exec.CommandContext
func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	if len(args) > 3 && args[3] == "settings" {
		if len(args) > 4 && args[4] == "--exact" {
			if args[5] == "test-key" {
				os.Stdout.WriteString("test-key test-value\n")
			}
		} else {
			os.Stdout.WriteString("test-key test-value\n")
		}
	}
	os.Exit(0)
}
