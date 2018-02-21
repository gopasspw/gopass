//GVariant : GVariant — strongly typed value datatype
// https://developer.gnome.org/glib/2.26/glib-GVariant.html

package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include "gvariant.go.h"
// #include "glib.go.h"
import "C"

import (
	"fmt"
	"unsafe"
)

/*
 * GVariant
 */

// IVariant is an interface type implemented by Variant and all types which embed
// an Variant.  It is meant to be used as a type for function arguments which
// require GVariants or any subclasses thereof.
type IVariant interface {
	ToGVariant() *C.GVariant
	ToVariant() *Variant
}

// A Variant is a representation of GLib's GVariant.
type Variant struct {
	GVariant *C.GVariant
}

// ToGVariant exposes the underlying *C.GVariant type for this Variant,
// necessary to implement IVariant.
func (v *Variant) ToGVariant() *C.GVariant {
	if v == nil {
		return nil
	}
	return v.native()
}

// ToVariant returns this Variant, necessary to implement IVariant.
func (v *Variant) ToVariant() *Variant {
	return v
}

// newVariant creates a new Variant from a GVariant pointer.
func newVariant(p *C.GVariant) *Variant {
	return &Variant{GVariant: p}
}

// VariantFromUnsafePointer returns a Variant from an unsafe pointer.
// XXX: unnecessary footgun?
//func VariantFromUnsafePointer(p unsafe.Pointer) *Variant {
//	return &Variant{C.toGVariant(p)}
//}

// native returns a pointer to the underlying GVariant.
func (v *Variant) native() *C.GVariant {
	if v == nil || v.GVariant == nil {
		return nil
	}
	return v.GVariant
}

// Native returns a pointer to the underlying GVariant.
func (v *Variant) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

// TypeString returns the g variant type string for this variant.
func (v *Variant) TypeString() string {
	// the string returned from this belongs to GVariant and must not be freed.
	return C.GoString((*C.char)(C.g_variant_get_type_string(v.native())))
}

// IsContainer returns true if the variant is a container and false otherwise.
func (v *Variant) IsContainer() bool {
	return gobool(C.g_variant_is_container(v.native()))
}

// IsFloating returns true if the variant has a floating reference count.
// XXX: this isn't useful without ref_sink/take_ref, which are themselves
// perhaps not useful for most Go code that may use variants.
//func (v *Variant) IsFloating() bool {
//	return gobool(C.g_variant_is_floating(v.native()))
//}

// GetBoolean returns the bool value of this variant.
func (v *Variant) GetBoolean() bool {
	return gobool(C.g_variant_get_boolean(v.native()))
}

// GetString returns the string value of the variant.
func (v *Variant) GetString() string {
	var len C.gsize
	gc := C.g_variant_get_string(v.native(), &len)
	defer C.g_free(C.gpointer(gc))
	return C.GoStringN((*C.char)(gc), (C.int)(len))
}

// GetStrv returns a slice of strings from this variant.  It wraps
// g_variant_get_strv, but returns copies of the strings instead.
func (v *Variant) GetStrv() []string {
	gstrv := C.g_variant_get_strv(v.native(), nil)
	// we do not own the memory for these strings, so we must not use strfreev
	// but we must free the actual pointer we receive.
	c := gstrv
	defer C.g_free(C.gpointer(gstrv))
	var strs []string

	for *c != nil {
		strs = append(strs, C.GoString((*C.char)(*c)))
		c = C.next_gcharptr(c)
	}
	return strs
}

// GetInt returns the int64 value of the variant if it is an integer type, and
// an error otherwise.  It wraps variouns `g_variant_get_*` functions dealing
// with integers of different sizes.
func (v *Variant) GetInt() (int64, error) {
	t := v.Type().String()
	var i int64
	switch t {
	case "y":
		i = int64(C.g_variant_get_byte(v.native()))
	case "n":
		i = int64(C.g_variant_get_int16(v.native()))
	case "q":
		i = int64(C.g_variant_get_uint16(v.native()))
	case "i":
		i = int64(C.g_variant_get_int32(v.native()))
	case "u":
		i = int64(C.g_variant_get_uint32(v.native()))
	case "x":
		i = int64(C.g_variant_get_int64(v.native()))
	case "t":
		i = int64(C.g_variant_get_uint64(v.native()))
	default:
		return 0, fmt.Errorf("variant type %s not an integer type", t)
	}
	return i, nil
}

// Type returns the VariantType for this variant.
func (v *Variant) Type() *VariantType {
	return newVariantType(C.g_variant_get_type(v.native()))
}

// IsType returns true if the variant's type matches t.
func (v *Variant) IsType(t *VariantType) bool {
	return gobool(C.g_variant_is_of_type(v.native(), t.native()))
}

// String wraps g_variant_print().  It returns a string understood
// by g_variant_parse().
func (v *Variant) String() string {
	gc := C.g_variant_print(v.native(), gbool(false))
	defer C.g_free(C.gpointer(gc))
	return C.GoString((*C.char)(gc))
}

// AnnotatedString wraps g_variant_print(), but returns a type-annotated
// string.
func (v *Variant) AnnotatedString() string {
	gc := C.g_variant_print(v.native(), gbool(true))
	defer C.g_free(C.gpointer(gc))
	return C.GoString((*C.char)(gc))
}

