// Copyright (c) 2013-2014 Conformal Systems <info@conformal.com>
//
// This file originated from: http://opensource.conformal.com/
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

// Go bindings for GDK 3.  Supports version 3.6 and later.
package gdk

// #cgo pkg-config: gdk-3.0
// #include <gdk/gdk.h>
// #include "gdk.go.h"
import "C"
import (
	"errors"
	"reflect"
	"runtime"
	"strconv"
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

func init() {
	tm := []glib.TypeMarshaler{
		// Enums
		{glib.Type(C.gdk_drag_action_get_type()), marshalDragAction},
		{glib.Type(C.gdk_colorspace_get_type()), marshalColorspace},
		{glib.Type(C.gdk_event_type_get_type()), marshalEventType},
		{glib.Type(C.gdk_interp_type_get_type()), marshalInterpType},
		{glib.Type(C.gdk_modifier_type_get_type()), marshalModifierType},
		{glib.Type(C.gdk_pixbuf_alpha_mode_get_type()), marshalPixbufAlphaMode},
		{glib.Type(C.gdk_event_mask_get_type()), marshalEventMask},

		// Objects/Interfaces
		{glib.Type(C.gdk_device_get_type()), marshalDevice},
		{glib.Type(C.gdk_cursor_get_type()), marshalCursor},
		{glib.Type(C.gdk_device_manager_get_type()), marshalDeviceManager},
		{glib.Type(C.gdk_display_get_type()), marshalDisplay},
		{glib.Type(C.gdk_drag_context_get_type()), marshalDragContext},
		{glib.Type(C.gdk_pixbuf_get_type()), marshalPixbuf},
		{glib.Type(C.gdk_rgba_get_type()), marshalRGBA},
		{glib.Type(C.gdk_screen_get_type()), marshalScreen},
		{glib.Type(C.gdk_visual_get_type()), marshalVisual},
		{glib.Type(C.gdk_window_get_type()), marshalWindow},

		// Boxed
		{glib.Type(C.gdk_event_get_type()), marshalEvent},
	}
	glib.RegisterGValueMarshalers(tm)
}

/*
 * Type conversions
 */

func gbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}
func gobool(b C.gboolean) bool {
	if b != 0 {
		return true
	}
	return false
}

/*
 * Unexported vars
 */

var nilPtrErr = errors.New("cgo returned unexpected nil pointer")

/*
 * Constants
 */

// DragAction is a representation of GDK's GdkDragAction.
type DragAction int

const (
	ACTION_DEFAULT DragAction = C.GDK_ACTION_DEFAULT
	ACTION_COPY    DragAction = C.GDK_ACTION_COPY
	ACTION_MOVE    DragAction = C.GDK_ACTION_MOVE
	ACTION_LINK    DragAction = C.GDK_ACTION_LINK
	ACTION_PRIVATE DragAction = C.GDK_ACTION_PRIVATE
	ACTION_ASK     DragAction = C.GDK_ACTION_ASK
)

func marshalDragAction(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return DragAction(c), nil
}

// Colorspace is a representation of GDK's GdkColorspace.
type Colorspace int

const (
	COLORSPACE_RGB Colorspace = C.GDK_COLORSPACE_RGB
)

func marshalColorspace(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return Colorspace(c), nil
}

// InterpType is a representation of GDK's GdkInterpType.
type InterpType int

const (
	INTERP_NEAREST  InterpType = C.GDK_INTERP_NEAREST
	INTERP_TILES    InterpType = C.GDK_INTERP_TILES
	INTERP_BILINEAR InterpType = C.GDK_INTERP_BILINEAR
	INTERP_HYPER    InterpType = C.GDK_INTERP_HYPER
)

// PixbufRotation is a representation of GDK's GdkPixbufRotation.
type PixbufRotation int

const (
	PIXBUF_ROTATE_NONE             PixbufRotation = C.GDK_PIXBUF_ROTATE_NONE
	PIXBUF_ROTATE_COUNTERCLOCKWISE PixbufRotation = C.GDK_PIXBUF_ROTATE_COUNTERCLOCKWISE
	PIXBUF_ROTATE_UPSIDEDOWN       PixbufRotation = C.GDK_PIXBUF_ROTATE_UPSIDEDOWN
	PIXBUF_ROTATE_CLOCKWISE        PixbufRotation = C.GDK_PIXBUF_ROTATE_CLOCKWISE
)

func marshalInterpType(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return InterpType(c), nil
}

// ModifierType is a representation of GDK's GdkModifierType.
type ModifierType uint

const (
	GDK_SHIFT_MASK    ModifierType = C.GDK_SHIFT_MASK
	GDK_LOCK_MASK                  = C.GDK_LOCK_MASK
	GDK_CONTROL_MASK               = C.GDK_CONTROL_MASK
	GDK_MOD1_MASK                  = C.GDK_MOD1_MASK
	GDK_MOD2_MASK                  = C.GDK_MOD2_MASK
	GDK_MOD3_MASK                  = C.GDK_MOD3_MASK
	GDK_MOD4_MASK                  = C.GDK_MOD4_MASK
	GDK_MOD5_MASK                  = C.GDK_MOD5_MASK
	GDK_BUTTON1_MASK               = C.GDK_BUTTON1_MASK
	GDK_BUTTON2_MASK               = C.GDK_BUTTON2_MASK
	GDK_BUTTON3_MASK               = C.GDK_BUTTON3_MASK
	GDK_BUTTON4_MASK               = C.GDK_BUTTON4_MASK
	GDK_BUTTON5_MASK               = C.GDK_BUTTON5_MASK
	GDK_SUPER_MASK                 = C.GDK_SUPER_MASK
	GDK_HYPER_MASK                 = C.GDK_HYPER_MASK
	GDK_META_MASK                  = C.GDK_META_MASK
	GDK_RELEASE_MASK               = C.GDK_RELEASE_MASK
	GDK_MODIFIER_MASK              = C.GDK_MODIFIER_MASK
)

func marshalModifierType(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return ModifierType(c), nil
}

// PixbufAlphaMode is a representation of GDK's GdkPixbufAlphaMode.
type PixbufAlphaMode int

const (
	GDK_PIXBUF_ALPHA_BILEVEL PixbufAlphaMode = C.GDK_PIXBUF_ALPHA_BILEVEL
	GDK_PIXBUF_ALPHA_FULL    PixbufAlphaMode = C.GDK_PIXBUF_ALPHA_FULL
)

func marshalPixbufAlphaMode(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return PixbufAlphaMode(c), nil
}

// Selections
const (
	SELECTION_PRIMARY       Atom = 1
	SELECTION_SECONDARY     Atom = 2
	SELECTION_CLIPBOARD     Atom = 69
	TARGET_BITMAP           Atom = 5
	TARGET_COLORMAP         Atom = 7
	TARGET_DRAWABLE         Atom = 17
	TARGET_PIXMAP           Atom = 20
	TARGET_STRING           Atom = 31
	SELECTION_TYPE_ATOM     Atom = 4
	SELECTION_TYPE_BITMAP   Atom = 5
	SELECTION_TYPE_COLORMAP Atom = 7
	SELECTION_TYPE_DRAWABLE Atom = 17
	SELECTION_TYPE_INTEGER  Atom = 19
	SELECTION_TYPE_PIXMAP   Atom = 20
	SELECTION_TYPE_WINDOW   Atom = 33
	SELECTION_TYPE_STRING   Atom = 31
)

