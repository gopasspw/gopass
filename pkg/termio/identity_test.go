package termio

import (
	"testing"

	"github.com/gopasspw/gitconfig"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
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

	cmd := &cli.Command{}
	cmd.Flags = append(cmd.Flags, &cli.StringFlag{Name: "name"})
	_ = cmd.Set("name", "bar")
	assert.Equal(t, "bar", DetectName(ctx, cmd))

	t.Setenv("GIT_AUTHOR_NAME", "Author Name")
	assert.Equal(t, "Author Name", DetectName(ctx, nil))
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

	cmd := &cli.Command{}
	cmd.Flags = append(cmd.Flags, &cli.StringFlag{Name: "email"})
	_ = cmd.Set("email", "bar@baz.de")
	assert.Equal(t, "bar@baz.de", DetectEmail(ctx, cmd))

	t.Setenv("GIT_AUTHOR_EMAIL", "bar@bar.bar")
	assert.Equal(t, "bar@bar.bar", DetectEmail(ctx, nil))
}

func TestDetectGitConfig(t *testing.T) {
	td := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", td)
	t.Setenv("GOPASS_HOMEDIR", td)

	t.Setenv("GIT_AUTHOR_NAME", "")
	t.Setenv("GIT_AUTHOR_EMAIL", " ")
	t.Setenv("DEBFULLNAME", "deb-name")
	t.Setenv("DEBEMAIL", "deb@example.com")
	t.Setenv("USER", "user-name")
	t.Setenv("EMAIL", "user@example.com")

	cfg := gitconfig.New().LoadAll(td)
	require.NoError(t, cfg.SetLocal("user.name", "Jane Doe"))
	require.NoError(t, cfg.SetLocal("user.email", "jane@example.com"))

	ctx := WithWorkdir(config.NewContextInMemory(), td)

	// Git config takes precedence over DEBFULLNAME/USER and DEBEMAIL/EMAIL.
	assert.Equal(t, "Jane Doe", DetectName(ctx, nil))
	assert.Equal(t, "jane@example.com", DetectEmail(ctx, nil))

	// GIT_AUTHOR_* env vars take precedence over git config.
	t.Setenv("GIT_AUTHOR_NAME", "Author Name")
	t.Setenv("GIT_AUTHOR_EMAIL", "author@example.com")
	assert.Equal(t, "Author Name", DetectName(ctx, nil))
	assert.Equal(t, "author@example.com", DetectEmail(ctx, nil))
}
