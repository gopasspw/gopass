// Same copyright and license as the rest of the files in this project
// This file contains accelerator related functions and structures

// +build !gtk_3_6,!gtk_3_8
// not use this: go build -tags gtk_3_8'. Otherwise, if no build tags are used, GTK 3.10

package gtk

// #cgo pkg-config: gtk+-3.0
// #include <stdlib.h>
// #include <gtk/gtk.h>
// #include "gtk_since_3_10.go.h"
import "C"

/*
 * GtkWindow
 */

// SetTitlebar is a wrapper around gtk_window_set_titlebar().
func (v *Window) SetTitlebar(titlebar IWidget) {
	C.gtk_window_set_titlebar(v.native(), titlebar.toWidget())
}

// Close is a wrapper around gtk_window_close().
func (v *Window) Close() {
	C.gtk_window_close(v.native())
}
