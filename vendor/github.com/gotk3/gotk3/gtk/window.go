// Same copyright and license as the rest of the files in this project
// This file contains accelerator related functions and structures

package gtk

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"errors"
	"unsafe"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
)

/*
 * GtkWindow
 */

// Window is a representation of GTK's GtkWindow.
type Window struct {
	Bin
}

// IWindow is an interface type implemented by all structs embedding a
// Window.  It is meant to be used as an argument type for wrapper
// functions that wrap around a C GTK function taking a GtkWindow.
type IWindow interface {
	toWindow() *C.GtkWindow
}

// native returns a pointer to the underlying GtkWindow.
func (v *Window) native() *C.GtkWindow {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkWindow(p)
}

func (v *Window) toWindow() *C.GtkWindow {
	if v == nil {
		return nil
	}
	return v.native()
}

func marshalWindow(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapWindow(obj), nil
}

func wrapWindow(obj *glib.Object) *Window {
	return &Window{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}}
}

// WindowNew is a wrapper around gtk_window_new().
func WindowNew(t WindowType) (*Window, error) {
	c := C.gtk_window_new(C.GtkWindowType(t))
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapWindow(glib.Take(unsafe.Pointer(c))), nil
}

// SetTitle is a wrapper around gtk_window_set_title().
func (v *Window) SetTitle(title string) {
	cstr := C.CString(title)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_window_set_title(v.native(), (*C.gchar)(cstr))
}

// SetResizable is a wrapper around gtk_window_set_resizable().
func (v *Window) SetResizable(resizable bool) {
	C.gtk_window_set_resizable(v.native(), gbool(resizable))
}

// GetResizable is a wrapper around gtk_window_get_resizable().
func (v *Window) GetResizable() bool {
	c := C.gtk_window_get_resizable(v.native())
	return gobool(c)
}

// ActivateFocus is a wrapper around gtk_window_activate_focus().
func (v *Window) ActivateFocus() bool {
	c := C.gtk_window_activate_focus(v.native())
	return gobool(c)
}

// ActivateDefault is a wrapper around gtk_window_activate_default().
func (v *Window) ActivateDefault() bool {
	c := C.gtk_window_activate_default(v.native())
	return gobool(c)
}

// SetModal is a wrapper around gtk_window_set_modal().
func (v *Window) SetModal(modal bool) {
	C.gtk_window_set_modal(v.native(), gbool(modal))
}

// SetDefaultSize is a wrapper around gtk_window_set_default_size().
func (v *Window) SetDefaultSize(width, height int) {
	C.gtk_window_set_default_size(v.native(), C.gint(width), C.gint(height))
}

