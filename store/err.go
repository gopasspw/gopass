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
)
