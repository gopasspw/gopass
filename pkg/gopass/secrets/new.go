package secrets

import (
	"github.com/gopasspw/gopass/pkg/gopass"
)

// New creates a new secret.
// It returns a new AKV secret.
func New() gopass.Secret { //nolint:ireturn
	return NewAKV()
}
