// +build gtk_3_6 gtk_3_8 gtk_3_10 gtk_3_12 gtk_3_14 gtk_3_16 gtk_3_18 gtk_3_20

package gtk

// #cgo pkg-config: gtk+-3.0
// #include <stdlib.h>
// #include <gtk/gtk.h>
import "C"
import "github.com/gotk3/gotk3/gdk"

// PopupAtPointer() is a wrapper for gtk_menu_popup_at_pointer(), on older versions it uses PopupAtMouseCursor
func (v *Menu) PopupAtPointer(_ *gdk.Event) {
	C.gtk_menu_popup(v.native(),
		nil,
		nil,
		nil,
		nil,
		C.guint(0),
		C.gtk_get_current_event_time())
}
