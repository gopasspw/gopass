// Same copyright and license as the rest of the files in this project
// This file contains accelerator related functions and structures

package gtk

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"runtime"
	"unsafe"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
)

/*
 * GtkTreeView
 */

// TreeView is a representation of GTK's GtkTreeView.
type TreeView struct {
	Container
}

// native returns a pointer to the underlying GtkTreeView.
func (v *TreeView) native() *C.GtkTreeView {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkTreeView(p)
}

func marshalTreeView(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapTreeView(obj), nil
}

func wrapTreeView(obj *glib.Object) *TreeView {
	return &TreeView{Container{Widget{glib.InitiallyUnowned{obj}}}}
}

func setupTreeView(c unsafe.Pointer) (*TreeView, error) {
	if c == nil {
		return nil, nilPtrErr
	}

	return wrapTreeView(glib.Take(c)), nil
}

// TreeViewNew is a wrapper around gtk_tree_view_new().
func TreeViewNew() (*TreeView, error) {
	return setupTreeView(unsafe.Pointer(C.gtk_tree_view_new()))
}

// TreeViewNewWithModel is a wrapper around gtk_tree_view_new_with_model().
func TreeViewNewWithModel(model ITreeModel) (*TreeView, error) {
	return setupTreeView(unsafe.Pointer(C.gtk_tree_view_new_with_model(model.toTreeModel())))
}

// GetModel is a wrapper around gtk_tree_view_get_model().
func (v *TreeView) GetModel() (*TreeModel, error) {
	c := C.gtk_tree_view_get_model(v.native())
	if c == nil {
		return nil, nil
	}
	return wrapTreeModel(glib.Take(unsafe.Pointer(c))), nil
}

// SetModel is a wrapper around gtk_tree_view_set_model().
func (v *TreeView) SetModel(model ITreeModel) {
	if model == nil {
		C.gtk_tree_view_set_model(v.native(), nil)
	} else {
		C.gtk_tree_view_set_model(v.native(), model.toTreeModel())
	}
}

// GetSelection is a wrapper around gtk_tree_view_get_selection().
func (v *TreeView) GetSelection() (*TreeSelection, error) {
	c := C.gtk_tree_view_get_selection(v.native())
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapTreeSelection(glib.Take(unsafe.Pointer(c))), nil
}

// AppendColumn is a wrapper around gtk_tree_view_append_column().
func (v *TreeView) AppendColumn(column *TreeViewColumn) int {
	c := C.gtk_tree_view_append_column(v.native(), column.native())
	return int(c)
}

// GetPathAtPos is a wrapper around gtk_tree_view_get_path_at_pos().
func (v *TreeView) GetPathAtPos(x, y int, path *TreePath, column *TreeViewColumn, cellX, cellY *int) bool {
	var ctp **C.GtkTreePath
	if path != nil {
		ctp = (**C.GtkTreePath)(unsafe.Pointer(&path.GtkTreePath))
	} else {
		ctp = nil
	}

	var pctvcol **C.GtkTreeViewColumn
	if column != nil {
		ctvcol := column.native()
		pctvcol = &ctvcol
	} else {
		pctvcol = nil
	}

	return 0 != C.gtk_tree_view_get_path_at_pos(
		v.native(),
		(C.gint)(x),
		(C.gint)(y),
		ctp,
		pctvcol,
		(*C.gint)(unsafe.Pointer(cellX)),
		(*C.gint)(unsafe.Pointer(cellY)))
}

// GetLevelIndentation is a wrapper around gtk_tree_view_get_level_indentation().
func (v *TreeView) GetLevelIndentation() int {
	return int(C.gtk_tree_view_get_level_indentation(v.native()))
}

// GetShowExpanders is a wrapper around gtk_tree_view_get_show_expanders().
func (v *TreeView) GetShowExpanders() bool {
	return gobool(C.gtk_tree_view_get_show_expanders(v.native()))
}

