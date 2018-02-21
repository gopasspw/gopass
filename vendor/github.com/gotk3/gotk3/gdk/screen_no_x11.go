// +build !linux no_x11

package gdk

func WorkspaceControlSupported() bool {
	return false
}

// GetScreenNumber is a wrapper around gdk_x11_screen_get_screen_number().
// It only works on GDK versions compiled with X11 support - its return value can't be used if WorkspaceControlSupported returns false
func (v *Screen) GetScreenNumber() int {
	return -1
}

// GetNumberOfDesktops is a wrapper around gdk_x11_screen_get_number_of_desktops().
// It only works on GDK versions compiled with X11 support - its return value can't be used if WorkspaceControlSupported returns false
func (v *Screen) GetNumberOfDesktops() uint32 {
	return 0
}

// GetCurrentDesktop is a wrapper around gdk_x11_screen_get_current_desktop().
// It only works on GDK versions compiled with X11 support - its return value can't be used if WorkspaceControlSupported returns false
func (v *Screen) GetCurrentDesktop() uint32 {
	return 0
}
