package apimock_test

import (
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/apimock"
)

var _ gopass.Store = apimock.New() //nolint:staticcheck
