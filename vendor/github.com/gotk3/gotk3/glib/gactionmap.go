package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <gio/gio.h>
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"
import "unsafe"

// ActionMap is a representation of glib's GActionMap GInterface
type ActionMap struct {
	*Object
}

// void g_action_map_add_action_entries (GActionMap *action_map, const GActionEntry *entries, gint n_entries, gpointer user_data);
// struct GActionEntry

// native() returns a pointer to the underlying GActionMap.
func (v *ActionMap) native() *C.GActionMap {
	if v == nil || v.GObject == nil {
		return nil
	}
	return C.toGActionMap(unsafe.Pointer(v.GObject))
}

func (v *ActionMap) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalActionMap(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return wrapActionMap(wrapObject(unsafe.Pointer(c))), nil
}

func wrapActionMap(obj *Object) *ActionMap {
	return &ActionMap{obj}
}

// LookupAction is a wrapper around g_action_map_lookup_action
func (v *ActionMap) LookupAction(actionName string) *Action {
	c := C.g_action_map_lookup_action(v.native(), (*C.gchar)(C.CString(actionName)))
	if c == nil {
		return nil
	}
	return wrapAction(wrapObject(unsafe.Pointer(c)))
}

// AddAction is a wrapper around g_action_map_add_action
func (v *ActionMap) AddAction(action IAction) {
	C.g_action_map_add_action(v.native(), action.toGAction())
}

// RemoveAction is a wrapper around g_action_map_remove_action
func (v *ActionMap) RemoveAction(actionName string) {
	C.g_action_map_remove_action(v.native(), (*C.gchar)(C.CString(actionName)))
}
