//go:build !windows
// +build !windows

package gitconfig

// SystemConfig is the location of the (optional) system-wide config defaults file.
var systemConfig = "/etc/gitconfig" // /etc/gopass/config
