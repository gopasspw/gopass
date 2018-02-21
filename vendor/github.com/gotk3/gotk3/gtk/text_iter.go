// Same copyright and license as the rest of the files in this project

package gtk

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"

import (
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

/*
 * GtkTextIter
 */

// TextIter is a representation of GTK's GtkTextIter
type TextIter C.GtkTextIter

// native returns a pointer to the underlying GtkTextIter.
func (v *TextIter) native() *C.GtkTextIter {
	if v == nil {
		return nil
	}
	return (*C.GtkTextIter)(v)
}

func marshalTextIter(p uintptr) (interface{}, error) {
	c := C.g_value_get_boxed((*C.GValue)(unsafe.Pointer(p)))
	return (*TextIter)(unsafe.Pointer(c)), nil
}

// GetBuffer is a wrapper around gtk_text_iter_get_buffer().
func (v *TextIter) GetBuffer() *TextBuffer {
	c := C.gtk_text_iter_get_buffer(v.native())
	if c == nil {
		return nil
	}
	return wrapTextBuffer(glib.Take(unsafe.Pointer(c)))
}

// GetOffset is a wrapper around gtk_text_iter_get_offset().
func (v *TextIter) GetOffset() int {
	return int(C.gtk_text_iter_get_offset(v.native()))
}

// GetLine is a wrapper around gtk_text_iter_get_line().
func (v *TextIter) GetLine() int {
	return int(C.gtk_text_iter_get_line(v.native()))
}

// GetLineOffset is a wrapper around gtk_text_iter_get_line_offset().
func (v *TextIter) GetLineOffset() int {
	return int(C.gtk_text_iter_get_line_offset(v.native()))
}

// GetLineIndex is a wrapper around gtk_text_iter_get_line_index().
func (v *TextIter) GetLineIndex() int {
	return int(C.gtk_text_iter_get_line_index(v.native()))
}

// GetVisibleLineOffset is a wrapper around gtk_text_iter_get_visible_line_offset().
func (v *TextIter) GetVisibleLineOffset() int {
	return int(C.gtk_text_iter_get_visible_line_offset(v.native()))
}

// GetVisibleLineIndex is a wrapper around gtk_text_iter_get_visible_line_index().
func (v *TextIter) GetVisibleLineIndex() int {
	return int(C.gtk_text_iter_get_visible_line_index(v.native()))
}

// GetChar is a wrapper around gtk_text_iter_get_char().
func (v *TextIter) GetChar() rune {
	return rune(C.gtk_text_iter_get_char(v.native()))
}

// GetSlice is a wrapper around gtk_text_iter_get_slice().
func (v *TextIter) GetSlice(end *TextIter) string {
	c := C.gtk_text_iter_get_slice(v.native(), end.native())
	return C.GoString((*C.char)(c))
}

// GetText is a wrapper around gtk_text_iter_get_text().
func (v *TextIter) GetText(end *TextIter) string {
	c := C.gtk_text_iter_get_text(v.native(), end.native())
	return C.GoString((*C.char)(c))
}

// GetVisibleSlice is a wrapper around gtk_text_iter_get_visible_slice().
func (v *TextIter) GetVisibleSlice(end *TextIter) string {
	c := C.gtk_text_iter_get_visible_slice(v.native(), end.native())
	return C.GoString((*C.char)(c))
}

// GetVisibleText is a wrapper around gtk_text_iter_get_visible_text().
func (v *TextIter) GetVisibleText(end *TextIter) string {
	c := C.gtk_text_iter_get_visible_text(v.native(), end.native())
	return C.GoString((*C.char)(c))
}

// EndsTag is a wrapper around gtk_text_iter_ends_tag().
func (v *TextIter) EndsTag(v1 *TextTag) bool {
	return gobool(C.gtk_text_iter_ends_tag(v.native(), v1.native()))
}

// TogglesTag is a wrapper around gtk_text_iter_toggles_tag().
func (v *TextIter) TogglesTag(v1 *TextTag) bool {
	return gobool(C.gtk_text_iter_toggles_tag(v.native(), v1.native()))
}

// HasTag is a wrapper around gtk_text_iter_has_tag().
func (v *TextIter) HasTag(v1 *TextTag) bool {
	return gobool(C.gtk_text_iter_has_tag(v.native(), v1.native()))
}

// Editable is a wrapper around gtk_text_iter_editable().
func (v *TextIter) Editable(v1 bool) bool {
	return gobool(C.gtk_text_iter_editable(v.native(), gbool(v1)))
}

// CanInsert is a wrapper around gtk_text_iter_can_insert().
func (v *TextIter) CanInsert(v1 bool) bool {
	return gobool(C.gtk_text_iter_can_insert(v.native(), gbool(v1)))
}

// StartsWord is a wrapper around gtk_text_iter_starts_word().
func (v *TextIter) StartsWord() bool {
	return gobool(C.gtk_text_iter_starts_word(v.native()))
}

// EndsWord is a wrapper around gtk_text_iter_ends_word().
func (v *TextIter) EndsWord() bool {
	return gobool(C.gtk_text_iter_ends_word(v.native()))
}

// InsideWord is a wrapper around gtk_text_iter_inside_word().
func (v *TextIter) InsideWord() bool {
	return gobool(C.gtk_text_iter_inside_word(v.native()))
}

// StartsLine is a wrapper around gtk_text_iter_starts_line().
func (v *TextIter) StartsLine() bool {
	return gobool(C.gtk_text_iter_starts_line(v.native()))
}

// EndsLine is a wrapper around gtk_text_iter_ends_line().
func (v *TextIter) EndsLine() bool {
	return gobool(C.gtk_text_iter_ends_line(v.native()))
}

// StartsSentence is a wrapper around gtk_text_iter_starts_sentence().
func (v *TextIter) StartsSentence() bool {
	return gobool(C.gtk_text_iter_starts_sentence(v.native()))
}

// EndsSentence is a wrapper around gtk_text_iter_ends_sentence().
func (v *TextIter) EndsSentence() bool {
	return gobool(C.gtk_text_iter_ends_sentence(v.native()))
}

// InsideSentence is a wrapper around gtk_text_iter_inside_sentence().
func (v *TextIter) InsideSentence() bool {
	return gobool(C.gtk_text_iter_inside_sentence(v.native()))
}

// IsCursorPosition is a wrapper around gtk_text_iter_is_cursor_position().
func (v *TextIter) IsCursorPosition() bool {
	return gobool(C.gtk_text_iter_is_cursor_position(v.native()))
}

// GetCharsInLine is a wrapper around gtk_text_iter_get_chars_in_line().
func (v *TextIter) GetCharsInLine() int {
	return int(C.gtk_text_iter_get_chars_in_line(v.native()))
}

// GetBytesInLine is a wrapper around gtk_text_iter_get_bytes_in_line().
func (v *TextIter) GetBytesInLine() int {
	return int(C.gtk_text_iter_get_bytes_in_line(v.native()))
}

// IsEnd is a wrapper around gtk_text_iter_is_end().
func (v *TextIter) IsEnd() bool {
	return gobool(C.gtk_text_iter_is_end(v.native()))
}

// IsStart is a wrapper around gtk_text_iter_is_start().
func (v *TextIter) IsStart() bool {
	return gobool(C.gtk_text_iter_is_start(v.native()))
}

// ForwardChar is a wrapper around gtk_text_iter_forward_char().
func (v *TextIter) ForwardChar() bool {
	return gobool(C.gtk_text_iter_forward_char(v.native()))
}

// BackwardChar is a wrapper around gtk_text_iter_backward_char().
func (v *TextIter) BackwardChar() bool {
	return gobool(C.gtk_text_iter_backward_char(v.native()))
}

// ForwardChars is a wrapper around gtk_text_iter_forward_chars().
func (v *TextIter) ForwardChars(v1 int) bool {
	return gobool(C.gtk_text_iter_forward_chars(v.native(), C.gint(v1)))
}

// BackwardChars is a wrapper around gtk_text_iter_backward_chars().
func (v *TextIter) BackwardChars(v1 int) bool {
	return gobool(C.gtk_text_iter_backward_chars(v.native(), C.gint(v1)))
}

// ForwardLine is a wrapper around gtk_text_iter_forward_line().
func (v *TextIter) ForwardLine() bool {
	return gobool(C.gtk_text_iter_forward_line(v.native()))
}

// BackwardLine is a wrapper around gtk_text_iter_backward_line().
func (v *TextIter) BackwardLine() bool {
	return gobool(C.gtk_text_iter_backward_line(v.native()))
}

// ForwardLines is a wrapper around gtk_text_iter_forward_lines().
func (v *TextIter) ForwardLines(v1 int) bool {
	return gobool(C.gtk_text_iter_forward_lines(v.native(), C.gint(v1)))
}

// BackwardLines is a wrapper around gtk_text_iter_backward_lines().
func (v *TextIter) BackwardLines(v1 int) bool {
	return gobool(C.gtk_text_iter_backward_lines(v.native(), C.gint(v1)))
}

// ForwardWordEnds is a wrapper around gtk_text_iter_forward_word_ends().
func (v *TextIter) ForwardWordEnds(v1 int) bool {
	return gobool(C.gtk_text_iter_forward_word_ends(v.native(), C.gint(v1)))
}

// ForwardWordEnd is a wrapper around gtk_text_iter_forward_word_end().
func (v *TextIter) ForwardWordEnd() bool {
	return gobool(C.gtk_text_iter_forward_word_end(v.native()))
}

// ForwardCursorPosition is a wrapper around gtk_text_iter_forward_cursor_position().
func (v *TextIter) ForwardCursorPosition() bool {
	return gobool(C.gtk_text_iter_forward_cursor_position(v.native()))
}

// BackwardCursorPosition is a wrapper around gtk_text_iter_backward_cursor_position().
func (v *TextIter) BackwardCursorPosition() bool {
	return gobool(C.gtk_text_iter_backward_cursor_position(v.native()))
}

// ForwardCursorPositions is a wrapper around gtk_text_iter_forward_cursor_positions().
func (v *TextIter) ForwardCursorPositions(v1 int) bool {
	return gobool(C.gtk_text_iter_forward_cursor_positions(v.native(), C.gint(v1)))
}

// BackwardCursorPositions is a wrapper around gtk_text_iter_backward_cursor_positions().
func (v *TextIter) BackwardCursorPositions(v1 int) bool {
	return gobool(C.gtk_text_iter_backward_cursor_positions(v.native(), C.gint(v1)))
}

// ForwardSentenceEnds is a wrapper around gtk_text_iter_forward_sentence_ends().
func (v *TextIter) ForwardSentenceEnds(v1 int) bool {
	return gobool(C.gtk_text_iter_forward_sentence_ends(v.native(), C.gint(v1)))
}

// ForwardSentenceEnd is a wrapper around gtk_text_iter_forward_sentence_end().
func (v *TextIter) ForwardSentenceEnd() bool {
	return gobool(C.gtk_text_iter_forward_sentence_end(v.native()))
}

// ForwardVisibleWordEnds is a wrapper around gtk_text_iter_forward_word_ends().
func (v *TextIter) ForwardVisibleWordEnds(v1 int) bool {
	return gobool(C.gtk_text_iter_forward_word_ends(v.native(), C.gint(v1)))
}

// ForwardVisibleWordEnd is a wrapper around gtk_text_iter_forward_visible_word_end().
func (v *TextIter) ForwardVisibleWordEnd() bool {
	return gobool(C.gtk_text_iter_forward_visible_word_end(v.native()))
}

// ForwardVisibleCursorPosition is a wrapper around gtk_text_iter_forward_visible_cursor_position().
func (v *TextIter) ForwardVisibleCursorPosition() bool {
	return gobool(C.gtk_text_iter_forward_visible_cursor_position(v.native()))
}

// BackwardVisibleCursorPosition is a wrapper around gtk_text_iter_backward_visible_cursor_position().
func (v *TextIter) BackwardVisibleCursorPosition() bool {
	return gobool(C.gtk_text_iter_backward_visible_cursor_position(v.native()))
}

// ForwardVisibleCursorPositions is a wrapper around gtk_text_iter_forward_visible_cursor_positions().
func (v *TextIter) ForwardVisibleCursorPositions(v1 int) bool {
	return gobool(C.gtk_text_iter_forward_visible_cursor_positions(v.native(), C.gint(v1)))
}

// BackwardVisibleCursorPositions is a wrapper around gtk_text_iter_backward_visible_cursor_positions().
func (v *TextIter) BackwardVisibleCursorPositions(v1 int) bool {
	return gobool(C.gtk_text_iter_backward_visible_cursor_positions(v.native(), C.gint(v1)))
}

// ForwardVisibleLine is a wrapper around gtk_text_iter_forward_visible_line().
func (v *TextIter) ForwardVisibleLine() bool {
	return gobool(C.gtk_text_iter_forward_visible_line(v.native()))
}

// BackwardVisibleLine is a wrapper around gtk_text_iter_backward_visible_line().
func (v *TextIter) BackwardVisibleLine() bool {
	return gobool(C.gtk_text_iter_backward_visible_line(v.native()))
}

// ForwardVisibleLines is a wrapper around gtk_text_iter_forward_visible_lines().
func (v *TextIter) ForwardVisibleLines(v1 int) bool {
	return gobool(C.gtk_text_iter_forward_visible_lines(v.native(), C.gint(v1)))
}

// BackwardVisibleLines is a wrapper around gtk_text_iter_backward_visible_lines().
func (v *TextIter) BackwardVisibleLines(v1 int) bool {
	return gobool(C.gtk_text_iter_backward_visible_lines(v.native(), C.gint(v1)))
}

// SetOffset is a wrapper around gtk_text_iter_set_offset().
func (v *TextIter) SetOffset(v1 int) {
	C.gtk_text_iter_set_offset(v.native(), C.gint(v1))
}

// SetLine is a wrapper around gtk_text_iter_set_line().
func (v *TextIter) SetLine(v1 int) {
	C.gtk_text_iter_set_line(v.native(), C.gint(v1))
}

// SetLineOffset is a wrapper around gtk_text_iter_set_line_offset().
func (v *TextIter) SetLineOffset(v1 int) {
	C.gtk_text_iter_set_line_offset(v.native(), C.gint(v1))
}

// SetLineIndex is a wrapper around gtk_text_iter_set_line_index().
func (v *TextIter) SetLineIndex(v1 int) {
	C.gtk_text_iter_set_line_index(v.native(), C.gint(v1))
}

// SetVisibleLineOffset is a wrapper around gtk_text_iter_set_visible_line_offset().
func (v *TextIter) SetVisibleLineOffset(v1 int) {
	C.gtk_text_iter_set_visible_line_offset(v.native(), C.gint(v1))
}

// SetVisibleLineIndex is a wrapper around gtk_text_iter_set_visible_line_index().
func (v *TextIter) SetVisibleLineIndex(v1 int) {
	C.gtk_text_iter_set_visible_line_index(v.native(), C.gint(v1))
}

// ForwardToEnd is a wrapper around gtk_text_iter_forward_to_end().
func (v *TextIter) ForwardToEnd() {
	C.gtk_text_iter_forward_to_end(v.native())
}

// ForwardToLineEnd is a wrapper around gtk_text_iter_forward_to_line_end().
func (v *TextIter) ForwardToLineEnd() bool {
	return gobool(C.gtk_text_iter_forward_to_line_end(v.native()))
}

// ForwardToTagToggle is a wrapper around gtk_text_iter_forward_to_tag_toggle().
func (v *TextIter) ForwardToTagToggle(v1 *TextTag) bool {
	return gobool(C.gtk_text_iter_forward_to_tag_toggle(v.native(), v1.native()))
}

// BackwardToTagToggle is a wrapper around gtk_text_iter_backward_to_tag_toggle().
func (v *TextIter) BackwardToTagToggle(v1 *TextTag) bool {
	return gobool(C.gtk_text_iter_backward_to_tag_toggle(v.native(), v1.native()))
}

// Equal is a wrapper around gtk_text_iter_equal().
func (v *TextIter) Equal(v1 *TextIter) bool {
	return gobool(C.gtk_text_iter_equal(v.native(), v1.native()))
}

// Compare is a wrapper around gtk_text_iter_compare().
func (v *TextIter) Compare(v1 *TextIter) int {
	return int(C.gtk_text_iter_compare(v.native(), v1.native()))
}

// InRange is a wrapper around gtk_text_iter_in_range().
func (v *TextIter) InRange(v1 *TextIter, v2 *TextIter) bool {
	return gobool(C.gtk_text_iter_in_range(v.native(), v1.native(), v2.native()))
}

// void 	gtk_text_iter_order ()
// gboolean 	(*GtkTextCharPredicate) ()
// gboolean 	gtk_text_iter_forward_find_char ()
// gboolean 	gtk_text_iter_backward_find_char ()
// gboolean 	gtk_text_iter_forward_search ()
// gboolean 	gtk_text_iter_backward_search ()
// gboolean 	gtk_text_iter_get_attributes ()
// GtkTextIter * 	gtk_text_iter_copy ()
// void 	gtk_text_iter_assign ()
// void 	gtk_text_iter_free ()
// GdkPixbuf * 	gtk_text_iter_get_pixbuf ()
// GSList * 	gtk_text_iter_get_marks ()
// GSList * 	gtk_text_iter_get_toggled_tags ()
// GtkTextChildAnchor * 	gtk_text_iter_get_child_anchor ()
// GSList * 	gtk_text_iter_get_tags ()
// PangoLanguage * 	gtk_text_iter_get_language ()
