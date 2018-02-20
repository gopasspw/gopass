// Same copyright and license as the rest of the files in this project
// This file contains accelerator related functions and structures

package gtk

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

/*
 * GtkTreeViewColumn
 */

// TreeViewColumns is a representation of GTK's GtkTreeViewColumn.
type TreeViewColumn struct {
	glib.InitiallyUnowned
}

// native returns a pointer to the underlying GtkTreeViewColumn.
func (v *TreeViewColumn) native() *C.GtkTreeViewColumn {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkTreeViewColumn(p)
}

func marshalTreeViewColumn(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapTreeViewColumn(obj), nil
}

func wrapTreeViewColumn(obj *glib.Object) *TreeViewColumn {
	return &TreeViewColumn{glib.InitiallyUnowned{obj}}
}

// TreeViewColumnNew() is a wrapper around gtk_tree_view_column_new().
func TreeViewColumnNew() (*TreeViewColumn, error) {
	c := C.gtk_tree_view_column_new()
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapTreeViewColumn(glib.Take(unsafe.Pointer(c))), nil
}

// TreeViewColumnNewWithAttribute() is a wrapper around
// gtk_tree_view_column_new_with_attributes() that only sets one
// attribute for one column.
func TreeViewColumnNewWithAttribute(title string, renderer ICellRenderer, attribute string, column int) (*TreeViewColumn, error) {
	t_cstr := C.CString(title)
	defer C.free(unsafe.Pointer(t_cstr))
	a_cstr := C.CString(attribute)
	defer C.free(unsafe.Pointer(a_cstr))
	c := C._gtk_tree_view_column_new_with_attributes_one((*C.gchar)(t_cstr),
		renderer.toCellRenderer(), (*C.gchar)(a_cstr), C.gint(column))
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapTreeViewColumn(glib.Take(unsafe.Pointer(c))), nil
}

// AddAttribute() is a wrapper around gtk_tree_view_column_add_attribute().
func (v *TreeViewColumn) AddAttribute(renderer ICellRenderer, attribute string, column int) {
	cstr := C.CString(attribute)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_tree_view_column_add_attribute(v.native(),
		renderer.toCellRenderer(), (*C.gchar)(cstr), C.gint(column))
}

// SetExpand() is a wrapper around gtk_tree_view_column_set_expand().
func (v *TreeViewColumn) SetExpand(expand bool) {
	C.gtk_tree_view_column_set_expand(v.native(), gbool(expand))
}

// GetExpand() is a wrapper around gtk_tree_view_column_get_expand().
func (v *TreeViewColumn) GetExpand() bool {
	c := C.gtk_tree_view_column_get_expand(v.native())
	return gobool(c)
}

// SetMinWidth() is a wrapper around gtk_tree_view_column_set_min_width().
func (v *TreeViewColumn) SetMinWidth(minWidth int) {
	C.gtk_tree_view_column_set_min_width(v.native(), C.gint(minWidth))
}

// GetMinWidth() is a wrapper around gtk_tree_view_column_get_min_width().
func (v *TreeViewColumn) GetMinWidth() int {
	c := C.gtk_tree_view_column_get_min_width(v.native())
	return int(c)
}

// PackStart() is a wrapper around gtk_tree_view_column_pack_start().
func (v *TreeViewColumn) PackStart(cell *CellRenderer, expand bool) {
	C.gtk_tree_view_column_pack_start(v.native(), cell.native(), gbool(expand))
}

// PackEnd() is a wrapper around gtk_tree_view_column_pack_end().
func (v *TreeViewColumn) PackEnd(cell *CellRenderer, expand bool) {
	C.gtk_tree_view_column_pack_end(v.native(), cell.native(), gbool(expand))
}

// Clear() is a wrapper around gtk_tree_view_column_clear().
func (v *TreeViewColumn) Clear() {
	C.gtk_tree_view_column_clear(v.native())
}

// ClearAttributes() is a wrapper around gtk_tree_view_column_clear_attributes().
func (v *TreeViewColumn) ClearAttributes(cell *CellRenderer) {
	C.gtk_tree_view_column_clear_attributes(v.native(), cell.native())
}

