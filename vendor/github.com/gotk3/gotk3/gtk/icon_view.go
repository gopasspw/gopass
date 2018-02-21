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
 * GtkIconView
 */

// IconView is a representation of GTK's GtkIconView.
type IconView struct {
	Container
}

// native returns a pointer to the underlying GtkIconView.
func (v *IconView) native() *C.GtkIconView {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkIconView(p)
}

func marshalIconView(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapIconView(obj), nil
}

func wrapIconView(obj *glib.Object) *IconView {
	return &IconView{Container{Widget{glib.InitiallyUnowned{obj}}}}
}

// IconViewNew is a wrapper around gtk_icon_view_new().
func IconViewNew() (*IconView, error) {
	c := C.gtk_icon_view_new()
	if c == nil {
		return nil, nilPtrErr
	}

	return wrapIconView(glib.Take(unsafe.Pointer(c))), nil
}

// IconViewNewWithModel is a wrapper around gtk_icon_view_new_with_model().
func IconViewNewWithModel(model ITreeModel) (*IconView, error) {
	c := C.gtk_icon_view_new_with_model(model.toTreeModel())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := glib.Take(unsafe.Pointer(c))
	return wrapIconView(obj), nil
}

// SetModel is a wrapper around gtk_icon_view_set_model().
func (v *IconView) SetModel(model ITreeModel) {
	C.gtk_icon_view_set_model(v.native(), model.toTreeModel())
}

