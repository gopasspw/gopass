/*
 * Copyright (c) 2015- terrak <terrak1975@gmail.com>
 *
 * This file originated from: http://www.terrak.net/
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package pango

// #cgo pkg-config: pango pangocairo
// #include <pango/pango.h>
// #include <cairo.h>
// #include <pango/pangocairo.h>
// #include "pango.go.h"
import "C"
import (
	//	"github.com/gotk3/gotk3/glib"
	"unsafe"

	"github.com/gotk3/gotk3/cairo"
)

func init() {
	//	tm := []glib.TypeMarshaler{
	//		// Enums
	//		{glib.Type(C.pango_alignement_get_type()), marshalAlignment},
	//		{glib.Type(C.pango_ellipsize_mode_get_type()), marshalEllipsizeMode},
	//		{glib.Type(C.pango_wrap_mode_get_type()), marshalWrapMode},
	//	}
	//	glib.RegisterGValueMarshalers(tm)
}

func cairo_context(cr *cairo.Context) *C.cairo_t {
	return (*C.cairo_t)(cr.GetCContext())
}

/* Convenience
 */
//PangoContext *pango_cairo_create_context (cairo_t   *cr);
func CairoCreateContext(cr *cairo.Context) *Context {
	c := C.pango_cairo_create_context(cairo_context(cr))
	context := new(Context)
	context.pangoContext = (*C.PangoContext)(c)
	return context
}

//PangoLayout *pango_cairo_create_layout (cairo_t     *cr);
func CairoCreateLayout(cr *cairo.Context) *Layout {
	c := C.pango_cairo_create_layout(cairo_context(cr))
	layout := new(Layout)
	layout.pangoLayout = (*C.PangoLayout)(c)
	return layout
}

//void         pango_cairo_update_layout (cairo_t     *cr,
//					PangoLayout *layout);
func CairoUpdateLayout(cr *cairo.Context, v *Layout) {
	C.pango_cairo_update_layout(cairo_context(cr), v.native())
}

/*
 * Rendering
 */
//void pango_cairo_show_glyph_string (cairo_t          *cr,
//				    PangoFont        *font,
//				    PangoGlyphString *glyphs);
func CairoShowGlyphString(cr *cairo.Context, font *Font, glyphs *GlyphString) {
	C.pango_cairo_show_glyph_string(cairo_context(cr), font.native(), glyphs.native())
}

//void pango_cairo_show_glyph_item   (cairo_t          *cr,
//				    const char       *text,
//				    PangoGlyphItem   *glyph_item);
func CairoShowGlyphItem(cr *cairo.Context, text string, glyph_item *GlyphItem) {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	C.pango_cairo_show_glyph_item(cairo_context(cr), (*C.char)(cstr), glyph_item.native())
}

//void pango_cairo_show_layout_line  (cairo_t          *cr,
//				    PangoLayoutLine  *line);
func CairoShowLayoutLine(cr *cairo.Context, line *LayoutLine) {
	C.pango_cairo_show_layout_line(cairo_context(cr), line.native())
}

//void pango_cairo_show_layout       (cairo_t          *cr,
//				    PangoLayout      *layout);
func CairoShowLayout(cr *cairo.Context, layout *Layout) {
	C.pango_cairo_show_layout(cairo_context(cr), layout.native())
}

//void pango_cairo_show_error_underline (cairo_t       *cr,
//				       double         x,
//				       double         y,
//				       double         width,
//				       double         height);

/*
 * Rendering to a path
 */

//void pango_cairo_glyph_string_path (cairo_t          *cr,
//				    PangoFont        *font,
//				    PangoGlyphString *glyphs);
func CairoGlyphStringPath(cr *cairo.Context, font *Font, glyphs *GlyphString) {
	C.pango_cairo_glyph_string_path(cairo_context(cr), font.native(), glyphs.native())
}

//void pango_cairo_layout_line_path  (cairo_t          *cr,
//				    PangoLayoutLine  *line);
func CairoLayoutLinePath(cr *cairo.Context, line *LayoutLine) {
	C.pango_cairo_layout_line_path(cairo_context(cr), line.native())
}

//void pango_cairo_layout_path       (cairo_t          *cr,
//				    PangoLayout      *layout);
func CairoLayoutPath(cr *cairo.Context, layout *Layout) {
	C.pango_cairo_layout_path(cairo_context(cr), layout.native())
}

//void pango_cairo_error_underline_path (cairo_t       *cr,
//				       double         x,
//				       double         y,
//				       double         width,
//				       double         height);
func CairoErrorUnderlinePath(cr *cairo.Context, x, y, width, height float64) {
	C.pango_cairo_error_underline_path(cairo_context(cr), C.double(x), C.double(y), C.double(width), C.double(height))
}
