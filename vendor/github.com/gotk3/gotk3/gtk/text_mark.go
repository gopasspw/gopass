// Same copyright and license as the rest of the files in this project

package gtk

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"

import "unsafe"

/*
 * GtkTextMark
 */

// TextMark is a representation of GTK's GtkTextMark
type TextMark C.GtkTextMark

// native returns a pointer to the underlying GtkTextMark.
func (v *TextMark) native() *C.GtkTextMark {
	if v == nil {
		return nil
	}
	return (*C.GtkTextMark)(v)
}

func marshalTextMark(p uintptr) (interface{}, error) {
	c := C.g_value_get_boxed((*C.GValue)(unsafe.Pointer(p)))
	return (*TextMark)(unsafe.Pointer(c)), nil
}
