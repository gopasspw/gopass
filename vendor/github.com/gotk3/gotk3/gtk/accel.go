// Same copyright and license as the rest of the files in this project
// This file contains accelerator related functions and structures

package gtk

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
)

// AccelFlags is a representation of GTK's GtkAccelFlags
type AccelFlags int

const (
	ACCEL_VISIBLE AccelFlags = C.GTK_ACCEL_VISIBLE
	ACCEL_LOCKED  AccelFlags = C.GTK_ACCEL_LOCKED
	ACCEL_MASK    AccelFlags = C.GTK_ACCEL_MASK
)

func marshalAccelFlags(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return AccelFlags(c), nil
}

// AcceleratorName is a wrapper around gtk_accelerator_name().
func AcceleratorName(key uint, mods gdk.ModifierType) string {
	c := C.gtk_accelerator_name(C.guint(key), C.GdkModifierType(mods))
	defer C.free(unsafe.Pointer(c))
	return C.GoString((*C.char)(c))
}

// AcceleratorValid is a wrapper around gtk_accelerator_valid().
func AcceleratorValid(key uint, mods gdk.ModifierType) bool {
	return gobool(C.gtk_accelerator_valid(C.guint(key), C.GdkModifierType(mods)))
}

// AcceleratorGetDefaultModMask is a wrapper around gtk_accelerator_get_default_mod_mask().
func AcceleratorGetDefaultModMask() gdk.ModifierType {
	return gdk.ModifierType(C.gtk_accelerator_get_default_mod_mask())
}

// AcceleratorParse is a wrapper around gtk_accelerator_parse().
func AcceleratorParse(acc string) (key uint, mods gdk.ModifierType) {
	cstr := C.CString(acc)
	defer C.free(unsafe.Pointer(cstr))

	k := C.guint(0)
	m := C.GdkModifierType(0)

	C.gtk_accelerator_parse((*C.gchar)(cstr), &k, &m)
	return uint(k), gdk.ModifierType(m)
}

// AcceleratorGetLabel is a wrapper around gtk_accelerator_get_label().
func AcceleratorGetLabel(key uint, mods gdk.ModifierType) string {
	c := C.gtk_accelerator_get_label(C.guint(key), C.GdkModifierType(mods))
	defer C.free(unsafe.Pointer(c))
	return C.GoString((*C.char)(c))
}

// AcceleratorSetDefaultModMask is a wrapper around gtk_accelerator_set_default_mod_mask().
func AcceleratorSetDefaultModMask(mods gdk.ModifierType) {
	C.gtk_accelerator_set_default_mod_mask(C.GdkModifierType(mods))
}

/*
 * GtkAccelGroup
 */

// AccelGroup is a representation of GTK's GtkAccelGroup.
type AccelGroup struct {
	*glib.Object
}

// native returns a pointer to the underlying GtkAccelGroup.
func (v *AccelGroup) native() *C.GtkAccelGroup {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkAccelGroup(p)
}

func marshalAccelGroup(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapAccelGroup(obj), nil
}

func wrapAccelGroup(obj *glib.Object) *AccelGroup {
	return &AccelGroup{obj}
}

// AccelGroup is a wrapper around gtk_accel_group_new().
func AccelGroupNew() (*AccelGroup, error) {
	c := C.gtk_accel_group_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := glib.Take(unsafe.Pointer(c))
	return wrapAccelGroup(obj), nil
}

// Connect is a wrapper around gtk_accel_group_connect().
func (v *AccelGroup) Connect(key uint, mods gdk.ModifierType, flags AccelFlags, f interface{}) {
	closure, _ := glib.ClosureNew(f)
	cl := (*C.struct__GClosure)(unsafe.Pointer(closure))
	C.gtk_accel_group_connect(
		v.native(),
		C.guint(key),
		C.GdkModifierType(mods),
		C.GtkAccelFlags(flags),
		cl)
}

