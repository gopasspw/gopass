package gtk

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
)

func init() {
	tm := []glib.TypeMarshaler{
		{glib.Type(C.gtk_color_chooser_get_type()), marshalColorChooser},
		{glib.Type(C.gtk_color_chooser_dialog_get_type()), marshalColorChooserDialog},
	}

	glib.RegisterGValueMarshalers(tm)

	WrapMap["GtkColorChooser"] = wrapColorChooser
	WrapMap["GtkColorChooserDialog"] = wrapColorChooserDialog
}

/*
 * GtkColorChooser
 */

// ColorChooser is a representation of GTK's GtkColorChooser GInterface.
type ColorChooser struct {
	*glib.Object
}

// IColorChooser is an interface type implemented by all structs
// embedding an ColorChooser. It is meant to be used as an argument type
// for wrapper functions that wrap around a C GTK function taking a
// GtkColorChooser.
type IColorChooser interface {
	toColorChooser() *C.GtkColorChooser
}

// native returns a pointer to the underlying GtkAppChooser.
func (v *ColorChooser) native() *C.GtkColorChooser {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkColorChooser(p)
}

func marshalColorChooser(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapColorChooser(obj), nil
}

func wrapColorChooser(obj *glib.Object) *ColorChooser {
	return &ColorChooser{obj}
}

func (v *ColorChooser) toColorChooser() *C.GtkColorChooser {
	if v == nil {
		return nil
	}
	return v.native()
}

// GetRGBA() is a wrapper around gtk_color_chooser_get_rgba().
func (v *ColorChooser) GetRGBA() *gdk.RGBA {
	gdkColor := gdk.NewRGBA()
	C.gtk_color_chooser_get_rgba(v.native(), (*C.GdkRGBA)(unsafe.Pointer(gdkColor.Native())))
	return gdkColor
}

// SetRGBA() is a wrapper around gtk_color_chooser_set_rgba().
func (v *ColorChooser) SetRGBA(gdkColor *gdk.RGBA) {
	C.gtk_color_chooser_set_rgba(v.native(), (*C.GdkRGBA)(unsafe.Pointer(gdkColor.Native())))
}

// GetUseAlpha() is a wrapper around gtk_color_chooser_get_use_alpha().
func (v *ColorChooser) GetUseAlpha() bool {
	return gobool(C.gtk_color_chooser_get_use_alpha(v.native()))
}

// SetUseAlpha() is a wrapper around gtk_color_chooser_set_use_alpha().
func (v *ColorChooser) SetUseAlpha(use_alpha bool) {
	C.gtk_color_chooser_set_use_alpha(v.native(), gbool(use_alpha))
}

// AddPalette() is a wrapper around gtk_color_chooser_add_palette().
func (v *ColorChooser) AddPalette(orientation Orientation, colors_per_line int, colors []*gdk.RGBA) {
	n_colors := len(colors)
	var c_colors []C.GdkRGBA
	for _, c := range colors {
		c_colors = append(c_colors, *(*C.GdkRGBA)(unsafe.Pointer(c.Native())))
	}
	C.gtk_color_chooser_add_palette(
		v.native(),
		C.GtkOrientation(orientation),
		C.gint(colors_per_line),
		C.gint(n_colors),
		&c_colors[0],
	)
}

/*
 * GtkColorChooserDialog
 */

// ColorChooserDialog is a representation of GTK's GtkColorChooserDialog.
type ColorChooserDialog struct {
	Dialog

	// Interfaces
	ColorChooser
}

// native returns a pointer to the underlying GtkColorChooserButton.
func (v *ColorChooserDialog) native() *C.GtkColorChooserDialog {
	if v == nil || v.GObject == nil {
		return nil
	}

	p := unsafe.Pointer(v.GObject)
	return C.toGtkColorChooserDialog(p)
}

func marshalColorChooserDialog(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return wrapColorChooserDialog(glib.Take(unsafe.Pointer(c))), nil
}

func wrapColorChooserDialog(obj *glib.Object) *ColorChooserDialog {
	dialog := wrapDialog(obj)
	cc := wrapColorChooser(obj)
	return &ColorChooserDialog{*dialog, *cc}
}

// ColorChooserDialogNew() is a wrapper around gtk_color_chooser_dialog_new().
func ColorChooserDialogNew(title string, parent *Window) (*ColorChooserDialog, error) {
	cstr := C.CString(title)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_color_chooser_dialog_new((*C.gchar)(cstr), parent.native())
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapColorChooserDialog(glib.Take(unsafe.Pointer(c))), nil
}
