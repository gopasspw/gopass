// +build !linux no_x11

package gdk

func (v *Window) MoveToCurrentDesktop() {
}

// GetDesktop is a wrapper around gdk_x11_window_get_desktop().
// It only works on GDK versions compiled with X11 support - its return value can't be used if WorkspaceControlSupported returns false
func (v *Window) GetDesktop() uint32 {
	return 0
}

// MoveToDesktop is a wrapper around gdk_x11_window_move_to_desktop().
// It only works on GDK versions compiled with X11 support - its return value can't be used if WorkspaceControlSupported returns false
func (v *Window) MoveToDesktop(d uint32) {
}
