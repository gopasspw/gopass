// Same copyright and license as the rest of the files in this project

package gtk

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
)

// TextWindowType is a representation of GTK's GtkTextWindowType.
type TextWindowType int

const (
	TEXT_WINDOW_WIDGET TextWindowType = C.GTK_TEXT_WINDOW_WIDGET
	TEXT_WINDOW_TEXT   TextWindowType = C.GTK_TEXT_WINDOW_TEXT
	TEXT_WINDOW_LEFT   TextWindowType = C.GTK_TEXT_WINDOW_LEFT
	TEXT_WINDOW_RIGHT  TextWindowType = C.GTK_TEXT_WINDOW_RIGHT
	TEXT_WINDOW_TOP    TextWindowType = C.GTK_TEXT_WINDOW_TOP
	TEXT_WINDOW_BOTTOM TextWindowType = C.GTK_TEXT_WINDOW_BOTTOM
)

/*
 * GtkTextView
 */

// TextView is a representation of GTK's GtkTextView
type TextView struct {
	Container
}

// native returns a pointer to the underlying GtkTextView.
func (v *TextView) native() *C.GtkTextView {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkTextView(p)
}

func marshalTextView(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapTextView(obj), nil
}

func wrapTextView(obj *glib.Object) *TextView {
	return &TextView{Container{Widget{glib.InitiallyUnowned{obj}}}}
}

// TextViewNew is a wrapper around gtk_text_view_new().
func TextViewNew() (*TextView, error) {
	c := C.gtk_text_view_new()
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapTextView(glib.Take(unsafe.Pointer(c))), nil
}

// TextViewNewWithBuffer is a wrapper around gtk_text_view_new_with_buffer().
func TextViewNewWithBuffer(buf *TextBuffer) (*TextView, error) {
	cbuf := buf.native()
	c := C.gtk_text_view_new_with_buffer(cbuf)
	return wrapTextView(glib.Take(unsafe.Pointer(c))), nil
}

// GetBuffer is a wrapper around gtk_text_view_get_buffer().
func (v *TextView) GetBuffer() (*TextBuffer, error) {
	c := C.gtk_text_view_get_buffer(v.native())
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapTextBuffer(glib.Take(unsafe.Pointer(c))), nil
}

// SetBuffer is a wrapper around gtk_text_view_set_buffer().
func (v *TextView) SetBuffer(buffer *TextBuffer) {
	C.gtk_text_view_set_buffer(v.native(), buffer.native())
}

// SetEditable is a wrapper around gtk_text_view_set_editable().
func (v *TextView) SetEditable(editable bool) {
	C.gtk_text_view_set_editable(v.native(), gbool(editable))
}

// GetEditable is a wrapper around gtk_text_view_get_editable().
func (v *TextView) GetEditable() bool {
	c := C.gtk_text_view_get_editable(v.native())
	return gobool(c)
}

// SetWrapMode is a wrapper around gtk_text_view_set_wrap_mode().
func (v *TextView) SetWrapMode(wrapMode WrapMode) {
	C.gtk_text_view_set_wrap_mode(v.native(), C.GtkWrapMode(wrapMode))
}

// GetWrapMode is a wrapper around gtk_text_view_get_wrap_mode().
func (v *TextView) GetWrapMode() WrapMode {
	return WrapMode(C.gtk_text_view_get_wrap_mode(v.native()))
}

// SetCursorVisible is a wrapper around gtk_text_view_set_cursor_visible().
func (v *TextView) SetCursorVisible(visible bool) {
	C.gtk_text_view_set_cursor_visible(v.native(), gbool(visible))
}

// GetCursorVisible is a wrapper around gtk_text_view_get_cursor_visible().
func (v *TextView) GetCursorVisible() bool {
	c := C.gtk_text_view_get_cursor_visible(v.native())
	return gobool(c)
}

