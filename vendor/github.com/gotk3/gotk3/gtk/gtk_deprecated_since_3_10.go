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

// This file includes wrapers for symbols deprecated beginning with GTK 3.10,
// and should only be included in a build targeted intended to target GTK
// 3.8 or earlier.  To target an earlier build build, use the build tag
// gtk_MAJOR_MINOR.  For example, to target GTK 3.8, run
// 'go build -tags gtk_3_8'.
// +build gtk_3_6 gtk_3_8

package gtk

// #cgo pkg-config: gtk+-3.0
// #include <stdlib.h>
// #include <gtk/gtk.h>
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

// ButtonNewFromStock is a wrapper around gtk_button_new_from_stock().
func ButtonNewFromStock(stock Stock) (*Button, error) {
	cstr := C.CString(string(stock))
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_button_new_from_stock((*C.gchar)(cstr))
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapButton(glib.Take(unsafe.Pointer(c))), nil
}

// SetUseStock is a wrapper around gtk_button_set_use_stock().
func (v *Button) SetUseStock(useStock bool) {
	C.gtk_button_set_use_stock(v.native(), gbool(useStock))
}

// GetUseStock is a wrapper around gtk_button_get_use_stock().
func (v *Button) GetUseStock() bool {
	c := C.gtk_button_get_use_stock(v.native())
	return gobool(c)
}

