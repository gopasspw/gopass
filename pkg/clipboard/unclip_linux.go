//go:build linux

package clipboard

import (
	"context"
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
)

func clearClipboardHistory(ctx context.Context) error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("failed to connect to session bus: %w", err)
	}

	obj := conn.Object("org.kde.klipper", "/klipper")
	call := obj.Call("org.kde.klipper.klipper.clearClipboardHistory", 0)

	if call.Err != nil {
		if strings.HasPrefix(call.Err.Error(), "The name org.kde.klipper was not provided") {
			return nil
		}

		if strings.HasPrefix(call.Err.Error(), "The name is not activatable") {
			return nil
		}

		return call.Err
	}

	return nil
}
