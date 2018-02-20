// Same copyright and license as the rest of the files in this project
// This file contains style related functions and structures

package gtk

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

/*
 * GtkApplicationWindow
 */

// ApplicationWindow is a representation of GTK's GtkApplicationWindow.
type ApplicationWindow struct {
	Window

	// Interfaces
	glib.ActionMap
	glib.ActionGroup
}

// native returns a pointer to the underlying GtkApplicationWindow.
func (v *ApplicationWindow) native() *C.GtkApplicationWindow {
	if v == nil || v.Window.GObject == nil { // v.Window is necessary because v.GObject would be ambiguous
		return nil
	}
	p := unsafe.Pointer(v.Window.GObject)
	return C.toGtkApplicationWindow(p)
}

func marshalApplicationWindow(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapApplicationWindow(obj), nil
}

func wrapApplicationWindow(obj *glib.Object) *ApplicationWindow {
	am := glib.ActionMap{obj}
	ag := glib.ActionGroup{obj}
	return &ApplicationWindow{Window{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}}, am, ag}
}

// ApplicationWindowNew is a wrapper around gtk_application_window_new().
func ApplicationWindowNew(app *Application) (*ApplicationWindow, error) {
	c := C.gtk_application_window_new(app.native())
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapApplicationWindow(glib.Take(unsafe.Pointer(c))), nil
}

// SetShowMenubar is a wrapper around gtk_application_window_set_show_menubar().
func (v *ApplicationWindow) SetShowMenubar(b bool) {
	C.gtk_application_window_set_show_menubar(v.native(), gbool(b))
}

// GetShowMenubar is a wrapper around gtk_application_window_get_show_menubar().
func (v *ApplicationWindow) GetShowMenubar() bool {
	return gobool(C.gtk_application_window_get_show_menubar(v.native()))
}

// GetID is a wrapper around gtk_application_window_get_id().
func (v *ApplicationWindow) GetID() uint {
	return uint(C.gtk_application_window_get_id(v.native()))
}
