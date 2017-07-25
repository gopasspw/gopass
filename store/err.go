package store

import "fmt"

var (
	// ErrExistsFailed is returend if we can't check for existence
	ErrExistsFailed = fmt.Errorf("Failed to check for existence")
	// ErrNotFound is returned if an entry was not found
	ErrNotFound = fmt.Errorf("Entry is not in the password store")
	// ErrEncrypt is returned if we failed to encrypt an entry
	ErrEncrypt = fmt.Errorf("Failed to encrypt")
	// ErrDecrypt is returned if we failed to decrypt and entry
	ErrDecrypt = fmt.Errorf("Failed to decrypt")
	// ErrSneaky is returned if the user passes a possible malicious path to gopass
	ErrSneaky = fmt.Errorf("you've attempted to pass a sneaky path to gopass. go home")
	// ErrGitInit is returned if git is already initialized
	ErrGitInit = fmt.Errorf("git is already initialized")
	// ErrGitNotInit is returned if git is not initialized
	ErrGitNotInit = fmt.Errorf("git is not initialized")
	// ErrGitNoRemote is returned if git has no origin remote
	ErrGitNoRemote = fmt.Errorf("git has no remote origin")
	// ErrGitNothingToCommit is returned if there are no staged changes
	ErrGitNothingToCommit = fmt.Errorf("git has nothing to commit")
	// ErrNoBody is returned if a secret exists but has no content beyond a password
	ErrNoBody = fmt.Errorf("no safe content to display, you can force display with show -f")
	// ErrNoPassword is returned is a secret exists but has no password, only a body
	ErrNoPassword = fmt.Errorf("no password to display")
	// ErrYAMLNoMark is returned if a secret contains no valid YAML document marker
	ErrYAMLNoMark = fmt.Errorf("no YAML document marker found")
	// ErrYAMLNoKey is returned if a YAML document doesn't contain a key
	ErrYAMLNoKey = fmt.Errorf("key not found in YAML document")
	// ErrYAMLValueUnsupported is returned is the user tries to unmarshal an nested struct
	ErrYAMLValueUnsupported = fmt.Errorf("can not unmarshal nested YAML value")
)