// SetOverwrite is a wrapper around gtk_text_view_set_overwrite().
func (v *TextView) SetOverwrite(overwrite bool) {
	C.gtk_text_view_set_overwrite(v.native(), gbool(overwrite))
}

// GetOverwrite is a wrapper around gtk_text_view_get_overwrite().
func (v *TextView) GetOverwrite() bool {
	c := C.gtk_text_view_get_overwrite(v.native())
	return gobool(c)
}

// SetJustification is a wrapper around gtk_text_view_set_justification().
func (v *TextView) SetJustification(justify Justification) {
	C.gtk_text_view_set_justification(v.native(), C.GtkJustification(justify))
}

// GetJustification is a wrapper around gtk_text_view_get_justification().
func (v *TextView) GetJustification() Justification {
	c := C.gtk_text_view_get_justification(v.native())
	return Justification(c)
}

// SetAcceptsTab is a wrapper around gtk_text_view_set_accepts_tab().
func (v *TextView) SetAcceptsTab(acceptsTab bool) {
	C.gtk_text_view_set_accepts_tab(v.native(), gbool(acceptsTab))
}

// GetAcceptsTab is a wrapper around gtk_text_view_get_accepts_tab().
func (v *TextView) GetAcceptsTab() bool {
	c := C.gtk_text_view_get_accepts_tab(v.native())
	return gobool(c)
}

// SetPixelsAboveLines is a wrapper around gtk_text_view_set_pixels_above_lines().
func (v *TextView) SetPixelsAboveLines(px int) {
	C.gtk_text_view_set_pixels_above_lines(v.native(), C.gint(px))
}

// GetPixelsAboveLines is a wrapper around gtk_text_view_get_pixels_above_lines().
func (v *TextView) GetPixelsAboveLines() int {
	c := C.gtk_text_view_get_pixels_above_lines(v.native())
	return int(c)
}

// SetPixelsBelowLines is a wrapper around gtk_text_view_set_pixels_below_lines().
func (v *TextView) SetPixelsBelowLines(px int) {
	C.gtk_text_view_set_pixels_below_lines(v.native(), C.gint(px))
}

// GetPixelsBelowLines is a wrapper around gtk_text_view_get_pixels_below_lines().
func (v *TextView) GetPixelsBelowLines() int {
	c := C.gtk_text_view_get_pixels_below_lines(v.native())
	return int(c)
}

// SetPixelsInsideWrap is a wrapper around gtk_text_view_set_pixels_inside_wrap().
func (v *TextView) SetPixelsInsideWrap(px int) {
	C.gtk_text_view_set_pixels_inside_wrap(v.native(), C.gint(px))
}

// GetPixelsInsideWrap is a wrapper around gtk_text_view_get_pixels_inside_wrap().
func (v *TextView) GetPixelsInsideWrap() int {
	c := C.gtk_text_view_get_pixels_inside_wrap(v.native())
	return int(c)
}

// SetLeftMargin is a wrapper around gtk_text_view_set_left_margin().
func (v *TextView) SetLeftMargin(margin int) {
	C.gtk_text_view_set_left_margin(v.native(), C.gint(margin))
}

// GetLeftMargin is a wrapper around gtk_text_view_get_left_margin().
func (v *TextView) GetLeftMargin() int {
	c := C.gtk_text_view_get_left_margin(v.native())
	return int(c)
}

// SetRightMargin is a wrapper around gtk_text_view_set_right_margin().
func (v *TextView) SetRightMargin(margin int) {
	C.gtk_text_view_set_right_margin(v.native(), C.gint(margin))
}

// GetRightMargin is a wrapper around gtk_text_view_get_right_margin().
func (v *TextView) GetRightMargin() int {
	c := C.gtk_text_view_get_right_margin(v.native())
	return int(c)
}

// SetIndent is a wrapper around gtk_text_view_set_indent().
func (v *TextView) SetIndent(indent int) {
	C.gtk_text_view_set_indent(v.native(), C.gint(indent))
}

