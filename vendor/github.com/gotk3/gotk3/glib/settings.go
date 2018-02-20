package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <gio/gio.h>
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"
import "unsafe"

// Settings is a representation of GSettings.
type Settings struct {
	*Object
}

// native() returns a pointer to the underlying GSettings.
func (v *Settings) native() *C.GSettings {
	if v == nil || v.GObject == nil {
		return nil
	}
	return C.toGSettings(unsafe.Pointer(v.GObject))
}

func (v *Settings) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalSettings(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return wrapSettings(wrapObject(unsafe.Pointer(c))), nil
}

func wrapSettings(obj *Object) *Settings {
	return &Settings{obj}
}

func wrapFullSettings(obj *C.GSettings) *Settings {
	if obj == nil {
		return nil
	}
	return wrapSettings(wrapObject(unsafe.Pointer(obj)))
}

// SettingsNew is a wrapper around g_settings_new().
func SettingsNew(schemaID string) *Settings {
	cstr := (*C.gchar)(C.CString(schemaID))
	defer C.free(unsafe.Pointer(cstr))

	return wrapFullSettings(C.g_settings_new(cstr))
}

// SettingsNewWithPath is a wrapper around g_settings_new_with_path().
func SettingsNewWithPath(schemaID, path string) *Settings {
	cstr1 := (*C.gchar)(C.CString(schemaID))
	defer C.free(unsafe.Pointer(cstr1))

	cstr2 := (*C.gchar)(C.CString(path))
	defer C.free(unsafe.Pointer(cstr2))

	return wrapFullSettings(C.g_settings_new_with_path(cstr1, cstr2))
}

// SettingsNewWithBackend is a wrapper around g_settings_new_with_backend().
func SettingsNewWithBackend(schemaID string, backend *SettingsBackend) *Settings {
	cstr1 := (*C.gchar)(C.CString(schemaID))
	defer C.free(unsafe.Pointer(cstr1))

	return wrapFullSettings(C.g_settings_new_with_backend(cstr1, backend.native()))
}

// SettingsNewWithBackendAndPath is a wrapper around g_settings_new_with_backend_and_path().
func SettingsNewWithBackendAndPath(schemaID string, backend *SettingsBackend, path string) *Settings {
	cstr1 := (*C.gchar)(C.CString(schemaID))
	defer C.free(unsafe.Pointer(cstr1))

	cstr2 := (*C.gchar)(C.CString(path))
	defer C.free(unsafe.Pointer(cstr2))

	return wrapFullSettings(C.g_settings_new_with_backend_and_path(cstr1, backend.native(), cstr2))
}

// SettingsNewFull is a wrapper around g_settings_new_full().
func SettingsNewFull(schema *SettingsSchema, backend *SettingsBackend, path string) *Settings {
	cstr1 := (*C.gchar)(C.CString(path))
	defer C.free(unsafe.Pointer(cstr1))

	return wrapFullSettings(C.g_settings_new_full(schema.native(), backend.native(), cstr1))
}

// SettingsSync is a wrapper around g_settings_sync().
func SettingsSync() {
	C.g_settings_sync()
}

// IsWritable is a wrapper around g_settings_is_writable().
func (v *Settings) IsWritable(name string) bool {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	return gobool(C.g_settings_is_writable(v.native(), cstr1))
}

// Delay is a wrapper around g_settings_delay().
func (v *Settings) Delay() {
	C.g_settings_delay(v.native())
}

// Apply is a wrapper around g_settings_apply().
func (v *Settings) Apply() {
	C.g_settings_apply(v.native())
}

// Revert is a wrapper around g_settings_revert().
func (v *Settings) Revert() {
	C.g_settings_revert(v.native())
}

// GetHasUnapplied is a wrapper around g_settings_get_has_unapplied().
func (v *Settings) GetHasUnapplied() bool {
	return gobool(C.g_settings_get_has_unapplied(v.native()))
}

// GetChild is a wrapper around g_settings_get_child().
func (v *Settings) GetChild(name string) *Settings {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	return wrapFullSettings(C.g_settings_get_child(v.native(), cstr1))
}

// Reset is a wrapper around g_settings_reset().
func (v *Settings) Reset(name string) {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	C.g_settings_reset(v.native(), cstr1)
}

// ListChildren is a wrapper around g_settings_list_children().
func (v *Settings) ListChildren() []string {
	return toGoStringArray(C.g_settings_list_children(v.native()))
}

// GetBoolean is a wrapper around g_settings_get_boolean().
func (v *Settings) GetBoolean(name string) bool {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	return gobool(C.g_settings_get_boolean(v.native(), cstr1))
}

