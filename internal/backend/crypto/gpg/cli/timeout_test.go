package cli

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/require"
)

// gpgTimeoutHelperEnv marks the re-executed helper process used to observe the
// configured deadline.
const gpgTimeoutHelperEnv = "GOPASS_TEST_GPG_TIMEOUT_HELPER"

func TestTimeoutFromContext(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		set  string
		want time.Duration
	}{
		{
			name: "default is one minute",
			want: Timeout,
		},
		{
			name: "override to ten minutes",
			set:  "600",
			want: 10 * time.Minute,
		},
		{
			name: "zero disables the deadline",
			set:  "0",
		},
		{
			name: "negative disables the deadline",
			set:  "-1",
		},
		{
			name: "garbage falls back to the default",
			set:  "yes",
			want: Timeout,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := config.NewContextInMemory()
			if tc.set != "" {
				cfg, _ := config.FromContext(ctx)
				require.NoError(t, cfg.Set("", "gpg.timeout", tc.set))
			}

			require.Equal(t, tc.want, timeoutFromContext(ctx))
		})
	}
}

func TestWithTimeout(t *testing.T) {
	t.Parallel()

	// a zero timeout must not add a deadline of its own
	ctx, cancel := withTimeout(t.Context(), 0)
	defer cancel()
	require.NoError(t, ctx.Err())

	// a positive timeout must bound the context
	ctx, cancel = withTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()
	require.NoError(t, ctx.Err())
	select {
	case <-ctx.Done():
		require.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
	case <-time.After(5 * time.Second):
		t.Fatal("context was not bounded by the timeout")
	}
}

// TestDecryptDeadlineMentionsGPGTimeout makes sure a GPG invocation that runs
// into the configured deadline reports that deadline, including how to raise
// it. The fake gpg is this test binary itself, re-executed as a helper that
// outlives the deadline on every platform (the appended --decrypt lands
// harmlessly behind the flag terminator).
func TestDecryptDeadlineMentionsGPGTimeout(t *testing.T) {
	t.Setenv(gpgTimeoutHelperEnv, "1")

	cfg := config.NewInMemory()
	require.NoError(t, cfg.Set("", "gpg.timeout", "1"))
	ctx := cfg.WithConfig(t.Context())

	g := &GPG{
		binary: os.Args[0],
		args:   []string{"-test.run=^TestGPGTimeoutHelperProcess$", "--"},
	}

	start := time.Now()
	_, err := g.Decrypt(ctx, []byte("foo"))
	require.Error(t, err)
	require.Less(t, time.Since(start), 10*time.Second)
	require.ErrorContains(t, err, "gpg.timeout")
}

// TestGPGTimeoutHelperProcess is never run directly by the go test runner of
// this package: it only runs in the re-executed helper process started by
// TestDecryptDeadlineMentionsGPGTimeout, where it stands in for a gpg
// invocation waiting on a passphrase prompt.
func TestGPGTimeoutHelperProcess(t *testing.T) {
	if os.Getenv(gpgTimeoutHelperEnv) != "1" {
		t.Skip("not the helper process")
	}

	select {
	case <-time.After(30 * time.Second):
	case <-t.Context().Done():
	}
}