// GetIndent is a wrapper around gtk_text_view_get_indent().
func (v *TextView) GetIndent() int {
	c := C.gtk_text_view_get_indent(v.native())
	return int(c)
}

// SetInputHints is a wrapper around gtk_text_view_set_input_hints().
func (v *TextView) SetInputHints(hints InputHints) {
	C.gtk_text_view_set_input_hints(v.native(), C.GtkInputHints(hints))
}

// GetInputHints is a wrapper around gtk_text_view_get_input_hints().
func (v *TextView) GetInputHints() InputHints {
	c := C.gtk_text_view_get_input_hints(v.native())
	return InputHints(c)
}

// SetInputPurpose is a wrapper around gtk_text_view_set_input_purpose().
func (v *TextView) SetInputPurpose(purpose InputPurpose) {
	C.gtk_text_view_set_input_purpose(v.native(),
		C.GtkInputPurpose(purpose))
}

// GetInputPurpose is a wrapper around gtk_text_view_get_input_purpose().
func (v *TextView) GetInputPurpose() InputPurpose {
	c := C.gtk_text_view_get_input_purpose(v.native())
	return InputPurpose(c)
}

// ScrollToMark is a wrapper around gtk_text_view_scroll_to_mark().
func (v *TextView) ScrollToMark(mark *TextMark, within_margin float64, use_align bool, xalign, yalign float64) {
	C.gtk_text_view_scroll_to_mark(v.native(), mark.native(), C.gdouble(within_margin), gbool(use_align), C.gdouble(xalign), C.gdouble(yalign))
}

// ScrollToIter is a wrapper around gtk_text_view_scroll_to_iter().
func (v *TextView) ScrollToIter(iter *TextIter, within_margin float64, use_align bool, xalign, yalign float64) bool {
	return gobool(C.gtk_text_view_scroll_to_iter(v.native(), iter.native(), C.gdouble(within_margin), gbool(use_align), C.gdouble(xalign), C.gdouble(yalign)))
}

// ScrollMarkOnscreen is a wrapper around gtk_text_view_scroll_mark_onscreen().
func (v *TextView) ScrollMarkOnscreen(mark *TextMark) {
	C.gtk_text_view_scroll_mark_onscreen(v.native(), mark.native())
}

// MoveMarkOnscreen is a wrapper around gtk_text_view_move_mark_onscreen().
func (v *TextView) MoveMarkOnscreen(mark *TextMark) bool {
	return gobool(C.gtk_text_view_move_mark_onscreen(v.native(), mark.native()))
}

// PlaceCursorOnscreen is a wrapper around gtk_text_view_place_cursor_onscreen().
func (v *TextView) PlaceCursorOnscreen() bool {
	return gobool(C.gtk_text_view_place_cursor_onscreen(v.native()))
}

// GetVisibleRect is a wrapper around gtk_text_view_get_visible_rect().
func (v *TextView) GetVisibleRect() *gdk.Rectangle {
	var rect C.GdkRectangle
	C.gtk_text_view_get_visible_rect(v.native(), &rect)
	return gdk.WrapRectangle(uintptr(unsafe.Pointer(&rect)))
}

// GetIterLocation is a wrapper around gtk_text_view_get_iter_location().
func (v *TextView) GetIterLocation(iter *TextIter) *gdk.Rectangle {
	var rect C.GdkRectangle
	C.gtk_text_view_get_iter_location(v.native(), iter.native(), &rect)
	return gdk.WrapRectangle(uintptr(unsafe.Pointer(&rect)))
}

// GetCursorLocations is a wrapper around gtk_text_view_get_cursor_locations().
func (v *TextView) GetCursorLocations(iter *TextIter) (strong, weak *gdk.Rectangle) {
	var strongRect, weakRect C.GdkRectangle
	C.gtk_text_view_get_cursor_locations(v.native(), iter.native(), &strongRect, &weakRect)
	return gdk.WrapRectangle(uintptr(unsafe.Pointer(&strongRect))), gdk.WrapRectangle(uintptr(unsafe.Pointer(&weakRect)))
}