// SetBoolean is a wrapper around g_settings_set_boolean().
func (v *Settings) SetBoolean(name string, value bool) bool {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	return gobool(C.g_settings_set_boolean(v.native(), cstr1, gbool(value)))
}

// GetInt is a wrapper around g_settings_get_int().
func (v *Settings) GetInt(name string) int {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	return int(C.g_settings_get_int(v.native(), cstr1))
}

// SetInt is a wrapper around g_settings_set_int().
func (v *Settings) SetInt(name string, value int) bool {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	return gobool(C.g_settings_set_int(v.native(), cstr1, C.gint(value)))
}

// GetUInt is a wrapper around g_settings_get_uint().
func (v *Settings) GetUInt(name string) uint {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	return uint(C.g_settings_get_uint(v.native(), cstr1))
}

// SetUInt is a wrapper around g_settings_set_uint().
func (v *Settings) SetUInt(name string, value uint) bool {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	return gobool(C.g_settings_set_uint(v.native(), cstr1, C.guint(value)))
}

// GetDouble is a wrapper around g_settings_get_double().
func (v *Settings) GetDouble(name string) float64 {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	return float64(C.g_settings_get_double(v.native(), cstr1))
}

// SetDouble is a wrapper around g_settings_set_double().
func (v *Settings) SetDouble(name string, value float64) bool {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	return gobool(C.g_settings_set_double(v.native(), cstr1, C.gdouble(value)))
}

// GetString is a wrapper around g_settings_get_string().
func (v *Settings) GetString(name string) string {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	return C.GoString((*C.char)(C.g_settings_get_string(v.native(), cstr1)))
}

// SetString is a wrapper around g_settings_set_string().
func (v *Settings) SetString(name string, value string) bool {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	cstr2 := (*C.gchar)(C.CString(value))
	defer C.free(unsafe.Pointer(cstr2))

	return gobool(C.g_settings_set_string(v.native(), cstr1, cstr2))
}

// GetEnum is a wrapper around g_settings_get_enum().
func (v *Settings) GetEnum(name string) int {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	return int(C.g_settings_get_enum(v.native(), cstr1))
}

// GetStrv is a wrapper around g_settings_get_strv().
func (v *Settings) GetStrv(name string) []string {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))
	return toGoStringArray(C.g_settings_get_strv(v.native(), cstr1))
}

// SetStrv is a wrapper around g_settings_set_strv().
func (v *Settings) SetStrv(name string, values []string) bool {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	cvalues := make([]*C.gchar, len(values))
	for i, accel := range values {
		cvalues[i] = (*C.gchar)(C.CString(accel))
		defer C.free(unsafe.Pointer(cvalues[i]))
	}
	cvalues = append(cvalues, nil)

	return gobool(C.g_settings_set_strv(v.native(), cstr1, &cvalues[0]))
}

// SetEnum is a wrapper around g_settings_set_enum().
func (v *Settings) SetEnum(name string, value int) bool {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	return gobool(C.g_settings_set_enum(v.native(), cstr1, C.gint(value)))
}

// GetFlags is a wrapper around g_settings_get_flags().
func (v *Settings) GetFlags(name string) uint {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	return uint(C.g_settings_get_flags(v.native(), cstr1))
}

// SetFlags is a wrapper around g_settings_set_flags().
func (v *Settings) SetFlags(name string, value uint) bool {
	cstr1 := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr1))

	return gobool(C.g_settings_set_flags(v.native(), cstr1, C.guint(value)))
}

func (v *Settings) GetValue(name string) *Variant {
	cstr := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr))
	return newVariant(C.g_settings_get_value(v.native(), cstr))
}

// GVariant * 	g_settings_get_value ()
// gboolean 	g_settings_set_value ()
// GVariant * 	g_settings_get_user_value ()
// GVariant * 	g_settings_get_default_value ()
// const gchar * const * 	g_settings_list_schemas ()
// const gchar * const * 	g_settings_list_relocatable_schemas ()
// gchar ** 	g_settings_list_keys ()
// GVariant * 	g_settings_get_range ()
// gboolean 	g_settings_range_check ()
// void 	g_settings_get ()
// gboolean 	g_settings_set ()
// gpointer 	g_settings_get_mapped ()
// void 	g_settings_bind ()
// void 	g_settings_bind_with_mapping ()
// void 	g_settings_bind_writable ()
// void 	g_settings_unbind ()
// gaction * 	g_settings_create_action ()
// gchar ** 	g_settings_get_strv ()
// gboolean 	g_settings_set_strv ()
