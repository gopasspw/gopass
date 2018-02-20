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

// This file includes wrapers for symbols included since GTK 3.12, and
// and should not be included in a build intended to target any older GTK
// versions.  To target an older build, such as 3.10, use
// 'go build -tags gtk_3_10'.  Otherwise, if no build tags are used, GTK 3.12
// is assumed and this file is built.

// +build !gtk_3_6,!gtk_3_8,!gtk_3_10

package gtk

// #cgo pkg-config: gtk+-3.0
// #include <gtk/gtk.h>
// #include "popover_since_3_12.go.h"
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

func init() {
	tm := []glib.TypeMarshaler{
		{glib.Type(C.gtk_popover_get_type()), marshalPopover},
	}

	glib.RegisterGValueMarshalers(tm)
	WrapMap["GtkPopover"] = wrapPopover
}

//TODO(sjon): Implement GtkPopover
//GtkPopover
type Popover struct {
	Bin
}

func (v *Popover) native() *C.GtkPopover {
	if v == nil || v.GObject == nil {
		return nil
	}

	p := unsafe.Pointer(v.GObject)
	return C.toGtkPopover(p)
}

func marshalPopover(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return wrapPopover(glib.Take(unsafe.Pointer(c))), nil
}

func wrapPopover(obj *glib.Object) *Popover {
	return &Popover{Bin{Container{Widget{glib.InitiallyUnowned{obj}}}}}
}

//gtk_popover_new()
func PopoverNew(relative IWidget) (*Popover, error) {
	//Takes relative to widget
	var c *C.struct__GtkWidget
	if relative == nil {
		c = C.gtk_popover_new(nil)
	} else {
		c = C.gtk_popover_new(relative.toWidget())
	}
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapPopover(glib.Take(unsafe.Pointer(c))), nil
}
