package clipboard

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gopasspw/clipboard"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/mitchellh/go-ps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotExistingClipboardCopyCommand(t *testing.T) {
	t.Setenv("GOPASS_NO_NOTIFY", "true")
	t.Setenv("GOPASS_CLIPBOARD_COPY_CMD", "not_existing_command")

	ctx, cancel := context.WithCancel(config.NewContextInMemory())
	defer cancel()

	maybeErr := CopyTo(ctx, "foo", []byte("bar"), 1)
	require.Error(t, maybeErr)
	assert.Contains(t, maybeErr.Error(), "\"not_existing_command\": executable file not found")
}

func TestUnsupportedCopyToClipboard(t *testing.T) {
	t.Setenv("GOPASS_NO_NOTIFY", "true")

	ctx, cancel := context.WithCancel(config.NewContextInMemory())
	defer cancel()

	clipboard.ForceUnsupported = true

	buf := &bytes.Buffer{}
	out.Stderr = buf

	require.NoError(t, CopyTo(ctx, "foo", []byte("bar"), 1))
	assert.Contains(t, buf.String(), "WARNING")
}

func TestClearClipboard(t *testing.T) {
	ctx, cancel := context.WithCancel(config.NewContextInMemory())
	require.NoError(t, clearClip(ctx, "foo", []byte("bar"), 0))
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func BenchmarkWalkProc(b *testing.B) {
	for b.Loop() {
		_ = filepath.Walk("/proc", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if strings.Count(path, "/") != 3 {
				return nil
			}
			if !strings.HasSuffix(path, "/status") {
				return nil
			}
			pid, err := strconv.Atoi(path[6:strings.LastIndex(path, "/")])
			if err != nil {
				return nil
			}
			walkFn(pid, func(int) {})

			return nil
		})
	}
}

func BenchmarkListProc(b *testing.B) {
	for b.Loop() {
		procs, err := ps.Processes()
		if err != nil {
			b.Fatalf("err: %s", err)
		}

		for _, proc := range procs {
			walkFn(proc.Pid(), func(int) {})
		}
	}
}
