package notify

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/require"
)

// to test cmd.exec correctly we use the same functionality as go itself see exec_test.go.
func TestDarwinNotify(t *testing.T) {
	ctx := config.NewNoWrites().WithConfig(context.Background())
	t.Setenv("GOPASS_NO_NOTIFY", "true")
	require.NoError(t, Notify(ctx, "foo", "bar"))
}

func TestLegacyNotification(t *testing.T) {
	ctx := config.NewNoWrites().WithConfig(context.Background())
	// override execCommand with mock
	execCommand = mockExecCommand
	defer func() {
		execCommand = exec.Command
	}()

	err := Notify(ctx, "foo", "bar")
	require.NoError(t, err)
}

func TestLegacyTerminalNotifierNotification(t *testing.T) {
	ctx := config.NewNoWrites().WithConfig(context.Background())
	// override execCommand with mock
	execCommand = mockExecCommand
	execLookPath = mockExecLookPathTerminalNotifier
	defer func() {
		execCommand = exec.Command
	}()

	err := Notify(ctx, "foo", "bar")
	require.NoError(t, err)
}

func TestNoExecutableFound(t *testing.T) {
	ctx := config.NewNoWrites().WithConfig(context.Background())
	// override execCommand with mock
	execCommand = mockExecCommand
	execLookPath = mockExecLookPath
	defer func() {
		execCommand = exec.Command
	}()

	err := Notify(ctx, "foo", "bar")
	require.Error(t, err)
}

func mockExecLookPath(_ string) (string, error) {
	return "", fmt.Errorf("no binary found")
}

func mockExecLookPathTerminalNotifier(command string) (string, error) {
	if command == terminalNotifier {
		return "", fmt.Errorf("no binary found")
	}

	return "", nil
}

func mockExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}

	return cmd
}
