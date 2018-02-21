// +build !gtk_3_6,!gtk_3_8,!gtk_3_10,!gtk_3_12,!gtk_3_14

// See: https://developer.gnome.org/gtk3/3.16/api-index-3-16.html

package gtk

// #cgo pkg-config: gtk+-3.0
// #include <gtk/gtk.h>
// #include "gtk_since_3_16.go.h"
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

func init() {
	tm := []glib.TypeMarshaler{

		// Objects/Interfaces
		{glib.Type(C.gtk_stack_sidebar_get_type()), marshalStackSidebar},
	}
	glib.RegisterGValueMarshalers(tm)

	//Contribute to casting
	for k, v := range map[string]WrapFn{
		"GtkStackSidebar": wrapStackSidebar,
	} {
		WrapMap[k] = v
	}
}

// SetOverlayScrolling is a wrapper around gtk_scrolled_window_set_overlay_scrolling().
func (v *ScrolledWindow) SetOverlayScrolling(scrolling bool) {
	C.gtk_scrolled_window_set_overlay_scrolling(v.native(), gbool(scrolling))
}

// GetOverlayScrolling is a wrapper around gtk_scrolled_window_get_overlay_scrolling().
func (v *ScrolledWindow) GetOverlayScrolling() bool {
	return gobool(C.gtk_scrolled_window_get_overlay_scrolling(v.native()))
}

// SetWideHandle is a wrapper around gtk_paned_set_wide_handle().
func (v *Paned) SetWideHandle(wide bool) {
	C.gtk_paned_set_wide_handle(v.native(), gbool(wide))
}

// GetWideHandle is a wrapper around gtk_paned_get_wide_handle().
func (v *Paned) GetWideHandle() bool {
	return gobool(C.gtk_paned_get_wide_handle(v.native()))
}

// GetXAlign is a wrapper around gtk_label_get_xalign().
func (v *Label) GetXAlign() float64 {
	c := C.gtk_label_get_xalign(v.native())
	return float64(c)
}

// GetYAlign is a wrapper around gtk_label_get_yalign().
func (v *Label) GetYAlign() float64 {
	c := C.gtk_label_get_yalign(v.native())
	return float64(c)
}

// SetXAlign is a wrapper around gtk_label_set_xalign().
func (v *Label) SetXAlign(n float64) {
	C.gtk_label_set_xalign(v.native(), C.gfloat(n))
}

// SetYAlign is a wrapper around gtk_label_set_yalign().
func (v *Label) SetYAlign(n float64) {
	C.gtk_label_set_yalign(v.native(), C.gfloat(n))
}

/*
 * GtkStackSidebar
 */

// StackSidebar is a representation of GTK's GtkStackSidebar.
type StackSidebar struct {
	Bin
}

// native returns a pointer to the underlying GtkStack.
func (v *StackSidebar) native() *C.GtkStackSidebar {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkStackSidebar(p)
}

func marshalStackSidebar(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapStackSidebar(obj), nil
}

func wrapStackSidebar(obj *glib.Object) *StackSidebar {
	return &StackSidebar{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}}
}

// StackSidebarNew is a wrapper around gtk_stack_sidebar_new().
func StackSidebarNew() (*StackSidebar, error) {
	c := C.gtk_stack_sidebar_new()
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapStackSidebar(glib.Take(unsafe.Pointer(c))), nil
}

func (v *StackSidebar) SetStack(stack *Stack) {
	C.gtk_stack_sidebar_set_stack(v.native(), stack.native())
}

func (v *StackSidebar) GetStack() *Stack {
	c := C.gtk_stack_sidebar_get_stack(v.native())
	if c == nil {
		return nil
	}
	return wrapStack(glib.Take(unsafe.Pointer(c)))
}