// ConnectByPath is a wrapper around gtk_accel_group_connect_by_path().
func (v *AccelGroup) ConnectByPath(path string, f interface{}) {
	closure, _ := glib.ClosureNew(f)
	cl := (*C.struct__GClosure)(unsafe.Pointer(closure))

	cstr := C.CString(path)
	defer C.free(unsafe.Pointer(cstr))

	C.gtk_accel_group_connect_by_path(
		v.native(),
		(*C.gchar)(cstr),
		cl)
}

// Disconnect is a wrapper around gtk_accel_group_disconnect().
func (v *AccelGroup) Disconnect(f interface{}) {
	closure, _ := glib.ClosureNew(f)
	cl := (*C.struct__GClosure)(unsafe.Pointer(closure))
	C.gtk_accel_group_disconnect(v.native(), cl)
}

// DisconnectKey is a wrapper around gtk_accel_group_disconnect_key().
func (v *AccelGroup) DisconnectKey(key uint, mods gdk.ModifierType) {
	C.gtk_accel_group_disconnect_key(v.native(), C.guint(key), C.GdkModifierType(mods))
}

// Lock is a wrapper around gtk_accel_group_lock().
func (v *AccelGroup) Lock() {
	C.gtk_accel_group_lock(v.native())
}

// Unlock is a wrapper around gtk_accel_group_unlock().
func (v *AccelGroup) Unlock() {
	C.gtk_accel_group_unlock(v.native())
}

// IsLocked is a wrapper around gtk_accel_group_get_is_locked().
func (v *AccelGroup) IsLocked() bool {
	return gobool(C.gtk_accel_group_get_is_locked(v.native()))
}

// AccelGroupFromClosure is a wrapper around gtk_accel_group_from_accel_closure().
func AccelGroupFromClosure(f interface{}) *AccelGroup {
	closure, _ := glib.ClosureNew(f)
	cl := (*C.struct__GClosure)(unsafe.Pointer(closure))
	c := C.gtk_accel_group_from_accel_closure(cl)
	if c == nil {
		return nil
	}
	return wrapAccelGroup(glib.Take(unsafe.Pointer(c)))
}

// GetModifierMask is a wrapper around gtk_accel_group_get_modifier_mask().
func (v *AccelGroup) GetModifierMask() gdk.ModifierType {
	return gdk.ModifierType(C.gtk_accel_group_get_modifier_mask(v.native()))
}

// AccelGroupsActivate is a wrapper around gtk_accel_groups_activate().
func AccelGroupsActivate(obj *glib.Object, key uint, mods gdk.ModifierType) bool {
	return gobool(C.gtk_accel_groups_activate((*C.GObject)(unsafe.Pointer(obj.Native())), C.guint(key), C.GdkModifierType(mods)))
}

// Activate is a wrapper around gtk_accel_group_activate().
func (v *AccelGroup) Activate(quark glib.Quark, acceleratable *glib.Object, key uint, mods gdk.ModifierType) bool {
	return gobool(C.gtk_accel_group_activate(v.native(), C.GQuark(quark), (*C.GObject)(unsafe.Pointer(acceleratable.Native())), C.guint(key), C.GdkModifierType(mods)))
}

// AccelGroupsFromObject is a wrapper around gtk_accel_groups_from_object().
func AccelGroupsFromObject(obj *glib.Object) *glib.SList {
	res := C.gtk_accel_groups_from_object((*C.GObject)(unsafe.Pointer(obj.Native())))
	if res == nil {
		return nil
	}
	return (*glib.SList)(unsafe.Pointer(res))
}

/*
 * GtkAccelMap
 */

// AccelMap is a representation of GTK's GtkAccelMap.
type AccelMap struct {
	*glib.Object
}

// native returns a pointer to the underlying GtkAccelMap.
func (v *AccelMap) native() *C.GtkAccelMap {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkAccelMap(p)
}

func marshalAccelMap(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapAccelMap(obj), nil
}

func wrapAccelMap(obj *glib.Object) *AccelMap {
	return &AccelMap{obj}
}

// AccelMapAddEntry is a wrapper around gtk_accel_map_add_entry().
func AccelMapAddEntry(path string, key uint, mods gdk.ModifierType) {
	cstr := C.CString(path)
	defer C.free(unsafe.Pointer(cstr))

	C.gtk_accel_map_add_entry((*C.gchar)(cstr), C.guint(key), C.GdkModifierType(mods))
}

