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

// Package cairo implements Go bindings for Cairo.  Supports version 1.10 and
// later.
package cairo

// #cgo pkg-config: cairo cairo-gobject
// #include <stdlib.h>
// #include <cairo.h>
// #include <cairo-gobject.h>
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

func init() {
	tm := []glib.TypeMarshaler{
		// Enums
		{glib.Type(C.cairo_gobject_antialias_get_type()), marshalAntialias},
		{glib.Type(C.cairo_gobject_content_get_type()), marshalContent},
		{glib.Type(C.cairo_gobject_fill_rule_get_type()), marshalFillRule},
		{glib.Type(C.cairo_gobject_line_cap_get_type()), marshalLineCap},
		{glib.Type(C.cairo_gobject_line_join_get_type()), marshalLineJoin},
		{glib.Type(C.cairo_gobject_operator_get_type()), marshalOperator},
		{glib.Type(C.cairo_gobject_status_get_type()), marshalStatus},
		{glib.Type(C.cairo_gobject_surface_type_get_type()), marshalSurfaceType},

		// Boxed
		{glib.Type(C.cairo_gobject_context_get_type()), marshalContext},
		{glib.Type(C.cairo_gobject_surface_get_type()), marshalSurface},
	}
	glib.RegisterGValueMarshalers(tm)
}

// Constants

// Content is a representation of Cairo's cairo_content_t.
type Content int

const (
	CONTENT_COLOR       Content = C.CAIRO_CONTENT_COLOR
	CONTENT_ALPHA       Content = C.CAIRO_CONTENT_ALPHA
	CONTENT_COLOR_ALPHA Content = C.CAIRO_CONTENT_COLOR_ALPHA
)

func marshalContent(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return Content(c), nil
}
