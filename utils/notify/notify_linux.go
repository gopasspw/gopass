// +build linux

package notify

import (
	"os"

	"github.com/godbus/dbus"
)

// Notify displays a desktop notification with dbus
func Notify(subj, msg string) error {
	if nn := os.Getenv("GOPASS_NO_NOTIFY"); nn != "" {
		return nil
	}
	conn, err := dbus.SessionBus()
	if err != nil {
		return err
	}

	obj := conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")
	call := obj.Call("org.freedesktop.Notifications.Notify", 0, "gopass", uint32(0), iconURI(), subj, msg, []string{}, map[string]dbus.Variant{}, int32(5000))
	if call.Err != nil {
		return err
	}

	return nil
}
