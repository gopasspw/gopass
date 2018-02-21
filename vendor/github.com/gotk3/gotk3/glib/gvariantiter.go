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
 * GVariantIter
 */

// VariantIter is a representation of GLib's GVariantIter.
type VariantIter struct {
	GVariantIter *C.GVariantIter
}

func (v *VariantIter) toGVariantIter() *C.GVariantIter {
	if v == nil {
		return nil
	}
	return v.native()
}

func (v *VariantIter) toVariantIter() *VariantIter {
	return v
}

// newVariantIter creates a new VariantIter from a GVariantIter pointer.
func newVariantIter(p *C.GVariantIter) *VariantIter {
	return &VariantIter{GVariantIter: p}
}

// native returns a pointer to the underlying GVariantIter.
func (v *VariantIter) native() *C.GVariantIter {
	if v == nil || v.GVariantIter == nil {
		return nil
	}
	p := unsafe.Pointer(v.GVariantIter)
	return C.toGVariantIter(p)
}

// Native returns a pointer to the underlying GVariantIter.
func (v *VariantIter) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}