// GetLineAtY is a wrapper around gtk_text_view_get_line_at_y().
func (v *TextView) GetLineAtY(y int) (*TextIter, int) {
	var iter TextIter
	var line_top C.gint
	iiter := (C.GtkTextIter)(iter)
	C.gtk_text_view_get_line_at_y(v.native(), &iiter, C.gint(y), &line_top)
	return &iter, int(line_top)
}

// GetLineYrange is a wrapper around gtk_text_view_get_line_yrange().
func (v *TextView) GetLineYrange(iter *TextIter) (y, height int) {
	var yx, heightx C.gint
	C.gtk_text_view_get_line_yrange(v.native(), iter.native(), &yx, &heightx)
	return int(yx), int(heightx)
}

// GetIterAtLocation is a wrapper around gtk_text_view_get_iter_at_location().
func (v *TextView) GetIterAtLocation(x, y int) *TextIter {
	var iter C.GtkTextIter
	C.gtk_text_view_get_iter_at_location(v.native(), &iter, C.gint(x), C.gint(y))
	return (*TextIter)(&iter)
}

// GetIterAtPosition is a wrapper around gtk_text_view_get_iter_at_position().
func (v *TextView) GetIterAtPosition(x, y int) (*TextIter, int) {
	var iter C.GtkTextIter
	var trailing C.gint
	C.gtk_text_view_get_iter_at_position(v.native(), &iter, &trailing, C.gint(x), C.gint(y))
	return (*TextIter)(&iter), int(trailing)
}

// BufferToWindowCoords is a wrapper around gtk_text_view_buffer_to_window_coords().
func (v *TextView) BufferToWindowCoords(win TextWindowType, buffer_x, buffer_y int) (window_x, window_y int) {
	var wx, wy C.gint
	C.gtk_text_view_buffer_to_window_coords(v.native(), C.GtkTextWindowType(win), C.gint(buffer_x), C.gint(buffer_y), &wx, &wy)
	return int(wx), int(wy)
}

// WindowToBufferCoords is a wrapper around gtk_text_view_window_to_buffer_coords().
func (v *TextView) WindowToBufferCoords(win TextWindowType, window_x, window_y int) (buffer_x, buffer_y int) {
	var bx, by C.gint
	C.gtk_text_view_window_to_buffer_coords(v.native(), C.GtkTextWindowType(win), C.gint(window_x), C.gint(window_y), &bx, &by)
	return int(bx), int(by)
}

// GetWindow is a wrapper around gtk_text_view_get_window().
func (v *TextView) GetWindow(win TextWindowType) *gdk.Window {
	c := C.gtk_text_view_get_window(v.native(), C.GtkTextWindowType(win))
	if c == nil {
		return nil
	}
	return &gdk.Window{glib.Take(unsafe.Pointer(c))}
}

// GetWindowType is a wrapper around gtk_text_view_get_window_type().
func (v *TextView) GetWindowType(w *gdk.Window) TextWindowType {
	return TextWindowType(C.gtk_text_view_get_window_type(v.native(), (*C.GdkWindow)(unsafe.Pointer(w.Native()))))
}

// SetBorderWindowSize is a wrapper around gtk_text_view_set_border_window_size().
func (v *TextView) SetBorderWindowSize(tp TextWindowType, size int) {
	C.gtk_text_view_set_border_window_size(v.native(), C.GtkTextWindowType(tp), C.gint(size))
}

// GetBorderWindowSize is a wrapper around gtk_text_view_get_border_window_size().
func (v *TextView) GetBorderWindowSize(tp TextWindowType) int {
	return int(C.gtk_text_view_get_border_window_size(v.native(), C.GtkTextWindowType(tp)))
}

