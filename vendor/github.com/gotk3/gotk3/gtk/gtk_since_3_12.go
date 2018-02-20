// +build !gtk_3_6,!gtk_3_8,!gtk_3_10
// not use this: go build -tags gtk_3_8'. Otherwise, if no build tags are used, GTK 3.10

package gtk

// #cgo pkg-config: gtk+-3.0
// #include <gtk/gtk.h>
// #include "gtk_since_3_12.go.h"
import "C"

import (
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

func init() {
	tm := []glib.TypeMarshaler{
		// Objects/Interfaces
		{glib.Type(C.gtk_flow_box_get_type()), marshalFlowBox},
		{glib.Type(C.gtk_flow_box_child_get_type()), marshalFlowBoxChild},
	}
	glib.RegisterGValueMarshalers(tm)

	WrapMap["GtkFlowBox"] = wrapFlowBox
	WrapMap["GtkFlowBoxChild"] = wrapFlowBoxChild
}

// SetPopover is a wrapper around gtk_menu_button_set_popover().
func (v *MenuButton) SetPopover(popover *Popover) {
	C.gtk_menu_button_set_popover(v.native(), popover.toWidget())
}

// GetPopover is a wrapper around gtk_menu_button_get_popover().
func (v *MenuButton) GetPopover() *Popover {
	c := C.gtk_menu_button_get_popover(v.native())
	if c == nil {
		return nil
	}
	return wrapPopover(glib.Take(unsafe.Pointer(c)))
}

/*
 * FlowBox
 */
type FlowBox struct {
	Container
}

func (fb *FlowBox) native() *C.GtkFlowBox {
	if fb == nil || fb.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(fb.GObject)
	return C.toGtkFlowBox(p)
}

func marshalFlowBox(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapFlowBox(obj), nil
}

func wrapFlowBox(obj *glib.Object) *FlowBox {
	return &FlowBox{Container{Widget{glib.InitiallyUnowned{obj}}}}
}

// FlowBoxNew is a wrapper around gtk_flow_box_new()
func FlowBoxNew() (*FlowBox, error) {
	c := C.gtk_flow_box_new()
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapFlowBox(glib.Take(unsafe.Pointer(c))), nil
}

// Insert is a wrapper around gtk_flow_box_insert()
func (fb *FlowBox) Insert(widget IWidget, position int) {
	C.gtk_flow_box_insert(fb.native(), widget.toWidget(), C.gint(position))
}

// GetChildAtIndex is a wrapper around gtk_flow_box_get_child_at_index()
func (fb *FlowBox) GetChildAtIndex(idx int) *FlowBoxChild {
	c := C.gtk_flow_box_get_child_at_index(fb.native(), C.gint(idx))
	if c == nil {
		return nil
	}
	return wrapFlowBoxChild(glib.Take(unsafe.Pointer(c)))
}

// TODO 3.22.6 gtk_flow_box_get_child_at_pos()

// SetHAdjustment is a wrapper around gtk_flow_box_set_hadjustment()
func (fb *FlowBox) SetHAdjustment(adjustment *Adjustment) {
	C.gtk_flow_box_set_hadjustment(fb.native(), adjustment.native())
}

// SetVAdjustment is a wrapper around gtk_flow_box_set_vadjustment()
func (fb *FlowBox) SetVAdjustment(adjustment *Adjustment) {
	C.gtk_flow_box_set_vadjustment(fb.native(), adjustment.native())
}

// SetHomogeneous is a wrapper around gtk_flow_box_set_homogeneous()
func (fb *FlowBox) SetHomogeneous(homogeneous bool) {
	C.gtk_flow_box_set_homogeneous(fb.native(), gbool(homogeneous))
}

// GetHomogeneous is a wrapper around gtk_flow_box_get_homogeneous()
func (fb *FlowBox) GetHomogeneous() bool {
	c := C.gtk_flow_box_get_homogeneous(fb.native())
	return gobool(c)
}

// SetRowSpacing is a wrapper around gtk_flow_box_set_row_spacing()
func (fb *FlowBox) SetRowSpacing(spacing uint) {
	C.gtk_flow_box_set_row_spacing(fb.native(), C.guint(spacing))
}

// GetRowSpacing is a wrapper around gtk_flow_box_get_row_spacing()
func (fb *FlowBox) GetRowSpacing() uint {
	c := C.gtk_flow_box_get_row_spacing(fb.native())
	return uint(c)
}

// SetColumnSpacing is a wrapper around gtk_flow_box_set_column_spacing()
func (fb *FlowBox) SetColumnSpacing(spacing uint) {
	C.gtk_flow_box_set_column_spacing(fb.native(), C.guint(spacing))
}

// GetColumnSpacing is a wrapper around gtk_flow_box_get_column_spacing()
func (fb *FlowBox) GetColumnSpacing() uint {
	c := C.gtk_flow_box_get_column_spacing(fb.native())
	return uint(c)
}

// SetMinChildrenPerLine is a wrapper around gtk_flow_box_set_min_children_per_line()
func (fb *FlowBox) SetMinChildrenPerLine(n_children uint) {
	C.gtk_flow_box_set_min_children_per_line(fb.native(), C.guint(n_children))
}

// GetMinChildrenPerLine is a wrapper around gtk_flow_box_get_min_children_per_line()
func (fb *FlowBox) GetMinChildrenPerLine() uint {
	c := C.gtk_flow_box_get_min_children_per_line(fb.native())
	return uint(c)
}

// SetMaxChildrenPerLine is a wrapper around gtk_flow_box_set_max_children_per_line()
func (fb *FlowBox) SetMaxChildrenPerLine(n_children uint) {
	C.gtk_flow_box_set_max_children_per_line(fb.native(), C.guint(n_children))
}

// GetMaxChildrenPerLine is a wrapper around gtk_flow_box_get_max_children_per_line()
func (fb *FlowBox) GetMaxChildrenPerLine() uint {
	c := C.gtk_flow_box_get_max_children_per_line(fb.native())
	return uint(c)
}

// SetActivateOnSingleClick is a wrapper around gtk_flow_box_set_activate_on_single_click()
func (fb *FlowBox) SetActivateOnSingleClick(single bool) {
	C.gtk_flow_box_set_activate_on_single_click(fb.native(), gbool(single))
}

// GetActivateOnSingleClick gtk_flow_box_get_activate_on_single_click()
func (fb *FlowBox) GetActivateOnSingleClick() bool {
	c := C.gtk_flow_box_get_activate_on_single_click(fb.native())
	return gobool(c)
}

// TODO: gtk_flow_box_selected_foreach()

// GetSelectedChildren is a wrapper around gtk_flow_box_get_selected_children()
func (fb *FlowBox) GetSelectedChildren() (rv []*FlowBoxChild) {
	c := C.gtk_flow_box_get_selected_children(fb.native())
	if c == nil {
		return
	}
	list := glib.WrapList(uintptr(unsafe.Pointer(c)))
	for l := list; l != nil; l = l.Next() {
		o := wrapFlowBoxChild(glib.Take(l.Data().(unsafe.Pointer)))
		rv = append(rv, o)
	}
	// We got a transfer container, so we must free the list.
	list.Free()

	return
}

// SelectChild is a wrapper around gtk_flow_box_select_child()
func (fb *FlowBox) SelectChild(child *FlowBoxChild) {
	C.gtk_flow_box_select_child(fb.native(), child.native())
}

// UnselectChild is a wrapper around gtk_flow_box_unselect_child()
func (fb *FlowBox) UnselectChild(child *FlowBoxChild) {
	C.gtk_flow_box_unselect_child(fb.native(), child.native())
}

// SelectAll is a wrapper around gtk_flow_box_select_all()
func (fb *FlowBox) SelectAll() {
	C.gtk_flow_box_select_all(fb.native())
}

// UnselectAll is a wrapper around gtk_flow_box_unselect_all()
func (fb *FlowBox) UnselectAll() {
	C.gtk_flow_box_unselect_all(fb.native())
}

// SetSelectionMode is a wrapper around gtk_flow_box_set_selection_mode()
func (fb *FlowBox) SetSelectionMode(mode SelectionMode) {
	C.gtk_flow_box_set_selection_mode(fb.native(), C.GtkSelectionMode(mode))
}

// GetSelectionMode is a wrapper around gtk_flow_box_get_selection_mode()
func (fb *FlowBox) GetSelectionMode() SelectionMode {
	c := C.gtk_flow_box_get_selection_mode(fb.native())
	return SelectionMode(c)
}

// TODO gtk_flow_box_set_filter_func()
// TODO gtk_flow_box_invalidate_filter()
// TODO gtk_flow_box_set_sort_func()
// TODO gtk_flow_box_invalidate_sort()
// TODO 3.18 gtk_flow_box_bind_model()

/*
 * FlowBoxChild
 */
type FlowBoxChild struct {
	Bin
}

func (fbc *FlowBoxChild) native() *C.GtkFlowBoxChild {
	if fbc == nil || fbc.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(fbc.GObject)
	return C.toGtkFlowBoxChild(p)
}

func marshalFlowBoxChild(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapFlowBoxChild(obj), nil
}

func wrapFlowBoxChild(obj *glib.Object) *FlowBoxChild {
	return &FlowBoxChild{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}}
}

// FlowBoxChildNew is a wrapper around gtk_flow_box_child_new()
func FlowBoxChildNew() (*FlowBoxChild, error) {
	c := C.gtk_flow_box_child_new()
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapFlowBoxChild(glib.Take(unsafe.Pointer(c))), nil
}

// GetIndex is a wrapper around gtk_flow_box_child_get_index()
func (fbc *FlowBoxChild) GetIndex() int {
	c := C.gtk_flow_box_child_get_index(fbc.native())
	return int(c)
}

// IsSelected is a wrapper around gtk_flow_box_child_is_selected()
func (fbc *FlowBoxChild) IsSelected() bool {
	c := C.gtk_flow_box_child_is_selected(fbc.native())
	return gobool(c)
}

// Changed is a wrapper around gtk_flow_box_child_changed()
func (fbc *FlowBoxChild) Changed() {
	C.gtk_flow_box_child_changed(fbc.native())
}
