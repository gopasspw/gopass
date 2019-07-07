// +build !windows

package manifest

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/mitchellh/go-homedir"
)

var (
	// WrapperName is the name of the gopass wrapper
	WrapperName = "gopass_wrapper.sh"

	globalManifestPath = map[string]map[string]string{
		"darwin": {
			"firefox":  "/Library/Application Support/Mozilla/NativeMessagingHosts",
			"chrome":   "/Library/Google/Chrome/NativeMessagingHosts",
			"chromium": "/Library/Application Support/Chromium/NativeMessagingHosts",
			"brave":    "/Library/Application Support/Brave/NativeMessagingHosts",
			"vivaldi":  "/Library/Application Support/Vivaldi/NativeMessagingHosts",
			"iridium":  "/Library/Application Support/Iridium/NativeMessagingHosts",
			"slimjet":  "/Library/Application Support/Slimjet/NativeMessagingHosts",
		},
		"linux": {
			"firefox":  "mozilla/native-messaging-hosts", // will be prefixed with the appropriate lib path
			"chrome":   "/etc/opt/chrome/native-messaging-hosts",
			"chromium": "/etc/chromium/native-messaging-hosts",
			"brave":    "/etc/opt/chrome/native-messaging-hosts",
			"vivaldi":  "/etc/opt/vivaldi/native-messaging-hosts",
			"iridium":  "/etc/iridium-browser/native-messaging-hosts",
			"slimjet":  "/etc/opt/slimjet/native-messaging-hosts",
		},
	}

	manifestPath = map[string]map[string]string{
		"darwin": {
			"firefox":  "~/Library/Application Support/Mozilla/NativeMessagingHosts",
			"chrome":   "~/Library/Application Support/Google/Chrome/NativeMessagingHosts",
			"chromium": "~/Library/Application Support/Chromium/NativeMessagingHosts",
			"brave":    "~/Library/Application Support/Brave/NativeMessagingHosts",
			"vivaldi":  "~/Library/Application Support/Vivaldi/NativeMessagingHosts",
			"iridium":  "~/Library/Application Support/Iridium/NativeMessagingHosts",
			"slimjet":  "~/Library/Application Support/Slimjet/NativeMessagingHosts",
		},
		"linux": {
			"firefox":  "~/.mozilla/native-messaging-hosts",
			"chrome":   "~/.config/google-chrome/NativeMessagingHosts",
			"chromium": "~/.config/chromium/NativeMessagingHosts",
			"brave":    "~/.config/BraveSoftware/Brave-Browser/NativeMessagingHosts",
			"vivaldi":  "~/.config/vivaldi/NativeMessagingHosts",
			"iridium":  "~/.config/iridium/NativeMessagingHosts",
			"slimjet":  "~/.config/slimjet/NativeMessagingHosts",
		},
	}
)

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

// Path returns the manifest file path
func Path(browser, libpath string, globalInstall bool) (string, error) {
	location, err := getLocation(browser, libpath, globalInstall)
	if err != nil {
		return "", err
	}

	expanded, err := homedir.Expand(location)
	if err != nil {
		return "", err
	}

	return filepath.Join(expanded, Name+".json"), nil
}

// getLocation returns only the manifest path
func getLocation(browser, libpath string, globalInstall bool) (string, error) {
	if globalInstall {
		return getGlobalLocation(browser, libpath)
	}

	pm, found := manifestPath[runtime.GOOS]
	if !found {
		return "", fmt.Errorf("platform %s is currently not supported", runtime.GOOS)
	}
	path, found := pm[browser]
	if !found {
		return "", fmt.Errorf("browser %s on %s is currently not supported", browser, runtime.GOOS)
	}
	return path, nil
}

func getGlobalLocation(browser, libpath string) (string, error) {
	pm, found := globalManifestPath[runtime.GOOS]
	if !found {
		return "", fmt.Errorf("platform %s is currently not supported", runtime.GOOS)
	}
	path, found := pm[browser]
	if !found {
		return "", fmt.Errorf("browser %s on %s is currently not supported", browser, runtime.GOOS)
	}
	if browser == "firefox" {
		path = libpath + "/" + path
	}
	return path, nil
}
