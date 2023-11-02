package gpg

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func genTestKey(args ...string) Key {
	first := "John"
	if len(args) > 0 && args[0] != "" {
		first = args[0]
	}

	nick := "johnny"
	if len(args) > 1 && args[1] != "" {
		nick = args[1]
	}

	last := "Doe"
	if len(args) > 2 && args[2] != "" {
		last = args[2]
	}

	email := "john.doe@example.org"
	if len(args) > 3 && args[3] != "" {
		email = args[3]
	}

	fp := "25FF1614B8F87B52FFFF99B962AF4031C82E0039"
	if len(args) > 4 && args[4] != "" {
		fp = args[4]
	}

	validity := "u"
	if len(args) > 5 && args[5] != "" {
		validity = args[5]
	}

	trust := "ultimate"
	if len(args) > 6 && args[6] != "" {
		trust = args[6]
	}

	creation := time.Date(2018, 1, 1, 1, 1, 1, 0, time.UTC)
	expiration := time.Date(2218, 1, 1, 1, 1, 1, 0, time.UTC)

	return Key{
		KeyType:        "sec",
		KeyLength:      2048,
		Validity:       validity,
		CreationDate:   creation,
		ExpirationDate: expiration,
		Ownertrust:     trust,
		Fingerprint:    fp,
		Identities: map[string]Identity{
			fmt.Sprintf("%s %s (%s) <%s>", first, last, nick, email): {
				Name:           fmt.Sprintf("%s %s", first, last),
				Comment:        nick,
				Email:          email,
				CreationDate:   creation,
				ExpirationDate: expiration,
			},
		},
		Caps: Capabilities{
			Encrypt:        true,
			Sign:           false,
			Certify:        false,
			Authentication: false,
			Deactivated:    false,
		},
	}
}

func TestKey(t *testing.T) {
	t.Parallel()

	k := Key{
		Identities: map[string]Identity{},
	}
	assert.Equal(t, "(invalid:)", k.OneLine())
	assert.Equal(t, "", k.Identity().Name)
	k = genTestKey()
	assert.True(t, k.IsUseable(false))
	assert.Equal(t, "sec   2048D/0x62AF4031C82E0039 2018-01-01 [expires: 2218-01-01]\n      Key fingerprint = 25FF1614B8F87B52FFFF99B962AF4031C82E0039\nuid                            John Doe (johnny) <john.doe@example.org>", k.String())
	assert.Equal(t, "0x62AF4031C82E0039 - John Doe (johnny) <john.doe@example.org>", k.OneLine())
	assert.Equal(t, "0x62AF4031C82E0039", k.ID())
}

func TestIdentitySort(t *testing.T) {
	t.Parallel()

	creation := time.Date(2017, 1, 1, 1, 1, 1, 0, time.UTC)
	expiration := time.Date(2018, 1, 1, 1, 1, 1, 0, time.UTC)

	k := genTestKey()
	k.Identities["Foo Bar"] = Identity{
		Name:           "Foo Bar",
		Comment:        "foo",
		Email:          "foo.bar@example.com",
		CreationDate:   creation,
		ExpirationDate: expiration,
	}
	assert.Equal(t, "0x62AF4031C82E0039 - John Doe (johnny) <john.doe@example.org>", k.OneLine())
}

func TestUseability(t *testing.T) {
	t.Parallel()

	// invalid
	for _, k := range []Key{
		{},
		{
			ExpirationDate: time.Now().Add(-time.Second),
			Caps:           Capabilities{Encrypt: true},
		},
		{
			ExpirationDate: time.Now().Add(time.Hour),
			Caps:           Capabilities{Encrypt: true},
			Validity:       "z",
		},
		{
			ExpirationDate: time.Now().Add(time.Hour),
			Caps:           Capabilities{Deactivated: true},
		},
		{
			ExpirationDate: time.Now().Add(time.Hour),
			Caps:           Capabilities{Encrypt: false},
		},
	} {
		assert.False(t, k.IsUseable(false))
	}
	// valid
	for _, k := range []Key{
		{
			ExpirationDate: time.Now().Add(time.Hour),
			Validity:       "m",
			Caps:           Capabilities{Encrypt: true},
		},
		{
			ExpirationDate: time.Now().Add(time.Hour),
			Validity:       "f",
			Caps:           Capabilities{Encrypt: true},
		},
		{
			ExpirationDate: time.Now().Add(time.Hour),
			Validity:       "u",
			Caps:           Capabilities{Encrypt: true},
		},
	} {
		assert.True(t, k.IsUseable(false))
	}
}