// SetLevelIndentation is a wrapper around gtk_tree_view_set_level_indentation().
func (v *TreeView) SetLevelIndentation(indent int) {
	C.gtk_tree_view_set_level_indentation(v.native(), C.gint(indent))
}

// SetShowExpanders is a wrapper around gtk_tree_view_set_show_expanders().
func (v *TreeView) SetShowExpanders(show bool) {
	C.gtk_tree_view_set_show_expanders(v.native(), gbool(show))
}

// GetHeadersVisible is a wrapper around gtk_tree_view_get_headers_visible().
func (v *TreeView) GetHeadersVisible() bool {
	return gobool(C.gtk_tree_view_get_headers_visible(v.native()))
}

// SetHeadersVisible is a wrapper around gtk_tree_view_set_headers_visible().
func (v *TreeView) SetHeadersVisible(show bool) {
	C.gtk_tree_view_set_headers_visible(v.native(), gbool(show))
}

// ColumnsAutosize is a wrapper around gtk_tree_view_columns_autosize().
func (v *TreeView) ColumnsAutosize() {
	C.gtk_tree_view_columns_autosize(v.native())
}

// GetHeadersClickable is a wrapper around gtk_tree_view_get_headers_clickable().
func (v *TreeView) GetHeadersClickable() bool {
	return gobool(C.gtk_tree_view_get_headers_clickable(v.native()))
}

// SetHeadersClickable is a wrapper around gtk_tree_view_set_headers_clickable().
func (v *TreeView) SetHeadersClickable(show bool) {
	C.gtk_tree_view_set_headers_clickable(v.native(), gbool(show))
}

// GetActivateOnSingleClick is a wrapper around gtk_tree_view_get_activate_on_single_click().
func (v *TreeView) GetActivateOnSingleClick() bool {
	return gobool(C.gtk_tree_view_get_activate_on_single_click(v.native()))
}

// SetActivateOnSingleClick is a wrapper around gtk_tree_view_set_activate_on_single_click().
func (v *TreeView) SetActivateOnSingleClick(show bool) {
	C.gtk_tree_view_set_activate_on_single_click(v.native(), gbool(show))
}

// RemoveColumn is a wrapper around gtk_tree_view_remove_column().
func (v *TreeView) RemoveColumn(column *TreeViewColumn) int {
	return int(C.gtk_tree_view_remove_column(v.native(), column.native()))
}

// InsertColumn is a wrapper around gtk_tree_view_insert_column().
func (v *TreeView) InsertColumn(column *TreeViewColumn, pos int) int {
	return int(C.gtk_tree_view_insert_column(v.native(), column.native(), C.gint(pos)))
}

// GetNColumns is a wrapper around gtk_tree_view_get_n_columns().
func (v *TreeView) GetNColumns() uint {
	return uint(C.gtk_tree_view_get_n_columns(v.native()))
}

// GetColumn is a wrapper around gtk_tree_view_get_column().
func (v *TreeView) GetColumn(n int) *TreeViewColumn {
	c := C.gtk_tree_view_get_column(v.native(), C.gint(n))
	if c == nil {
		return nil
	}
	return wrapTreeViewColumn(glib.Take(unsafe.Pointer(c)))
}

// MoveColumnAfter is a wrapper around gtk_tree_view_move_column_after().
func (v *TreeView) MoveColumnAfter(column *TreeViewColumn, baseColumn *TreeViewColumn) {
	C.gtk_tree_view_move_column_after(v.native(), column.native(), baseColumn.native())
}

// SetExpanderColumn is a wrapper around gtk_tree_view_set_expander_column().
func (v *TreeView) SetExpanderColumn(column *TreeViewColumn) {
	C.gtk_tree_view_set_expander_column(v.native(), column.native())
}

// GetExpanderColumn is a wrapper around gtk_tree_view_get_expander_column().
func (v *TreeView) GetExpanderColumn() *TreeViewColumn {
	c := C.gtk_tree_view_get_expander_column(v.native())
	if c == nil {
		return nil
	}
	return wrapTreeViewColumn(glib.Take(unsafe.Pointer(c)))
}

