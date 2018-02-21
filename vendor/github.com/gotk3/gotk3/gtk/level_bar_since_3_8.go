// +build !gtk_3_6

package gtk

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"

// SetInverted() is a wrapper around gtk_level_bar_set_inverted().
func (v *LevelBar) SetInverted(inverted bool) {
	C.gtk_level_bar_set_inverted(v.native(), gbool(inverted))
}

// GetInverted() is a wrapper around gtk_level_bar_get_inverted().
func (v *LevelBar) GetInverted() bool {
	c := C.gtk_level_bar_get_inverted(v.native())
	return gobool(c)
}