// added by terrak
// EventMask is a representation of GDK's GdkEventMask.
type EventMask int

const (
	EXPOSURE_MASK            EventMask = C.GDK_EXPOSURE_MASK
	POINTER_MOTION_MASK      EventMask = C.GDK_POINTER_MOTION_MASK
	POINTER_MOTION_HINT_MASK EventMask = C.GDK_POINTER_MOTION_HINT_MASK
	BUTTON_MOTION_MASK       EventMask = C.GDK_BUTTON_MOTION_MASK
	BUTTON1_MOTION_MASK      EventMask = C.GDK_BUTTON1_MOTION_MASK
	BUTTON2_MOTION_MASK      EventMask = C.GDK_BUTTON2_MOTION_MASK
	BUTTON3_MOTION_MASK      EventMask = C.GDK_BUTTON3_MOTION_MASK
	BUTTON_PRESS_MASK        EventMask = C.GDK_BUTTON_PRESS_MASK
	BUTTON_RELEASE_MASK      EventMask = C.GDK_BUTTON_RELEASE_MASK
	KEY_PRESS_MASK           EventMask = C.GDK_KEY_PRESS_MASK
	KEY_RELEASE_MASK         EventMask = C.GDK_KEY_RELEASE_MASK
	ENTER_NOTIFY_MASK        EventMask = C.GDK_ENTER_NOTIFY_MASK
	LEAVE_NOTIFY_MASK        EventMask = C.GDK_LEAVE_NOTIFY_MASK
	FOCUS_CHANGE_MASK        EventMask = C.GDK_FOCUS_CHANGE_MASK
	STRUCTURE_MASK           EventMask = C.GDK_STRUCTURE_MASK
	PROPERTY_CHANGE_MASK     EventMask = C.GDK_PROPERTY_CHANGE_MASK
	VISIBILITY_NOTIFY_MASK   EventMask = C.GDK_VISIBILITY_NOTIFY_MASK
	PROXIMITY_IN_MASK        EventMask = C.GDK_PROXIMITY_IN_MASK
	PROXIMITY_OUT_MASK       EventMask = C.GDK_PROXIMITY_OUT_MASK
	SUBSTRUCTURE_MASK        EventMask = C.GDK_SUBSTRUCTURE_MASK
	SCROLL_MASK              EventMask = C.GDK_SCROLL_MASK
	TOUCH_MASK               EventMask = C.GDK_TOUCH_MASK
	SMOOTH_SCROLL_MASK       EventMask = C.GDK_SMOOTH_SCROLL_MASK
	ALL_EVENTS_MASK          EventMask = C.GDK_ALL_EVENTS_MASK
)

func marshalEventMask(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return EventMask(c), nil
}

// added by lazyshot
// ScrollDirection is a representation of GDK's GdkScrollDirection

type ScrollDirection int

const (
	SCROLL_UP     ScrollDirection = C.GDK_SCROLL_UP
	SCROLL_DOWN   ScrollDirection = C.GDK_SCROLL_DOWN
	SCROLL_LEFT   ScrollDirection = C.GDK_SCROLL_LEFT
	SCROLL_RIGHT  ScrollDirection = C.GDK_SCROLL_RIGHT
	SCROLL_SMOOTH ScrollDirection = C.GDK_SCROLL_SMOOTH
)

// WindowState is a representation of GDK's GdkWindowState
type WindowState int

const (
	WINDOW_STATE_WITHDRAWN  WindowState = C.GDK_WINDOW_STATE_WITHDRAWN
	WINDOW_STATE_ICONIFIED  WindowState = C.GDK_WINDOW_STATE_ICONIFIED
	WINDOW_STATE_MAXIMIZED  WindowState = C.GDK_WINDOW_STATE_MAXIMIZED
	WINDOW_STATE_STICKY     WindowState = C.GDK_WINDOW_STATE_STICKY
	WINDOW_STATE_FULLSCREEN WindowState = C.GDK_WINDOW_STATE_FULLSCREEN
	WINDOW_STATE_ABOVE      WindowState = C.GDK_WINDOW_STATE_ABOVE
	WINDOW_STATE_BELOW      WindowState = C.GDK_WINDOW_STATE_BELOW
	WINDOW_STATE_FOCUSED    WindowState = C.GDK_WINDOW_STATE_FOCUSED
	WINDOW_STATE_TILED      WindowState = C.GDK_WINDOW_STATE_TILED
)

// WindowTypeHint is a representation of GDK's GdkWindowTypeHint
type WindowTypeHint int

const (
	WINDOW_TYPE_HINT_NORMAL        WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_NORMAL
	WINDOW_TYPE_HINT_DIALOG        WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_DIALOG
	WINDOW_TYPE_HINT_MENU          WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_MENU
	WINDOW_TYPE_HINT_TOOLBAR       WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_TOOLBAR
	WINDOW_TYPE_HINT_SPLASHSCREEN  WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_SPLASHSCREEN
	WINDOW_TYPE_HINT_UTILITY       WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_UTILITY
	WINDOW_TYPE_HINT_DOCK          WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_DOCK
	WINDOW_TYPE_HINT_DESKTOP       WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_DESKTOP
	WINDOW_TYPE_HINT_DROPDOWN_MENU WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_DROPDOWN_MENU
	WINDOW_TYPE_HINT_POPUP_MENU    WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_POPUP_MENU
	WINDOW_TYPE_HINT_TOOLTIP       WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_TOOLTIP
	WINDOW_TYPE_HINT_NOTIFICATION  WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_NOTIFICATION
	WINDOW_TYPE_HINT_COMBO         WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_COMBO
	WINDOW_TYPE_HINT_DND           WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_DND
)

// CURRENT_TIME is a representation of GDK_CURRENT_TIME

const CURRENT_TIME = C.GDK_CURRENT_TIME

// GrabStatus is a representation of GdkGrabStatus

type GrabStatus int

const (
	GRAB_SUCCESS         GrabStatus = C.GDK_GRAB_SUCCESS
	GRAB_ALREADY_GRABBED GrabStatus = C.GDK_GRAB_ALREADY_GRABBED
	GRAB_INVALID_TIME    GrabStatus = C.GDK_GRAB_INVALID_TIME
	GRAB_FROZEN          GrabStatus = C.GDK_GRAB_FROZEN
	// Only exists since 3.16
	// GRAB_FAILED GrabStatus = C.GDK_GRAB_FAILED
	GRAB_FAILED GrabStatus = 5
)

// GrabOwnership is a representation of GdkGrabOwnership

type GrabOwnership int

const (
	OWNERSHIP_NONE        GrabOwnership = C.GDK_OWNERSHIP_NONE
	OWNERSHIP_WINDOW      GrabOwnership = C.GDK_OWNERSHIP_WINDOW
	OWNERSHIP_APPLICATION GrabOwnership = C.GDK_OWNERSHIP_APPLICATION
)

// DeviceType is a representation of GdkDeviceType

type DeviceType int

const (
	DEVICE_TYPE_MASTER   DeviceType = C.GDK_DEVICE_TYPE_MASTER
	DEVICE_TYPE_SLAVE    DeviceType = C.GDK_DEVICE_TYPE_SLAVE
	DEVICE_TYPE_FLOATING DeviceType = C.GDK_DEVICE_TYPE_FLOATING
)

// EventPropagation constants

