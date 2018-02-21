package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <gio/gio.h>
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"
import "unsafe"

// ActionGroup is a representation of glib's GActionGroup GInterface
type ActionGroup struct {
	*Object
}

// g_action_group_list_actions()
// g_action_group_query_action()
// should only called from implementations:
// g_action_group_action_added
// g_action_group_action_removed
// g_action_group_action_enabled_changed
// g_action_group_action_state_changed

// native() returns a pointer to the underlying GActionGroup.
func (v *ActionGroup) native() *C.GActionGroup {
	if v == nil || v.GObject == nil {
		return nil
	}
	return C.toGActionGroup(unsafe.Pointer(v.GObject))
}

func (v *ActionGroup) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalActionGroup(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return wrapActionGroup(wrapObject(unsafe.Pointer(c))), nil
}

func wrapActionGroup(obj *Object) *ActionGroup {
	return &ActionGroup{obj}
}

// HasAction is a wrapper around g_action_group_has_action().
func (v *ActionGroup) HasAction(actionName string) bool {
	return gobool(C.g_action_group_has_action(v.native(), (*C.gchar)(C.CString(actionName))))
}

// GetActionEnabled is a wrapper around g_action_group_get_action_enabled().
func (v *ActionGroup) GetActionEnabled(actionName string) bool {
	return gobool(C.g_action_group_get_action_enabled(v.native(), (*C.gchar)(C.CString(actionName))))
}

// GetActionParameterType is a wrapper around g_action_group_get_action_parameter_type().
func (v *ActionGroup) GetActionParameterType(actionName string) *VariantType {
	c := C.g_action_group_get_action_parameter_type(v.native(), (*C.gchar)(C.CString(actionName)))
	if c == nil {
		return nil
	}
	return newVariantType((*C.GVariantType)(c))
}

// GetActionStateType is a wrapper around g_action_group_get_action_state_type().
func (v *ActionGroup) GetActionStateType(actionName string) *VariantType {
	c := C.g_action_group_get_action_state_type(v.native(), (*C.gchar)(C.CString(actionName)))
	if c == nil {
		return nil
	}
	return newVariantType((*C.GVariantType)(c))
}

// GetActionState is a wrapper around g_action_group_get_action_state().
func (v *ActionGroup) GetActionState(actionName string) *Variant {
	c := C.g_action_group_get_action_state(v.native(), (*C.gchar)(C.CString(actionName)))
	if c == nil {
		return nil
	}
	return newVariant((*C.GVariant)(c))
}

// GetActionStateHint is a wrapper around g_action_group_get_action_state_hint().
func (v *ActionGroup) GetActionStateHint(actionName string) *Variant {
	c := C.g_action_group_get_action_state_hint(v.native(), (*C.gchar)(C.CString(actionName)))
	if c == nil {
		return nil
	}
	return newVariant((*C.GVariant)(c))
}

// ChangeActionState is a wrapper around g_action_group_change_action_state
func (v *ActionGroup) ChangeActionState(actionName string, value *Variant) {
	C.g_action_group_change_action_state(v.native(), (*C.gchar)(C.CString(actionName)), value.native())
}

// Activate is a wrapper around g_action_group_activate_action
func (v *ActionGroup) Activate(actionName string, parameter *Variant) {
	C.g_action_group_activate_action(v.native(), (*C.gchar)(C.CString(actionName)), parameter.native())
}
