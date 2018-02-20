// Same copyright and license as the rest of the files in this project

//GVariant : GVariant — strongly typed value datatype
// https://developer.gnome.org/glib/2.26/glib-GVariant.html

package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <glib.h>
// #include "gvarianttype.go.h"
import "C"

// A VariantType is a wrapper for the GVariantType, which encodes type
// information for GVariants.
type VariantType struct {
	GVariantType *C.GVariantType
}

func (v *VariantType) native() *C.GVariantType {
	if v == nil {
		return nil
	}
	return v.GVariantType
}

// String returns a copy of this VariantType's type string.
func (v *VariantType) String() string {
	ch := C.g_variant_type_dup_string(v.native())
	defer C.g_free(C.gpointer(ch))
	return C.GoString((*C.char)(ch))
}

func newVariantType(v *C.GVariantType) *VariantType {
	return &VariantType{v}
}

// Variant types for comparing between them.  Cannot be const because
// they are pointers.
var (
	VARIANT_TYPE_BOOLEAN           = newVariantType(C._G_VARIANT_TYPE_BOOLEAN)
	VARIANT_TYPE_BYTE              = newVariantType(C._G_VARIANT_TYPE_BYTE)
	VARIANT_TYPE_INT16             = newVariantType(C._G_VARIANT_TYPE_INT16)
	VARIANT_TYPE_UINT16            = newVariantType(C._G_VARIANT_TYPE_UINT16)
	VARIANT_TYPE_INT32             = newVariantType(C._G_VARIANT_TYPE_INT32)
	VARIANT_TYPE_UINT32            = newVariantType(C._G_VARIANT_TYPE_UINT32)
	VARIANT_TYPE_INT64             = newVariantType(C._G_VARIANT_TYPE_INT64)
	VARIANT_TYPE_UINT64            = newVariantType(C._G_VARIANT_TYPE_UINT64)
	VARIANT_TYPE_HANDLE            = newVariantType(C._G_VARIANT_TYPE_HANDLE)
	VARIANT_TYPE_DOUBLE            = newVariantType(C._G_VARIANT_TYPE_DOUBLE)
	VARIANT_TYPE_STRING            = newVariantType(C._G_VARIANT_TYPE_STRING)
	VARIANT_TYPE_ANY               = newVariantType(C._G_VARIANT_TYPE_ANY)
	VARIANT_TYPE_BASIC             = newVariantType(C._G_VARIANT_TYPE_BASIC)
	VARIANT_TYPE_TUPLE             = newVariantType(C._G_VARIANT_TYPE_TUPLE)
	VARIANT_TYPE_UNIT              = newVariantType(C._G_VARIANT_TYPE_UNIT)
	VARIANT_TYPE_DICTIONARY        = newVariantType(C._G_VARIANT_TYPE_DICTIONARY)
	VARIANT_TYPE_STRING_ARRAY      = newVariantType(C._G_VARIANT_TYPE_STRING_ARRAY)
	VARIANT_TYPE_OBJECT_PATH_ARRAY = newVariantType(C._G_VARIANT_TYPE_OBJECT_PATH_ARRAY)
	VARIANT_TYPE_BYTESTRING        = newVariantType(C._G_VARIANT_TYPE_BYTESTRING)
	VARIANT_TYPE_BYTESTRING_ARRAY  = newVariantType(C._G_VARIANT_TYPE_BYTESTRING_ARRAY)
	VARIANT_TYPE_VARDICT           = newVariantType(C._G_VARIANT_TYPE_VARDICT)
)
