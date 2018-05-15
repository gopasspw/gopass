package manifest

import (
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
)

// Render returns the rendered wrapper and manifest
func Render(browser, wrapperPath, libPath, binPath string, global bool) ([]byte, []byte, error) {
	mf, err := getManifestContent(browser, wrapperPath)
	if err != nil {
		return nil, nil, err
	}

	if binPath == "" {
		binPath = gopassPath(global)
	}
	wrap, err := getWrapperContent(binPath)
	if err != nil {
		return nil, nil, err
	}

	return wrap, mf, nil
}

// ValidBrowser returns true if the given browser is supported on this platform
func ValidBrowser(name string) bool {
	_, found := manifestPath[runtime.GOOS][name]
	return found
}

// ValidBrowsers are all browsers for which the manifest can be currently installed
func ValidBrowsers() []string {
	keys := make([]string, 0, len(manifestPath[runtime.GOOS]))
	for k := range manifestPath[runtime.GOOS] {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func gopassPath(global bool) string {
	if !global {
		if hd, err := homedir.Dir(); err == nil {
			if gpp, err := os.Executable(); err == nil && strings.HasPrefix(gpp, hd) {
				return gpp
			}
		}
	}
	if gpp, err := exec.LookPath("gopass"); err == nil {
		return gpp
	}
	return "gopass"
}
