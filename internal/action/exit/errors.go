package exit

import (
	"fmt"

	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/urfave/cli/v2"
)

const (
	// OK means no error (status code 0).
	OK = iota
	// Unknown is used if we can't determine the exact exit cause.
	Unknown
	// Usage is used if there was some kind of invocation error.
	Usage
	// Aborted is used if the user willingly aborted an action.
	Aborted
	// Unsupported is used if an operation is not supported by gopass.
	Unsupported
	// AlreadyInitialized is used if someone is trying to initialize.
	// an already initialized store.
	AlreadyInitialized
	// NotInitialized is used if someone is trying to use an unitialized.
	// store.
	NotInitialized
	// Git is used if any git errors are encountered.
	Git
	// Mount is used if a substore mount operation fails.
	Mount
	// NoName is used when no name was provided for a named entry.
	NoName
	// NotFound is used if a requested secret is not found.
	NotFound
	// Decrypt is used when reading/decrypting a secret failed.
	Decrypt
	// Encrypt is used when writing/encrypting of a secret fails.
	Encrypt
	// List is used when listing the store content fails.
	List
	// Audit is used when audit report possible issues.
	Audit
	// Fsck is used when the integrity check fails.
	Fsck
	// Config is used when config errors occur.
	Config
	// Recipients is used when a recipient operation fails.
	Recipients
	// IO is used for misc. I/O errors.
	IO
	// GPG is used for misc. gpg errors.
	GPG
)

// Error returns a user friendly CLI error.
func Error(exitCode int, err error, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	if err != nil {
		debug.LogN(1, "%s - stacktrace: %+v", msg, err)
	}
	return cli.Exit(msg, exitCode)
}