// SetSpacing() is a wrapper around gtk_tree_view_column_set_spacing().
func (v *TreeViewColumn) SetSpacing(spacing int) {
	C.gtk_tree_view_column_set_spacing(v.native(), C.gint(spacing))
}

// GetSpacing() is a wrapper around gtk_tree_view_column_get_spacing().
func (v *TreeViewColumn) GetSpacing() int {
	return int(C.gtk_tree_view_column_get_spacing(v.native()))
}

// SetVisible() is a wrapper around gtk_tree_view_column_set_visible().
func (v *TreeViewColumn) SetVisible(visible bool) {
	C.gtk_tree_view_column_set_visible(v.native(), gbool(visible))
}

// GetVisible() is a wrapper around gtk_tree_view_column_get_visible().
func (v *TreeViewColumn) GetVisible() bool {
	return gobool(C.gtk_tree_view_column_get_visible(v.native()))
}

// SetResizable() is a wrapper around gtk_tree_view_column_set_resizable().
func (v *TreeViewColumn) SetResizable(resizable bool) {
	C.gtk_tree_view_column_set_resizable(v.native(), gbool(resizable))
}

// GetResizable() is a wrapper around gtk_tree_view_column_get_resizable().
func (v *TreeViewColumn) GetResizable() bool {
	return gobool(C.gtk_tree_view_column_get_resizable(v.native()))
}

// GetWidth() is a wrapper around gtk_tree_view_column_get_width().
func (v *TreeViewColumn) GetWidth() int {
	return int(C.gtk_tree_view_column_get_width(v.native()))
}

// SetFixedWidth() is a wrapper around gtk_tree_view_column_set_fixed_width().
func (v *TreeViewColumn) SetFixedWidth(w int) {
	C.gtk_tree_view_column_set_fixed_width(v.native(), C.gint(w))
}

// GetFixedWidth() is a wrapper around gtk_tree_view_column_get_fixed_width().
func (v *TreeViewColumn) GetFixedWidth() int {
	return int(C.gtk_tree_view_column_get_fixed_width(v.native()))
}

// SetMaxWidth() is a wrapper around gtk_tree_view_column_set_max_width().
func (v *TreeViewColumn) SetMaxWidth(w int) {
	C.gtk_tree_view_column_set_max_width(v.native(), C.gint(w))
}

// GetMaxWidth() is a wrapper around gtk_tree_view_column_get_max_width().
func (v *TreeViewColumn) GetMaxWidth() int {
	return int(C.gtk_tree_view_column_get_max_width(v.native()))
}

// Clicked() is a wrapper around gtk_tree_view_column_clicked().
func (v *TreeViewColumn) Clicked() {
	C.gtk_tree_view_column_clicked(v.native())
}

// SetTitle() is a wrapper around gtk_tree_view_column_set_title().
func (v *TreeViewColumn) SetTitle(t string) {
	cstr := (*C.gchar)(C.CString(t))
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_tree_view_column_set_title(v.native(), cstr)
}

// GetTitle() is a wrapper around gtk_tree_view_column_get_title().
func (v *TreeViewColumn) GetTitle() string {
	return C.GoString((*C.char)(C.gtk_tree_view_column_get_title(v.native())))
}

// SetClickable() is a wrapper around gtk_tree_view_column_set_clickable().
func (v *TreeViewColumn) SetClickable(clickable bool) {
	C.gtk_tree_view_column_set_clickable(v.native(), gbool(clickable))
}

// GetClickable() is a wrapper around gtk_tree_view_column_get_clickable().
func (v *TreeViewColumn) GetClickable() bool {
	return gobool(C.gtk_tree_view_column_get_clickable(v.native()))
}

// SetReorderable() is a wrapper around gtk_tree_view_column_set_reorderable().
func (v *TreeViewColumn) SetReorderable(reorderable bool) {
	C.gtk_tree_view_column_set_reorderable(v.native(), gbool(reorderable))
}

