package termio

import (
	"context"
	"os"
	"testing"

	"gotest.tools/assert"
)

func TestDetectName(t *testing.T) {
	ctx := context.Background()

	oga := os.Getenv("GIT_AUTHOR_NAME")
	odf := os.Getenv("DEBFULLNAME")
	ous := os.Getenv("USER")
	defer func() {
		os.Setenv("GIT_AUTHOR_NAME", oga)
		os.Setenv("DEBFULLNAME", odf)
		os.Setenv("USER", ous)
	}()

	os.Unsetenv("GIT_AUTHOR_NAME")
	os.Unsetenv("DEBFULLNAME")
	os.Unsetenv("USER")
	assert.Equal(t, "", DetectName(ctx, nil))
	os.Setenv("USER", "foo")
	assert.Equal(t, "foo", DetectName(ctx, nil))
}

func TestDetectEmail(t *testing.T) {
	ctx := context.Background()

	oga := os.Getenv("GIT_AUTHOR_EMAIL")
	odf := os.Getenv("DEBEMAIL")
	ous := os.Getenv("EMAIL")
	defer func() {
		os.Setenv("GIT_AUTHOR_EMAIL", oga)
		os.Setenv("DEBEMAIL", odf)
		os.Setenv("EMAIL", ous)
	}()

	os.Unsetenv("GIT_AUTHOR_EMAIL")
	os.Unsetenv("DEBEMAIL")
	os.Unsetenv("EMAIL")
	assert.Equal(t, "", DetectEmail(ctx, nil))
}
