package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <gio/gio.h>
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"
import "unsafe"

// MenuModel is a representation of GMenuModel.
type MenuModel struct {
	*Object
}

// native() returns a pointer to the underlying GMenuModel.
func (v *MenuModel) native() *C.GMenuModel {
	if v == nil || v.GObject == nil {
		return nil
	}
	return C.toGMenuModel(unsafe.Pointer(v.GObject))
}

func (v *MenuModel) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalMenuModel(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return wrapMenuModel(wrapObject(unsafe.Pointer(c))), nil
}

func wrapMenuModel(obj *Object) *MenuModel {
	return &MenuModel{obj}
}

// IsMutable is a wrapper around g_menu_model_is_mutable().
func (v *MenuModel) IsMutable() bool {
	return gobool(C.g_menu_model_is_mutable(v.native()))
}

// GetNItems is a wrapper around g_menu_model_get_n_items().
func (v *MenuModel) GetNItems() int {
	return int(C.g_menu_model_get_n_items(v.native()))
}

// GetItemLink is a wrapper around g_menu_model_get_item_link().
func (v *MenuModel) GetItemLink(index int, link string) *MenuModel {
	cstr := (*C.gchar)(C.CString(link))
	defer C.free(unsafe.Pointer(cstr))
	c := C.g_menu_model_get_item_link(v.native(), C.gint(index), cstr)
	if c == nil {
		return nil
	}
	return wrapMenuModel(wrapObject(unsafe.Pointer(c)))
}

// ItemsChanged is a wrapper around g_menu_model_items_changed().
func (v *MenuModel) ItemsChanged(position, removed, added int) {
	C.g_menu_model_items_changed(v.native(), C.gint(position), C.gint(removed), C.gint(added))
}

// GVariant * 	g_menu_model_get_item_attribute_value ()
// gboolean 	g_menu_model_get_item_attribute ()
// GMenuAttributeIter * 	g_menu_model_iterate_item_attributes ()
// GMenuLinkIter * 	g_menu_model_iterate_item_links ()

// Menu is a representation of GMenu.
type Menu struct {
	MenuModel
}

// native() returns a pointer to the underlying GMenu.
func (m *Menu) native() *C.GMenu {
	if m == nil || m.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(m.GObject)
	return C.toGMenu(p)
}

func marshalMenu(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return wrapMenu(wrapObject(unsafe.Pointer(c))), nil
}

func wrapMenu(obj *Object) *Menu {
	return &Menu{MenuModel{obj}}
}

// MenuNew is a wrapper around g_menu_new().
func MenuNew() *Menu {
	c := C.g_menu_new()
	if c == nil {
		return nil
	}
	return wrapMenu(wrapObject(unsafe.Pointer(c)))
}

// Freeze is a wrapper around g_menu_freeze().
func (v *Menu) Freeze() {
	C.g_menu_freeze(v.native())
}

// Insert is a wrapper around g_menu_insert().
func (v *Menu) Insert(position int, label, detailed_action string) {
	cstr1 := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer(cstr1))

	cstr2 := (*C.gchar)(C.CString(detailed_action))
	defer C.free(unsafe.Pointer(cstr2))

	C.g_menu_insert(v.native(), C.gint(position), cstr1, cstr2)
}

// Prepend is a wrapper around g_menu_prepend().
func (v *Menu) Prepend(label, detailed_action string) {
	cstr1 := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer(cstr1))

	cstr2 := (*C.gchar)(C.CString(detailed_action))
	defer C.free(unsafe.Pointer(cstr2))

	C.g_menu_prepend(v.native(), cstr1, cstr2)
}

// Append is a wrapper around g_menu_append().
func (v *Menu) Append(label, detailed_action string) {
	cstr1 := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer(cstr1))

	cstr2 := (*C.gchar)(C.CString(detailed_action))
	defer C.free(unsafe.Pointer(cstr2))

	C.g_menu_append(v.native(), cstr1, cstr2)
}

// InsertItem is a wrapper around g_menu_insert_item().
func (v *Menu) InsertItem(position int, item *MenuItem) {
	C.g_menu_insert_item(v.native(), C.gint(position), item.native())
}

// AppendItem is a wrapper around g_menu_append_item().
func (v *Menu) AppendItem(item *MenuItem) {
	C.g_menu_append_item(v.native(), item.native())
}

// PrependItem is a wrapper around g_menu_prepend_item().
func (v *Menu) PrependItem(item *MenuItem) {
	C.g_menu_prepend_item(v.native(), item.native())
}

// InsertSection is a wrapper around g_menu_insert_section().
func (v *Menu) InsertSection(position int, label string, section *MenuModel) {
	cstr1 := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer(cstr1))

	C.g_menu_insert_section(v.native(), C.gint(position), cstr1, section.native())
}

// PrependSection is a wrapper around g_menu_prepend_section().
func (v *Menu) PrependSection(label string, section *MenuModel) {
	cstr1 := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer(cstr1))

	C.g_menu_prepend_section(v.native(), cstr1, section.native())
}

// AppendSection is a wrapper around g_menu_append_section().
func (v *Menu) AppendSection(label string, section *MenuModel) {
	cstr1 := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer(cstr1))

	C.g_menu_append_section(v.native(), cstr1, section.native())
}

