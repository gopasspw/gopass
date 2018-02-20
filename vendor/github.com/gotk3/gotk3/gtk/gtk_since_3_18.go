// +build !gtk_3_6,!gtk_3_8,!gtk_3_10,!gtk_3_12,!gtk_3_14,!gtk_3_16,gtk_3_18

// See: https://developer.gnome.org/gtk3/3.18/api-index-3-18.html

// For gtk_overlay_reorder_overlay():
// See: https://git.gnome.org/browse/gtk+/tree/gtk/gtkoverlay.h?h=gtk-3-18

package gtk

// #cgo pkg-config: gtk+-3.0
// #include <gtk/gtk.h>
import "C"

// ReorderOverlay() is a wrapper around gtk_overlay_reorder_overlay().
func (v *Overlay) ReorderOverlay(child IWidget, position int) {
	C.gtk_overlay_reorder_overlay(v.native(), child.toWidget(), C.gint(position))
}

// GetOverlayPassThrough() is a wrapper around
// gtk_overlay_get_overlay_pass_through().
func (v *Overlay) GetOverlayPassThrough(widget IWidget) bool {
	c := C.gtk_overlay_get_overlay_pass_through(v.native(), widget.toWidget())
	return gobool(c)
}

// SetOverlayPassThrough() is a wrapper around
// gtk_overlay_set_overlay_pass_through().
func (v *Overlay) SetOverlayPassThrough(widget IWidget, passThrough bool) {
	C.gtk_overlay_set_overlay_pass_through(v.native(), widget.toWidget(), gbool(passThrough))
}