// ScrollToPoint is a wrapper around gtk_tree_view_scroll_to_point().
func (v *TreeView) ScrollToPoint(treeX, treeY int) {
	C.gtk_tree_view_scroll_to_point(v.native(), C.gint(treeX), C.gint(treeY))
}

// SetCursor is a wrapper around gtk_tree_view_set_cursor().
func (v *TreeView) SetCursor(path *TreePath, focusColumn *TreeViewColumn, startEditing bool) {
	C.gtk_tree_view_set_cursor(v.native(), path.native(), focusColumn.native(), gbool(startEditing))
}

// SetCursorOnCell is a wrapper around gtk_tree_view_set_cursor_on_cell().
func (v *TreeView) SetCursorOnCell(path *TreePath, focusColumn *TreeViewColumn, focusCell *CellRenderer, startEditing bool) {
	C.gtk_tree_view_set_cursor_on_cell(v.native(), path.native(), focusColumn.native(), focusCell.native(), gbool(startEditing))
}

// GetCursor is a wrapper around gtk_tree_view_get_cursor().
func (v *TreeView) GetCursor() (p *TreePath, c *TreeViewColumn) {
	var path *C.GtkTreePath
	var col *C.GtkTreeViewColumn

	C.gtk_tree_view_get_cursor(v.native(), &path, &col)

	if path != nil {
		p = &TreePath{path}
		runtime.SetFinalizer(p, (*TreePath).free)
	}

	if col != nil {
		c = wrapTreeViewColumn(glib.Take(unsafe.Pointer(col)))
	}

	return
}

// RowActivated is a wrapper around gtk_tree_view_row_activated().
func (v *TreeView) RowActivated(path *TreePath, column *TreeViewColumn) {
	C.gtk_tree_view_row_activated(v.native(), path.native(), column.native())
}

// ExpandAll is a wrapper around gtk_tree_view_expand_all().
func (v *TreeView) ExpandAll() {
	C.gtk_tree_view_expand_all(v.native())
}

// CollapseAll is a wrapper around gtk_tree_view_collapse_all().
func (v *TreeView) CollapseAll() {
	C.gtk_tree_view_collapse_all(v.native())
}

// ExpandToPath is a wrapper around gtk_tree_view_expand_to_path().
func (v *TreeView) ExpandToPath(path *TreePath) {
	C.gtk_tree_view_expand_to_path(v.native(), path.native())
}

// ExpandRow is a wrapper around gtk_tree_view_expand_row().
func (v *TreeView) ExpandRow(path *TreePath, openAll bool) bool {
	return gobool(C.gtk_tree_view_expand_row(v.native(), path.native(), gbool(openAll)))
}

// CollapseRow is a wrapper around gtk_tree_view_collapse_row().
func (v *TreeView) CollapseRow(path *TreePath) bool {
	return gobool(C.gtk_tree_view_collapse_row(v.native(), path.native()))
}

// RowExpanded is a wrapper around gtk_tree_view_row_expanded().
func (v *TreeView) RowExpanded(path *TreePath) bool {
	return gobool(C.gtk_tree_view_row_expanded(v.native(), path.native()))
}

// SetReorderable is a wrapper around gtk_tree_view_set_reorderable().
func (v *TreeView) SetReorderable(b bool) {
	C.gtk_tree_view_set_reorderable(v.native(), gbool(b))
}

// GetReorderable is a wrapper around gtk_tree_view_get_reorderable().
func (v *TreeView) GetReorderable() bool {
	return gobool(C.gtk_tree_view_get_reorderable(v.native()))
}

// GetBinWindow is a wrapper around gtk_tree_view_get_bin_window().
func (v *TreeView) GetBinWindow() *gdk.Window {
	c := C.gtk_tree_view_get_bin_window(v.native())
	if c == nil {
		return nil
	}

	w := &gdk.Window{glib.Take(unsafe.Pointer(c))}
	return w
}

