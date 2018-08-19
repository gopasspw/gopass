// +build windows

package manifest

import (
	"fmt"
	"runtime"
	"sort"
)

var (
	// NativeHostExeName is the name of the gopass wrapper binary
	NativeHostExeName = "gopass_native_host.exe"

	// windows stores the path to the manifests in the registry
	registryPaths = map[string]string{
		"firefox":  `Software\Mozilla\NativeMessagingHosts\` + Name,
		"chrome":   `Software\Google\Chrome\NativeMessagingHosts\` + Name,
		"chromium": `Software\Google\Chrome\NativeMessagingHosts\` + Name,
	}
)

// ValidBrowser returns true if the given browser is supported on this platform
func ValidBrowser(name string) bool {
	_, found := registryPaths[name]
	return found
}

// ValidBrowsers are all browsers for which the manifest can be currently installed
func ValidBrowsers() []string {
	keys := make([]string, 0, len(registryPaths))
	for k := range registryPaths {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// GetRegistryPath returns the relative registry path to use in the windows registry key
func GetRegistryPath(browser string) (string, error) {
	path, found := registryPaths[browser]
	if !found {
		return "", fmt.Errorf("browser %s on %s is currently not supported", browser, runtime.GOOS)
	}
	return path, nil
}
