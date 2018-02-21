// Same copyright and license as the rest of the files in this project
// This file contains accelerator related functions and structures

// +build !gtk_3_6,!gtk_3_8,!gtk_3_10
// not use this: go build -tags gtk_3_8'. Otherwise, if no build tags are used, GTK 3.10

package gtk

// #cgo pkg-config: gtk+-3.0
// #include <stdlib.h>
// #include <gtk/gtk.h>
// #include "gtk_since_3_10.go.h"
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

// GetChildByName is a wrapper around gtk_stack_get_child_by_name().
func (v *Stack) GetChildByName(name string) *Widget {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_stack_get_child_by_name(v.native(), (*C.gchar)(cstr))
	if c == nil {
		return nil
	}
	return wrapWidget(glib.Take(unsafe.Pointer(c)))
}
