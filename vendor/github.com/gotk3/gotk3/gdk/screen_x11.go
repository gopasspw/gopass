// +build linux
// +build !no_x11

package gdk

// #cgo pkg-config: gdk-x11-3.0
// #include <gdk/gdk.h>
// #include <gdk/gdkx.h>
import "C"

func WorkspaceControlSupported() bool {
	return true
}

// GetScreenNumber is a wrapper around gdk_x11_screen_get_screen_number().
// It only works on GDK versions compiled with X11 support - its return value can't be used if WorkspaceControlSupported returns false
func (v *Screen) GetScreenNumber() int {
	return int(C.gdk_x11_screen_get_screen_number(v.native()))
}

// GetNumberOfDesktops is a wrapper around gdk_x11_screen_get_number_of_desktops().
// It only works on GDK versions compiled with X11 support - its return value can't be used if WorkspaceControlSupported returns false
func (v *Screen) GetNumberOfDesktops() uint32 {
	return uint32(C.gdk_x11_screen_get_number_of_desktops(v.native()))
}

// GetCurrentDesktop is a wrapper around gdk_x11_screen_get_current_desktop().
// It only works on GDK versions compiled with X11 support - its return value can't be used if WorkspaceControlSupported returns false
func (v *Screen) GetCurrentDesktop() uint32 {
	return uint32(C.gdk_x11_screen_get_current_desktop(v.native()))
}
