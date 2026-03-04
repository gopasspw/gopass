package termio

import (
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestDetectName(t *testing.T) {
	ctx := config.NewContextInMemory()
	td := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", td)
	t.Setenv("GOPASS_HOMEDIR", td)

	t.Setenv("GIT_AUTHOR_NAME", "")
	t.Setenv("DEBFULLNAME", "")
	t.Setenv("USER", "")

	assert.Empty(t, DetectName(ctx, nil))

	t.Setenv("USER", "foo")
	assert.Equal(t, "foo", DetectName(ctx, nil))
}

func TestDetectEmail(t *testing.T) {
	ctx := config.NewContextInMemory()
	td := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", td)
	t.Setenv("GOPASS_HOMEDIR", td)

	t.Setenv("GIT_AUTHOR_EMAIL", "")
	t.Setenv("DEBEMAIL", "")
	t.Setenv("EMAIL", "")

	assert.Empty(t, DetectEmail(ctx, nil))

	t.Setenv("EMAIL", "foo@bar.de")
	assert.Equal(t, "foo@bar.de", DetectEmail(ctx, nil))
}
