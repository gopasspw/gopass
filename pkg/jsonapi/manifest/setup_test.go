package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	manifestGolden = `{
    "name": "com.justwatch.gopass",
    "description": "Gopass wrapper to search and return passwords",
    "path": "/tmp/",
    "type": "stdio",
    "allowed_origins": [
        "chrome-extension://kkhfnlkhiapbiehimabddjbimfaijdhk/"
    ]
}`
)

func TestRender(t *testing.T) {
	w, m, err := Render("chrome", "/tmp/", "gopass", true)
	assert.NoError(t, err)
	assert.Equal(t, wrapperGolden, string(w))
	assert.Equal(t, manifestGolden, string(m))
}

func TestValidBrowser(t *testing.T) {
	for _, b := range []string{"chrome", "chromium", "firefox"} {
		assert.Equal(t, true, ValidBrowser(b))
	}
}

func TestValidBrowsers(t *testing.T) {
	assert.Equal(t, []string{"chrome", "chromium", "firefox"}, ValidBrowsers())
}
