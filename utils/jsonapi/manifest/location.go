package manifest

import (
	"fmt"
	"path"
	"runtime"
)

func getLocation(browser, libpath string, globalInstall bool) (string, error) {
	switch platform := runtime.GOOS; platform {
	case "darwin":
		{
			switch browser {
			case "firefox":
				{
					if globalInstall {
						return "/Library/Application Support/Mozilla/NativeMessagingHosts/%s.json", nil
					}
					return "~/Library/Application Support/Mozilla/NativeMessagingHosts/%s.json", nil
				}
			case "chrome":
				{
					if globalInstall {
						return "/Library/Google/Chrome/NativeMessagingHosts/%s.json", nil
					}
					return "~/Library/Application Support/Google/Chrome/NativeMessagingHosts/%s.json", nil
				}
			case "chromium":
				{
					if globalInstall {
						return "/Library/Application Support/Chromium/NativeMessagingHosts/%s.json", nil
					}
					return "~/Library/Application Support/Chromium/NativeMessagingHosts/%s.json", nil
				}
			}
		}
	case "linux":
		{
			switch browser {
			case "firefox":
				{
					if globalInstall {
						return path.Join(libpath, "mozilla/native-messaging-hosts/%s.json"), nil
					}
					return "~/.mozilla/native-messaging-hosts/%s.json", nil
				}
			case "chrome":
				{
					if globalInstall {
						return "/etc/opt/chrome/native-messaging-hosts/%s.json", nil
					}
					return "~/.config/google-chrome/NativeMessagingHosts/%s.json", nil
				}
			case "chromium":
				{
					if globalInstall {
						return "/etc/chromium/native-messaging-hosts/%s.json", nil
					}
					return "~/.config/chromium/NativeMessagingHosts/%s.json", nil
				}
			}
		}
	default:
		{
			return "", fmt.Errorf("platform %s is currently not supported", platform)
		}
	}
	return "", nil
}