const (
	GDK_EVENT_PROPAGATE bool = C.GDK_EVENT_PROPAGATE != 0
	GDK_EVENT_STOP      bool = C.GDK_EVENT_STOP != 0
)

/*
 * GdkAtom
 */

// Atom is a representation of GDK's GdkAtom.
type Atom uintptr

// native returns the underlying GdkAtom.
func (v Atom) native() C.GdkAtom {
	return C.toGdkAtom(unsafe.Pointer(uintptr(v)))
}

func (v Atom) Name() string {
	c := C.gdk_atom_name(v.native())
	defer C.g_free(C.gpointer(c))
	return C.GoString((*C.char)(c))
}

// GdkAtomIntern is a wrapper around gdk_atom_intern
func GdkAtomIntern(atomName string, onlyIfExists bool) Atom {
	cstr := C.CString(atomName)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gdk_atom_intern((*C.gchar)(cstr), gbool(onlyIfExists))
	return Atom(uintptr(unsafe.Pointer(c)))
}

/*
 * GdkDevice
 */

// Device is a representation of GDK's GdkDevice.
type Device struct {
	*glib.Object
}

// native returns a pointer to the underlying GdkDevice.
func (v *Device) native() *C.GdkDevice {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGdkDevice(p)
}

// Native returns a pointer to the underlying GdkDevice.
func (v *Device) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalDevice(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	return &Device{obj}, nil
}

/*
 * GdkCursor
 */

// Cursor is a representation of GdkCursor.
type Cursor struct {
	*glib.Object
}

// CursorNewFromName is a wrapper around gdk_cursor_new_from_name().
func CursorNewFromName(display *Display, name string) (*Cursor, error) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gdk_cursor_new_from_name(display.native(), (*C.gchar)(cstr))
	if c == nil {
		return nil, nilPtrErr
	}

	return &Cursor{glib.Take(unsafe.Pointer(c))}, nil
}

// native returns a pointer to the underlying GdkCursor.
func (v *Cursor) native() *C.GdkCursor {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGdkCursor(p)
}

// Native returns a pointer to the underlying GdkCursor.
func (v *Cursor) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalCursor(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	return &Cursor{obj}, nil
}

/*
 * GdkDeviceManager
 */

// DeviceManager is a representation of GDK's GdkDeviceManager.
type DeviceManager struct {
	*glib.Object
}

// native returns a pointer to the underlying GdkDeviceManager.
func (v *DeviceManager) native() *C.GdkDeviceManager {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGdkDeviceManager(p)
}

// Native returns a pointer to the underlying GdkDeviceManager.
func (v *DeviceManager) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalDeviceManager(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	return &DeviceManager{obj}, nil
}

// GetDisplay() is a wrapper around gdk_device_manager_get_display().
func (v *DeviceManager) GetDisplay() (*Display, error) {
	c := C.gdk_device_manager_get_display(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	return &Display{glib.Take(unsafe.Pointer(c))}, nil
}

/*
 * GdkDisplay
 */

// Display is a representation of GDK's GdkDisplay.
type Display struct {
	*glib.Object
}

// native returns a pointer to the underlying GdkDisplay.
func (v *Display) native() *C.GdkDisplay {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGdkDisplay(p)
}

// Native returns a pointer to the underlying GdkDisplay.
func (v *Display) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalDisplay(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	return &Display{obj}, nil
}

func toDisplay(s *C.GdkDisplay) (*Display, error) {
	if s == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(s))}
	return &Display{obj}, nil
}

// DisplayOpen() is a wrapper around gdk_display_open().
func DisplayOpen(displayName string) (*Display, error) {
	cstr := C.CString(displayName)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gdk_display_open((*C.gchar)(cstr))
	if c == nil {
		return nil, nilPtrErr
	}

	return &Display{glib.Take(unsafe.Pointer(c))}, nil
}

// DisplayGetDefault() is a wrapper around gdk_display_get_default().
func DisplayGetDefault() (*Display, error) {
	c := C.gdk_display_get_default()
	if c == nil {
		return nil, nilPtrErr
	}

	return &Display{glib.Take(unsafe.Pointer(c))}, nil
}