// InsertSubmenu is a wrapper around g_menu_insert_submenu().
func (v *Menu) InsertSubmenu(position int, label string, submenu *MenuModel) {
	cstr1 := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer(cstr1))

	C.g_menu_insert_submenu(v.native(), C.gint(position), cstr1, submenu.native())
}

// PrependSubmenu is a wrapper around g_menu_prepend_submenu().
func (v *Menu) PrependSubmenu(label string, submenu *MenuModel) {
	cstr1 := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer(cstr1))

	C.g_menu_prepend_submenu(v.native(), cstr1, submenu.native())
}

// AppendSubmenu is a wrapper around g_menu_append_submenu().
func (v *Menu) AppendSubmenu(label string, submenu *MenuModel) {
	cstr1 := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer(cstr1))

	C.g_menu_append_submenu(v.native(), cstr1, submenu.native())
}

// Remove is a wrapper around g_menu_remove().
func (v *Menu) Remove(position int) {
	C.g_menu_remove(v.native(), C.gint(position))
}

// RemoveAll is a wrapper around g_menu_remove_all().
func (v *Menu) RemoveAll() {
	C.g_menu_remove_all(v.native())
}

// MenuItem is a representation of GMenuItem.
type MenuItem struct {
	*Object
}

// native() returns a pointer to the underlying GMenuItem.
func (m *MenuItem) native() *C.GMenuItem {
	if m == nil || m.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(m.GObject)
	return C.toGMenuItem(p)
}

func marshalMenuItem(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return wrapMenuItem(wrapObject(unsafe.Pointer(c))), nil
}

func wrapMenuItem(obj *Object) *MenuItem {
	return &MenuItem{obj}
}

// MenuItemNew is a wrapper around g_menu_item_new().
func MenuItemNew(label, detailed_action string) *MenuItem {
	cstr1 := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer(cstr1))

	cstr2 := (*C.gchar)(C.CString(detailed_action))
	defer C.free(unsafe.Pointer(cstr2))

	c := C.g_menu_item_new(cstr1, cstr2)
	if c == nil {
		return nil
	}
	return wrapMenuItem(wrapObject(unsafe.Pointer(c)))
}

// MenuItemNewSection is a wrapper around g_menu_item_new_section().
func MenuItemNewSection(label string, section *MenuModel) *MenuItem {
	cstr1 := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer(cstr1))

	c := C.g_menu_item_new_section(cstr1, section.native())
	if c == nil {
		return nil
	}
	return wrapMenuItem(wrapObject(unsafe.Pointer(c)))
}

// MenuItemNewSubmenu is a wrapper around g_menu_item_new_submenu().
func MenuItemNewSubmenu(label string, submenu *MenuModel) *MenuItem {
	cstr1 := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer(cstr1))

	c := C.g_menu_item_new_submenu(cstr1, submenu.native())
	if c == nil {
		return nil
	}
	return wrapMenuItem(wrapObject(unsafe.Pointer(c)))
}

// MenuItemNewFromModel is a wrapper around g_menu_item_new_from_model().
func MenuItemNewFromModel(model *MenuModel, index int) *MenuItem {
	c := C.g_menu_item_new_from_model(model.native(), C.gint(index))
	if c == nil {
		return nil
	}
	return wrapMenuItem(wrapObject(unsafe.Pointer(c)))
}

//SetLabel is a wrapper around g_menu_item_set_label().
func (v *MenuItem) SetLabel(label string) {
	cstr1 := (*C.gchar)(C.CString(label))
	defer C.free(unsafe.Pointer(cstr1))

	C.g_menu_item_set_label(v.native(), cstr1)
}

//SetDetailedAction is a wrapper around g_menu_item_set_detailed_action().
func (v *MenuItem) SetDetailedAction(act string) {
	cstr1 := (*C.gchar)(C.CString(act))
	defer C.free(unsafe.Pointer(cstr1))

	C.g_menu_item_set_detailed_action(v.native(), cstr1)
}

//SetSection is a wrapper around g_menu_item_set_section().
func (v *MenuItem) SetSection(section *MenuModel) {
	C.g_menu_item_set_section(v.native(), section.native())
}

//SetSubmenu is a wrapper around g_menu_item_set_submenu().
func (v *MenuItem) SetSubmenu(submenu *MenuModel) {
	C.g_menu_item_set_submenu(v.native(), submenu.native())
}

//GetLink is a wrapper around g_menu_item_get_link().
func (v *MenuItem) GetLink(link string) *MenuModel {
	cstr1 := (*C.gchar)(C.CString(link))
	defer C.free(unsafe.Pointer(cstr1))

	c := C.g_menu_item_get_link(v.native(), cstr1)
	if c == nil {
		return nil
	}
	return wrapMenuModel(wrapObject(unsafe.Pointer(c)))
}

//SetLink is a wrapper around g_menu_item_Set_link().
func (v *MenuItem) SetLink(link string, model *MenuModel) {
	cstr1 := (*C.gchar)(C.CString(link))
	defer C.free(unsafe.Pointer(cstr1))

	C.g_menu_item_set_link(v.native(), cstr1, model.native())
}

// void 	g_menu_item_set_action_and_target_value ()
// void 	g_menu_item_set_action_and_target ()
// GVariant * 	g_menu_item_get_attribute_value ()
// gboolean 	g_menu_item_get_attribute ()
// void 	g_menu_item_set_attribute_value ()
// void 	g_menu_item_set_attribute ()
