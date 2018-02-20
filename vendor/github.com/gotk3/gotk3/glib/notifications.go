package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <gio/gio.h>
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"
import "unsafe"

// Only available from 2.42
// // NotificationPriority is a representation of GLib's GNotificationPriority.
// type NotificationPriority int

// const (
// 	NOTIFICATION_PRIORITY_NORMAL NotificationPriority = C.G_NOTIFICATION_PRIORITY_NORMAL
// 	NOTIFICATION_PRIORITY_LOW    NotificationPriority = C.G_NOTIFICATION_PRIORITY_LOW
// 	NOTIFICATION_PRIORITY_HIGH   NotificationPriority = C.G_NOTIFICATION_PRIORITY_HIGH
// 	NOTIFICATION_PRIORITY_URGENT NotificationPriority = C.G_NOTIFICATION_PRIORITY_URGENT
// )

// Notification is a representation of GNotification.
type Notification struct {
	*Object
}

// native() returns a pointer to the underlying GNotification.
func (v *Notification) native() *C.GNotification {
	if v == nil || v.GObject == nil {
		return nil
	}
	return C.toGNotification(unsafe.Pointer(v.GObject))
}

func (v *Notification) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalNotification(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return wrapNotification(wrapObject(unsafe.Pointer(c))), nil
}

func wrapNotification(obj *Object) *Notification {
	return &Notification{obj}
}

// NotificationNew is a wrapper around g_notification_new().
func NotificationNew(title string) *Notification {
	cstr1 := (*C.gchar)(C.CString(title))
	defer C.free(unsafe.Pointer(cstr1))

	c := C.g_notification_new(cstr1)
	if c == nil {
		return nil
	}
	return wrapNotification(wrapObject(unsafe.Pointer(c)))
}

// SetTitle is a wrapper around g_notification_set_title().
func (v *Notification) SetTitle(title string) {
	cstr1 := (*C.gchar)(C.CString(title))
	defer C.free(unsafe.Pointer(cstr1))

	C.g_notification_set_title(v.native(), cstr1)
}

// SetBody is a wrapper around g_notification_set_body().
func (v *Notification) SetBody(body string) {
	cstr1 := (*C.gchar)(C.CString(body))
	defer C.free(unsafe.Pointer(cstr1))

	C.g_notification_set_body(v.native(), cstr1)
}

// Only available from 2.42
// // SetPriority is a wrapper around g_notification_set_priority().
// func (v *Notification) SetPriority(prio NotificationPriority) {
// 	C.g_notification_set_priority(v.native(), C.GNotificationPriority(prio))
// }

// SetDefaultAction is a wrapper around g_notification_set_default_action().
func (v *Notification) SetDefaultAction(detailedAction string) {
	cstr1 := (*C.gchar)(C.CString(detailedAction))
	defer C.free(unsafe.Pointer(cstr1))

	C.g_notification_set_default_action(v.native(), cstr1)
}

// AddButton is a wrapper around g_notification_add_button().
func (v *Notification) AddButton(label, detailedAction string) {
	cstr1 := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer(cstr1))

	cstr2 := (*C.gchar)(C.CString(detailedAction))
	defer C.free(unsafe.Pointer(cstr2))

	C.g_notification_add_button(v.native(), cstr1, cstr2)
}

// void 	g_notification_set_default_action_and_target () // requires varargs
// void 	g_notification_set_default_action_and_target_value () // requires variant
// void 	g_notification_add_button_with_target () // requires varargs
// void 	g_notification_add_button_with_target_value () //requires variant
// void 	g_notification_set_urgent () // Deprecated, so not implemented
// void 	g_notification_set_icon () // Requires support for GIcon, which we don't have yet.
