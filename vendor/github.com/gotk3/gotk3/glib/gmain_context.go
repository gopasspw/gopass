package glib

// #cgo pkg-config: glib-2.0 gobject-2.0 gio-2.0
// #include <gio/gio.h>
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"

type MainContext C.GMainContext

// native returns a pointer to the underlying GMainContext.
func (v *MainContext) native() *C.GMainContext {
	if v == nil {
		return nil
	}
	return (*C.GMainContext)(v)
}

// MainContextDefault is a wrapper around g_main_context_default().
func MainContextDefault() *MainContext {
	c := C.g_main_context_default()
	if c == nil {
		return nil
	}
	return (*MainContext)(c)
}

// MainDepth is a wrapper around g_main_depth().
func MainDepth() int {
	return int(C.g_main_depth())
}
