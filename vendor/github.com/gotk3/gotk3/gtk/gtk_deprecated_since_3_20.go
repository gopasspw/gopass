//+build gtk_3_6 gtk_3_8 gtk_3_10 gtk_3_12 gtk_3_14 gtk_3_16 gtk_3_18

package gtk

// #cgo pkg-config: gtk+-3.0
// #include <gtk/gtk.h>
// #include <stdlib.h>
import "C"

// GetFocusOnClick() is a wrapper around gtk_button_get_focus_on_click().
func (v *Button) GetFocusOnClick() bool {
	c := C.gtk_button_get_focus_on_click(v.native())
	return gobool(c)
}

// BeginsTag is a wrapper around gtk_text_iter_begins_tag().
func (v *TextIter) BeginsTag(v1 *TextTag) bool {
	return gobool(C.gtk_text_iter_begins_tag(v.native(), v1.native()))
}

// ResizeToGeometry is a wrapper around gtk_window_resize_to_geometry().
func (v *Window) ResizeToGeometry(width, height int) {
	C.gtk_window_resize_to_geometry(v.native(), C.gint(width), C.gint(height))
}

// SetDefaultGeometry is a wrapper around gtk_window_set_default_geometry().
func (v *Window) SetDefaultGeometry(width, height int) {
	C.gtk_window_set_default_geometry(v.native(), C.gint(width),
		C.gint(height))
}

// SetFocusOnClick() is a wrapper around gtk_button_set_focus_on_click().
func (v *Button) SetFocusOnClick(focusOnClick bool) {
	C.gtk_button_set_focus_on_click(v.native(), gbool(focusOnClick))
}
