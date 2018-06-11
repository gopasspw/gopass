// +build linux freebsd

package notify

import (
	"context"
	"os"

	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/godbus/dbus"
)

// Notify displays a desktop notification with dbus
func Notify(ctx context.Context, subj, msg string) error {
	if os.Getenv("GOPASS_NO_NOTIFY") != "" || !ctxutil.IsNotifications(ctx) {
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
