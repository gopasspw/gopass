package passkey

import (
	"context"
	"fmt"
)

// Consent is the outcome of an authorization gesture: what the authenticator
// established about the user during a single ceremony.
//
// See https://www.w3.org/TR/webauthn-3/#authorization-gesture
type Consent struct {
	// Present reports that the user explicitly approved this ceremony. This
	// is the test of user presence.
	Present bool
	// Verified reports that the identity of the user was verified during
	// this ceremony.
	Verified bool
}

// Operation is what the user is asked to consent to.
type Operation int

const (
	// OperationRegister is the creation of a new passkey.
	OperationRegister Operation = iota
	// OperationAuthenticate is the use of a passkey to sign in.
	OperationAuthenticate
	// OperationDisclose is telling a Relying Party that a passkey it knows
	// is already stored. No passkey is created or used.
	OperationDisclose
)

// String returns a description of the operation.
func (o Operation) String() string {
	switch o {
	case OperationRegister:
		return "register"
	case OperationAuthenticate:
		return "authenticate"
	case OperationDisclose:
		return "disclose"
	default:
		return fmt.Sprintf("Operation(%d)", int(o))
	}
}

// ConsentRequest describes a ceremony the user is asked to consent to.
type ConsentRequest struct {
	// Operation is what is about to happen.
	Operation Operation
	// RPID is the ID of the Relying Party.
	RPID string
	// RPName is the name the Relying Party gave itself. It is only known
	// for registrations and must not be trusted to identify the site.
	RPName string
	// Origin is the origin the request comes from.
	Origin string
	// TopOrigin is the origin of the top-level document, if the request
	// comes from a document that is embedded in a document of another
	// origin. It is the site the user sees in the address bar.
	TopOrigin string
	// UserName is the name of the user account.
	UserName string
	// UserDisplayName is the friendly name of the user account.
	UserDisplayName string
	// Entry is the name of the secret holding the passkey. For registrations
	// it is the secret that already holds a passkey of the user account, or
	// empty if there is none.
	Entry string
	// VerificationRequired reports that the Relying Party requires the user
	// to be verified. The ceremony fails if the consent does not say so.
	VerificationRequired bool
}

// ConsentFunc collects the authorization gesture for a ceremony: it asks the
// user to approve what the request describes and reports what was
// established. A zero Consent means the user declined. An error means the
// user could not be asked.
//
// See [ConsentByConfirmation] for a ConsentFunc that asks a yes/no question.
type ConsentFunc func(ctx context.Context, req ConsentRequest) (Consent, error)

// ConfirmFunc asks the user to confirm what a request describes and returns
// the answer.
type ConfirmFunc func(ctx context.Context, req ConsentRequest) (bool, error)

// ConsentByConfirmation returns a ConsentFunc for an authenticator whose
// only way to involve the user is to ask for a confirmation.
//
// A confirmation is a test of user presence. It is not a verification of the
// identity of the user, and gopass has no way to perform one on its own: it
// can not tell whether a passphrase was entered to unlock the store or
// whether an agent still had the key. A user who protects the store well
// enough to consider every use of it as verified can say so with the config
// key "passkey.user-verification". Only then is a confirmed ceremony reported
// as user verified.
//
// A request that requires a verification that will not be reported fails
// with [ErrUnsuitable] before the user is asked.
func ConsentByConfirmation(confirm ConfirmFunc) ConsentFunc {
	return func(ctx context.Context, req ConsentRequest) (Consent, error) {
		verified := UserVerificationEnabled(ctx)

		if req.VerificationRequired && !verified {
			return Consent{}, fmt.Errorf(
				"%w: the site requires user verification, which gopass only reports if passkey.user-verification is enabled",
				ErrUnsuitable)
		}

		confirmed, err := confirm(ctx, req)
		if err != nil || !confirmed {
			return Consent{}, err
		}

		return Consent{Present: true, Verified: verified}, nil
	}
}
