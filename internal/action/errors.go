package action

import (
	"fmt"

	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/urfave/cli/v2"
)

const (
	// ExitOK means no error (status code 0)
	ExitOK = iota
	// ExitUnknown is used if we can't determine the exact exit cause
	ExitUnknown
	// ExitUsage is used if there was some kind of invocation error
	ExitUsage
	// ExitAborted is used if the user willingly aborted an action
	ExitAborted
	// ExitUnsupported is used if an operation is not supported by gopass
	ExitUnsupported
	// ExitAlreadyInitialized is used if someone is trying to initialize
	// an already initialized store
	ExitAlreadyInitialized
	// ExitNotInitialized is used if someone is trying to use an unitialized
	// store
	ExitNotInitialized
	// ExitGit is used if any git errors are encountered
	ExitGit
	// ExitMount is used if a substore mount operation fails
	ExitMount
	// ExitNoName is used when no name was provided for a named entry
	ExitNoName
	// ExitNotFound is used if a requested secret is not found
	ExitNotFound
	// ExitDecrypt is used when reading/decrypting a secret failed
	ExitDecrypt
	// ExitEncrypt is used when writing/encrypting of a secret fails
	ExitEncrypt
	// ExitList is used when listing the store content fails
	ExitList
	// ExitAudit is used when audit report possible issues
	ExitAudit
	// ExitFsck is used when the integrity check fails
	ExitFsck
	// ExitConfig is used when config errors occur
	ExitConfig
	// ExitRecipients is used when a recipient operation fails
	ExitRecipients
	// ExitIO is used for misc. I/O errors
	ExitIO
	// ExitGPG is used for misc. gpg errors
	ExitGPG
)

// ExitError returns a user friendly CLI error
func ExitError(exitCode int, err error, format string, args ...interface{}) error {
	if err != nil {
		debug.Log("Stacktrace: %+v", err)
	}
	return cli.Exit(fmt.Sprintf(format, args...), exitCode)
}