type AccelKey struct {
	key   uint
	mods  gdk.ModifierType
	flags uint16
}

func (v *AccelKey) native() *C.struct__GtkAccelKey {
	if v == nil {
		return nil
	}

	var val C.struct__GtkAccelKey
	val.accel_key = C.guint(v.key)
	val.accel_mods = C.GdkModifierType(v.mods)
	val.accel_flags = v.flags
	return &val
}

func wrapAccelKey(obj *C.struct__GtkAccelKey) *AccelKey {
	var v AccelKey

	v.key = uint(obj.accel_key)
	v.mods = gdk.ModifierType(obj.accel_mods)
	v.flags = uint16(obj.accel_flags)

	return &v
}

// AccelMapLookupEntry is a wrapper around gtk_accel_map_lookup_entry().
func AccelMapLookupEntry(path string) *AccelKey {
	cstr := C.CString(path)
	defer C.free(unsafe.Pointer(cstr))

	var v *C.struct__GtkAccelKey

	C.gtk_accel_map_lookup_entry((*C.gchar)(cstr), v)
	return wrapAccelKey(v)
}

// AccelMapChangeEntry is a wrapper around gtk_accel_map_change_entry().
func AccelMapChangeEntry(path string, key uint, mods gdk.ModifierType, replace bool) bool {
	cstr := C.CString(path)
	defer C.free(unsafe.Pointer(cstr))

	return gobool(C.gtk_accel_map_change_entry((*C.gchar)(cstr), C.guint(key), C.GdkModifierType(mods), gbool(replace)))
}

// AccelMapLoad is a wrapper around gtk_accel_map_load().
func AccelMapLoad(fileName string) {
	cstr := C.CString(fileName)
	defer C.free(unsafe.Pointer(cstr))

	C.gtk_accel_map_load((*C.gchar)(cstr))
}

// AccelMapSave is a wrapper around gtk_accel_map_save().
func AccelMapSave(fileName string) {
	cstr := C.CString(fileName)
	defer C.free(unsafe.Pointer(cstr))

	C.gtk_accel_map_save((*C.gchar)(cstr))
}

// AccelMapLoadFD is a wrapper around gtk_accel_map_load_fd().
func AccelMapLoadFD(fd int) {
	C.gtk_accel_map_load_fd(C.gint(fd))
}

// AccelMapSaveFD is a wrapper around gtk_accel_map_save_fd().
func AccelMapSaveFD(fd int) {
	C.gtk_accel_map_save_fd(C.gint(fd))
}

// AccelMapAddFilter is a wrapper around gtk_accel_map_add_filter().
func AccelMapAddFilter(filter string) {
	cstr := C.CString(filter)
	defer C.free(unsafe.Pointer(cstr))

	C.gtk_accel_map_add_filter((*C.gchar)(cstr))
}

// AccelMapGet is a wrapper around gtk_accel_map_get().
func AccelMapGet() *AccelMap {
	c := C.gtk_accel_map_get()
	if c == nil {
		return nil
	}
	return wrapAccelMap(glib.Take(unsafe.Pointer(c)))
}

// AccelMapLockPath is a wrapper around gtk_accel_map_lock_path().
func AccelMapLockPath(path string) {
	cstr := C.CString(path)
	defer C.free(unsafe.Pointer(cstr))

	C.gtk_accel_map_lock_path((*C.gchar)(cstr))
}

// AccelMapUnlockPath is a wrapper around gtk_accel_map_unlock_path().
func AccelMapUnlockPath(path string) {
	cstr := C.CString(path)
	defer C.free(unsafe.Pointer(cstr))

	C.gtk_accel_map_unlock_path((*C.gchar)(cstr))
}

// SetAccelGroup is a wrapper around gtk_menu_set_accel_group().
func (v *Menu) SetAccelGroup(accelGroup *AccelGroup) {
	C.gtk_menu_set_accel_group(v.native(), accelGroup.native())
}

