// +build !gtk_3_6,!gtk_3_8,!gtk_3_10,!gtk_3_12,!gtk_3_14,!gtk_3_16,gtk_3_18

// See: https://developer.gnome.org/gtk3/3.18/api-index-3-18.html

package gtk

// #cgo pkg-config: gtk+-3.0
// #include <gtk/gtk.h>
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

//void
//gtk_popover_set_default_widget (GtkPopover *popover, GtkWidget *widget);
func (p *Popover) SetDefaultWidget(widget IWidget) {
	C.gtk_popover_set_default_widget(p.native(), widget.toWidget())
}

//GtkWidget *
//gtk_popover_get_default_widget (GtkPopover *popover);
func (p *Popover) GetDefaultWidget() *Widget {
	w := C.gtk_popover_get_default_widget(p.native())
	if w == nil {
		return nil
	}
	return &Widget{glib.InitiallyUnowned{glib.Take(unsafe.Pointer(w))}}
}
