package store

import "github.com/pkg/errors"

var (
	// ErrExistsFailed is returend if we can't check for existence
	ErrExistsFailed = errors.Errorf("Failed to check for existence")
	// ErrNotFound is returned if an entry was not found
	ErrNotFound = errors.Errorf("Entry is not in the password store")
	// ErrEncrypt is returned if we failed to encrypt an entry
	ErrEncrypt = errors.Errorf("Failed to encrypt")
	// ErrDecrypt is returned if we failed to decrypt and entry
	ErrDecrypt = errors.Errorf("Failed to decrypt")
	// ErrSneaky is returned if the user passes a possible malicious path to gopass
	ErrSneaky = errors.Errorf("you've attempted to pass a sneaky path to gopass. go home")
	// ErrGitInit is returned if git is already initialized
	ErrGitInit = errors.Errorf("git is already initialized")
	// ErrGitNotInit is returned if git is not initialized
	ErrGitNotInit = errors.Errorf("git is not initialized")
	// ErrGitNoRemote is returned if git has no origin remote
	ErrGitNoRemote = errors.Errorf("git has no remote origin")
	// ErrGitNothingToCommit is returned if there are no staged changes
	ErrGitNothingToCommit = errors.Errorf("git has nothing to commit")
	// ErrNoBody is returned if a secret exists but has no content beyond a password
	ErrNoBody = errors.Errorf("no safe content to display, you can force display with show -f")
	// ErrNoPassword is returned is a secret exists but has no password, only a body
	ErrNoPassword = errors.Errorf("no password to display")
	// ErrYAMLNoMark is returned if a secret contains no valid YAML document marker
	ErrYAMLNoMark = errors.Errorf("no YAML document marker found")
	// ErrYAMLNoKey is returned if a YAML document doesn't contain a key
	ErrYAMLNoKey = errors.Errorf("key not found in YAML document")
	// ErrYAMLValueUnsupported is returned is the user tries to unmarshal an nested struct
	ErrYAMLValueUnsupported = errors.Errorf("can not unmarshal nested YAML value")
)
