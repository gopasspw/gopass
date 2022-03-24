package gpg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIdentity(t *testing.T) {
	t.Parallel()

	id := Identity{
		Name:           "John Doe",
		Comment:        "johnny",
		Email:          "john.doe@example.org",
		CreationDate:   time.Now(),
		ExpirationDate: time.Now().Add(time.Hour),
	}

	assert.Equal(t, id.ID(), "John Doe (johnny) <john.doe@example.org>")
	assert.Equal(t, id.String(), "uid                            "+id.ID())
}
