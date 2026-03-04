// Package store provides the interface for the gopass password store.
// It defines the methods and types used to interact with the password store.
package store

import (
	"context"
)

// RecipientCallback is a callback to verify the list of recipients.
type RecipientCallback func(context.Context, string, []string) ([]string, error)

// ImportCallback is a callback to ask the user if they want to import
// a certain recipients public key into their keystore.
type ImportCallback func(context.Context, string, []string) bool

// FsckCallback is a callback to ask the user to confirm certain fsck
// corrective actions.
type FsckCallback func(context.Context, string) bool
