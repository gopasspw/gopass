package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <gio/gio.h>
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"
import "unsafe"

type BindingFlags int

const (
	BINDING_DEFAULT        BindingFlags = C.G_BINDING_DEFAULT
	BINDING_BIDIRECTIONAL  BindingFlags = C.G_BINDING_BIDIRECTIONAL
	BINDING_SYNC_CREATE                 = C.G_BINDING_SYNC_CREATE
	BINDING_INVERT_BOOLEAN              = C.G_BINDING_INVERT_BOOLEAN
)

type Binding struct {
	*Object
}

func (v *Binding) native() *C.GBinding {
	if v == nil || v.GObject == nil {
		return nil
	}
	return C.toGBinding(unsafe.Pointer(v.GObject))
}

func marshalBinding(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return &Binding{wrapObject(unsafe.Pointer(c))}, nil
}

// Creates a binding between source property on source and target property on
// target . Whenever the source property is changed the target_property is
// updated using the same value.
func BindProperty(source *Object, sourceProperty string,
	target *Object, targetProperty string,
	flags BindingFlags) *Binding {
	srcStr := (*C.gchar)(C.CString(sourceProperty))
	defer C.free(unsafe.Pointer(srcStr))
	tgtStr := (*C.gchar)(C.CString(targetProperty))
	defer C.free(unsafe.Pointer(tgtStr))
	obj := C.g_object_bind_property(
		C.gpointer(source.GObject), srcStr,
		C.gpointer(target.GObject), tgtStr,
		C.GBindingFlags(flags),
	)
	if obj == nil {
		return nil
	}
	return &Binding{wrapObject(unsafe.Pointer(obj))}
}

// Explicitly releases the binding between the source and the target property
// expressed by Binding
func (v *Binding) Unbind() {
	C.g_binding_unbind(v.native())
}

// Retrieves the GObject instance used as the source of the binding
func (v *Binding) GetSource() *Object {
	obj := C.g_binding_get_source(v.native())
	if obj == nil {
		return nil
	}
	return wrapObject(unsafe.Pointer(obj))
}

// Retrieves the name of the property of “source” used as the source of
// the binding.
func (v *Binding) GetSourceProperty() string {
	s := C.g_binding_get_source_property(v.native())
	return C.GoString((*C.char)(s))
}

// Retrieves the GObject instance used as the target of the binding.
func (v *Binding) GetTarget() *Object {
	obj := C.g_binding_get_target(v.native())
	if obj == nil {
		return nil
	}
	return wrapObject(unsafe.Pointer(obj))
}

// Retrieves the name of the property of “target” used as the target of
// the binding.
func (v *Binding) GetTargetProperty() string {
	s := C.g_binding_get_target_property(v.native())
	return C.GoString((*C.char)(s))
}

// Retrieves the flags passed when constructing the GBinding.
func (v *Binding) GetFlags() BindingFlags {
	flags := C.g_binding_get_flags(v.native())
	return BindingFlags(flags)
}
