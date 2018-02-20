package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <gio/gio.h>
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"
import "unsafe"

// SettingsSchema is a representation of GSettingsSchema.
type SettingsSchema struct {
	schema *C.GSettingsSchema
}

func wrapSettingsSchema(obj *C.GSettingsSchema) *SettingsSchema {
	if obj == nil {
		return nil
	}
	return &SettingsSchema{obj}
}

func (v *SettingsSchema) Native() uintptr {
	return uintptr(unsafe.Pointer(v.schema))
}

func (v *SettingsSchema) native() *C.GSettingsSchema {
	if v == nil || v.schema == nil {
		return nil
	}
	return v.schema
}

// Ref() is a wrapper around g_settings_schema_ref().
func (v *SettingsSchema) Ref() *SettingsSchema {
	return wrapSettingsSchema(C.g_settings_schema_ref(v.native()))
}

// Unref() is a wrapper around g_settings_schema_unref().
func (v *SettingsSchema) Unref() {
	C.g_settings_schema_unref(v.native())
}

// GetID() is a wrapper around g_settings_schema_get_id().
func (v *SettingsSchema) GetID() string {
	return C.GoString((*C.char)(C.g_settings_schema_get_id(v.native())))
}

// GetPath() is a wrapper around g_settings_schema_get_path().
func (v *SettingsSchema) GetPath() string {
	return C.GoString((*C.char)(C.g_settings_schema_get_path(v.native())))
}

// HasKey() is a wrapper around g_settings_schema_has_key().
func (v *SettingsSchema) HasKey(v1 string) bool {
	cstr := (*C.gchar)(C.CString(v1))
	defer C.free(unsafe.Pointer(cstr))

	return gobool(C.g_settings_schema_has_key(v.native(), cstr))
}

func toGoStringArray(c **C.gchar) []string {
	var strs []string
	originalc := c
	defer C.g_strfreev(originalc)

	for *c != nil {
		strs = append(strs, C.GoString((*C.char)(*c)))
		c = C.next_gcharptr(c)
	}

	return strs

}

// // ListChildren() is a wrapper around g_settings_schema_list_children().
// func (v *SettingsSchema) ListChildren() []string {
// 	return toGoStringArray(C.g_settings_schema_list_children(v.native()))
// }

// // ListKeys() is a wrapper around g_settings_schema_list_keys().
// func (v *SettingsSchema) ListKeys() []string {
// 	return toGoStringArray(C.g_settings_schema_list_keys(v.native()))
// }

// const GVariantType * 	g_settings_schema_key_get_value_type ()
// GVariant * 	g_settings_schema_key_get_default_value ()
// GVariant * 	g_settings_schema_key_get_range ()
// gboolean 	g_settings_schema_key_range_check ()
// const gchar * 	g_settings_schema_key_get_name ()
// const gchar * 	g_settings_schema_key_get_summary ()
// const gchar * 	g_settings_schema_key_get_description ()

// GSettingsSchemaKey * 	g_settings_schema_get_key ()
// GSettingsSchemaKey * 	g_settings_schema_key_ref ()
// void 	g_settings_schema_key_unref ()
