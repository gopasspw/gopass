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

/*
 * GVariantClass
 */

type VariantClass int

const (
	VARIANT_CLASS_BOOLEAN     VariantClass = C.G_VARIANT_CLASS_BOOLEAN     //The GVariant is a boolean.
	VARIANT_CLASS_BYTE        VariantClass = C.G_VARIANT_CLASS_BYTE        //The GVariant is a byte.
	VARIANT_CLASS_INT16       VariantClass = C.G_VARIANT_CLASS_INT16       //The GVariant is a signed 16 bit integer.
	VARIANT_CLASS_UINT16      VariantClass = C.G_VARIANT_CLASS_UINT16      //The GVariant is an unsigned 16 bit integer.
	VARIANT_CLASS_INT32       VariantClass = C.G_VARIANT_CLASS_INT32       //The GVariant is a signed 32 bit integer.
	VARIANT_CLASS_UINT32      VariantClass = C.G_VARIANT_CLASS_UINT32      //The GVariant is an unsigned 32 bit integer.
	VARIANT_CLASS_INT64       VariantClass = C.G_VARIANT_CLASS_INT64       //The GVariant is a signed 64 bit integer.
	VARIANT_CLASS_UINT64      VariantClass = C.G_VARIANT_CLASS_UINT64      //The GVariant is an unsigned 64 bit integer.
	VARIANT_CLASS_HANDLE      VariantClass = C.G_VARIANT_CLASS_HANDLE      //The GVariant is a file handle index.
	VARIANT_CLASS_DOUBLE      VariantClass = C.G_VARIANT_CLASS_DOUBLE      //The GVariant is a double precision floating point value.
	VARIANT_CLASS_STRING      VariantClass = C.G_VARIANT_CLASS_STRING      //The GVariant is a normal string.
	VARIANT_CLASS_OBJECT_PATH VariantClass = C.G_VARIANT_CLASS_OBJECT_PATH //The GVariant is a D-Bus object path string.
	VARIANT_CLASS_SIGNATURE   VariantClass = C.G_VARIANT_CLASS_SIGNATURE   //The GVariant is a D-Bus signature string.
	VARIANT_CLASS_VARIANT     VariantClass = C.G_VARIANT_CLASS_VARIANT     //The GVariant is a variant.
	VARIANT_CLASS_MAYBE       VariantClass = C.G_VARIANT_CLASS_MAYBE       //The GVariant is a maybe-typed value.
	VARIANT_CLASS_ARRAY       VariantClass = C.G_VARIANT_CLASS_ARRAY       //The GVariant is an array.
	VARIANT_CLASS_TUPLE       VariantClass = C.G_VARIANT_CLASS_TUPLE       //The GVariant is a tuple.
	VARIANT_CLASS_DICT_ENTRY  VariantClass = C.G_VARIANT_CLASS_DICT_ENTRY  //The GVariant is a dictionary entry.
)