// GetName() is a wrapper around gdk_display_get_name().
func (v *Display) GetName() (string, error) {
	c := C.gdk_display_get_name(v.native())
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

// GetDefaultScreen() is a wrapper around gdk_display_get_default_screen().
func (v *Display) GetDefaultScreen() (*Screen, error) {
	c := C.gdk_display_get_default_screen(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	return &Screen{glib.Take(unsafe.Pointer(c))}, nil
}

// DeviceIsGrabbed() is a wrapper around gdk_display_device_is_grabbed().
func (v *Display) DeviceIsGrabbed(device *Device) bool {
	c := C.gdk_display_device_is_grabbed(v.native(), device.native())
	return gobool(c)
}

// Beep() is a wrapper around gdk_display_beep().
func (v *Display) Beep() {
	C.gdk_display_beep(v.native())
}

// Sync() is a wrapper around gdk_display_sync().
func (v *Display) Sync() {
	C.gdk_display_sync(v.native())
}

// Flush() is a wrapper around gdk_display_flush().
func (v *Display) Flush() {
	C.gdk_display_flush(v.native())
}

// Close() is a wrapper around gdk_display_close().
func (v *Display) Close() {
	C.gdk_display_close(v.native())
}

// IsClosed() is a wrapper around gdk_display_is_closed().
func (v *Display) IsClosed() bool {
	c := C.gdk_display_is_closed(v.native())
	return gobool(c)
}

// GetEvent() is a wrapper around gdk_display_get_event().
func (v *Display) GetEvent() (*Event, error) {
	c := C.gdk_display_get_event(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	//The finalizer is not on the glib.Object but on the event.
	e := &Event{c}
	runtime.SetFinalizer(e, (*Event).free)
	return e, nil
}

// PeekEvent() is a wrapper around gdk_display_peek_event().
func (v *Display) PeekEvent() (*Event, error) {
	c := C.gdk_display_peek_event(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	//The finalizer is not on the glib.Object but on the event.
	e := &Event{c}
	runtime.SetFinalizer(e, (*Event).free)
	return e, nil
}

// PutEvent() is a wrapper around gdk_display_put_event().
func (v *Display) PutEvent(event *Event) {
	C.gdk_display_put_event(v.native(), event.native())
}

// HasPending() is a wrapper around gdk_display_has_pending().
func (v *Display) HasPending() bool {
	c := C.gdk_display_has_pending(v.native())
	return gobool(c)
}

// SetDoubleClickTime() is a wrapper around gdk_display_set_double_click_time().
func (v *Display) SetDoubleClickTime(msec uint) {
	C.gdk_display_set_double_click_time(v.native(), C.guint(msec))
}

// SetDoubleClickDistance() is a wrapper around gdk_display_set_double_click_distance().
func (v *Display) SetDoubleClickDistance(distance uint) {
	C.gdk_display_set_double_click_distance(v.native(), C.guint(distance))
}

// SupportsColorCursor() is a wrapper around gdk_display_supports_cursor_color().
func (v *Display) SupportsColorCursor() bool {
	c := C.gdk_display_supports_cursor_color(v.native())
	return gobool(c)
}

// SupportsCursorAlpha() is a wrapper around gdk_display_supports_cursor_alpha().
func (v *Display) SupportsCursorAlpha() bool {
	c := C.gdk_display_supports_cursor_alpha(v.native())
	return gobool(c)
}

// GetDefaultCursorSize() is a wrapper around gdk_display_get_default_cursor_size().
func (v *Display) GetDefaultCursorSize() uint {
	c := C.gdk_display_get_default_cursor_size(v.native())
	return uint(c)
}

// GetMaximalCursorSize() is a wrapper around gdk_display_get_maximal_cursor_size().
func (v *Display) GetMaximalCursorSize() (width, height uint) {
	var w, h C.guint
	C.gdk_display_get_maximal_cursor_size(v.native(), &w, &h)
	return uint(w), uint(h)
}

// GetDefaultGroup() is a wrapper around gdk_display_get_default_group().
func (v *Display) GetDefaultGroup() (*Window, error) {
	c := C.gdk_display_get_default_group(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	return &Window{glib.Take(unsafe.Pointer(c))}, nil
}

// SupportsSelectionNotification() is a wrapper around
// gdk_display_supports_selection_notification().
func (v *Display) SupportsSelectionNotification() bool {
	c := C.gdk_display_supports_selection_notification(v.native())
	return gobool(c)
}

// RequestSelectionNotification() is a wrapper around
// gdk_display_request_selection_notification().
func (v *Display) RequestSelectionNotification(selection Atom) bool {
	c := C.gdk_display_request_selection_notification(v.native(),
		selection.native())
	return gobool(c)
}

// SupportsClipboardPersistence() is a wrapper around
// gdk_display_supports_clipboard_persistence().
func (v *Display) SupportsClipboardPersistence() bool {
	c := C.gdk_display_supports_clipboard_persistence(v.native())
	return gobool(c)
}

// TODO(jrick)
func (v *Display) StoreClipboard(clipboardWindow *Window, time uint32, targets ...Atom) {
	panic("Not implemented")
}

// SupportsShapes() is a wrapper around gdk_display_supports_shapes().
func (v *Display) SupportsShapes() bool {
	c := C.gdk_display_supports_shapes(v.native())
	return gobool(c)
}

// SupportsInputShapes() is a wrapper around gdk_display_supports_input_shapes().
func (v *Display) SupportsInputShapes() bool {
	c := C.gdk_display_supports_input_shapes(v.native())
	return gobool(c)
}

// TODO(jrick) glib.AppLaunchContext GdkAppLaunchContext
func (v *Display) GetAppLaunchContext() {
	panic("Not implemented")
}

// NotifyStartupComplete() is a wrapper around gdk_display_notify_startup_complete().
func (v *Display) NotifyStartupComplete(startupID string) {
	cstr := C.CString(startupID)
	defer C.free(unsafe.Pointer(cstr))
	C.gdk_display_notify_startup_complete(v.native(), (*C.gchar)(cstr))
}

// EventType is a representation of GDK's GdkEventType.
// Do not confuse these event types with the signals that GTK+ widgets emit
type EventType int

func marshalEventType(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return EventType(c), nil
}

const (
	EVENT_NOTHING             EventType = C.GDK_NOTHING
	EVENT_DELETE              EventType = C.GDK_DELETE
	EVENT_DESTROY             EventType = C.GDK_DESTROY
	EVENT_EXPOSE              EventType = C.GDK_EXPOSE
	EVENT_MOTION_NOTIFY       EventType = C.GDK_MOTION_NOTIFY
	EVENT_BUTTON_PRESS        EventType = C.GDK_BUTTON_PRESS
	EVENT_2BUTTON_PRESS       EventType = C.GDK_2BUTTON_PRESS
	EVENT_DOUBLE_BUTTON_PRESS EventType = C.GDK_DOUBLE_BUTTON_PRESS
	EVENT_3BUTTON_PRESS       EventType = C.GDK_3BUTTON_PRESS
	EVENT_TRIPLE_BUTTON_PRESS EventType = C.GDK_TRIPLE_BUTTON_PRESS
	EVENT_BUTTON_RELEASE      EventType = C.GDK_BUTTON_RELEASE
	EVENT_KEY_PRESS           EventType = C.GDK_KEY_PRESS
	EVENT_KEY_RELEASE         EventType = C.GDK_KEY_RELEASE
	EVENT_LEAVE_NOTIFY        EventType = C.GDK_ENTER_NOTIFY
	EVENT_FOCUS_CHANGE        EventType = C.GDK_FOCUS_CHANGE
	EVENT_CONFIGURE           EventType = C.GDK_CONFIGURE
	EVENT_MAP                 EventType = C.GDK_MAP
	EVENT_UNMAP               EventType = C.GDK_UNMAP
	EVENT_PROPERTY_NOTIFY     EventType = C.GDK_PROPERTY_NOTIFY
	EVENT_SELECTION_CLEAR     EventType = C.GDK_SELECTION_CLEAR
	EVENT_SELECTION_REQUEST   EventType = C.GDK_SELECTION_REQUEST
	EVENT_SELECTION_NOTIFY    EventType = C.GDK_SELECTION_NOTIFY
	EVENT_PROXIMITY_IN        EventType = C.GDK_PROXIMITY_IN
	EVENT_PROXIMITY_OUT       EventType = C.GDK_PROXIMITY_OUT
	EVENT_DRAG_ENTER          EventType = C.GDK_DRAG_ENTER
	EVENT_DRAG_LEAVE          EventType = C.GDK_DRAG_LEAVE
	EVENT_DRAG_MOTION         EventType = C.GDK_DRAG_MOTION
	EVENT_DRAG_STATUS         EventType = C.GDK_DRAG_STATUS
	EVENT_DROP_START          EventType = C.GDK_DROP_START
	EVENT_DROP_FINISHED       EventType = C.GDK_DROP_FINISHED
	EVENT_CLIENT_EVENT        EventType = C.GDK_CLIENT_EVENT
	EVENT_VISIBILITY_NOTIFY   EventType = C.GDK_VISIBILITY_NOTIFY
	EVENT_SCROLL              EventType = C.GDK_SCROLL
	EVENT_WINDOW_STATE        EventType = C.GDK_WINDOW_STATE
	EVENT_SETTING             EventType = C.GDK_SETTING
	EVENT_OWNER_CHANGE        EventType = C.GDK_OWNER_CHANGE
	EVENT_GRAB_BROKEN         EventType = C.GDK_GRAB_BROKEN
	EVENT_DAMAGE              EventType = C.GDK_DAMAGE
	EVENT_TOUCH_BEGIN         EventType = C.GDK_TOUCH_BEGIN
	EVENT_TOUCH_UPDATE        EventType = C.GDK_TOUCH_UPDATE
	EVENT_TOUCH_END           EventType = C.GDK_TOUCH_END
	EVENT_TOUCH_CANCEL        EventType = C.GDK_TOUCH_CANCEL
	EVENT_LAST                EventType = C.GDK_EVENT_LAST
)

/*
 * GDK Keyval
 */

// KeyvalFromName() is a wrapper around gdk_keyval_from_name().
func KeyvalFromName(keyvalName string) uint {
	str := (*C.gchar)(C.CString(keyvalName))
	defer C.free(unsafe.Pointer(str))
	return uint(C.gdk_keyval_from_name(str))
}

func KeyvalConvertCase(v uint) (lower, upper uint) {
	var l, u C.guint
	l = 0
	u = 0
	C.gdk_keyval_convert_case(C.guint(v), &l, &u)
	return uint(l), uint(u)
}

func KeyvalIsLower(v uint) bool {
	return gobool(C.gdk_keyval_is_lower(C.guint(v)))
}

func KeyvalIsUpper(v uint) bool {
	return gobool(C.gdk_keyval_is_upper(C.guint(v)))
}

func KeyvalToLower(v uint) uint {
	return uint(C.gdk_keyval_to_lower(C.guint(v)))
}

func KeyvalToUpper(v uint) uint {
	return uint(C.gdk_keyval_to_upper(C.guint(v)))
}

func KeyvalToUnicode(v uint) rune {
	return rune(C.gdk_keyval_to_unicode(C.guint(v)))
}

func UnicodeToKeyval(v rune) uint {
	return uint(C.gdk_unicode_to_keyval(C.guint32(v)))
}

/*
 * GdkDragContext
 */

// DragContext is a representation of GDK's GdkDragContext.
type DragContext struct {
	*glib.Object
}

// native returns a pointer to the underlying GdkDragContext.
func (v *DragContext) native() *C.GdkDragContext {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGdkDragContext(p)
}

// Native returns a pointer to the underlying GdkDragContext.
func (v *DragContext) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalDragContext(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	return &DragContext{obj}, nil
}

func (v *DragContext) ListTargets() *glib.List {
	c := C.gdk_drag_context_list_targets(v.native())
	return glib.WrapList(uintptr(unsafe.Pointer(c)))
}

/*
 * GdkEvent
 */

// Event is a representation of GDK's GdkEvent.
type Event struct {
	GdkEvent *C.GdkEvent
}

// native returns a pointer to the underlying GdkEvent.
func (v *Event) native() *C.GdkEvent {
	if v == nil {
		return nil
	}
	return v.GdkEvent
}

// Native returns a pointer to the underlying GdkEvent.
func (v *Event) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalEvent(p uintptr) (interface{}, error) {
	c := C.g_value_get_boxed((*C.GValue)(unsafe.Pointer(p)))
	return &Event{(*C.GdkEvent)(unsafe.Pointer(c))}, nil
}

func (v *Event) free() {
	C.gdk_event_free(v.native())
}

/*
 * GdkEventButton
 */

// EventButton is a representation of GDK's GdkEventButton.
type EventButton struct {
	*Event
}

func EventButtonNew() *EventButton {
	ee := (*C.GdkEvent)(unsafe.Pointer(&C.GdkEventButton{}))
	ev := Event{ee}
	return &EventButton{&ev}
}

// EventButtonNewFromEvent returns an EventButton from an Event.
//
// Using widget.Connect() for a key related signal such as
// "button-press-event" results in a *Event being passed as
// the callback's second argument. The argument is actually a
// *EventButton. EventButtonNewFromEvent provides a means of creating
// an EventKey from the Event.
func EventButtonNewFromEvent(event *Event) *EventButton {
	ee := (*C.GdkEvent)(unsafe.Pointer(event.native()))
	ev := Event{ee}
	return &EventButton{&ev}
}

// Native returns a pointer to the underlying GdkEventButton.
func (v *EventButton) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *EventButton) native() *C.GdkEventButton {
	return (*C.GdkEventButton)(unsafe.Pointer(v.Event.native()))
}

func (v *EventButton) X() float64 {
	c := v.native().x
	return float64(c)
}

func (v *EventButton) Y() float64 {
	c := v.native().y
	return float64(c)
}

// XRoot returns the x coordinate of the pointer relative to the root of the screen.
func (v *EventButton) XRoot() float64 {
	c := v.native().x_root
	return float64(c)
}

// YRoot returns the y coordinate of the pointer relative to the root of the screen.
func (v *EventButton) YRoot() float64 {
	c := v.native().y_root
	return float64(c)
}

func (v *EventButton) Button() uint {
	c := v.native().button
	return uint(c)
}

func (v *EventButton) State() uint {
	c := v.native().state
	return uint(c)
}

// Time returns the time of the event in milliseconds.
func (v *EventButton) Time() uint32 {
	c := v.native().time
	return uint32(c)
}

func (v *EventButton) Type() EventType {
	c := v.native()._type
	return EventType(c)
}

func (v *EventButton) MotionVal() (float64, float64) {
	x := v.native().x
	y := v.native().y
	return float64(x), float64(y)
}

func (v *EventButton) MotionValRoot() (float64, float64) {
	x := v.native().x_root
	y := v.native().y_root
	return float64(x), float64(y)
}

func (v *EventButton) ButtonVal() uint {
	c := v.native().button
	return uint(c)
}

/*
 * GdkEventKey
 */

// EventKey is a representation of GDK's GdkEventKey.
type EventKey struct {
	*Event
}

func EventKeyNew() *EventKey {
	ee := (*C.GdkEvent)(unsafe.Pointer(&C.GdkEventKey{}))
	ev := Event{ee}
	return &EventKey{&ev}
}

// EventKeyNewFromEvent returns an EventKey from an Event.
//
// Using widget.Connect() for a key related signal such as
// "key-press-event" results in a *Event being passed as
// the callback's second argument. The argument is actually a
// *EventKey. EventKeyNewFromEvent provides a means of creating
// an EventKey from the Event.
func EventKeyNewFromEvent(event *Event) *EventKey {
	ee := (*C.GdkEvent)(unsafe.Pointer(event.native()))
	ev := Event{ee}
	return &EventKey{&ev}
}

// Native returns a pointer to the underlying GdkEventKey.
func (v *EventKey) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *EventKey) native() *C.GdkEventKey {
	return (*C.GdkEventKey)(unsafe.Pointer(v.Event.native()))
}

func (v *EventKey) KeyVal() uint {
	c := v.native().keyval
	return uint(c)
}

func (v *EventKey) Type() EventType {
	c := v.native()._type
	return EventType(c)
}

func (v *EventKey) State() uint {
	c := v.native().state
	return uint(c)
}

/*
 * GdkEventMotion
 */

type EventMotion struct {
	*Event
}

func EventMotionNew() *EventMotion {
	ee := (*C.GdkEvent)(unsafe.Pointer(&C.GdkEventMotion{}))
	ev := Event{ee}
	return &EventMotion{&ev}
}

// EventMotionNewFromEvent returns an EventMotion from an Event.
//
// Using widget.Connect() for a key related signal such as
// "button-press-event" results in a *Event being passed as
// the callback's second argument. The argument is actually a
// *EventMotion. EventMotionNewFromEvent provides a means of creating
// an EventKey from the Event.
func EventMotionNewFromEvent(event *Event) *EventMotion {
	ee := (*C.GdkEvent)(unsafe.Pointer(event.native()))
	ev := Event{ee}
	return &EventMotion{&ev}
}

// Native returns a pointer to the underlying GdkEventMotion.
func (v *EventMotion) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *EventMotion) native() *C.GdkEventMotion {
	return (*C.GdkEventMotion)(unsafe.Pointer(v.Event.native()))
}

func (v *EventMotion) MotionVal() (float64, float64) {
	x := v.native().x
	y := v.native().y
	return float64(x), float64(y)
}

func (v *EventMotion) MotionValRoot() (float64, float64) {
	x := v.native().x_root
	y := v.native().y_root
	return float64(x), float64(y)
}

/*
 * GdkEventScroll
 */

// EventScroll is a representation of GDK's GdkEventScroll.
type EventScroll struct {
	*Event
}

func EventScrollNew() *EventScroll {
	ee := (*C.GdkEvent)(unsafe.Pointer(&C.GdkEventScroll{}))
	ev := Event{ee}
	return &EventScroll{&ev}
}

// EventScrollNewFromEvent returns an EventScroll from an Event.
//
// Using widget.Connect() for a key related signal such as
// "button-press-event" results in a *Event being passed as
// the callback's second argument. The argument is actually a
// *EventScroll. EventScrollNewFromEvent provides a means of creating
// an EventKey from the Event.
func EventScrollNewFromEvent(event *Event) *EventScroll {
	ee := (*C.GdkEvent)(unsafe.Pointer(event.native()))
	ev := Event{ee}
	return &EventScroll{&ev}
}

// Native returns a pointer to the underlying GdkEventScroll.
func (v *EventScroll) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *EventScroll) native() *C.GdkEventScroll {
	return (*C.GdkEventScroll)(unsafe.Pointer(v.Event.native()))
}

func (v *EventScroll) DeltaX() float64 {
	return float64(v.native().delta_x)
}

func (v *EventScroll) DeltaY() float64 {
	return float64(v.native().delta_y)
}

func (v *EventScroll) X() float64 {
	return float64(v.native().x)
}

func (v *EventScroll) Y() float64 {
	return float64(v.native().y)
}

func (v *EventScroll) Type() EventType {
	c := v.native()._type
	return EventType(c)
}

func (v *EventScroll) Direction() ScrollDirection {
	c := v.native().direction
	return ScrollDirection(c)
}

/*
 * GdkEventWindowState
 */

// EventWindowState is a representation of GDK's GdkEventWindowState.
type EventWindowState struct {
	*Event
}

func EventWindowStateNew() *EventWindowState {
	ee := (*C.GdkEvent)(unsafe.Pointer(&C.GdkEventWindowState{}))
	ev := Event{ee}
	return &EventWindowState{&ev}
}

// EventWindowStateNewFromEvent returns an EventWindowState from an Event.
//
// Using widget.Connect() for the
// "window-state-event" signal results in a *Event being passed as
// the callback's second argument. The argument is actually a
// *EventWindowState. EventWindowStateNewFromEvent provides a means of creating
// an EventWindowState from the Event.
func EventWindowStateNewFromEvent(event *Event) *EventWindowState {
	ee := (*C.GdkEvent)(unsafe.Pointer(event.native()))
	ev := Event{ee}
	return &EventWindowState{&ev}
}

// Native returns a pointer to the underlying GdkEventWindowState.
func (v *EventWindowState) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *EventWindowState) native() *C.GdkEventWindowState {
	return (*C.GdkEventWindowState)(unsafe.Pointer(v.Event.native()))
}

