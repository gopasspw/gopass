package passkey

import (
	"errors"
	"fmt"
)

// ErrorName is the name of the DOMException a WebAuthn client reports to the
// Relying Party when a ceremony fails.
//
// See https://www.w3.org/TR/webauthn-3/#sctn-create-request-exceptions
type ErrorName string

const (
	// NotAllowedError is the catch-all error. It covers a user declining a
	// ceremony as well as there being no usable credential, without telling
	// the Relying Party which one it was.
	NotAllowedError ErrorName = "NotAllowedError"
	// InvalidStateError reports that a credential listed in excludeCredentials
	// is already stored, and that the user agreed to disclose that.
	InvalidStateError ErrorName = "InvalidStateError"
	// NotSupportedError reports that none of the requested credential types
	// or algorithms is supported.
	NotSupportedError ErrorName = "NotSupportedError"
	// SecurityError reports that the origin is not allowed to use the RP ID.
	SecurityError ErrorName = "SecurityError"
	// ConstraintError reports that a requirement of the Relying Party can not
	// be met by this authenticator.
	ConstraintError ErrorName = "ConstraintError"
	// TypeError reports malformed options.
	TypeError ErrorName = "TypeError"
	// UnknownError reports an internal failure.
	UnknownError ErrorName = "UnknownError"
)

var (
	// ErrNoCredentials is returned if no stored credential can satisfy a
	// request. A client can use it to fall back to another authenticator.
	ErrNoCredentials = errors.New("no matching passkey found")
	// ErrDeclined is returned if the user did not consent to a ceremony.
	ErrDeclined = errors.New("ceremony was not confirmed by the user")
	// ErrSelectionRequired is returned if more than one credential can
	// satisfy a request and none was selected.
	ErrSelectionRequired = errors.New("more than one passkey matches, a selection is required")
	// ErrUnsuitable is returned if the Relying Party asked for properties this
	// authenticator does not have, e.g. a roaming authenticator. A client can
	// use it to fall back to another authenticator.
	ErrUnsuitable = errors.New("request is not suitable for this authenticator")
	// ErrUnsupportedKey is returned for private keys of a type or with
	// parameters that are not supported.
	ErrUnsupportedKey = errors.New("unsupported key")
	// ErrInvalidCredential is returned if a stored credential is malformed.
	ErrInvalidCredential = errors.New("invalid passkey")
	// ErrNotSynced is returned together with a result if a credential was
	// written to the store, but the store reported a failure afterwards,
	// e.g. because it could not push to its remote. The credential is safe
	// where it is. It is not available on other devices yet.
	ErrNotSynced = errors.New("stored, but the store reported a failure")
)

// Error is a failed ceremony. Name is what a client should report to the
// Relying Party. Err carries the detail, which is meant for the user and must
// not be disclosed to the Relying Party: it can reveal whether a credential
// exists.
type Error struct {
	Name ErrorName
	Err  error
}

// Error implements error.
func (e *Error) Error() string {
	if e.Err == nil {
		return string(e.Name)
	}

	return fmt.Sprintf("%s: %s", e.Name, e.Err)
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.Err
}

// NameOf returns the DOMException name a client should report for err. Errors
// that did not originate in a ceremony are reported as UnknownError.
func NameOf(err error) ErrorName {
	var pe *Error
	if errors.As(err, &pe) {
		return pe.Name
	}

	return UnknownError
}

func newError(name ErrorName, err error) *Error {
	return &Error{Name: name, Err: err}
}

func errorf(name ErrorName, format string, args ...any) *Error {
	return &Error{Name: name, Err: fmt.Errorf(format, args...)}
}
