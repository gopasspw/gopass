package manifest

import (
	"fmt"
	"runtime"
)

var globalLocations = map[string]map[string]string{
	"darwin": {
		"firefox":  "/Library/Application Support/Mozilla/NativeMessagingHosts/%s.json",
		"chrome":   "/Library/Google/Chrome/NativeMessagingHosts/%s.json",
		"chromium": "/Library/Application Support/Chromium/NativeMessagingHosts/%s.json",
	},
	"linux": {
		"firefox":  "mozilla/native-messaging-hosts/%s.json",
		"chrome":   "/etc/opt/chrome/native-messaging-hosts/%s.json",
		"chromium": "/etc/chromium/native-messaging-hosts/%s.json",
	},
}

var locations = map[string]map[string]string{
	"darwin": {
		"firefox":  "~/Library/Application Support/Mozilla/NativeMessagingHosts/%s.json",
		"chrome":   "~/Library/Application Support/Google/Chrome/NativeMessagingHosts/%s.json",
		"chromium": "~/Library/Application Support/Chromium/NativeMessagingHosts/%s.json",
	},
	"linux": {
		"firefox":  "~/.mozilla/native-messaging-hosts/%s.json",
		"chrome":   "~/.config/google-chrome/NativeMessagingHosts/%s.json",
		"chromium": "~/.config/chromium/NativeMessagingHosts/%s.json",
	},
}

func getLocation(browser, libpath string, globalInstall bool) (string, error) {
	platform := runtime.GOOS
	if globalInstall {
		pm, found := globalLocations[platform]
		if !found {
			return "", fmt.Errorf("platform %s is currently not supported", platform)
		}
		path, found := pm[browser]
		if !found {
			return "", fmt.Errorf("browser %s on %s is currently not supported", browser, platform)
		}
		if browser == "firefox" {
			path = libpath + "/" + path
		}
		return path, nil
	}

	pm, found := locations[platform]
	if !found {
		return "", fmt.Errorf("platform %s is currently not supported", platform)
	}
	path, found := pm[browser]
	if !found {
		return "", fmt.Errorf("browser %s on %s is currently not supported", browser, platform)
	}
	return path, nil
}