func (v *EventWindowState) Type() EventType {
	c := v.native()._type
	return EventType(c)
}

func (v *EventWindowState) ChangedMask() WindowState {
	c := v.native().changed_mask
	return WindowState(c)
}

func (v *EventWindowState) NewWindowState() WindowState {
	c := v.native().new_window_state
	return WindowState(c)
}

/*
 * GdkGravity
 */
type GdkGravity int

const (
	GDK_GRAVITY_NORTH_WEST = C.GDK_GRAVITY_NORTH_WEST
	GDK_GRAVITY_NORTH      = C.GDK_GRAVITY_NORTH
	GDK_GRAVITY_NORTH_EAST = C.GDK_GRAVITY_NORTH_EAST
	GDK_GRAVITY_WEST       = C.GDK_GRAVITY_WEST
	GDK_GRAVITY_CENTER     = C.GDK_GRAVITY_CENTER
	GDK_GRAVITY_EAST       = C.GDK_GRAVITY_EAST
	GDK_GRAVITY_SOUTH_WEST = C.GDK_GRAVITY_SOUTH_WEST
	GDK_GRAVITY_SOUTH      = C.GDK_GRAVITY_SOUTH
	GDK_GRAVITY_SOUTH_EAST = C.GDK_GRAVITY_SOUTH_EAST
	GDK_GRAVITY_STATIC     = C.GDK_GRAVITY_STATIC
)

