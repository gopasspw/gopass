package gtk

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
)

/*
 * GtkTooltip
 */

type Tooltip struct {
	Widget
}

// native returns a pointer to the underlying GtkIconView.
func (t *Tooltip) native() *C.GtkTooltip {
	if t == nil || t.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(t.GObject)
	return C.toGtkTooltip(p)
}

func marshalTooltip(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapTooltip(obj), nil
}

func wrapTooltip(obj *glib.Object) *Tooltip {
	return &Tooltip{Widget{glib.InitiallyUnowned{obj}}}
}

// SetMarkup is a wrapper around gtk_tooltip_set_markup().
func (t *Tooltip) SetMarkup(str string) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_tooltip_set_markup(t.native(), (*C.gchar)(cstr))
}

// SetText is a wrapper around gtk_tooltip_set_text().
func (t *Tooltip) SetText(str string) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_tooltip_set_text(t.native(), (*C.gchar)(cstr))
}

// SetIcon is a wrapper around gtk_tooltip_set_icon().
func (t *Tooltip) SetIcon(pixbuf *gdk.Pixbuf) {
	C.gtk_tooltip_set_icon(t.native(),
		(*C.GdkPixbuf)(unsafe.Pointer(pixbuf.Native())))
}

// SetIconFromIconName is a wrapper around gtk_tooltip_set_icon_from_icon_name().
func (t *Tooltip) SetIconFromIconName(iconName string, size IconSize) {
	cstr := C.CString(iconName)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_tooltip_set_icon_from_icon_name(t.native(),
		(*C.gchar)(cstr),
		C.GtkIconSize(size))
}

// func (t *Tooltip) SetIconFromGIcon() { }

// SetCustom is a wrapper around gtk_tooltip_set_custom().
func (t *Tooltip) SetCustom(w *Widget) {
	C.gtk_tooltip_set_custom(t.native(), w.native())
}

// SetTipArea is a wrapper around gtk_tooltip_set_tip_area().
func (t *Tooltip) SetTipArea(rect gdk.Rectangle) {
	C.gtk_tooltip_set_tip_area(t.native(), nativeGdkRectangle(rect))
}
