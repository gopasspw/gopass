package termio

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectName(t *testing.T) {
	ctx := context.Background()

	oga := os.Getenv("GIT_AUTHOR_NAME")
	odf := os.Getenv("DEBFULLNAME")
	ous := os.Getenv("USER")

	defer func() {
		_ = os.Setenv("GIT_AUTHOR_NAME", oga)
		_ = os.Setenv("DEBFULLNAME", odf)
		_ = os.Setenv("USER", ous)
	}()

	_ = os.Unsetenv("GIT_AUTHOR_NAME")
	_ = os.Unsetenv("DEBFULLNAME")
	_ = os.Unsetenv("USER")

	assert.Equal(t, "", DetectName(ctx, nil))

	t.Setenv("USER", "foo")

	assert.Equal(t, "foo", DetectName(ctx, nil))
}

func TestDetectEmail(t *testing.T) {
	ctx := context.Background()

	oga := os.Getenv("GIT_AUTHOR_EMAIL")
	odf := os.Getenv("DEBEMAIL")
	ous := os.Getenv("EMAIL")

	defer func() {
		_ = os.Setenv("GIT_AUTHOR_EMAIL", oga)
		_ = os.Setenv("DEBEMAIL", odf)
		_ = os.Setenv("EMAIL", ous)
	}()

	_ = os.Unsetenv("GIT_AUTHOR_EMAIL")
	_ = os.Unsetenv("DEBEMAIL")
	_ = os.Unsetenv("EMAIL")

	assert.Equal(t, "", DetectEmail(ctx, nil))
}