// GetScreen is a wrapper around gtk_window_get_screen().
func (v *Window) GetScreen() (*gdk.Screen, error) {
	c := C.gtk_window_get_screen(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	s := &gdk.Screen{glib.Take(unsafe.Pointer(c))}
	return s, nil
}

// SetIcon is a wrapper around gtk_window_set_icon().
func (v *Window) SetIcon(icon *gdk.Pixbuf) {
	iconPtr := (*C.GdkPixbuf)(unsafe.Pointer(icon.Native()))
	C.gtk_window_set_icon(v.native(), iconPtr)
}

// WindowSetDefaultIcon is a wrapper around gtk_window_set_default_icon().
func WindowSetDefaultIcon(icon *gdk.Pixbuf) {
	iconPtr := (*C.GdkPixbuf)(unsafe.Pointer(icon.Native()))
	C.gtk_window_set_default_icon(iconPtr)
}

// TODO(jrick) GdkGeometry GdkWindowHints.
/*
func (v *Window) SetGeometryHints() {
}
*/

// SetGravity is a wrapper around gtk_window_set_gravity().
func (v *Window) SetGravity(gravity gdk.GdkGravity) {
	C.gtk_window_set_gravity(v.native(), C.GdkGravity(gravity))
}

// TODO(jrick) GdkGravity.
/*
func (v *Window) GetGravity() {
}
*/

// SetPosition is a wrapper around gtk_window_set_position().
func (v *Window) SetPosition(position WindowPosition) {
	C.gtk_window_set_position(v.native(), C.GtkWindowPosition(position))
}

// SetTransientFor is a wrapper around gtk_window_set_transient_for().
func (v *Window) SetTransientFor(parent IWindow) {
	var pw *C.GtkWindow = nil
	if parent != nil {
		pw = parent.toWindow()
	}
	C.gtk_window_set_transient_for(v.native(), pw)
}

// SetDestroyWithParent is a wrapper around
// gtk_window_set_destroy_with_parent().
func (v *Window) SetDestroyWithParent(setting bool) {
	C.gtk_window_set_destroy_with_parent(v.native(), gbool(setting))
}

// SetHideTitlebarWhenMaximized is a wrapper around
// gtk_window_set_hide_titlebar_when_maximized().
func (v *Window) SetHideTitlebarWhenMaximized(setting bool) {
	C.gtk_window_set_hide_titlebar_when_maximized(v.native(),
		gbool(setting))
}

// IsActive is a wrapper around gtk_window_is_active().
func (v *Window) IsActive() bool {
	c := C.gtk_window_is_active(v.native())
	return gobool(c)
}

// HasToplevelFocus is a wrapper around gtk_window_has_toplevel_focus().
func (v *Window) HasToplevelFocus() bool {
	c := C.gtk_window_has_toplevel_focus(v.native())
	return gobool(c)
}

// GetFocus is a wrapper around gtk_window_get_focus().
func (v *Window) GetFocus() (*Widget, error) {
	c := C.gtk_window_get_focus(v.native())
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapWidget(glib.Take(unsafe.Pointer(c))), nil
}

// SetFocus is a wrapper around gtk_window_set_focus().
func (v *Window) SetFocus(w *Widget) {
	C.gtk_window_set_focus(v.native(), w.native())
}

// GetDefaultWidget is a wrapper arround gtk_window_get_default_widget().
func (v *Window) GetDefaultWidget() *Widget {
	c := C.gtk_window_get_default_widget(v.native())
	if c == nil {
		return nil
	}
	obj := glib.Take(unsafe.Pointer(c))
	return wrapWidget(obj)
}

// SetDefault is a wrapper arround gtk_window_set_default().
func (v *Window) SetDefault(widget IWidget) {
	C.gtk_window_set_default(v.native(), widget.toWidget())
}

// Present is a wrapper around gtk_window_present().
func (v *Window) Present() {
	C.gtk_window_present(v.native())
}

// PresentWithTime is a wrapper around gtk_window_present_with_time().
func (v *Window) PresentWithTime(ts uint32) {
	C.gtk_window_present_with_time(v.native(), C.guint32(ts))
}

// Iconify is a wrapper around gtk_window_iconify().
func (v *Window) Iconify() {
	C.gtk_window_iconify(v.native())
}

// Deiconify is a wrapper around gtk_window_deiconify().
func (v *Window) Deiconify() {
	C.gtk_window_deiconify(v.native())
}

// Stick is a wrapper around gtk_window_stick().
func (v *Window) Stick() {
	C.gtk_window_stick(v.native())
}

// Unstick is a wrapper around gtk_window_unstick().
func (v *Window) Unstick() {
	C.gtk_window_unstick(v.native())
}

// Maximize is a wrapper around gtk_window_maximize().
func (v *Window) Maximize() {
	C.gtk_window_maximize(v.native())
}

// Unmaximize is a wrapper around gtk_window_unmaximize().
func (v *Window) Unmaximize() {
	C.gtk_window_unmaximize(v.native())
}

// Fullscreen is a wrapper around gtk_window_fullscreen().
func (v *Window) Fullscreen() {
	C.gtk_window_fullscreen(v.native())
}

// Unfullscreen is a wrapper around gtk_window_unfullscreen().
func (v *Window) Unfullscreen() {
	C.gtk_window_unfullscreen(v.native())
}

// SetKeepAbove is a wrapper around gtk_window_set_keep_above().
func (v *Window) SetKeepAbove(setting bool) {
	C.gtk_window_set_keep_above(v.native(), gbool(setting))
}

// SetKeepBelow is a wrapper around gtk_window_set_keep_below().
func (v *Window) SetKeepBelow(setting bool) {
	C.gtk_window_set_keep_below(v.native(), gbool(setting))
}

// SetDecorated is a wrapper around gtk_window_set_decorated().
func (v *Window) SetDecorated(setting bool) {
	C.gtk_window_set_decorated(v.native(), gbool(setting))
}

// SetDeletable is a wrapper around gtk_window_set_deletable().
func (v *Window) SetDeletable(setting bool) {
	C.gtk_window_set_deletable(v.native(), gbool(setting))
}

// SetTypeHint is a wrapper around gtk_window_set_type_hint().
func (v *Window) SetTypeHint(typeHint gdk.WindowTypeHint) {
	C.gtk_window_set_type_hint(v.native(), C.GdkWindowTypeHint(typeHint))
}

// SetSkipTaskbarHint is a wrapper around gtk_window_set_skip_taskbar_hint().
func (v *Window) SetSkipTaskbarHint(setting bool) {
	C.gtk_window_set_skip_taskbar_hint(v.native(), gbool(setting))
}

// SetSkipPagerHint is a wrapper around gtk_window_set_skip_pager_hint().
func (v *Window) SetSkipPagerHint(setting bool) {
	C.gtk_window_set_skip_pager_hint(v.native(), gbool(setting))
}

// SetUrgencyHint is a wrapper around gtk_window_set_urgency_hint().
func (v *Window) SetUrgencyHint(setting bool) {
	C.gtk_window_set_urgency_hint(v.native(), gbool(setting))
}

// SetAcceptFocus is a wrapper around gtk_window_set_accept_focus().
func (v *Window) SetAcceptFocus(setting bool) {
	C.gtk_window_set_accept_focus(v.native(), gbool(setting))
}

// SetFocusOnMap is a wrapper around gtk_window_set_focus_on_map().
func (v *Window) SetFocusOnMap(setting bool) {
	C.gtk_window_set_focus_on_map(v.native(), gbool(setting))
}

// SetStartupID is a wrapper around gtk_window_set_startup_id().
func (v *Window) SetStartupID(sid string) {
	cstr := (*C.gchar)(C.CString(sid))
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_window_set_startup_id(v.native(), cstr)
}

// SetRole is a wrapper around gtk_window_set_role().
func (v *Window) SetRole(s string) {
	cstr := (*C.gchar)(C.CString(s))
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_window_set_role(v.native(), cstr)
}

// GetDecorated is a wrapper around gtk_window_get_decorated().
func (v *Window) GetDecorated() bool {
	c := C.gtk_window_get_decorated(v.native())
	return gobool(c)
}

// GetDeletable is a wrapper around gtk_window_get_deletable().
func (v *Window) GetDeletable() bool {
	c := C.gtk_window_get_deletable(v.native())
	return gobool(c)
}

// WindowGetDefaultIconName is a wrapper around gtk_window_get_default_icon_name().
func WindowGetDefaultIconName() (string, error) {
	return stringReturn(C.gtk_window_get_default_icon_name())
}

// GetDefaultSize is a wrapper around gtk_window_get_default_size().
func (v *Window) GetDefaultSize() (width, height int) {
	var w, h C.gint
	C.gtk_window_get_default_size(v.native(), &w, &h)
	return int(w), int(h)
}

// GetDestroyWithParent is a wrapper around
// gtk_window_get_destroy_with_parent().
func (v *Window) GetDestroyWithParent() bool {
	c := C.gtk_window_get_destroy_with_parent(v.native())
	return gobool(c)
}

// GetHideTitlebarWhenMaximized is a wrapper around
// gtk_window_get_hide_titlebar_when_maximized().
func (v *Window) GetHideTitlebarWhenMaximized() bool {
	c := C.gtk_window_get_hide_titlebar_when_maximized(v.native())
	return gobool(c)
}

// GetIcon is a wrapper around gtk_window_get_icon().
func (v *Window) GetIcon() (*gdk.Pixbuf, error) {
	c := C.gtk_window_get_icon(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	p := &gdk.Pixbuf{glib.Take(unsafe.Pointer(c))}
	return p, nil
}

// GetIconName is a wrapper around gtk_window_get_icon_name().
func (v *Window) GetIconName() (string, error) {
	return stringReturn(C.gtk_window_get_icon_name(v.native()))
}

// GetModal is a wrapper around gtk_window_get_modal().
func (v *Window) GetModal() bool {
	c := C.gtk_window_get_modal(v.native())
	return gobool(c)
}

// GetPosition is a wrapper around gtk_window_get_position().
func (v *Window) GetPosition() (root_x, root_y int) {
	var x, y C.gint
	C.gtk_window_get_position(v.native(), &x, &y)
	return int(x), int(y)
}

func stringReturn(c *C.gchar) (string, error) {
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

// GetRole is a wrapper around gtk_window_get_role().
func (v *Window) GetRole() (string, error) {
	return stringReturn(C.gtk_window_get_role(v.native()))
}

// GetSize is a wrapper around gtk_window_get_size().
func (v *Window) GetSize() (width, height int) {
	var w, h C.gint
	C.gtk_window_get_size(v.native(), &w, &h)
	return int(w), int(h)
}

// GetTitle is a wrapper around gtk_window_get_title().
func (v *Window) GetTitle() (string, error) {
	return stringReturn(C.gtk_window_get_title(v.native()))
}

// GetTransientFor is a wrapper around gtk_window_get_transient_for().
func (v *Window) GetTransientFor() (*Window, error) {
	c := C.gtk_window_get_transient_for(v.native())
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapWindow(glib.Take(unsafe.Pointer(c))), nil
}

// GetAttachedTo is a wrapper around gtk_window_get_attached_to().
func (v *Window) GetAttachedTo() (*Widget, error) {
	c := C.gtk_window_get_attached_to(v.native())
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapWidget(glib.Take(unsafe.Pointer(c))), nil
}

// GetTypeHint is a wrapper around gtk_window_get_type_hint().
func (v *Window) GetTypeHint() gdk.WindowTypeHint {
	c := C.gtk_window_get_type_hint(v.native())
	return gdk.WindowTypeHint(c)
}

// GetSkipTaskbarHint is a wrapper around gtk_window_get_skip_taskbar_hint().
func (v *Window) GetSkipTaskbarHint() bool {
	c := C.gtk_window_get_skip_taskbar_hint(v.native())
	return gobool(c)
}

// GetSkipPagerHint is a wrapper around gtk_window_get_skip_pager_hint().
func (v *Window) GetSkipPagerHint() bool {
	c := C.gtk_window_get_skip_taskbar_hint(v.native())
	return gobool(c)
}

// GetUrgencyHint is a wrapper around gtk_window_get_urgency_hint().
func (v *Window) GetUrgencyHint() bool {
	c := C.gtk_window_get_urgency_hint(v.native())
	return gobool(c)
}

// GetAcceptFocus is a wrapper around gtk_window_get_accept_focus().
func (v *Window) GetAcceptFocus() bool {
	c := C.gtk_window_get_accept_focus(v.native())
	return gobool(c)
}

// GetFocusOnMap is a wrapper around gtk_window_get_focus_on_map().
func (v *Window) GetFocusOnMap() bool {
	c := C.gtk_window_get_focus_on_map(v.native())
	return gobool(c)
}

// HasGroup is a wrapper around gtk_window_has_group().
func (v *Window) HasGroup() bool {
	c := C.gtk_window_has_group(v.native())
	return gobool(c)
}

// Move is a wrapper around gtk_window_move().
func (v *Window) Move(x, y int) {
	C.gtk_window_move(v.native(), C.gint(x), C.gint(y))
}

// Resize is a wrapper around gtk_window_resize().
func (v *Window) Resize(width, height int) {
	C.gtk_window_resize(v.native(), C.gint(width), C.gint(height))
}

// WindowSetDefaultIconFromFile is a wrapper around gtk_window_set_default_icon_from_file().
func WindowSetDefaultIconFromFile(file string) error {
	cstr := C.CString(file)
	defer C.free(unsafe.Pointer(cstr))
	var err *C.GError = nil
	res := C.gtk_window_set_default_icon_from_file((*C.gchar)(cstr), &err)
	if res == 0 {
		defer C.g_error_free(err)
		return errors.New(C.GoString((*C.char)(err.message)))
	}
	return nil
}

// WindowSetDefaultIconName is a wrapper around gtk_window_set_default_icon_name().
func WindowSetDefaultIconName(s string) {
	cstr := (*C.gchar)(C.CString(s))
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_window_set_default_icon_name(cstr)
}

// SetIconFromFile is a wrapper around gtk_window_set_icon_from_file().
func (v *Window) SetIconFromFile(file string) error {
	cstr := C.CString(file)
	defer C.free(unsafe.Pointer(cstr))
	var err *C.GError = nil
	res := C.gtk_window_set_icon_from_file(v.native(), (*C.gchar)(cstr), &err)
	if res == 0 {
		defer C.g_error_free(err)
		return errors.New(C.GoString((*C.char)(err.message)))
	}
	return nil
}

// SetIconName is a wrapper around gtk_window_set_icon_name().
func (v *Window) SetIconName(name string) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_window_set_icon_name(v.native(), (*C.gchar)(cstr))
}

// SetAutoStartupNotification is a wrapper around
// gtk_window_set_auto_startup_notification().
// This doesn't seem write.  Might need to rethink?
/*
func (v *Window) SetAutoStartupNotification(setting bool) {
	C.gtk_window_set_auto_startup_notification(gbool(setting))
}
*/

// GetMnemonicsVisible is a wrapper around
// gtk_window_get_mnemonics_visible().
func (v *Window) GetMnemonicsVisible() bool {
	c := C.gtk_window_get_mnemonics_visible(v.native())
	return gobool(c)
}

// SetMnemonicsVisible is a wrapper around
// gtk_window_get_mnemonics_visible().
func (v *Window) SetMnemonicsVisible(setting bool) {
	C.gtk_window_set_mnemonics_visible(v.native(), gbool(setting))
}

// GetFocusVisible is a wrapper around gtk_window_get_focus_visible().
func (v *Window) GetFocusVisible() bool {
	c := C.gtk_window_get_focus_visible(v.native())
	return gobool(c)
}

// SetFocusVisible is a wrapper around gtk_window_set_focus_visible().
func (v *Window) SetFocusVisible(setting bool) {
	C.gtk_window_set_focus_visible(v.native(), gbool(setting))
}

// GetApplication is a wrapper around gtk_window_get_application().
func (v *Window) GetApplication() (*Application, error) {
	c := C.gtk_window_get_application(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	return wrapApplication(glib.Take(unsafe.Pointer(c))), nil
}

// SetApplication is a wrapper around gtk_window_set_application().
func (v *Window) SetApplication(a *Application) {
	C.gtk_window_set_application(v.native(), a.native())
}

// ActivateKey is a wrapper around gtk_window_activate_key().
func (v *Window) ActivateKey(event *gdk.EventKey) bool {
	c := C.gtk_window_activate_key(v.native(), (*C.GdkEventKey)(unsafe.Pointer(event.Native())))
	return gobool(c)
}

// AddMnemonic is a wrapper around gtk_window_add_mnemonic().
func (v *Window) AddMnemonic(keyval uint, target *Widget) {
	C.gtk_window_add_mnemonic(v.native(), C.guint(keyval), target.native())
}

// RemoveMnemonic is a wrapper around gtk_window_remove_mnemonic().
func (v *Window) RemoveMnemonic(keyval uint, target *Widget) {
	C.gtk_window_remove_mnemonic(v.native(), C.guint(keyval), target.native())
}

// ActivateMnemonic is a wrapper around gtk_window_mnemonic_activate().
func (v *Window) ActivateMnemonic(keyval uint, mods gdk.ModifierType) bool {
	c := C.gtk_window_mnemonic_activate(v.native(), C.guint(keyval), C.GdkModifierType(mods))
	return gobool(c)
}

// GetMnemonicModifier is a wrapper around gtk_window_get_mnemonic_modifier().
func (v *Window) GetMnemonicModifier() gdk.ModifierType {
	c := C.gtk_window_get_mnemonic_modifier(v.native())
	return gdk.ModifierType(c)
}

// SetMnemonicModifier is a wrapper around gtk_window_set_mnemonic_modifier().
func (v *Window) SetMnemonicModifier(mods gdk.ModifierType) {
	C.gtk_window_set_mnemonic_modifier(v.native(), C.GdkModifierType(mods))
}

// TODO gtk_window_begin_move_drag().
// TODO gtk_window_begin_resize_drag().
// TODO gtk_window_get_default_icon_list().
// TODO gtk_window_get_group().
// TODO gtk_window_get_icon_list().
// TODO gtk_window_get_window_type().
// TODO gtk_window_list_toplevels().
// TODO gtk_window_parse_geometry().
// TODO gtk_window_propogate_key_event().
// TODO gtk_window_set_attached_to().
// TODO gtk_window_set_default_icon_list().
// TODO gtk_window_set_icon_list().
// TODO gtk_window_set_screen().
// TODO gtk_window_get_resize_grip_area().
