package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
)

const gpgStatusPrefix = "[GNUPG:] "

// Decrypt will try to decrypt the given file.
func (g *GPG) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	timeout := timeoutFromContext(ctx)
	ctx, cancel := withTimeout(ctx, timeout)
	defer cancel()

	args := make([]string, 0, len(g.args)+5)
	args = append(args, g.args...)
	args = append(args, "--status-fd", "2", "--decrypt")
	// Useful information may appear there
	if debug.IsEnabled() {
		args = append(args, "--verbose", "--verbose")
	}
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	// If gopass-jsonapi is used, there is no way to reach this os.Stderr, so
	// we write this stderr to the log file as well. Status lines from
	// --status-fd 2 are stripped from os.Stderr and parsed to detect whether
	// decryption actually succeeded despite a non-zero exit code.
	statusFilter := &gpgStatusFilter{dst: os.Stderr}
	cmd.Stderr = io.MultiWriter(debug.LogWriter, statusFilter)

	debug.V(1).Log("Running %s %+v", cmd.Path, cmd.Args)
	stdout, err := cmd.Output()
	_ = statusFilter.Flush()
	if err != nil {
		if ctx.Err() == nil && isExitCode(err, 2) && decryptionSucceeded(statusFilter.status) {
			debug.Log("GPG decrypt exited with status 2 despite DECRYPTION_OKAY (likely multiple ECDH keys with throw-keyids); treating as success")

			return stdout, nil
		}

		debug.Log("GPG decrypt failed: %s %+v: %+v", cmd.Path, cmd.Args, err)
		if timeout > 0 && errors.Is(ctx.Err(), context.DeadlineExceeded) {
			// GPG was killed because it did not finish within the deadline.
			// Most often that is a passphrase prompt nobody answered in time,
			// so point at the knob that governs the wait.
			err = fmt.Errorf("gpg decrypt did not finish within %s. If you need more time to enter your GPG passphrase, raise the timeout with `gopass config gpg.timeout <seconds>`, e.g. `gopass config gpg.timeout 600` for ten minutes, or set it to 0 to disable the deadline: %w", timeout, err)
		}
	}

	return stdout, err
}

func isExitCode(err error, code int) bool {
	var exitErr *exec.ExitError

	return errors.As(err, &exitErr) && exitErr.ExitCode() == code
}

// decryptionSucceeded inspects GPG --status-fd lines (with the "[GNUPG:] "
// prefix already stripped) and returns true only when GPG reported a complete,
// clean decryption with no decryption, MDC, signature, or non-exit failures.
//
// GnuPG 2.4.4+ exits with status 2 and emits "FAILURE gpg-exit 33554433" when
// decrypting an anonymous-recipient (--throw-keyids) message in a keyring with
// multiple ECDH keys, because each failed key trial increments GPG's internal
// error counter even after a subsequent key succeeds.
func decryptionSucceeded(statusLines []string) bool {
	var beginCount, okayCount, endCount int

	for _, line := range statusLines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "BEGIN_DECRYPTION":
			beginCount++
		case "DECRYPTION_OKAY":
			okayCount++
		case "END_DECRYPTION":
			endCount++
		case "DECRYPTION_FAILED", "BADMDC", "NODATA", "BADSIG":
			return false
		case "FAILURE":
			if len(fields) < 2 || fields[1] != "gpg-exit" {
				return false
			}
		}
	}

	return okayCount > 0 && okayCount == beginCount && okayCount == endCount
}

// gpgStatusFilter splits GPG --status-fd 2 lines out of stderr so they can be
// inspected without cluttering the user's stderr, while forwarding all other
// stderr lines to dst.
type gpgStatusFilter struct {
	dst    io.Writer
	buf    bytes.Buffer
	status []string
}

func (f *gpgStatusFilter) Write(p []byte) (int, error) {
	n := len(p)
	for len(p) > 0 {
		idx := bytes.IndexByte(p, '\n')
		if idx < 0 {
			_, _ = f.buf.Write(p)

			break
		}
		_, _ = f.buf.Write(p[:idx+1])
		if err := f.flushLine(); err != nil {
			return n, err
		}
		p = p[idx+1:]
	}

	return n, nil
}

func (f *gpgStatusFilter) Flush() error {
	if f.buf.Len() == 0 {
		return nil
	}

	return f.flushLine()
}

func (f *gpgStatusFilter) flushLine() error {
	raw := f.buf.Bytes()
	line := strings.TrimRight(string(raw), "\r\n")
	if rest, ok := strings.CutPrefix(line, gpgStatusPrefix); ok {
		f.status = append(f.status, rest)
		f.buf.Reset()

		return nil
	}

	var err error
	if f.dst != nil {
		_, err = f.dst.Write(raw)
	}
	f.buf.Reset()

	return err
}
