package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <gio/gio.h>
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"
import "unsafe"

// SettingsSchemaSource is a representation of GSettingsSchemaSource.
type SettingsSchemaSource struct {
	source *C.GSettingsSchemaSource
}

func wrapSettingsSchemaSource(obj *C.GSettingsSchemaSource) *SettingsSchemaSource {
	if obj == nil {
		return nil
	}
	return &SettingsSchemaSource{obj}
}

func (v *SettingsSchemaSource) Native() uintptr {
	return uintptr(unsafe.Pointer(v.source))
}

func (v *SettingsSchemaSource) native() *C.GSettingsSchemaSource {
	if v == nil || v.source == nil {
		return nil
	}
	return v.source
}

// SettingsSchemaSourceGetDefault is a wrapper around g_settings_schema_source_get_default().
func SettingsSchemaSourceGetDefault() *SettingsSchemaSource {
	return wrapSettingsSchemaSource(C.g_settings_schema_source_get_default())
}

// Ref() is a wrapper around g_settings_schema_source_ref().
func (v *SettingsSchemaSource) Ref() *SettingsSchemaSource {
	return wrapSettingsSchemaSource(C.g_settings_schema_source_ref(v.native()))
}

// Unref() is a wrapper around g_settings_schema_source_unref().
func (v *SettingsSchemaSource) Unref() {
	C.g_settings_schema_source_unref(v.native())
}

// SettingsSchemaSourceNewFromDirectory() is a wrapper around g_settings_schema_source_new_from_directory().
func SettingsSchemaSourceNewFromDirectory(dir string, parent *SettingsSchemaSource, trusted bool) *SettingsSchemaSource {
	cstr := (*C.gchar)(C.CString(dir))
	defer C.free(unsafe.Pointer(cstr))

	return wrapSettingsSchemaSource(C.g_settings_schema_source_new_from_directory(cstr, parent.native(), gbool(trusted), nil))
}

// Lookup() is a wrapper around g_settings_schema_source_lookup().
func (v *SettingsSchemaSource) Lookup(schema string, recursive bool) *SettingsSchema {
	cstr := (*C.gchar)(C.CString(schema))
	defer C.free(unsafe.Pointer(cstr))

	return wrapSettingsSchema(C.g_settings_schema_source_lookup(v.native(), cstr, gbool(recursive)))
}

// ListSchemas is a wrapper around 	g_settings_schema_source_list_schemas().
func (v *SettingsSchemaSource) ListSchemas(recursive bool) (nonReolcatable, relocatable []string) {
	var nonRel, rel **C.gchar
	C.g_settings_schema_source_list_schemas(v.native(), gbool(recursive), &nonRel, &rel)
	return toGoStringArray(nonRel), toGoStringArray(rel)
}
