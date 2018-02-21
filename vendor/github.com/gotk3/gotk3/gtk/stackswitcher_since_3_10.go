// Same copyright and license as the rest of the files in this project
// This file contains accelerator related functions and structures

// +build !gtk_3_6,!gtk_3_8
// not use this: go build -tags gtk_3_8'. Otherwise, if no build tags are used, GTK 3.10

package gtk

// #cgo pkg-config: gtk+-3.0
// #include <stdlib.h>
// #include <gtk/gtk.h>
// #include "gtk_since_3_10.go.h"
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

func init() {
	//Contribute to casting
	for k, v := range map[string]WrapFn{
		"GtkStackSwitcher": wrapStackSwitcher,
	} {
		WrapMap[k] = v
	}
}

/*
 * GtkStackSwitcher
 */

// StackSwitcher is a representation of GTK's GtkStackSwitcher
type StackSwitcher struct {
	Box
}

// native returns a pointer to the underlying GtkStackSwitcher.
func (v *StackSwitcher) native() *C.GtkStackSwitcher {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkStackSwitcher(p)
}

func marshalStackSwitcher(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapStackSwitcher(obj), nil
}

func wrapStackSwitcher(obj *glib.Object) *StackSwitcher {
	return &StackSwitcher{Box{Container{Widget{glib.InitiallyUnowned{obj}}}}}
}

// StackSwitcherNew is a wrapper around gtk_stack_switcher_new().
func StackSwitcherNew() (*StackSwitcher, error) {
	c := C.gtk_stack_switcher_new()
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapStackSwitcher(glib.Take(unsafe.Pointer(c))), nil
}

// SetStack is a wrapper around gtk_stack_switcher_set_stack().
func (v *StackSwitcher) SetStack(stack *Stack) {
	C.gtk_stack_switcher_set_stack(v.native(), stack.native())
}

// GetStack is a wrapper around gtk_stack_switcher_get_stack().
func (v *StackSwitcher) GetStack() *Stack {
	c := C.gtk_stack_switcher_get_stack(v.native())
	if c == nil {
		return nil
	}
	return wrapStack(glib.Take(unsafe.Pointer(c)))
}