/*
 * GdkPixbuf
 */

// Pixbuf is a representation of GDK's GdkPixbuf.
type Pixbuf struct {
	*glib.Object
}

// native returns a pointer to the underlying GdkPixbuf.
func (v *Pixbuf) native() *C.GdkPixbuf {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGdkPixbuf(p)
}

// Native returns a pointer to the underlying GdkPixbuf.
func (v *Pixbuf) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalPixbuf(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	return &Pixbuf{obj}, nil
}

// GetColorspace is a wrapper around gdk_pixbuf_get_colorspace().
func (v *Pixbuf) GetColorspace() Colorspace {
	c := C.gdk_pixbuf_get_colorspace(v.native())
	return Colorspace(c)
}

// GetNChannels is a wrapper around gdk_pixbuf_get_n_channels().
func (v *Pixbuf) GetNChannels() int {
	c := C.gdk_pixbuf_get_n_channels(v.native())
	return int(c)
}

// GetHasAlpha is a wrapper around gdk_pixbuf_get_has_alpha().
func (v *Pixbuf) GetHasAlpha() bool {
	c := C.gdk_pixbuf_get_has_alpha(v.native())
	return gobool(c)
}

// GetBitsPerSample is a wrapper around gdk_pixbuf_get_bits_per_sample().
func (v *Pixbuf) GetBitsPerSample() int {
	c := C.gdk_pixbuf_get_bits_per_sample(v.native())
	return int(c)
}

// GetPixels is a wrapper around gdk_pixbuf_get_pixels_with_length().
// A Go slice is used to represent the underlying Pixbuf data array, one
// byte per channel.
func (v *Pixbuf) GetPixels() (channels []byte) {
	var length C.guint
	c := C.gdk_pixbuf_get_pixels_with_length(v.native(), &length)
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&channels))
	sliceHeader.Data = uintptr(unsafe.Pointer(c))
	sliceHeader.Len = int(length)
	sliceHeader.Cap = int(length)

	// To make sure the slice doesn't outlive the Pixbuf, add a reference
	v.Ref()
	runtime.SetFinalizer(&channels, func(_ *[]byte) {
		v.Unref()
	})
	return
}

// GetWidth is a wrapper around gdk_pixbuf_get_width().
func (v *Pixbuf) GetWidth() int {
	c := C.gdk_pixbuf_get_width(v.native())
	return int(c)
}

// GetHeight is a wrapper around gdk_pixbuf_get_height().
func (v *Pixbuf) GetHeight() int {
	c := C.gdk_pixbuf_get_height(v.native())
	return int(c)
}

// GetRowstride is a wrapper around gdk_pixbuf_get_rowstride().
func (v *Pixbuf) GetRowstride() int {
	c := C.gdk_pixbuf_get_rowstride(v.native())
	return int(c)
}

// GetByteLength is a wrapper around gdk_pixbuf_get_byte_length().
func (v *Pixbuf) GetByteLength() int {
	c := C.gdk_pixbuf_get_byte_length(v.native())
	return int(c)
}

// GetOption is a wrapper around gdk_pixbuf_get_option().  ok is true if
// the key has an associated value.
func (v *Pixbuf) GetOption(key string) (value string, ok bool) {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gdk_pixbuf_get_option(v.native(), (*C.gchar)(cstr))
	if c == nil {
		return "", false
	}
	return C.GoString((*C.char)(c)), true
}

