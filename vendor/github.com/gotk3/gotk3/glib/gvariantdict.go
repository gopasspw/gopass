// Same copyright and license as the rest of the files in this project

//GVariant : GVariant — strongly typed value datatype
// https://developer.gnome.org/glib/2.26/glib-GVariant.html

package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
// #include "gvariant.go.h"
import "C"
import "unsafe"

/*
 * GVariantDict
 */

// VariantDict is a representation of GLib's VariantDict.
type VariantDict struct {
	GVariantDict *C.GVariantDict
}

func (v *VariantDict) toGVariantDict() *C.GVariantDict {
	if v == nil {
		return nil
	}
	return v.native()
}

func (v *VariantDict) toVariantDict() *VariantDict {
	return v
}

// newVariantDict creates a new VariantDict from a GVariantDict pointer.
func newVariantDict(p *C.GVariantDict) *VariantDict {
	return &VariantDict{GVariantDict: p}
}

// native returns a pointer to the underlying GVariantDict.
func (v *VariantDict) native() *C.GVariantDict {
	if v == nil || v.GVariantDict == nil {
		return nil
	}
	p := unsafe.Pointer(v.GVariantDict)
	return C.toGVariantDict(p)
}

// Native returns a pointer to the underlying GVariantDict.
func (v *VariantDict) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}