// GetReorderable() is a wrapper around gtk_tree_view_column_get_reorderable().
func (v *TreeViewColumn) GetReorderable() bool {
	return gobool(C.gtk_tree_view_column_get_reorderable(v.native()))
}

// SetSortIndicator() is a wrapper around gtk_tree_view_column_set_sort_indicator().
func (v *TreeViewColumn) SetSortIndicator(reorderable bool) {
	C.gtk_tree_view_column_set_sort_indicator(v.native(), gbool(reorderable))
}

// GetSortIndicator() is a wrapper around gtk_tree_view_column_get_sort_indicator().
func (v *TreeViewColumn) GetSortIndicator() bool {
	return gobool(C.gtk_tree_view_column_get_sort_indicator(v.native()))
}

// SetSortColumnID() is a wrapper around gtk_tree_view_column_set_sort_column_id().
func (v *TreeViewColumn) SetSortColumnID(w int) {
	C.gtk_tree_view_column_set_sort_column_id(v.native(), C.gint(w))
}

// GetSortColumnID() is a wrapper around gtk_tree_view_column_get_sort_column_id().
func (v *TreeViewColumn) GetSortColumnID() int {
	return int(C.gtk_tree_view_column_get_sort_column_id(v.native()))
}

// CellIsVisible() is a wrapper around gtk_tree_view_column_cell_is_visible().
func (v *TreeViewColumn) CellIsVisible() bool {
	return gobool(C.gtk_tree_view_column_cell_is_visible(v.native()))
}

// FocusCell() is a wrapper around gtk_tree_view_column_focus_cell().
func (v *TreeViewColumn) FocusCell(cell *CellRenderer) {
	C.gtk_tree_view_column_focus_cell(v.native(), cell.native())
}

// QueueResize() is a wrapper around gtk_tree_view_column_queue_resize().
func (v *TreeViewColumn) QueueResize() {
	C.gtk_tree_view_column_queue_resize(v.native())
}

// GetXOffset() is a wrapper around gtk_tree_view_column_get_x_offset().
func (v *TreeViewColumn) GetXOffset() int {
	return int(C.gtk_tree_view_column_get_x_offset(v.native()))
}

// GtkTreeViewColumn * 	gtk_tree_view_column_new_with_area ()
// void 	gtk_tree_view_column_set_attributes ()
// void 	gtk_tree_view_column_set_cell_data_func ()

type TreeViewColumnSizing int

const (
	TREE_VIEW_COLUMN_GROW_ONLY int = C.GTK_TREE_VIEW_COLUMN_GROW_ONLY
	TREE_VIEW_COLUMN_AUTOSIZE      = C.GTK_TREE_VIEW_COLUMN_AUTOSIZE
	TREE_VIEW_COLUMN_FIXED         = C.GTK_TREE_VIEW_COLUMN_FIXED
)

// void 	gtk_tree_view_column_set_sizing ()
func (v *TreeViewColumn) SetSizing(sizing TreeViewColumnSizing) {
	C.gtk_tree_view_column_set_sizing(v.native(), C.GtkTreeViewColumnSizing(sizing))
}

// GtkTreeViewColumnSizing 	gtk_tree_view_column_get_sizing ()
func (v *TreeViewColumn) GetSizing() TreeViewColumnSizing {
	return TreeViewColumnSizing(C.gtk_tree_view_column_get_sizing(v.native()))
}

// void 	gtk_tree_view_column_set_widget ()
// GtkWidget * 	gtk_tree_view_column_get_widget ()
// GtkWidget * 	gtk_tree_view_column_get_button ()
// void 	gtk_tree_view_column_set_alignment ()
// gfloat 	gtk_tree_view_column_get_alignment ()
// void 	gtk_tree_view_column_set_sort_order ()
// GtkSortType 	gtk_tree_view_column_get_sort_order ()
// void 	gtk_tree_view_column_cell_set_cell_data ()
// void 	gtk_tree_view_column_cell_get_size ()
// gboolean 	gtk_tree_view_column_cell_get_position ()
// GtkWidget * 	gtk_tree_view_column_get_tree_view ()
