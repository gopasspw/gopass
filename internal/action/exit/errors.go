package exit

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/urfave/cli/v2"
)

// Exit code constants. Values are fixed and must never be re-ordered or
// renumbered — callers and scripts depend on their stability.
const (
	// OK means no error (status code 0).
	OK = 0
	// Unknown is used if we can't determine the exact exit cause.
	Unknown = 1
	// Usage is used if there was some kind of invocation error.
	Usage = 2
	// Aborted is used if the user willingly aborted an action.
	Aborted = 3
	// Unsupported is used if an operation is not supported by gopass.
	Unsupported = 4
	// AlreadyInitialized is used if someone is trying to initialize
	// an already initialized store.
	AlreadyInitialized = 5
	// NotInitialized is used if someone is trying to use an uninitialized
	// store.
	NotInitialized = 6
	// Git is used if any git errors are encountered.
	Git = 7
	// Mount is used if a substore mount operation fails.
	Mount = 8
	// NoName is used when no name was provided for a named entry.
	NoName = 9
	// NotFound is used if a requested secret is not found.
	NotFound = 10
	// Decrypt is used when reading/decrypting a secret failed.
	Decrypt = 11
	// Encrypt is used when writing/encrypting of a secret fails.
	Encrypt = 12
	// List is used when listing the store content fails.
	List = 13
	// Audit is used when audit reports possible issues.
	Audit = 14
	// Fsck is used when the integrity check fails.
	Fsck = 15
	// Config is used when config errors occur.
	Config = 16
	// Recipients is used when a recipient operation fails.
	Recipients = 17
	// IO is used for misc. I/O errors.
	IO = 18
	// GPG is used for misc. gpg errors.
	GPG = 19
	// Hook is used for Hook failures.
	Hook = 20
	// Doctor is used when the doctor command finds failing checks.
	Doctor = 21
)

// exitCodeDescriptions lists every defined exit code together with a short
// description suitable for human consumption.
var exitCodeDescriptions = []struct {
	Code int
	Name string
	Desc string
}{
	{OK, "OK", "Success — no error"},
	{Unknown, "Unknown", "Unclassified or unexpected error"},
	{Usage, "Usage", "Bad invocation: wrong arguments or flags"},
	{Aborted, "Aborted", "User deliberately aborted the operation"},
	{Unsupported, "Unsupported", "Operation is not supported"},
	{AlreadyInitialized, "AlreadyInitialized", "Store is already initialized"},
	{NotInitialized, "NotInitialized", "Store is not initialized"},
	{Git, "Git", "Git operation failed"},
	{Mount, "Mount", "Substore mount operation failed"},
	{NoName, "NoName", "No name provided for the entry"},
	{NotFound, "NotFound", "Requested secret not found"},
	{Decrypt, "Decrypt", "Reading or decrypting a secret failed"},
	{Encrypt, "Encrypt", "Writing or encrypting a secret failed"},
	{List, "List", "Listing store contents failed"},
	{Audit, "Audit", "Audit found one or more issues"},
	{Fsck, "Fsck", "Integrity check found errors"},
	{Config, "Config", "Configuration error (reserved)"},
	{Recipients, "Recipients", "Recipient operation failed"},
	{IO, "IO", "Miscellaneous I/O error"},
	{GPG, "GPG", "Miscellaneous GPG error (reserved)"},
	{Hook, "Hook", "Hook execution failed"},
	{Doctor, "Doctor", "Doctor found one or more failing checks"},
}

// PrintExitCodes writes a human-readable table of all defined exit codes to w.
func PrintExitCodes(w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "Code\tName\tDescription")
	fmt.Fprintln(tw, "----\t----\t-----------")
	for _, e := range exitCodeDescriptions {
		fmt.Fprintf(tw, "%d\t%s\t%s\n", e.Code, e.Name, e.Desc)
	}
	_ = tw.Flush()
}

// Error returns a user friendly CLI error.
func Error(exitCode int, err error, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	if err != nil {
		debug.LogN(1, "%s - stacktrace: %+v", msg, err)
	}

	return cli.Exit(msg, exitCode)
}