// GetModel is a wrapper around gtk_icon_view_get_model().
func (v *IconView) GetModel() (*TreeModel, error) {
	c := C.gtk_icon_view_get_model(v.native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := glib.Take(unsafe.Pointer(c))
	return wrapTreeModel(obj), nil
}

// SetTextColumn is a wrapper around gtk_icon_view_set_text_column().
func (v *IconView) SetTextColumn(column int) {
	C.gtk_icon_view_set_text_column(v.native(), C.gint(column))
}

// GetTextColumn is a wrapper around gtk_icon_view_get_text_column().
func (v *IconView) GetTextColumn() int {
	return int(C.gtk_icon_view_get_text_column(v.native()))
}

// SetMarkupColumn is a wrapper around gtk_icon_view_set_markup_column().
func (v *IconView) SetMarkupColumn(column int) {
	C.gtk_icon_view_set_markup_column(v.native(), C.gint(column))
}

// GetMarkupColumn is a wrapper around gtk_icon_view_get_markup_column().
func (v *IconView) GetMarkupColumn() int {
	return int(C.gtk_icon_view_get_markup_column(v.native()))
}

// SetPixbufColumn is a wrapper around gtk_icon_view_set_pixbuf_column().
func (v *IconView) SetPixbufColumn(column int) {
	C.gtk_icon_view_set_pixbuf_column(v.native(), C.gint(column))
}

// GetPixbufColumn is a wrapper around gtk_icon_view_get_pixbuf_column().
func (v *IconView) GetPixbufColumn() int {
	return int(C.gtk_icon_view_get_pixbuf_column(v.native()))
}

// GetPathAtPos is a wrapper around gtk_icon_view_get_path_at_pos().
func (v *IconView) GetPathAtPos(x, y int) *TreePath {
	var (
		cpath *C.GtkTreePath
		path  *TreePath
	)

	cpath = C.gtk_icon_view_get_path_at_pos(v.native(), C.gint(x), C.gint(y))

	if cpath != nil {
		path = &TreePath{cpath}
		runtime.SetFinalizer(path, (*TreePath).free)
	}

	return path
}

// GetItemAtPos is a wrapper around gtk_icon_view_get_item_at_pos().
func (v *IconView) GetItemAtPos(x, y int) (*TreePath, *CellRenderer) {
	var (
		cpath *C.GtkTreePath
		ccell *C.GtkCellRenderer
		path  *TreePath
		cell  *CellRenderer
	)

	C.gtk_icon_view_get_item_at_pos(v.native(), C.gint(x), C.gint(y), &cpath, &ccell)

	if cpath != nil {
		path = &TreePath{cpath}
		runtime.SetFinalizer(path, (*TreePath).free)
	}

	if ccell != nil {
		cell = wrapCellRenderer(glib.Take(unsafe.Pointer(ccell)))
	}

	return path, cell
}

// ConvertWidgetToBinWindowCoords is a wrapper around gtk_icon_view_convert_widget_to_bin_window_coords().
func (v *IconView) ConvertWidgetToBinWindowCoords(x, y int) (int, int) {
	var bx, by C.gint

	C.gtk_icon_view_convert_widget_to_bin_window_coords(v.native(), C.gint(x), C.gint(y), &bx, &by)

	return int(bx), int(by)
}

// SetCursor is a wrapper around gtk_icon_view_set_selection_mode().
func (v *IconView) SetCursor(path *TreePath, cell *CellRenderer, startEditing bool) {
	C.gtk_icon_view_set_cursor(v.native(), path.native(), cell.native(), gbool(startEditing))
}

// GetCursor is a wrapper around gtk_icon_view_get_cursor().
func (v *IconView) GetCursor() (*TreePath, *CellRenderer) {
	var (
		cpath *C.GtkTreePath
		ccell *C.GtkCellRenderer
		path  *TreePath
		cell  *CellRenderer
	)

	C.gtk_icon_view_get_cursor(v.native(), &cpath, &ccell)

	if cpath != nil {
		path = &TreePath{cpath}
		runtime.SetFinalizer(path, (*TreePath).free)
	}

	if ccell != nil {
		cell = wrapCellRenderer(glib.Take(unsafe.Pointer(ccell)))
	}

	return path, cell
}

// func (v *IconView) SelectedForeach() {}

// SetSelectionMode is a wrapper around gtk_icon_view_set_selection_mode().
func (v *IconView) SetSelectionMode(mode SelectionMode) {
	C.gtk_icon_view_set_selection_mode(v.native(), C.GtkSelectionMode(mode))
}

// GetSelectionMode is a wrapper around gtk_icon_view_get_selection_mode().
func (v *IconView) GetSelectionMode() SelectionMode {
	return SelectionMode(C.gtk_icon_view_get_selection_mode(v.native()))
}

// SetItemOrientation is a wrapper around gtk_icon_view_set_item_orientation().
func (v *IconView) SetItemOrientation(orientation Orientation) {
	C.gtk_icon_view_set_item_orientation(v.native(), C.GtkOrientation(orientation))
}

// GetItemOrientation is a wrapper around gtk_icon_view_get_item_orientation().
func (v *IconView) GetItemOrientation() Orientation {
	return Orientation(C.gtk_icon_view_get_item_orientation(v.native()))
}

// SetColumns is a wrapper around gtk_icon_view_set_columns().
func (v *IconView) SetColumns(columns int) {
	C.gtk_icon_view_set_columns(v.native(), C.gint(columns))
}

// GetColumns is a wrapper around gtk_icon_view_get_columns().
func (v *IconView) GetColumns() int {
	return int(C.gtk_icon_view_get_columns(v.native()))
}

// SetItemWidth is a wrapper around gtk_icon_view_set_item_width().
func (v *IconView) SetItemWidth(width int) {
	C.gtk_icon_view_set_item_width(v.native(), C.gint(width))
}

// GetItemWidth is a wrapper around gtk_icon_view_get_item_width().
func (v *IconView) GetItemWidth() int {
	return int(C.gtk_icon_view_get_item_width(v.native()))
}

// SetSpacing is a wrapper around gtk_icon_view_set_spacing().
func (v *IconView) SetSpacing(spacing int) {
	C.gtk_icon_view_set_spacing(v.native(), C.gint(spacing))
}

// GetSpacing is a wrapper around gtk_icon_view_get_spacing().
func (v *IconView) GetSpacing() int {
	return int(C.gtk_icon_view_get_spacing(v.native()))
}

// SetRowSpacing is a wrapper around gtk_icon_view_set_row_spacing().
func (v *IconView) SetRowSpacing(rowSpacing int) {
	C.gtk_icon_view_set_row_spacing(v.native(), C.gint(rowSpacing))
}

// GetRowSpacing is a wrapper around gtk_icon_view_get_row_spacing().
func (v *IconView) GetRowSpacing() int {
	return int(C.gtk_icon_view_get_row_spacing(v.native()))
}

// SetColumnSpacing is a wrapper around gtk_icon_view_set_column_spacing().
func (v *IconView) SetColumnSpacing(columnSpacing int) {
	C.gtk_icon_view_set_column_spacing(v.native(), C.gint(columnSpacing))
}

// GetColumnSpacing is a wrapper around gtk_icon_view_get_column_spacing().
func (v *IconView) GetColumnSpacing() int {
	return int(C.gtk_icon_view_get_column_spacing(v.native()))
}

// SetMargin is a wrapper around gtk_icon_view_set_margin().
func (v *IconView) SetMargin(margin int) {
	C.gtk_icon_view_set_margin(v.native(), C.gint(margin))
}

// GetMargin is a wrapper around gtk_icon_view_get_margin().
func (v *IconView) GetMargin() int {
	return int(C.gtk_icon_view_get_margin(v.native()))
}

// SetItemPadding is a wrapper around gtk_icon_view_set_item_padding().
func (v *IconView) SetItemPadding(itemPadding int) {
	C.gtk_icon_view_set_item_padding(v.native(), C.gint(itemPadding))
}

// GetItemPadding is a wrapper around gtk_icon_view_get_item_padding().
func (v *IconView) GetItemPadding() int {
	return int(C.gtk_icon_view_get_item_padding(v.native()))
}

// SetActivateOnSingleClick is a wrapper around gtk_icon_view_set_activate_on_single_click().
func (v *IconView) SetActivateOnSingleClick(single bool) {
	C.gtk_icon_view_set_activate_on_single_click(v.native(), gbool(single))
}

// ActivateOnSingleClick is a wrapper around gtk_icon_view_get_activate_on_single_click().
func (v *IconView) ActivateOnSingleClick() bool {
	return gobool(C.gtk_icon_view_get_activate_on_single_click(v.native()))
}

// GetCellRect is a wrapper around gtk_icon_view_get_cell_rect().
func (v *IconView) GetCellRect(path *TreePath, cell *CellRenderer) *gdk.Rectangle {
	var crect C.GdkRectangle

	C.gtk_icon_view_get_cell_rect(v.native(), path.native(), cell.native(), &crect)

	return gdk.WrapRectangle(uintptr(unsafe.Pointer(&crect)))
}

// SelectPath is a wrapper around gtk_icon_view_select_path().
func (v *IconView) SelectPath(path *TreePath) {
	C.gtk_icon_view_select_path(v.native(), path.native())
}

// UnselectPath is a wrapper around gtk_icon_view_unselect_path().
func (v *IconView) UnselectPath(path *TreePath) {
	C.gtk_icon_view_unselect_path(v.native(), path.native())
}

// PathIsSelected is a wrapper around gtk_icon_view_path_is_selected().
func (v *IconView) PathIsSelected(path *TreePath) bool {
	return gobool(C.gtk_icon_view_path_is_selected(v.native(), path.native()))
}

// GetSelectedItems is a wrapper around gtk_icon_view_unselect_path().
func (v *IconView) GetSelectedItems() *glib.List {
	clist := C.gtk_icon_view_get_selected_items(v.native())
	if clist == nil {
		return nil
	}

	glist := glib.WrapList(uintptr(unsafe.Pointer(clist)))
	glist.DataWrapper(func(ptr unsafe.Pointer) interface{} {
		return &TreePath{(*C.GtkTreePath)(ptr)}
	})
	runtime.SetFinalizer(glist, func(glist *glib.List) {
		glist.FreeFull(func(item interface{}) {
			path := item.(*TreePath)
			C.gtk_tree_path_free(path.GtkTreePath)
		})
	})

	return glist
}

// SelectAll is a wrapper around gtk_icon_view_select_all().
func (v *IconView) SelectAll() {
	C.gtk_icon_view_select_all(v.native())
}

// UnselectAll is a wrapper around gtk_icon_view_unselect_all().
func (v *IconView) UnselectAll() {
	C.gtk_icon_view_unselect_all(v.native())
}

// ItemActivated is a wrapper around gtk_icon_view_item_activated().
func (v *IconView) ItemActivated(path *TreePath) {
	C.gtk_icon_view_item_activated(v.native(), path.native())
}

// ScrollToPath is a wrapper around gtk_icon_view_scroll_to_path().
func (v *IconView) ScrollToPath(path *TreePath, useAlign bool, rowAlign, colAlign float64) {
	C.gtk_icon_view_scroll_to_path(v.native(), path.native(), gbool(useAlign),
		C.gfloat(rowAlign), C.gfloat(colAlign))
}

// GetVisibleRange is a wrapper around gtk_icon_view_get_visible_range().
func (v *IconView) GetVisibleRange() (*TreePath, *TreePath) {
	var (
		cpathStart, cpathEnd *C.GtkTreePath
		pathStart, pathEnd   *TreePath
	)

	C.gtk_icon_view_get_visible_range(v.native(), &cpathStart, &cpathEnd)

	if cpathStart != nil {
		pathStart = &TreePath{cpathStart}
		runtime.SetFinalizer(pathStart, (*TreePath).free)
	}

	if cpathEnd != nil {
		pathEnd = &TreePath{cpathEnd}
		runtime.SetFinalizer(pathEnd, (*TreePath).free)
	}

	return pathStart, pathEnd
}

// SetTooltipItem is a wrapper around gtk_icon_view_set_tooltip_item().
func (v *IconView) SetTooltipItem(tooltip *Tooltip, path *TreePath) {
	C.gtk_icon_view_set_tooltip_item(v.native(), tooltip.native(), path.native())
}

// SetTooltipCell is a wrapper around gtk_icon_view_set_tooltip_cell().
func (v *IconView) SetTooltipCell(tooltip *Tooltip, path *TreePath, cell *CellRenderer) {
	C.gtk_icon_view_set_tooltip_cell(v.native(), tooltip.native(), path.native(), cell.native())
}

// GetTooltipContext is a wrapper around gtk_icon_view_get_tooltip_context().
func (v *IconView) GetTooltipContext(x, y int, keyboardTip bool) (*TreeModel, *TreePath, *TreeIter) {
	var (
		cmodel *C.GtkTreeModel
		cpath  *C.GtkTreePath
		citer  *C.GtkTreeIter
		model  *TreeModel
		path   *TreePath
		iter   *TreeIter
	)

	px := C.gint(x)
	py := C.gint(y)
	if !gobool(C.gtk_icon_view_get_tooltip_context(v.native(),
		&px,
		&py,
		gbool(keyboardTip),
		&cmodel,
		&cpath,
		citer,
	)) {
		return nil, nil, nil
	}

	if cmodel != nil {
		model = wrapTreeModel(glib.Take(unsafe.Pointer(cmodel)))
	}

	if cpath != nil {
		path = &TreePath{cpath}
		runtime.SetFinalizer(path, (*TreePath).free)
	}

	if citer != nil {
		iter = &TreeIter{*citer}
		runtime.SetFinalizer(iter, (*TreeIter).free)
	}

	return model, path, iter
}

// SetTooltipColumn is a wrapper around gtk_icon_view_set_tooltip_column().
func (v *IconView) SetTooltipColumn(column int) {
	C.gtk_icon_view_set_tooltip_column(v.native(), C.gint(column))
}

// GetTooltipColumn is a wrapper around gtk_icon_view_get_tooltip_column().
func (v *IconView) GetTooltipColumn() int {
	return int(C.gtk_icon_view_get_tooltip_column(v.native()))
}

// GetItemRow is a wrapper around gtk_icon_view_get_item_row().
func (v *IconView) GetItemRow(path *TreePath) int {
	return int(C.gtk_icon_view_get_item_row(v.native(), path.native()))
}

/*
func (v *IconView) EnableModelDragSource() {}

func (v *IconView) EnableModelDragDest() {}

func (v *IconView) UnsetModelDragSource() {}

func (v *IconView) UnsetModelDragDest() {}
*/

// SetReorderable is a wrapper around gtk_icon_view_set_reorderable().
func (v *IconView) SetReorderable(reorderable bool) {
	C.gtk_icon_view_set_reorderable(v.native(), gbool(reorderable))
}

// GetReorderable is a wrapper around gtk_icon_view_get_reorderable().
func (v *IconView) GetReorderable() bool {
	return gobool(C.gtk_icon_view_get_reorderable(v.native()))
}

/*
func (v *IconView) SetDragDestItem() {}

func (v *IconView) GetDragDestItem() {}

func (v *IconView) GetDestItemAtPos() {}

func (v *IconView) CreateDragIcon() {}
*/
