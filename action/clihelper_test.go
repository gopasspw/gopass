package action

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_askForGitConfigUser(t *testing.T) {
	s := New("test-init")
	s.isTerm = true

	_, _, err := s.askForGitConfigUser()
	if err == nil {
		t.Error("Did not return error")
	}
}

func Test_askForGitConfigUserNonInteractive(t *testing.T) {
	s := New("test-init")
	// for explicitness
	s.isTerm = false

	keyList, err := s.gpg.ListPrivateKeys()
	if err != nil {
		t.Error(err.Error())
	}

	name, email, _ := s.askForGitConfigUser()

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
