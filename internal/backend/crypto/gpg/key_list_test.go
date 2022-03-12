package gpg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyList(t *testing.T) {
	t.Parallel()

	kl := KeyList{
		genTestKey("John", "johnny", "Doe", "john.doe@example.org"),
		genTestKey("Jane", "jane", "Doe", "jane.doe@example.org", "25FF1614B8F87B52FFFF99B962AF4031C82E0019"),
		genTestKey("Jim", "jimmy", "Doe", "jim.doe@example.org", "25FF1614B8F87B52FFFF99B962AF4031C82E2019", "z", "none"),
	}
	kl[2].SubKeys = map[string]struct{}{
		"0xDEADBEEF": {},
	}

	assert.Equal(t, []string{
		"0x62AF4031C82E0019",
		"0x62AF4031C82E0039",
		"0x62AF4031C82E2019",
	}, kl.Recipients())
	assert.Equal(t, []string{
		"0x62AF4031C82E0019",
		"0x62AF4031C82E0039",
	}, kl.UseableKeys(false).Recipients())
	assert.Equal(t, []string{
		"0x62AF4031C82E2019",
	}, kl.UnusableKeys(false).Recipients())

	// search by email
	k, err := kl.FindKey("jim.doe@example.org")
	assert.NoError(t, err)
	assert.Equal(t, "0x62AF4031C82E2019", k.ID())

	// search by fp
	k, err = kl.FindKey("25FF1614B8F87B52FFFF99B962AF4031C82E2019")
	assert.NoError(t, err)
	assert.Equal(t, "0x62AF4031C82E2019", k.ID())

	// search by id
	k, err = kl.FindKey("0x62AF4031C82E2019")
	assert.NoError(t, err)
	assert.Equal(t, "0x62AF4031C82E2019", k.ID())

	// search for non existing key
	k, err = kl.FindKey("0x62AF4091C82E2019")
	assert.Error(t, err)
	assert.Equal(t, "", k.ID())

	// search by full name
	k, err = kl.FindKey("John Doe")
	assert.NoError(t, err)
	assert.Equal(t, "0x62AF4031C82E0039", k.ID())

	// search by subkey id
	k, err = kl.FindKey("0xDEADBEEF")
	assert.NoError(t, err)
	assert.Equal(t, "0x62AF4031C82E2019", k.ID())
}