// SetEnableSearch is a wrapper around gtk_tree_view_set_enable_search().
func (v *TreeView) SetEnableSearch(b bool) {
	C.gtk_tree_view_set_enable_search(v.native(), gbool(b))
}

// GetEnableSearch is a wrapper around gtk_tree_view_get_enable_search().
func (v *TreeView) GetEnableSearch() bool {
	return gobool(C.gtk_tree_view_get_enable_search(v.native()))
}

// SetSearchColumn is a wrapper around gtk_tree_view_set_search_column().
func (v *TreeView) SetSearchColumn(c int) {
	C.gtk_tree_view_set_search_column(v.native(), C.gint(c))
}

// GetSearchColumn is a wrapper around gtk_tree_view_get_search_column().
func (v *TreeView) GetSearchColumn() int {
	return int(C.gtk_tree_view_get_search_column(v.native()))
}

// GetSearchEntry is a wrapper around gtk_tree_view_get_search_entry().
func (v *TreeView) GetSearchEntry() *Entry {
	c := C.gtk_tree_view_get_search_entry(v.native())
	if c == nil {
		return nil
	}
	return wrapEntry(glib.Take(unsafe.Pointer(c)))
}

// SetSearchEntry is a wrapper around gtk_tree_view_set_search_entry().
func (v *TreeView) SetSearchEntry(e *Entry) {
	C.gtk_tree_view_set_search_entry(v.native(), e.native())
}

// SetSearchEqualSubstringMatch sets TreeView to search by substring match.
func (v *TreeView) SetSearchEqualSubstringMatch() {
	C.gtk_tree_view_set_search_equal_func(
		v.native(),
		(C.GtkTreeViewSearchEqualFunc)(unsafe.Pointer(C.substring_match_equal_func)),
		nil,
		nil)
}

// SetFixedHeightMode is a wrapper around gtk_tree_view_set_fixed_height_mode().
func (v *TreeView) SetFixedHeightMode(b bool) {
	C.gtk_tree_view_set_fixed_height_mode(v.native(), gbool(b))
}

// GetFixedHeightMode is a wrapper around gtk_tree_view_get_fixed_height_mode().
func (v *TreeView) GetFixedHeightMode() bool {
	return gobool(C.gtk_tree_view_get_fixed_height_mode(v.native()))
}

// SetHoverSelection is a wrapper around gtk_tree_view_set_hover_selection().
func (v *TreeView) SetHoverSelection(b bool) {
	C.gtk_tree_view_set_hover_selection(v.native(), gbool(b))
}

// GetHoverSelection is a wrapper around gtk_tree_view_get_hover_selection().
func (v *TreeView) GetHoverSelection() bool {
	return gobool(C.gtk_tree_view_get_hover_selection(v.native()))
}

// SetHoverExpand is a wrapper around gtk_tree_view_set_hover_expand().
func (v *TreeView) SetHoverExpand(b bool) {
	C.gtk_tree_view_set_hover_expand(v.native(), gbool(b))
}

// GetHoverExpand is a wrapper around gtk_tree_view_get_hover_expand().
func (v *TreeView) GetHoverExpand() bool {
	return gobool(C.gtk_tree_view_get_hover_expand(v.native()))
}

// SetRubberBanding is a wrapper around gtk_tree_view_set_rubber_banding().
func (v *TreeView) SetRubberBanding(b bool) {
	C.gtk_tree_view_set_rubber_banding(v.native(), gbool(b))
}

// GetRubberBanding is a wrapper around gtk_tree_view_get_rubber_banding().
func (v *TreeView) GetRubberBanding() bool {
	return gobool(C.gtk_tree_view_get_rubber_banding(v.native()))
}

