package manifest

var (
	// DefaultBrowser to select when no browser is specified
	DefaultBrowser = "chrome"

	// Name is the name of the manifest
	Name = "com.justwatch.gopass"
	// WrapperName is the name of the gopass wrapper
	WrapperName = "gopass_wrapper.sh"
	// WrapperNameWindows is the name of the gopass wrapper binary
	WrapperNameWindows = "gopass_native_host.exe"

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
		// windows stores the path to the manifests in the registry
		"windows": {
			"firefox":  `Software\Mozilla\NativeMessagingHosts\` + Name,
			"chrome":   `Software\Google\Chrome\NativeMessagingHosts\` + Name,
			"chromium": `Software\Google\Chrome\NativeMessagingHosts\` + Name,
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
		"windows": {
			"firefox":  `Software\Mozilla\NativeMessagingHosts\` + Name,
			"chrome":   `Software\Google\Chrome\NativeMessagingHosts\` + Name,
			"chromium": `Software\Google\Chrome\NativeMessagingHosts\` + Name,
		},
	}
)
