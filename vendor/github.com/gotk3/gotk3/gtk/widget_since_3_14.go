// +build !gtk_3_6,!gtk_3_8,!gtk_3_10,!gtk_3_12

// See: https://developer.gnome.org/gtk3/3.14/api-index-3-14.html

package gtk

// #cgo pkg-config: gtk+-3.0
// #include <gtk/gtk.h>
import "C"

// GetClip is a wrapper around gtk_widget_get_clip().
func (v *Widget) GetClip() *Allocation {
	var clip Allocation
	C.gtk_widget_get_clip(v.native(), clip.native())
	return &clip
}

// SetClip is a wrapper around gtk_widget_set_clip().
func (v *Widget) SetClip(clip *Allocation) {
	C.gtk_widget_set_clip(v.native(), clip.native())
}