// PixbufNew is a wrapper around gdk_pixbuf_new().
func PixbufNew(colorspace Colorspace, hasAlpha bool, bitsPerSample, width, height int) (*Pixbuf, error) {
	c := C.gdk_pixbuf_new(C.GdkColorspace(colorspace), gbool(hasAlpha),
		C.int(bitsPerSample), C.int(width), C.int(height))
	if c == nil {
		return nil, nilPtrErr
	}

	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	p := &Pixbuf{obj}
	//obj.Ref()
	runtime.SetFinalizer(p, func(_ interface{}) { obj.Unref() })
	return p, nil
}

// PixbufCopy is a wrapper around gdk_pixbuf_copy().
func PixbufCopy(v *Pixbuf) (*Pixbuf, error) {
	c := C.gdk_pixbuf_copy(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	p := &Pixbuf{obj}
	//obj.Ref()
	runtime.SetFinalizer(p, func(_ interface{}) { obj.Unref() })
	return p, nil
}

// PixbufNewFromFile is a wrapper around gdk_pixbuf_new_from_file().
func PixbufNewFromFile(filename string) (*Pixbuf, error) {
	cstr := C.CString(filename)
	defer C.free(unsafe.Pointer(cstr))

	var err *C.GError
	c := C.gdk_pixbuf_new_from_file((*C.char)(cstr), &err)
	if c == nil {
		defer C.g_error_free(err)
		return nil, errors.New(C.GoString((*C.char)(err.message)))
	}

	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	p := &Pixbuf{obj}
	//obj.Ref()
	runtime.SetFinalizer(p, func(_ interface{}) { obj.Unref() })
	return p, nil
}

// PixbufNewFromFileAtSize is a wrapper around gdk_pixbuf_new_from_file_at_size().
func PixbufNewFromFileAtSize(filename string, width, height int) (*Pixbuf, error) {
	cstr := C.CString(filename)
	defer C.free(unsafe.Pointer(cstr))

	var err *C.GError = nil
	c := C.gdk_pixbuf_new_from_file_at_size(cstr, C.int(width), C.int(height), &err)
	if err != nil {
		defer C.g_error_free(err)
		return nil, errors.New(C.GoString((*C.char)(err.message)))
	}

	if c == nil {
		return nil, nilPtrErr
	}

	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	p := &Pixbuf{obj}
	//obj.Ref()
	runtime.SetFinalizer(p, func(_ interface{}) { obj.Unref() })
	return p, nil
}

// PixbufNewFromFileAtScale is a wrapper around gdk_pixbuf_new_from_file_at_scale().
func PixbufNewFromFileAtScale(filename string, width, height int, preserveAspectRatio bool) (*Pixbuf, error) {
	cstr := C.CString(filename)
	defer C.free(unsafe.Pointer(cstr))

	var err *C.GError = nil
	c := C.gdk_pixbuf_new_from_file_at_scale(cstr, C.int(width), C.int(height),
		gbool(preserveAspectRatio), &err)
	if err != nil {
		defer C.g_error_free(err)
		return nil, errors.New(C.GoString((*C.char)(err.message)))
	}

	if c == nil {
		return nil, nilPtrErr
	}

	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	p := &Pixbuf{obj}
	//obj.Ref()
	runtime.SetFinalizer(p, func(_ interface{}) { obj.Unref() })
	return p, nil
}

// ScaleSimple is a wrapper around gdk_pixbuf_scale_simple().
func (v *Pixbuf) ScaleSimple(destWidth, destHeight int, interpType InterpType) (*Pixbuf, error) {
	c := C.gdk_pixbuf_scale_simple(v.native(), C.int(destWidth),
		C.int(destHeight), C.GdkInterpType(interpType))
	if c == nil {
		return nil, nilPtrErr
	}

	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	p := &Pixbuf{obj}
	//obj.Ref()
	runtime.SetFinalizer(p, func(_ interface{}) { obj.Unref() })
	return p, nil
}

// RotateSimple is a wrapper around gdk_pixbuf_rotate_simple().
func (v *Pixbuf) RotateSimple(angle PixbufRotation) (*Pixbuf, error) {
	c := C.gdk_pixbuf_rotate_simple(v.native(), C.GdkPixbufRotation(angle))
	if c == nil {
		return nil, nilPtrErr
	}

	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	p := &Pixbuf{obj}
	//obj.Ref()
	runtime.SetFinalizer(p, func(_ interface{}) { obj.Unref() })
	return p, nil
}

// ApplyEmbeddedOrientation is a wrapper around gdk_pixbuf_apply_embedded_orientation().
func (v *Pixbuf) ApplyEmbeddedOrientation() (*Pixbuf, error) {
	c := C.gdk_pixbuf_apply_embedded_orientation(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	p := &Pixbuf{obj}
	//obj.Ref()
	runtime.SetFinalizer(p, func(_ interface{}) { obj.Unref() })
	return p, nil
}

// Flip is a wrapper around gdk_pixbuf_flip().
func (v *Pixbuf) Flip(horizontal bool) (*Pixbuf, error) {
	c := C.gdk_pixbuf_flip(v.native(), gbool(horizontal))
	if c == nil {
		return nil, nilPtrErr
	}

	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	p := &Pixbuf{obj}
	//obj.Ref()
	runtime.SetFinalizer(p, func(_ interface{}) { obj.Unref() })
	return p, nil
}

// SaveJPEG is a wrapper around gdk_pixbuf_save().
// Quality is a number between 0...100
func (v *Pixbuf) SaveJPEG(path string, quality int) error {
	cpath := C.CString(path)
	cquality := C.CString(strconv.Itoa(quality))
	defer C.free(unsafe.Pointer(cpath))
	defer C.free(unsafe.Pointer(cquality))

	var err *C.GError
	c := C._gdk_pixbuf_save_jpeg(v.native(), cpath, &err, cquality)
	if !gobool(c) {
		defer C.g_error_free(err)
		return errors.New(C.GoString((*C.char)(err.message)))
	}

	return nil
}

// SavePNG is a wrapper around gdk_pixbuf_save().
// Compression is a number between 0...9
func (v *Pixbuf) SavePNG(path string, compression int) error {
	cpath := C.CString(path)
	ccompression := C.CString(strconv.Itoa(compression))
	defer C.free(unsafe.Pointer(cpath))
	defer C.free(unsafe.Pointer(ccompression))

	var err *C.GError
	c := C._gdk_pixbuf_save_png(v.native(), cpath, &err, ccompression)
	if !gobool(c) {
		defer C.g_error_free(err)
		return errors.New(C.GoString((*C.char)(err.message)))
	}
	return nil
}

// PixbufGetFileInfo is a wrapper around gdk_pixbuf_get_file_info().
// TODO: need to wrap the returned format to GdkPixbufFormat.
func PixbufGetFileInfo(filename string) (format interface{}, width, height int) {
	cstr := C.CString(filename)
	defer C.free(unsafe.Pointer(cstr))
	var cw, ch C.gint
	format = C.gdk_pixbuf_get_file_info((*C.gchar)(cstr), &cw, &ch)
	// TODO: need to wrap the returned format to GdkPixbufFormat.
	return format, int(cw), int(ch)
}

/*
 * GdkPixbufLoader
 */

// PixbufLoader is a representation of GDK's GdkPixbufLoader.
// Users of PixbufLoader are expected to call Close() when they are finished.
type PixbufLoader struct {
	*glib.Object
}

// native() returns a pointer to the underlying GdkPixbufLoader.
func (v *PixbufLoader) native() *C.GdkPixbufLoader {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGdkPixbufLoader(p)
}

// PixbufLoaderNew() is a wrapper around gdk_pixbuf_loader_new().
func PixbufLoaderNew() (*PixbufLoader, error) {
	c := C.gdk_pixbuf_loader_new()
	if c == nil {
		return nil, nilPtrErr
	}

	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	p := &PixbufLoader{obj}
	obj.Ref()
	runtime.SetFinalizer(p, func(_ interface{}) { obj.Unref() })
	return p, nil
}

// PixbufLoaderNewWithType() is a wrapper around gdk_pixbuf_loader_new_with_type().
func PixbufLoaderNewWithType(t string) (*PixbufLoader, error) {
	var err *C.GError

	cstr := C.CString(t)
	defer C.free(unsafe.Pointer(cstr))

	c := C.gdk_pixbuf_loader_new_with_type((*C.char)(cstr), &err)
	if err != nil {
		defer C.g_error_free(err)
		return nil, errors.New(C.GoString((*C.char)(err.message)))
	}

	if c == nil {
		return nil, nilPtrErr
	}

	return &PixbufLoader{glib.Take(unsafe.Pointer(c))}, nil
}

// Write() is a wrapper around gdk_pixbuf_loader_write().  The
// function signature differs from the C equivalent to satisify the
// io.Writer interface.
func (v *PixbufLoader) Write(data []byte) (int, error) {
	// n is set to 0 on error, and set to len(data) otherwise.
	// This is a tiny hacky to satisfy io.Writer and io.WriteCloser,
	// which would allow access to all io and ioutil goodies,
	// and play along nice with go environment.

	if len(data) == 0 {
		return 0, nil
	}

	var err *C.GError
	c := C.gdk_pixbuf_loader_write(v.native(),
		(*C.guchar)(unsafe.Pointer(&data[0])), C.gsize(len(data)),
		&err)

	if !gobool(c) {
		defer C.g_error_free(err)
		return 0, errors.New(C.GoString((*C.char)(err.message)))
	}

	return len(data), nil
}

// Close is a wrapper around gdk_pixbuf_loader_close().  An error is
// returned instead of a bool like the native C function to support the
// io.Closer interface.
func (v *PixbufLoader) Close() error {
	var err *C.GError

	if ok := gobool(C.gdk_pixbuf_loader_close(v.native(), &err)); !ok {
		defer C.g_error_free(err)
		return errors.New(C.GoString((*C.char)(err.message)))
	}
	return nil
}

// SetSize is a wrapper around gdk_pixbuf_loader_set_size().
func (v *PixbufLoader) SetSize(width, height int) {
	C.gdk_pixbuf_loader_set_size(v.native(), C.int(width), C.int(height))
}

// GetPixbuf is a wrapper around gdk_pixbuf_loader_get_pixbuf().
func (v *PixbufLoader) GetPixbuf() (*Pixbuf, error) {
	c := C.gdk_pixbuf_loader_get_pixbuf(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	p := &Pixbuf{obj}
	//obj.Ref() // Don't call Ref here, gdk_pixbuf_loader_get_pixbuf already did that for us.
	runtime.SetFinalizer(p, func(_ interface{}) { obj.Unref() })
	return p, nil
}

type RGBA struct {
	rgba *C.GdkRGBA
}

func marshalRGBA(p uintptr) (interface{}, error) {
	c := C.g_value_get_boxed((*C.GValue)(unsafe.Pointer(p)))
	c2 := (*C.GdkRGBA)(unsafe.Pointer(c))
	return wrapRGBA(c2), nil
}

func wrapRGBA(obj *C.GdkRGBA) *RGBA {
	return &RGBA{obj}
}

func NewRGBA(values ...float64) *RGBA {
	cval := C.GdkRGBA{}
	c := &RGBA{&cval}
	if len(values) > 0 {
		c.rgba.red = C.gdouble(values[0])
	}
	if len(values) > 1 {
		c.rgba.green = C.gdouble(values[1])
	}
	if len(values) > 2 {
		c.rgba.blue = C.gdouble(values[2])
	}
	if len(values) > 3 {
		c.rgba.alpha = C.gdouble(values[3])
	}
	return c
}

func (c *RGBA) Floats() []float64 {
	return []float64{float64(c.rgba.red), float64(c.rgba.green), float64(c.rgba.blue), float64(c.rgba.alpha)}
}

func (v *RGBA) Native() uintptr {
	return uintptr(unsafe.Pointer(v.rgba))
}

// Parse is a representation of gdk_rgba_parse().
func (v *RGBA) Parse(spec string) bool {
	cstr := (*C.gchar)(C.CString(spec))
	defer C.free(unsafe.Pointer(cstr))

	return gobool(C.gdk_rgba_parse(v.rgba, cstr))
}

// String is a representation of gdk_rgba_to_string().
func (v *RGBA) String() string {
	return C.GoString((*C.char)(C.gdk_rgba_to_string(v.rgba)))
}

// GdkRGBA * 	gdk_rgba_copy ()
// void 	gdk_rgba_free ()
// gboolean 	gdk_rgba_equal ()
// guint 	gdk_rgba_hash ()

// PixbufGetType is a wrapper around gdk_pixbuf_get_type().
func PixbufGetType() glib.Type {
	return glib.Type(C.gdk_pixbuf_get_type())
}

/*
 * GdkRectangle
 */

// Rectangle is a representation of GDK's GdkRectangle type.
type Rectangle struct {
	GdkRectangle C.GdkRectangle
}

func WrapRectangle(p uintptr) *Rectangle {
	return wrapRectangle((*C.GdkRectangle)(unsafe.Pointer(p)))
}

func wrapRectangle(obj *C.GdkRectangle) *Rectangle {
	if obj == nil {
		return nil
	}
	return &Rectangle{*obj}
}

// Native() returns a pointer to the underlying GdkRectangle.
func (r *Rectangle) native() *C.GdkRectangle {
	return &r.GdkRectangle
}

// GetX returns x field of the underlying GdkRectangle.
func (r *Rectangle) GetX() int {
	return int(r.native().x)
}

// GetY returns y field of the underlying GdkRectangle.
func (r *Rectangle) GetY() int {
	return int(r.native().y)
}

// GetWidth returns width field of the underlying GdkRectangle.
func (r *Rectangle) GetWidth() int {
	return int(r.native().width)
}

// GetHeight returns height field of the underlying GdkRectangle.
func (r *Rectangle) GetHeight() int {
	return int(r.native().height)
}

/*
 * GdkVisual
 */

// Visual is a representation of GDK's GdkVisual.
type Visual struct {
	*glib.Object
}

func (v *Visual) native() *C.GdkVisual {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGdkVisual(p)
}

func (v *Visual) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalVisual(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	return &Visual{obj}, nil
}

/*
 * GdkWindow
 */

// Window is a representation of GDK's GdkWindow.
type Window struct {
	*glib.Object
}

// SetCursor is a wrapper around gdk_window_set_cursor().
func (v *Window) SetCursor(cursor *Cursor) {
	C.gdk_window_set_cursor(v.native(), cursor.native())
}

// native returns a pointer to the underlying GdkWindow.
func (v *Window) native() *C.GdkWindow {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGdkWindow(p)
}

// Native returns a pointer to the underlying GdkWindow.
func (v *Window) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalWindow(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	return &Window{obj}, nil
}

func toWindow(s *C.GdkWindow) (*Window, error) {
	if s == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(s))}
	return &Window{obj}, nil
}