// IsRubberBandingActive is a wrapper around gtk_tree_view_is_rubber_banding_active().
func (v *TreeView) IsRubberBandingActive() bool {
	return gobool(C.gtk_tree_view_is_rubber_banding_active(v.native()))
}

// SetEnableTreeLines is a wrapper around gtk_tree_view_set_enable_tree_lines().
func (v *TreeView) SetEnableTreeLines(b bool) {
	C.gtk_tree_view_set_enable_tree_lines(v.native(), gbool(b))
}

// GetEnableTreeLines is a wrapper around gtk_tree_view_get_enable_tree_lines().
func (v *TreeView) GetEnableTreeLines() bool {
	return gobool(C.gtk_tree_view_get_enable_tree_lines(v.native()))
}

// GetTooltipColumn is a wrapper around gtk_tree_view_get_tooltip_column().
func (v *TreeView) GetTooltipColumn() int {
	return int(C.gtk_tree_view_get_tooltip_column(v.native()))
}

// SetTooltipColumn is a wrapper around gtk_tree_view_set_tooltip_column().
func (v *TreeView) SetTooltipColumn(c int) {
	C.gtk_tree_view_set_tooltip_column(v.native(), C.gint(c))
}

// void 	gtk_tree_view_set_tooltip_row ()
// void 	gtk_tree_view_set_tooltip_cell ()
// gboolean 	gtk_tree_view_get_tooltip_context ()
// void 	gtk_tree_view_set_grid_lines ()
// GtkTreeViewGridLines 	gtk_tree_view_get_grid_lines ()
// void 	(*GtkTreeDestroyCountFunc) ()
// void 	gtk_tree_view_set_destroy_count_func ()
// gboolean 	(*GtkTreeViewRowSeparatorFunc) ()
// GtkTreeViewRowSeparatorFunc 	gtk_tree_view_get_row_separator_func ()
// void 	gtk_tree_view_set_row_separator_func ()
// void 	(*GtkTreeViewSearchPositionFunc) ()
// GtkTreeViewSearchPositionFunc 	gtk_tree_view_get_search_position_func ()
// void 	gtk_tree_view_set_search_position_func ()
// void 	gtk_tree_view_set_search_equal_func ()
// GtkTreeViewSearchEqualFunc 	gtk_tree_view_get_search_equal_func ()
// void 	gtk_tree_view_map_expanded_rows ()
// GList * 	gtk_tree_view_get_columns ()
// gint 	gtk_tree_view_insert_column_with_attributes ()
// gint 	gtk_tree_view_insert_column_with_data_func ()
// void 	gtk_tree_view_set_column_drag_function ()
// void 	gtk_tree_view_scroll_to_cell ()
// gboolean 	gtk_tree_view_is_blank_at_pos ()
// void 	gtk_tree_view_get_cell_area ()
// void 	gtk_tree_view_get_background_area ()
// void 	gtk_tree_view_get_visible_rect ()
// gboolean 	gtk_tree_view_get_visible_range ()
// void 	gtk_tree_view_convert_bin_window_to_tree_coords ()
// void 	gtk_tree_view_convert_bin_window_to_widget_coords ()
// void 	gtk_tree_view_convert_tree_to_bin_window_coords ()
// void 	gtk_tree_view_convert_tree_to_widget_coords ()
// void 	gtk_tree_view_convert_widget_to_bin_window_coords ()
// void 	gtk_tree_view_convert_widget_to_tree_coords ()
// void 	gtk_tree_view_enable_model_drag_dest ()
// void 	gtk_tree_view_enable_model_drag_source ()
// void 	gtk_tree_view_unset_rows_drag_source ()
// void 	gtk_tree_view_unset_rows_drag_dest ()
// void 	gtk_tree_view_set_drag_dest_row ()
// void 	gtk_tree_view_get_drag_dest_row ()
// gboolean 	gtk_tree_view_get_dest_row_at_pos ()
// cairo_surface_t * 	gtk_tree_view_create_row_drag_icon ()
