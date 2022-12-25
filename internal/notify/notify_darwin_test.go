package notify

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

// to test cmd.exec correctly we use the same functionality as go itself see exec_test.go.
func TestDarwinNotify(t *testing.T) {
	ctx := context.Background()
	t.Setenv("GOPASS_NO_NOTIFY", "true")
	assert.NoError(t, Notify(ctx, "foo", "bar"))
}

func TestLegacyNotification(t *testing.T) {
	ctx := context.Background()
	// override execCommand with mock
	execCommand = mockExecCommand
	defer func() {
		execCommand = exec.Command
	}()

	err := Notify(ctx, "foo", "bar")
	assert.NoError(t, err)
}

func TestLegacyTerminalNotifierNotification(t *testing.T) {
	ctx := context.Background()
	// override execCommand with mock
	execCommand = mockExecCommand
	execLookPath = mockExecLookPathTerminalNotifier
	defer func() {
		execCommand = exec.Command
	}()

	err := Notify(ctx, "foo", "bar")
	assert.NoError(t, err)
}

func TestNoExecutableFound(t *testing.T) {
	ctx := context.Background()
	// override execCommand with mock
	execCommand = mockExecCommand
	execLookPath = mockExecLookPath
	defer func() {
		execCommand = exec.Command
	}()

	err := Notify(ctx, "foo", "bar")
	assert.Error(t, err)
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
