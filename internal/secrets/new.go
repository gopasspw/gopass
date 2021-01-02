package secrets

import (
	"github.com/gopasspw/gopass/pkg/gopass"
)

// New creates a new secret
func New() gopass.Secret {
	return NewKV()
}
