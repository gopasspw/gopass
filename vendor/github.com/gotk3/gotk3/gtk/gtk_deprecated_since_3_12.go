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

// This file includes wrapers for symbols deprecated beginning with GTK 3.12,
// and should only be included in a build targeted intended to target GTK
// 3.10 or earlier.  To target an earlier build build, use the build tag
// gtk_MAJOR_MINOR.  For example, to target GTK 3.8, run
// 'go build -tags gtk_3_8'.
// +build gtk_3_6 gtk_3_8 gtk_3_10

package gtk

// #cgo pkg-config: gtk+-3.0
// #include <gtk/gtk.h>
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

/*
 * GtkDialog
 */

// GetActionArea() is a wrapper around gtk_dialog_get_action_area().
func (v *Dialog) GetActionArea() (*Widget, error) {
	c := C.gtk_dialog_get_action_area(v.native())
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapWidget(glib.Take(unsafe.Pointer(c))), nil
}

/*
 * GtkMessageDialog
 */

// GetImage is a wrapper around gtk_message_dialog_get_image().
func (v *MessageDialog) GetImage() (*Widget, error) {
	c := C.gtk_message_dialog_get_image(v.native())
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapWidget(glib.Take(unsafe.Pointer(c))), nil
}

// SetImage is a wrapper around gtk_message_dialog_set_image().
func (v *MessageDialog) SetImage(image IWidget) {
	C.gtk_message_dialog_set_image(v.native(), image.toWidget())
}

/*
 * GtkWidget
 */

// GetMarginLeft is a wrapper around gtk_widget_get_margin_left().
func (v *Widget) GetMarginLeft() int {
	c := C.gtk_widget_get_margin_left(v.native())
	return int(c)
}

// SetMarginLeft is a wrapper around gtk_widget_set_margin_left().
func (v *Widget) SetMarginLeft(margin int) {
	C.gtk_widget_set_margin_left(v.native(), C.gint(margin))
}

// GetMarginRight is a wrapper around gtk_widget_get_margin_right().
func (v *Widget) GetMarginRight() int {
	c := C.gtk_widget_get_margin_right(v.native())
	return int(c)
}

// SetMarginRight is a wrapper around gtk_widget_set_margin_right().
func (v *Widget) SetMarginRight(margin int) {
	C.gtk_widget_set_margin_right(v.native(), C.gint(margin))
}
