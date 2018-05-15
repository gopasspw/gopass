package manifest

var (
	// DefaultBrowser to select when no browser is specified
	DefaultBrowser = "chrome"

	// DefaultWrapperPath where the gopass wrapper shell script is installed to
	DefaultWrapperPath = "/usr/local/bin"

	// Name is the name of the manifest
	Name = "com.justwatch.gopass"
	// WrapperName is the name of the gopass wrapper
	WrapperName = "gopass_wrapper.sh"

	description    = "Gopass wrapper to search and return passwords"
	connectionType = "stdio"
	chromeOrigins  = []string{
		"chrome-extension://kkhfnlkhiapbiehimabddjbimfaijdhk/", // gopassbridge
	}
	firefoxOrigins = []string{
		"{eec37db0-22ad-4bf1-9068-5ae08df8c7e9}", // gopassbridge
	}

	globalManifestPath = map[string]map[string]string{
		"darwin": {
			"firefox":  "/Library/Application Support/Mozilla/NativeMessagingHosts",
			"chrome":   "/Library/Google/Chrome/NativeMessagingHosts",
			"chromium": "/Library/Application Support/Chromium/NativeMessagingHosts",
		},
		"linux": {
			"firefox":  "mozilla/native-messaging-hosts", // will be prefixed with the appropriate lib path
			"chrome":   "/etc/opt/chrome/native-messaging-hosts",
			"chromium": "/etc/chromium/native-messaging-hosts",
		},
	}

	manifestPath = map[string]map[string]string{
		"darwin": {
			"firefox":  "~/Library/Application Support/Mozilla/NativeMessagingHosts",
			"chrome":   "~/Library/Application Support/Google/Chrome/NativeMessagingHosts",
			"chromium": "~/Library/Application Support/Chromium/NativeMessagingHosts",
		},
		"linux": {
			"firefox":  "~/.mozilla/native-messaging-hosts",
			"chrome":   "~/.config/google-chrome/NativeMessagingHosts",
			"chromium": "~/.config/chromium/NativeMessagingHosts",
		},
	}
)
