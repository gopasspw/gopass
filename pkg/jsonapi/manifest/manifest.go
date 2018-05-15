package manifest

import (
	"encoding/json"
	"fmt"
)

func getManifestContent(browser, wrapperPath string) ([]byte, error) {
	switch browser {
	case "firefox":
		return newFirefoxManifest(wrapperPath).Format()
	case "chrome":
		fallthrough
	case "chromium":
		return newChromeManifest(wrapperPath).Format()
	default:
		return nil, fmt.Errorf("no manifest template for browser %s", browser)
	}
}

type chromeManifest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Path           string   `json:"path"`
	Type           string   `json:"type"`
	AllowedOrigins []string `json:"allowed_origins"`
}

func newChromeManifest(path string) *chromeManifest {
	return &chromeManifest{
		Name:           Name,
		Type:           connectionType,
		Path:           path,
		Description:    description,
		AllowedOrigins: chromeOrigins,
	}
}

func (c *chromeManifest) Format() ([]byte, error) {
	return json.MarshalIndent(c, "", "    ")
}

type firefoxManifest struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Path              string   `json:"path"`
	Type              string   `json:"type"`
	AllowedExtensions []string `json:"allowed_extensions"`
}

func newFirefoxManifest(path string) *firefoxManifest {
	return &firefoxManifest{
		Name:              Name,
		Type:              connectionType,
		Path:              path,
		Description:       description,
		AllowedExtensions: firefoxOrigins,
	}
}

func (f *firefoxManifest) Format() ([]byte, error) {
	return json.MarshalIndent(f, "", "    ")
}