// GetIconStock is a wrapper around gtk_entry_get_icon_stock().
func (v *Entry) GetIconStock(iconPos EntryIconPosition) (string, error) {
	c := C.gtk_entry_get_icon_stock(v.native(),
		C.GtkEntryIconPosition(iconPos))
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

// SetIconFromStock is a wrapper around gtk_entry_set_icon_from_stock().
func (v *Entry) SetIconFromStock(iconPos EntryIconPosition, stockID string) {
	cstr := C.CString(stockID)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_entry_set_icon_from_stock(v.native(),
		C.GtkEntryIconPosition(iconPos), (*C.gchar)(cstr))
}

// ImageNewFromStock is a wrapper around gtk_image_new_from_stock().
func ImageNewFromStock(stock Stock, size IconSize) (*Image, error) {
	cstr := C.CString(string(stock))
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_image_new_from_stock((*C.gchar)(cstr), C.GtkIconSize(size))
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapImage(glib.Take(unsafe.Pointer(c))), nil
}

// SetFromStock is a wrapper around gtk_image_set_from_stock().
func (v *Image) SetFromStock(stock Stock, size IconSize) {
	cstr := C.CString(string(stock))
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_image_set_from_stock(v.native(), (*C.gchar)(cstr),
		C.GtkIconSize(size))
}

// Stock is a special type that does not have an equivalent type in
// GTK.  It is the type used as a parameter anytime an identifier for
// stock icons are needed.  A Stock must be type converted to string when
// function parameters may take a Stock, but when other string values are
// valid as well.
type Stock string

const (
	STOCK_ABOUT                         Stock = C.GTK_STOCK_ABOUT
	STOCK_ADD                           Stock = C.GTK_STOCK_ADD
	STOCK_APPLY                         Stock = C.GTK_STOCK_APPLY
	STOCK_BOLD                          Stock = C.GTK_STOCK_BOLD
	STOCK_CANCEL                        Stock = C.GTK_STOCK_CANCEL
	STOCK_CAPS_LOCK_WARNING             Stock = C.GTK_STOCK_CAPS_LOCK_WARNING
	STOCK_CDROM                         Stock = C.GTK_STOCK_CDROM
	STOCK_CLEAR                         Stock = C.GTK_STOCK_CLEAR
	STOCK_CLOSE                         Stock = C.GTK_STOCK_CLOSE
	STOCK_COLOR_PICKER                  Stock = C.GTK_STOCK_COLOR_PICKER
	STOCK_CONNECT                       Stock = C.GTK_STOCK_CONNECT
	STOCK_CONVERT                       Stock = C.GTK_STOCK_CONVERT
	STOCK_COPY                          Stock = C.GTK_STOCK_COPY
	STOCK_CUT                           Stock = C.GTK_STOCK_CUT
	STOCK_DELETE                        Stock = C.GTK_STOCK_DELETE
	STOCK_DIALOG_AUTHENTICATION         Stock = C.GTK_STOCK_DIALOG_AUTHENTICATION
	STOCK_DIALOG_INFO                   Stock = C.GTK_STOCK_DIALOG_INFO
	STOCK_DIALOG_WARNING                Stock = C.GTK_STOCK_DIALOG_WARNING
	STOCK_DIALOG_ERROR                  Stock = C.GTK_STOCK_DIALOG_ERROR
	STOCK_DIALOG_QUESTION               Stock = C.GTK_STOCK_DIALOG_QUESTION
	STOCK_DIRECTORY                     Stock = C.GTK_STOCK_DIRECTORY
	STOCK_DISCARD                       Stock = C.GTK_STOCK_DISCARD
	STOCK_DISCONNECT                    Stock = C.GTK_STOCK_DISCONNECT
	STOCK_DND                           Stock = C.GTK_STOCK_DND
	STOCK_DND_MULTIPLE                  Stock = C.GTK_STOCK_DND_MULTIPLE
	STOCK_EDIT                          Stock = C.GTK_STOCK_EDIT
	STOCK_EXECUTE                       Stock = C.GTK_STOCK_EXECUTE
	STOCK_FILE                          Stock = C.GTK_STOCK_FILE
	STOCK_FIND                          Stock = C.GTK_STOCK_FIND
	STOCK_FIND_AND_REPLACE              Stock = C.GTK_STOCK_FIND_AND_REPLACE
	STOCK_FLOPPY                        Stock = C.GTK_STOCK_FLOPPY
	STOCK_FULLSCREEN                    Stock = C.GTK_STOCK_FULLSCREEN
	STOCK_GOTO_BOTTOM                   Stock = C.GTK_STOCK_GOTO_BOTTOM
	STOCK_GOTO_FIRST                    Stock = C.GTK_STOCK_GOTO_FIRST
	STOCK_GOTO_LAST                     Stock = C.GTK_STOCK_GOTO_LAST
	STOCK_GOTO_TOP                      Stock = C.GTK_STOCK_GOTO_TOP
	STOCK_GO_BACK                       Stock = C.GTK_STOCK_GO_BACK
	STOCK_GO_DOWN                       Stock = C.GTK_STOCK_GO_DOWN
	STOCK_GO_FORWARD                    Stock = C.GTK_STOCK_GO_FORWARD
	STOCK_GO_UP                         Stock = C.GTK_STOCK_GO_UP
	STOCK_HARDDISK                      Stock = C.GTK_STOCK_HARDDISK
	STOCK_HELP                          Stock = C.GTK_STOCK_HELP
	STOCK_HOME                          Stock = C.GTK_STOCK_HOME
	STOCK_INDEX                         Stock = C.GTK_STOCK_INDEX
	STOCK_INDENT                        Stock = C.GTK_STOCK_INDENT
	STOCK_INFO                          Stock = C.GTK_STOCK_INFO
	STOCK_ITALIC                        Stock = C.GTK_STOCK_ITALIC
	STOCK_JUMP_TO                       Stock = C.GTK_STOCK_JUMP_TO
	STOCK_JUSTIFY_CENTER                Stock = C.GTK_STOCK_JUSTIFY_CENTER
	STOCK_JUSTIFY_FILL                  Stock = C.GTK_STOCK_JUSTIFY_FILL
	STOCK_JUSTIFY_LEFT                  Stock = C.GTK_STOCK_JUSTIFY_LEFT
	STOCK_JUSTIFY_RIGHT                 Stock = C.GTK_STOCK_JUSTIFY_RIGHT
	STOCK_LEAVE_FULLSCREEN              Stock = C.GTK_STOCK_LEAVE_FULLSCREEN
	STOCK_MISSING_IMAGE                 Stock = C.GTK_STOCK_MISSING_IMAGE
	STOCK_MEDIA_FORWARD                 Stock = C.GTK_STOCK_MEDIA_FORWARD
	STOCK_MEDIA_NEXT                    Stock = C.GTK_STOCK_MEDIA_NEXT
	STOCK_MEDIA_PAUSE                   Stock = C.GTK_STOCK_MEDIA_PAUSE
	STOCK_MEDIA_PLAY                    Stock = C.GTK_STOCK_MEDIA_PLAY
	STOCK_MEDIA_PREVIOUS                Stock = C.GTK_STOCK_MEDIA_PREVIOUS
	STOCK_MEDIA_RECORD                  Stock = C.GTK_STOCK_MEDIA_RECORD
	STOCK_MEDIA_REWIND                  Stock = C.GTK_STOCK_MEDIA_REWIND
	STOCK_MEDIA_STOP                    Stock = C.GTK_STOCK_MEDIA_STOP
	STOCK_NETWORK                       Stock = C.GTK_STOCK_NETWORK
	STOCK_NEW                           Stock = C.GTK_STOCK_NEW
	STOCK_NO                            Stock = C.GTK_STOCK_NO
	STOCK_OK                            Stock = C.GTK_STOCK_OK
	STOCK_OPEN                          Stock = C.GTK_STOCK_OPEN
	STOCK_ORIENTATION_PORTRAIT          Stock = C.GTK_STOCK_ORIENTATION_PORTRAIT
	STOCK_ORIENTATION_LANDSCAPE         Stock = C.GTK_STOCK_ORIENTATION_LANDSCAPE
	STOCK_ORIENTATION_REVERSE_LANDSCAPE Stock = C.GTK_STOCK_ORIENTATION_REVERSE_LANDSCAPE
	STOCK_ORIENTATION_REVERSE_PORTRAIT  Stock = C.GTK_STOCK_ORIENTATION_REVERSE_PORTRAIT
	STOCK_PAGE_SETUP                    Stock = C.GTK_STOCK_PAGE_SETUP
	STOCK_PASTE                         Stock = C.GTK_STOCK_PASTE
	STOCK_PREFERENCES                   Stock = C.GTK_STOCK_PREFERENCES
	STOCK_PRINT                         Stock = C.GTK_STOCK_PRINT
	STOCK_PRINT_ERROR                   Stock = C.GTK_STOCK_PRINT_ERROR
	STOCK_PRINT_PAUSED                  Stock = C.GTK_STOCK_PRINT_PAUSED
	STOCK_PRINT_PREVIEW                 Stock = C.GTK_STOCK_PRINT_PREVIEW
	STOCK_PRINT_REPORT                  Stock = C.GTK_STOCK_PRINT_REPORT
	STOCK_PRINT_WARNING                 Stock = C.GTK_STOCK_PRINT_WARNING
	STOCK_PROPERTIES                    Stock = C.GTK_STOCK_PROPERTIES
	STOCK_QUIT                          Stock = C.GTK_STOCK_QUIT
	STOCK_REDO                          Stock = C.GTK_STOCK_REDO
	STOCK_REFRESH                       Stock = C.GTK_STOCK_REFRESH
	STOCK_REMOVE                        Stock = C.GTK_STOCK_REMOVE
	STOCK_REVERT_TO_SAVED               Stock = C.GTK_STOCK_REVERT_TO_SAVED
	STOCK_SAVE                          Stock = C.GTK_STOCK_SAVE
	STOCK_SAVE_AS                       Stock = C.GTK_STOCK_SAVE_AS
	STOCK_SELECT_ALL                    Stock = C.GTK_STOCK_SELECT_ALL
	STOCK_SELECT_COLOR                  Stock = C.GTK_STOCK_SELECT_COLOR
	STOCK_SELECT_FONT                   Stock = C.GTK_STOCK_SELECT_FONT
	STOCK_SORT_ASCENDING                Stock = C.GTK_STOCK_SORT_ASCENDING
	STOCK_SORT_DESCENDING               Stock = C.GTK_STOCK_SORT_DESCENDING
	STOCK_SPELL_CHECK                   Stock = C.GTK_STOCK_SPELL_CHECK
	STOCK_STOP                          Stock = C.GTK_STOCK_STOP
	STOCK_STRIKETHROUGH                 Stock = C.GTK_STOCK_STRIKETHROUGH
	STOCK_UNDELETE                      Stock = C.GTK_STOCK_UNDELETE
	STOCK_UNDERLINE                     Stock = C.GTK_STOCK_UNDERLINE
	STOCK_UNDO                          Stock = C.GTK_STOCK_UNDO
	STOCK_UNINDENT                      Stock = C.GTK_STOCK_UNINDENT
	STOCK_YES                           Stock = C.GTK_STOCK_YES
	STOCK_ZOOM_100                      Stock = C.GTK_STOCK_ZOOM_100
	STOCK_ZOOM_FIT                      Stock = C.GTK_STOCK_ZOOM_FIT
	STOCK_ZOOM_IN                       Stock = C.GTK_STOCK_ZOOM_IN
	STOCK_ZOOM_OUT                      Stock = C.GTK_STOCK_ZOOM_OUT
)

// ReshowWithInitialSize is a wrapper around
// gtk_window_reshow_with_initial_size().
func (v *Window) ReshowWithInitialSize() {
	C.gtk_window_reshow_with_initial_size(v.native())
}
