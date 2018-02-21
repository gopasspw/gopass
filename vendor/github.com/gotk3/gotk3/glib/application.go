package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <gio/gio.h>
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"
import "unsafe"

// Application is a representation of GApplication.
type Application struct {
	*Object

	// Interfaces
	ActionMap
	ActionGroup
}

// native() returns a pointer to the underlying GApplication.
func (v *Application) native() *C.GApplication {
	if v == nil || v.GObject == nil {
		return nil
	}
	return C.toGApplication(unsafe.Pointer(v.GObject))
}

func (v *Application) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalApplication(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return wrapApplication(wrapObject(unsafe.Pointer(c))), nil
}

func wrapApplication(obj *Object) *Application {
	am := wrapActionMap(obj)
	ag := wrapActionGroup(obj)
	return &Application{obj, *am, *ag}
}

// ApplicationIDIsValid is a wrapper around g_application_id_is_valid().
func ApplicationIDIsValid(id string) bool {
	cstr1 := (*C.gchar)(C.CString(id))
	defer C.free(unsafe.Pointer(cstr1))

	return gobool(C.g_application_id_is_valid(cstr1))
}

// ApplicationNew is a wrapper around g_application_new().
func ApplicationNew(appID string, flags ApplicationFlags) *Application {
	cstr1 := (*C.gchar)(C.CString(appID))
	defer C.free(unsafe.Pointer(cstr1))

	c := C.g_application_new(cstr1, C.GApplicationFlags(flags))
	if c == nil {
		return nil
	}
	return wrapApplication(wrapObject(unsafe.Pointer(c)))
}

// GetApplicationID is a wrapper around g_application_get_application_id().
func (v *Application) GetApplicationID() string {
	c := C.g_application_get_application_id(v.native())

	return C.GoString((*C.char)(c))
}

// SetApplicationID is a wrapper around g_application_set_application_id().
func (v *Application) SetApplicationID(id string) {
	cstr1 := (*C.gchar)(C.CString(id))
	defer C.free(unsafe.Pointer(cstr1))

	C.g_application_set_application_id(v.native(), cstr1)
}

// GetInactivityTimeout is a wrapper around g_application_get_inactivity_timeout().
func (v *Application) GetInactivityTimeout() uint {
	return uint(C.g_application_get_inactivity_timeout(v.native()))
}

// SetInactivityTimeout is a wrapper around g_application_set_inactivity_timeout().
func (v *Application) SetInactivityTimeout(timeout uint) {
	C.g_application_set_inactivity_timeout(v.native(), C.guint(timeout))
}

// GetFlags is a wrapper around g_application_get_flags().
func (v *Application) GetFlags() ApplicationFlags {
	return ApplicationFlags(C.g_application_get_flags(v.native()))
}

// SetFlags is a wrapper around g_application_set_flags().
func (v *Application) SetFlags(flags ApplicationFlags) {
	C.g_application_set_flags(v.native(), C.GApplicationFlags(flags))
}

// Only available in GLib 2.42+
// // GetResourceBasePath is a wrapper around g_application_get_resource_base_path().
// func (v *Application) GetResourceBasePath() string {
// 	c := C.g_application_get_resource_base_path(v.native())

// 	return C.GoString((*C.char)(c))
// }

// Only available in GLib 2.42+
// // SetResourceBasePath is a wrapper around g_application_set_resource_base_path().
// func (v *Application) SetResourceBasePath(bp string) {
// 	cstr1 := (*C.gchar)(C.CString(bp))
// 	defer C.free(unsafe.Pointer(cstr1))

// 	C.g_application_set_resource_base_path(v.native(), cstr1)
// }

// GetDbusObjectPath is a wrapper around g_application_get_dbus_object_path().
func (v *Application) GetDbusObjectPath() string {
	c := C.g_application_get_dbus_object_path(v.native())

	return C.GoString((*C.char)(c))
}

// GetIsRegistered is a wrapper around g_application_get_is_registered().
func (v *Application) GetIsRegistered() bool {
	return gobool(C.g_application_get_is_registered(v.native()))
}

// GetIsRemote is a wrapper around g_application_get_is_remote().
func (v *Application) GetIsRemote() bool {
	return gobool(C.g_application_get_is_remote(v.native()))
}

// Hold is a wrapper around g_application_hold().
func (v *Application) Hold() {
	C.g_application_hold(v.native())
}

// Release is a wrapper around g_application_release().
func (v *Application) Release() {
	C.g_application_release(v.native())
}

// Quit is a wrapper around g_application_quit().
func (v *Application) Quit() {
	C.g_application_quit(v.native())
}

// Activate is a wrapper around g_application_activate().
func (v *Application) Activate() {
	C.g_application_activate(v.native())
}

// SendNotification is a wrapper around g_application_send_notification().
func (v *Application) SendNotification(id string, notification *Notification) {
	cstr1 := (*C.gchar)(C.CString(id))
	defer C.free(unsafe.Pointer(cstr1))

	C.g_application_send_notification(v.native(), cstr1, notification.native())
}

// WithdrawNotification is a wrapper around g_application_withdraw_notification().
func (v *Application) WithdrawNotification(id string) {
	cstr1 := (*C.gchar)(C.CString(id))
	defer C.free(unsafe.Pointer(cstr1))

	C.g_application_withdraw_notification(v.native(), cstr1)
}

// SetDefault is a wrapper around g_application_set_default().
func (v *Application) SetDefault() {
	C.g_application_set_default(v.native())
}

// ApplicationGetDefault is a wrapper around g_application_get_default().
func ApplicationGetDefault() *Application {
	c := C.g_application_get_default()
	if c == nil {
		return nil
	}
	return wrapApplication(wrapObject(unsafe.Pointer(c)))
}

// MarkBusy is a wrapper around g_application_mark_busy().
func (v *Application) MarkBusy() {
	C.g_application_mark_busy(v.native())
}

// UnmarkBusy is a wrapper around g_application_unmark_busy().
func (v *Application) UnmarkBusy() {
	C.g_application_unmark_busy(v.native())
}

// Run is a wrapper around g_application_run().
func (v *Application) Run(args []string) int {
	cargs := C.make_strings(C.int(len(args)))
	defer C.destroy_strings(cargs)

	for i, arg := range args {
		cstr := C.CString(arg)
		defer C.free(unsafe.Pointer(cstr))
		C.set_string(cargs, C.int(i), (*C.char)(cstr))
	}

	return int(C.g_application_run(v.native(), C.int(len(args)), cargs))
}

// Only available in GLib 2.44+
// // GetIsBusy is a wrapper around g_application_get_is_busy().
// func (v *Application) GetIsBusy() bool {
// 	return gobool(C.g_application_get_is_busy(v.native()))
// }

// void 	g_application_bind_busy_property ()
// void 	g_application_unbind_busy_property ()
// gboolean 	g_application_register () // requires GCancellable
// void 	g_application_set_action_group () // Deprecated since 2.32
// GDBusConnection * 	g_application_get_dbus_connection () // No support for GDBusConnection
// void 	g_application_open () // Needs GFile
// void 	g_application_add_main_option_entries () //Needs GOptionEntry
// void 	g_application_add_main_option () //Needs GOptionFlags and GOptionArg
// void 	g_application_add_option_group () // Needs GOptionGroup
