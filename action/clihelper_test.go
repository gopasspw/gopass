package action

import (
	"context"
	"testing"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/stretchr/testify/assert"
)

func Test_askForGitConfigUser(t *testing.T) {
	ctx := context.Background()

	s := New(semver.Version{})

	ctx = ctxutil.WithTerminal(ctx, true)

	_, _, err := s.askForGitConfigUser(ctx)
	if err == nil {
		t.Error("Did not return error")
	}
}

func Test_askForGitConfigUserNonInteractive(t *testing.T) {
	ctx := context.Background()

	s := New(semver.Version{})

	ctx = ctxutil.WithTerminal(ctx, false)

	keyList, err := s.gpg.ListPrivateKeys(ctx)
	if err != nil {
		t.Error(err.Error())
	}

	name, email, _ := s.askForGitConfigUser(ctx)

	// unit tests cannot know whether keyList returned empty or not.
	// a better distinction would require mocking/patching
	// calls to s.gpg.ListPrivateKeys()
	if len(keyList) > 0 {
		assert.NotEqual(t, "", name)
		assert.NotEqual(t, "", email)
	} else {
		assert.Equal(t, "", name)
		assert.Equal(t, "", email)
	}
}
