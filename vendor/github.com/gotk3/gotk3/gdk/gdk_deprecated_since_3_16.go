//+build gtk_3_6 gtk_3_8 gtk_3_10 gtk_3_12 gtk_3_14

package gdk

// #cgo pkg-config: gdk-3.0
// #include <gdk/gdk.h>
import "C"

// SupportsComposite() is a wrapper around gdk_display_supports_composite().
func (v *Display) SupportsComposite() bool {
	c := C.gdk_display_supports_composite(v.native())
	return gobool(c)
}
