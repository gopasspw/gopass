// Same copyright and license as the rest of the files in this project

// GVariant : GVariant — strongly typed value datatype
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
 * GVariantBuilder
 */

// VariantBuilder is a representation of GLib's VariantBuilder.
type VariantBuilder struct {
	GVariantBuilder *C.GVariantBuilder
}

func (v *VariantBuilder) toGVariantBuilder() *C.GVariantBuilder {
	if v == nil {
		return nil
	}
	return v.native()
}

func (v *VariantBuilder) toVariantBuilder() *VariantBuilder {
	return v
}

// newVariantBuilder creates a new VariantBuilder from a GVariantBuilder pointer.
func newVariantBuilder(p *C.GVariantBuilder) *VariantBuilder {
	return &VariantBuilder{GVariantBuilder: p}
}

// native returns a pointer to the underlying GVariantBuilder.
func (v *VariantBuilder) native() *C.GVariantBuilder {
	if v == nil || v.GVariantBuilder == nil {
		return nil
	}
	p := unsafe.Pointer(v.GVariantBuilder)
	return C.toGVariantBuilder(p)
}

// Native returns a pointer to the underlying GVariantBuilder.
func (v *VariantBuilder) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}
