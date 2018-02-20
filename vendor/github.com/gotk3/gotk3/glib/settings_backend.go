package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"
import "unsafe"

// SettingsBackend is a representation of GSettingsBackend.
type SettingsBackend struct {
	*Object
}

// native() returns a pointer to the underlying GSettingsBackend.
func (v *SettingsBackend) native() *C.GSettingsBackend {
	if v == nil || v.GObject == nil {
		return nil
	}
	return C.toGSettingsBackend(unsafe.Pointer(v.GObject))
}

func (v *SettingsBackend) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalSettingsBackend(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return wrapSettingsBackend(wrapObject(unsafe.Pointer(c))), nil
}

func wrapSettingsBackend(obj *Object) *SettingsBackend {
	return &SettingsBackend{obj}
}

// SettingsBackendGetDefault is a wrapper around g_settings_backend_get_default().
func SettingsBackendGetDefault() *SettingsBackend {
	return wrapSettingsBackend(wrapObject(unsafe.Pointer(C.g_settings_backend_get_default())))
}

// KeyfileSettingsBackendNew is a wrapper around g_keyfile_settings_backend_new().
func KeyfileSettingsBackendNew(filename, rootPath, rootGroup string) *SettingsBackend {
	cstr1 := (*C.gchar)(C.CString(filename))
	defer C.free(unsafe.Pointer(cstr1))

	cstr2 := (*C.gchar)(C.CString(rootPath))
	defer C.free(unsafe.Pointer(cstr2))

	cstr3 := (*C.gchar)(C.CString(rootGroup))
	defer C.free(unsafe.Pointer(cstr3))

	return wrapSettingsBackend(wrapObject(unsafe.Pointer(C.g_keyfile_settings_backend_new(cstr1, cstr2, cstr3))))
}

// MemorySettingsBackendNew is a wrapper around g_memory_settings_backend_new().
func MemorySettingsBackendNew() *SettingsBackend {
	return wrapSettingsBackend(wrapObject(unsafe.Pointer(C.g_memory_settings_backend_new())))
}

// NullSettingsBackendNew is a wrapper around g_null_settings_backend_new().
func NullSettingsBackendNew() *SettingsBackend {
	return wrapSettingsBackend(wrapObject(unsafe.Pointer(C.g_null_settings_backend_new())))
}

// void 	g_settings_backend_changed ()
// void 	g_settings_backend_path_changed ()
// void 	g_settings_backend_keys_changed ()
// void 	g_settings_backend_path_writable_changed ()
// void 	g_settings_backend_writable_changed ()
// void 	g_settings_backend_changed_tree ()
// void 	g_settings_backend_flatten_tree ()
