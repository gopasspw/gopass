package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <gio/gio.h>
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"
import "unsafe"

// Action is a representation of glib's GAction GInterface.
type Action struct {
	*Object
}

// IAction is an interface type implemented by all structs
// embedding an Action.  It is meant to be used as an argument type
// for wrapper functions that wrap around a C function taking a
// GAction.
type IAction interface {
	toGAction() *C.GAction
	toAction() *Action
}

func (v *Action) toGAction() *C.GAction {
	if v == nil {
		return nil
	}
	return v.native()
}

func (v *Action) toAction() *Action {
	return v
}

// gboolean g_action_parse_detailed_name (const gchar *detailed_name, gchar **action_name, GVariant **target_value, GError **error);
// gchar * g_action_print_detailed_name (const gchar *action_name, GVariant *target_value);

// native() returns a pointer to the underlying GAction.
func (v *Action) native() *C.GAction {
	if v == nil || v.GObject == nil {
		return nil
	}
	return C.toGAction(unsafe.Pointer(v.GObject))
}

func (v *Action) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalAction(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return wrapAction(wrapObject(unsafe.Pointer(c))), nil
}

func wrapAction(obj *Object) *Action {
	return &Action{obj}
}

// ActionNameIsValid is a wrapper around g_action_name_is_valid
func ActionNameIsValid(actionName string) bool {
	cstr := (*C.gchar)(C.CString(actionName))
	return gobool(C.g_action_name_is_valid(cstr))
}

// GetName is a wrapper around g_action_get_name
func (v *Action) GetName() string {
	return C.GoString((*C.char)(C.g_action_get_name(v.native())))
}

// GetEnabled is a wrapper around g_action_get_enabled
func (v *Action) GetEnabled() bool {
	return gobool(C.g_action_get_enabled(v.native()))
}

// GetState is a wrapper around g_action_get_state
func (v *Action) GetState() *Variant {
	c := C.g_action_get_state(v.native())
	if c == nil {
		return nil
	}
	return newVariant((*C.GVariant)(c))
}

// GetStateHint is a wrapper around g_action_get_state_hint
func (v *Action) GetStateHint() *Variant {
	c := C.g_action_get_state_hint(v.native())
	if c == nil {
		return nil
	}
	return newVariant((*C.GVariant)(c))
}

// GetParameterType is a wrapper around g_action_get_parameter_type
func (v *Action) GetParameterType() *VariantType {
	c := C.g_action_get_parameter_type(v.native())
	if c == nil {
		return nil
	}
	return newVariantType((*C.GVariantType)(c))
}

// GetStateType is a wrapper around g_action_get_state_type
func (v *Action) GetStateType() *VariantType {
	c := C.g_action_get_state_type(v.native())
	if c == nil {
		return nil
	}
	return newVariantType((*C.GVariantType)(c))
}

// ChangeState is a wrapper around g_action_change_state
func (v *Action) ChangeState(value *Variant) {
	C.g_action_change_state(v.native(), value.native())
}

// Activate is a wrapper around g_action_activate
func (v *Action) Activate(parameter *Variant) {
	C.g_action_activate(v.native(), parameter.native())
}

// SimpleAction is a representation of GSimpleAction
type SimpleAction struct {
	Action
}

func (v *SimpleAction) native() *C.GSimpleAction {
	if v == nil || v.GObject == nil {
		return nil
	}
	return C.toGSimpleAction(unsafe.Pointer(v.GObject))
}

func (v *SimpleAction) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalSimpleAction(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return wrapSimpleAction(wrapObject(unsafe.Pointer(c))), nil
}

func wrapSimpleAction(obj *Object) *SimpleAction {
	return &SimpleAction{Action{obj}}
}

// SimpleActionNew is a wrapper around g_simple_action_new
func SimpleActionNew(name string, parameterType *VariantType) *SimpleAction {
	c := C.g_simple_action_new((*C.gchar)(C.CString(name)), parameterType.native())
	if c == nil {
		return nil
	}
	return wrapSimpleAction(wrapObject(unsafe.Pointer(c)))
}

// SimpleActionNewStateful is a wrapper around g_simple_action_new_stateful
func SimpleActionNewStateful(name string, parameterType *VariantType, state *Variant) *SimpleAction {
	c := C.g_simple_action_new_stateful((*C.gchar)(C.CString(name)), parameterType.native(), state.native())
	if c == nil {
		return nil
	}
	return wrapSimpleAction(wrapObject(unsafe.Pointer(c)))
}

// SetEnabled is a wrapper around g_simple_action_set_enabled
func (v *SimpleAction) SetEnabled(enabled bool) {
	C.g_simple_action_set_enabled(v.native(), gbool(enabled))
}

// SetState is a wrapper around g_simple_action_set_state
// This should only be called by the implementor of the action.
// Users of the action should not attempt to directly modify the 'state' property.
// Instead, they should call ChangeState [g_action_change_state()] to request the change.
func (v *SimpleAction) SetState(value *Variant) {
	C.g_simple_action_set_state(v.native(), value.native())
}

// SetStateHint is a wrapper around g_simple_action_set_state_hint
// GLib 2.44 only (currently no build tags, so commented out)
/*func (v *SimpleAction) SetStateHint(stateHint *Variant) {
	C.g_simple_action_set_state_hint(v.native(), stateHint.native())
}*/

// PropertyAction is a representation of GPropertyAction
type PropertyAction struct {
	Action
}

func (v *PropertyAction) native() *C.GPropertyAction {
	if v == nil || v.GObject == nil {
		return nil
	}
	return C.toGPropertyAction(unsafe.Pointer(v.GObject))
}

func (v *PropertyAction) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalPropertyAction(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return wrapPropertyAction(wrapObject(unsafe.Pointer(c))), nil
}

func wrapPropertyAction(obj *Object) *PropertyAction {
	return &PropertyAction{Action{obj}}
}

// PropertyActionNew is a wrapper around g_property_action_new
func PropertyActionNew(name string, object *Object, propertyName string) *PropertyAction {
	c := C.g_property_action_new((*C.gchar)(C.CString(name)), C.gpointer(unsafe.Pointer(object.native())), (*C.gchar)(C.CString(propertyName)))
	if c == nil {
		return nil
	}
	return wrapPropertyAction(wrapObject(unsafe.Pointer(c)))
}