//void	g_variant_unref ()
//GVariant *	g_variant_ref ()
//GVariant *	g_variant_ref_sink ()
//GVariant *	g_variant_take_ref ()
//gint	g_variant_compare ()
//GVariantClass	g_variant_classify ()
//gboolean	g_variant_check_format_string ()
//void	g_variant_get ()
//void	g_variant_get_va ()
//GVariant *	g_variant_new ()
//GVariant *	g_variant_new_va ()
//GVariant *	g_variant_new_boolean ()
//GVariant *	g_variant_new_byte ()
//GVariant *	g_variant_new_int16 ()
//GVariant *	g_variant_new_uint16 ()
//GVariant *	g_variant_new_int32 ()
//GVariant *	g_variant_new_uint32 ()
//GVariant *	g_variant_new_int64 ()
//GVariant *	g_variant_new_uint64 ()
//GVariant *	g_variant_new_handle ()
//GVariant *	g_variant_new_double ()
//GVariant *	g_variant_new_string ()
//GVariant *	g_variant_new_take_string ()
//GVariant *	g_variant_new_printf ()
//GVariant *	g_variant_new_object_path ()
//gboolean	g_variant_is_object_path ()
//GVariant *	g_variant_new_signature ()
//gboolean	g_variant_is_signature ()
//GVariant *	g_variant_new_variant ()
//GVariant *	g_variant_new_strv ()
//GVariant *	g_variant_new_objv ()
//GVariant *	g_variant_new_bytestring ()
//GVariant *	g_variant_new_bytestring_array ()
//guchar	g_variant_get_byte ()
//gint16	g_variant_get_int16 ()
//guint16	g_variant_get_uint16 ()
//gint32	g_variant_get_int32 ()
//guint32	g_variant_get_uint32 ()
//gint64	g_variant_get_int64 ()
//guint64	g_variant_get_uint64 ()
//gint32	g_variant_get_handle ()
//gdouble	g_variant_get_double ()
//const gchar *	g_variant_get_string ()
//gchar *	g_variant_dup_string ()
//GVariant *	g_variant_get_variant ()
//const gchar **	g_variant_get_strv ()
//gchar **	g_variant_dup_strv ()
//const gchar **	g_variant_get_objv ()
//gchar **	g_variant_dup_objv ()
//const gchar *	g_variant_get_bytestring ()
//gchar *	g_variant_dup_bytestring ()
//const gchar **	g_variant_get_bytestring_array ()
//gchar **	g_variant_dup_bytestring_array ()
//GVariant *	g_variant_new_maybe ()
//GVariant *	g_variant_new_array ()
//GVariant *	g_variant_new_tuple ()
//GVariant *	g_variant_new_dict_entry ()
//GVariant *	g_variant_new_fixed_array ()
//GVariant *	g_variant_get_maybe ()
//gsize	g_variant_n_children ()
//GVariant *	g_variant_get_child_value ()
//void	g_variant_get_child ()
//GVariant *	g_variant_lookup_value ()
//gboolean	g_variant_lookup ()
//gconstpointer	g_variant_get_fixed_array ()
//gsize	g_variant_get_size ()
//gconstpointer	g_variant_get_data ()
//GBytes *	g_variant_get_data_as_bytes ()
//void	g_variant_store ()
//GVariant *	g_variant_new_from_data ()
//GVariant *	g_variant_new_from_bytes ()
//GVariant *	g_variant_byteswap ()
//GVariant *	g_variant_get_normal_form ()
//gboolean	g_variant_is_normal_form ()
//guint	g_variant_hash ()
//gboolean	g_variant_equal ()
//gchar *	g_variant_print ()
//GString *	g_variant_print_string ()
//GVariantIter *	g_variant_iter_copy ()
//void	g_variant_iter_free ()
//gsize	g_variant_iter_init ()
//gsize	g_variant_iter_n_children ()
//GVariantIter *	g_variant_iter_new ()
//GVariant *	g_variant_iter_next_value ()
//gboolean	g_variant_iter_next ()
//gboolean	g_variant_iter_loop ()
//void	g_variant_builder_unref ()
//GVariantBuilder *	g_variant_builder_ref ()
//GVariantBuilder *	g_variant_builder_new ()
//void	g_variant_builder_init ()
//void	g_variant_builder_clear ()
//void	g_variant_builder_add_value ()
//void	g_variant_builder_add ()
//void	g_variant_builder_add_parsed ()
//GVariant *	g_variant_builder_end ()
//void	g_variant_builder_open ()
//void	g_variant_builder_close ()
//void	g_variant_dict_unref ()
//GVariantDict *	g_variant_dict_ref ()
//GVariantDict *	g_variant_dict_new ()
//void	g_variant_dict_init ()
//void	g_variant_dict_clear ()
//gboolean	g_variant_dict_contains ()
//gboolean	g_variant_dict_lookup ()
//GVariant *	g_variant_dict_lookup_value ()
//void	g_variant_dict_insert ()
//void	g_variant_dict_insert_value ()
//gboolean	g_variant_dict_remove ()
//GVariant *	g_variant_dict_end ()
//#define	G_VARIANT_PARSE_ERROR
//GVariant *	g_variant_parse ()
//GVariant *	g_variant_new_parsed_va ()
//GVariant *	g_variant_new_parsed ()
//gchar *	g_variant_parse_error_print_context ()
