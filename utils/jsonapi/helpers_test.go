package jsonapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_checkIsPublicSuffix(t *testing.T) {
	a := assert.New(t)

	a.True(isPublicSuffix("co.uk"))
	a.False(isPublicSuffix("amazon.co.uk"))
	a.True(isPublicSuffix("dyndns.org"))
	a.False(isPublicSuffix("foo.dyndns.org"))
}
