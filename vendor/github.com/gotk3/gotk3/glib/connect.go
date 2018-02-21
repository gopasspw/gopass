package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"
import (
	"errors"
	"reflect"
	"unsafe"
)

/*
 * Events
 */

type SignalHandle uint

func (v *Object) connectClosure(after bool, detailedSignal string, f interface{}, userData ...interface{}) (SignalHandle, error) {
	if len(userData) > 1 {
		return 0, errors.New("userData len must be 0 or 1")
	}

	cstr := C.CString(detailedSignal)
	defer C.free(unsafe.Pointer(cstr))

	closure, err := ClosureNew(f, userData...)
	if err != nil {
		return 0, err
	}

	C._g_closure_add_finalize_notifier(closure)

	c := C.g_signal_connect_closure(C.gpointer(v.native()),
		(*C.gchar)(cstr), closure, gbool(after))
	handle := SignalHandle(c)

	// Map the signal handle to the closure.
	signals[handle] = closure

	return handle, nil
}

// Connect is a wrapper around g_signal_connect_closure().  f must be
// a function with a signaure matching the callback signature for
// detailedSignal.  userData must either 0 or 1 elements which can
// be optionally passed to f.  If f takes less arguments than it is
// passed from the GLib runtime, the extra arguments are ignored.
//
// Arguments for f must be a matching Go equivalent type for the
// C callback, or an interface type which the value may be packed in.
// If the type is not suitable, a runtime panic will occur when the
// signal is emitted.
func (v *Object) Connect(detailedSignal string, f interface{}, userData ...interface{}) (SignalHandle, error) {
	return v.connectClosure(false, detailedSignal, f, userData...)
}

// ConnectAfter is a wrapper around g_signal_connect_closure().  f must be
// a function with a signaure matching the callback signature for
// detailedSignal.  userData must either 0 or 1 elements which can
// be optionally passed to f.  If f takes less arguments than it is
// passed from the GLib runtime, the extra arguments are ignored.
//
// Arguments for f must be a matching Go equivalent type for the
// C callback, or an interface type which the value may be packed in.
// If the type is not suitable, a runtime panic will occur when the
// signal is emitted.
//
// The difference between Connect and ConnectAfter is that the latter
// will be invoked after the default handler, not before.
func (v *Object) ConnectAfter(detailedSignal string, f interface{}, userData ...interface{}) (SignalHandle, error) {
	return v.connectClosure(true, detailedSignal, f, userData...)
}

// ClosureNew creates a new GClosure and adds its callback function
// to the internally-maintained map. It's exported for visibility to other
// gotk3 packages and shouldn't be used in application code.
func ClosureNew(f interface{}, marshalData ...interface{}) (*C.GClosure, error) {
	// Create a reflect.Value from f.  This is called when the
	// returned GClosure runs.
	rf := reflect.ValueOf(f)

	// Create closure context which points to the reflected func.
	cc := closureContext{rf: rf}

	// Closures can only be created from funcs.
	if rf.Type().Kind() != reflect.Func {
		return nil, errors.New("value is not a func")
	}

	if len(marshalData) > 0 {
		cc.userData = reflect.ValueOf(marshalData[0])
	}

	c := C._g_closure_new()

	// Associate the GClosure with rf.  rf will be looked up in this
	// map by the closure when the closure runs.
	closures.Lock()
	closures.m[c] = cc
	closures.Unlock()

	return c, nil
}

// removeClosure removes a closure from the internal closures map.  This is
// needed to prevent a leak where Go code can access the closure context
// (along with rf and userdata) even after an object has been destroyed and
// the GClosure is invalidated and will never run.
//
//export removeClosure
func removeClosure(_ C.gpointer, closure *C.GClosure) {
	closures.Lock()
	delete(closures.m, closure)
	closures.Unlock()
}
