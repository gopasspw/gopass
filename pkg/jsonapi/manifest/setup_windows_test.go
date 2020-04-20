package manifest

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	binDir := "C:\\My\\bin"
	manifestGolden := `{
    "name": "com.justwatch.gopass",
    "description": "Gopass wrapper to search and return passwords",
    "path": "` + strings.Replace(binDir, "\\", "\\\\", -1) + `",
    "type": "stdio",
    "allowed_origins": [
        "chrome-extension://kkhfnlkhiapbiehimabddjbimfaijdhk/"
    ]
}`
	w, m, err := Render("chrome", binDir, "gopass", true)
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