// GetAccelGroup is a wrapper around gtk_menu_get_accel_group().
func (v *Menu) GetAccelGroup() *AccelGroup {
	c := C.gtk_menu_get_accel_group(v.native())
	if c == nil {
		return nil
	}
	return wrapAccelGroup(glib.Take(unsafe.Pointer(c)))
}

// SetAccelPath is a wrapper around gtk_menu_set_accel_path().
func (v *Menu) SetAccelPath(path string) {
	cstr := C.CString(path)
	defer C.free(unsafe.Pointer(cstr))

	C.gtk_menu_set_accel_path(v.native(), (*C.gchar)(cstr))
}

// GetAccelPath is a wrapper around gtk_menu_get_accel_path().
func (v *Menu) GetAccelPath() string {
	c := C.gtk_menu_get_accel_path(v.native())
	return C.GoString((*C.char)(c))
}

// SetAccelPath is a wrapper around gtk_menu_item_set_accel_path().
func (v *MenuItem) SetAccelPath(path string) {
	cstr := C.CString(path)
	defer C.free(unsafe.Pointer(cstr))

	C.gtk_menu_item_set_accel_path(v.native(), (*C.gchar)(cstr))
}

// GetAccelPath is a wrapper around gtk_menu_item_get_accel_path().
func (v *MenuItem) GetAccelPath() string {
	c := C.gtk_menu_item_get_accel_path(v.native())
	return C.GoString((*C.char)(c))
}

// AddAccelerator is a wrapper around gtk_widget_add_accelerator().
func (v *Widget) AddAccelerator(signal string, group *AccelGroup, key uint, mods gdk.ModifierType, flags AccelFlags) {
	csignal := (*C.gchar)(C.CString(signal))
	defer C.free(unsafe.Pointer(csignal))

	C.gtk_widget_add_accelerator(v.native(),
		csignal,
		group.native(),
		C.guint(key),
		C.GdkModifierType(mods),
		C.GtkAccelFlags(flags))
}

// RemoveAccelerator is a wrapper around gtk_widget_remove_accelerator().
func (v *Widget) RemoveAccelerator(group *AccelGroup, key uint, mods gdk.ModifierType) bool {
	return gobool(C.gtk_widget_remove_accelerator(v.native(),
		group.native(),
		C.guint(key),
		C.GdkModifierType(mods)))
}

// SetAccelPath is a wrapper around gtk_widget_set_accel_path().
func (v *Widget) SetAccelPath(path string, group *AccelGroup) {
	cstr := (*C.gchar)(C.CString(path))
	defer C.free(unsafe.Pointer(cstr))

	C.gtk_widget_set_accel_path(v.native(), cstr, group.native())
}

// CanActivateAccel is a wrapper around gtk_widget_can_activate_accel().
func (v *Widget) CanActivateAccel(signalId uint) bool {
	return gobool(C.gtk_widget_can_activate_accel(v.native(), C.guint(signalId)))
}

// AddAccelGroup() is a wrapper around gtk_window_add_accel_group().
func (v *Window) AddAccelGroup(accelGroup *AccelGroup) {
	C.gtk_window_add_accel_group(v.native(), accelGroup.native())
}

// RemoveAccelGroup() is a wrapper around gtk_window_add_accel_group().
func (v *Window) RemoveAccelGroup(accelGroup *AccelGroup) {
	C.gtk_window_remove_accel_group(v.native(), accelGroup.native())
}

// These three functions are for system level access - thus not as high priority to implement
// TODO: void 	gtk_accelerator_parse_with_keycode ()
// TODO: gchar * 	gtk_accelerator_name_with_keycode ()
// TODO: gchar * 	gtk_accelerator_get_label_with_keycode ()

// TODO: GtkAccelKey * 	gtk_accel_group_find ()   - this function uses a function type - I don't know how to represent it in cgo
// TODO: gtk_accel_map_foreach_unfiltered  - can't be done without a function type
// TODO: gtk_accel_map_foreach  - can't be done without a function type

// TODO: gtk_accel_map_load_scanner
// TODO: gtk_widget_list_accel_closures