// ForwardDisplayLine is a wrapper around gtk_text_view_forward_display_line().
func (v *TextView) ForwardDisplayLine(iter *TextIter) bool {
	return gobool(C.gtk_text_view_forward_display_line(v.native(), iter.native()))
}

// BackwardDisplayLine is a wrapper around gtk_text_view_backward_display_line().
func (v *TextView) BackwardDisplayLine(iter *TextIter) bool {
	return gobool(C.gtk_text_view_backward_display_line(v.native(), iter.native()))
}

// ForwardDisplayLineEnd is a wrapper around gtk_text_view_forward_display_line_end().
func (v *TextView) ForwardDisplayLineEnd(iter *TextIter) bool {
	return gobool(C.gtk_text_view_forward_display_line_end(v.native(), iter.native()))
}

// BackwardDisplayLineStart is a wrapper around gtk_text_view_backward_display_line_start().
func (v *TextView) BackwardDisplayLineStart(iter *TextIter) bool {
	return gobool(C.gtk_text_view_backward_display_line_start(v.native(), iter.native()))
}

// StartsDisplayLine is a wrapper around gtk_text_view_starts_display_line().
func (v *TextView) StartsDisplayLine(iter *TextIter) bool {
	return gobool(C.gtk_text_view_starts_display_line(v.native(), iter.native()))
}

// MoveVisually is a wrapper around gtk_text_view_move_visually().
func (v *TextView) MoveVisually(iter *TextIter, count int) bool {
	return gobool(C.gtk_text_view_move_visually(v.native(), iter.native(), C.gint(count)))
}

// AddChildInWindow is a wrapper around gtk_text_view_add_child_in_window().
func (v *TextView) AddChildInWindow(child IWidget, tp TextWindowType, xpos, ypos int) {
	C.gtk_text_view_add_child_in_window(v.native(), child.toWidget(), C.GtkTextWindowType(tp), C.gint(xpos), C.gint(ypos))
}

// MoveChild is a wrapper around gtk_text_view_move_child().
func (v *TextView) MoveChild(child IWidget, xpos, ypos int) {
	C.gtk_text_view_move_child(v.native(), child.toWidget(), C.gint(xpos), C.gint(ypos))
}

// ImContextFilterKeypress is a wrapper around gtk_text_view_im_context_filter_keypress().
func (v *TextView) ImContextFilterKeypress(event *gdk.EventKey) bool {
	return gobool(C.gtk_text_view_im_context_filter_keypress(v.native(), (*C.GdkEventKey)(unsafe.Pointer(event.Native()))))
}

// ResetImContext is a wrapper around gtk_text_view_reset_im_context().
func (v *TextView) ResetImContext() {
	C.gtk_text_view_reset_im_context(v.native())
}

// GtkAdjustment * 	gtk_text_view_get_hadjustment ()  -- DEPRECATED
// GtkAdjustment * 	gtk_text_view_get_vadjustment ()  -- DEPRECATED
// void 	gtk_text_view_add_child_at_anchor ()
// GtkTextChildAnchor * 	gtk_text_child_anchor_new ()
// GList * 	gtk_text_child_anchor_get_widgets ()
// gboolean 	gtk_text_child_anchor_get_deleted ()
// void 	gtk_text_view_set_top_margin () -- SINCE 3.18
// gint 	gtk_text_view_get_top_margin () -- SINCE 3.18
// void 	gtk_text_view_set_bottom_margin ()  -- SINCE 3.18
// gint 	gtk_text_view_get_bottom_margin ()  -- SINCE 3.18
// void 	gtk_text_view_set_tabs () -- PangoTabArray
// PangoTabArray * 	gtk_text_view_get_tabs () -- PangoTabArray
// GtkTextAttributes * 	gtk_text_view_get_default_attributes () -- GtkTextAttributes
// void 	gtk_text_view_set_monospace () -- SINCE 3.16
// gboolean 	gtk_text_view_get_monospace () -- SINCE 3.16
