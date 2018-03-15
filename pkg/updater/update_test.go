package updater

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckHost(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		in string
		ok bool
	}{
		{
			in: "https://github.com/justwatchcom/gopass/releases/download/v1.6.8/gopass-1.6.8-linux-amd64.tar.gz",
			ok: true,
		},
		{
			in: "http://localhost:8080/foo/bar.tar.gz",
			ok: true,
		},
	} {
		u, err := url.Parse(tc.in)
		assert.NoError(t, err)
		err = updateCheckHost(ctx, u)
		if tc.ok {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}
