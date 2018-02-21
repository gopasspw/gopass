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

// #cgo pkg-config: pango
// #include <pango/pango.h>
// #include "pango.go.h"
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

func init() {
	tm := []glib.TypeMarshaler{
	// Enums
	//		{glib.Type(C.pango_alignment_get_type()), marshalAlignment},
	//		{glib.Type(C.pango_ellipsize_mode_get_type()), marshalEllipsizeMode},
	//		{glib.Type(C.pango_wrap_mode_get_type()), marshalWrapMode},

	// Objects/Interfaces
	// {glib.Type(C.pango_context_get_type()), marshalContext},
	}
	glib.RegisterGValueMarshalers(tm)
}

// Context is a representation of PangoContext.
type Context struct {
	pangoContext *C.PangoContext
}

// Native returns a pointer to the underlying PangoLayout.
func (v *Context) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *Context) native() *C.PangoContext {
	return (*C.PangoContext)(unsafe.Pointer(v.pangoContext))
}

/*
func marshalContext(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := wrapObject(unsafe.Pointer(c))
	return wrapContext(obj), nil
}

func wrapContext(obj *glib.Object) *Context {
	return &Context{obj}
}
*/
func WrapContext(p uintptr) *Context {
	context := new(Context)
	context.pangoContext = (*C.PangoContext)(unsafe.Pointer(p))
	return context
}

//PangoContext *pango_context_new           (void);
func ContextNew() *Context {
	c := C.pango_context_new()

	context := new(Context)
	context.pangoContext = (*C.PangoContext)(c)

	return context
}

//void          pango_context_changed       (PangoContext                 *context);
//void          pango_context_set_font_map  (PangoContext                 *context,
//					   PangoFontMap                 *font_map);
//PangoFontMap *pango_context_get_font_map  (PangoContext                 *context);
//guint         pango_context_get_serial    (PangoContext                 *context);
//void          pango_context_list_families (PangoContext                 *context,
//					   PangoFontFamily            ***families,
//					   int                          *n_families);
//PangoFont *   pango_context_load_font     (PangoContext                 *context,
//					   const PangoFontDescription   *desc);
//PangoFontset *pango_context_load_fontset  (PangoContext                 *context,
//					   const PangoFontDescription   *desc,
//					   PangoLanguage                *language);
//
//PangoFontMetrics *pango_context_get_metrics   (PangoContext                 *context,
//					       const PangoFontDescription   *desc,
//					       PangoLanguage                *language);
//
//void                      pango_context_set_font_description (PangoContext               *context,
//							      const PangoFontDescription *desc);
//PangoFontDescription *    pango_context_get_font_description (PangoContext               *context);
//PangoLanguage            *pango_context_get_language         (PangoContext               *context);
//void                      pango_context_set_language         (PangoContext               *context,
//							      PangoLanguage              *language);
//void                      pango_context_set_base_dir         (PangoContext               *context,
//							      PangoDirection              direction);
//PangoDirection            pango_context_get_base_dir         (PangoContext               *context);
//void                      pango_context_set_base_gravity     (PangoContext               *context,
//							      PangoGravity                gravity);
//PangoGravity              pango_context_get_base_gravity     (PangoContext               *context);
//PangoGravity              pango_context_get_gravity          (PangoContext               *context);
//void                      pango_context_set_gravity_hint     (PangoContext               *context,
//							      PangoGravityHint            hint);
//PangoGravityHint          pango_context_get_gravity_hint     (PangoContext               *context);
//
//void                      pango_context_set_matrix           (PangoContext      *context,
//						              const PangoMatrix *matrix);
//const PangoMatrix *       pango_context_get_matrix           (PangoContext      *context);

/* Break a string of Unicode characters into segments with
 * consistent shaping/language engine and bidrectional level.
 * Returns a #GList of #PangoItem's
 */
//GList *pango_itemize                (PangoContext      *context,
//				     const char        *text,
//				     int                start_index,
//				     int                length,
//				     PangoAttrList     *attrs,
//				     PangoAttrIterator *cached_iter);
//GList *pango_itemize_with_base_dir  (PangoContext      *context,
//				     PangoDirection     base_dir,
//				     const char        *text,
//				     int                start_index,
//				     int                length,
//				     PangoAttrList     *attrs,
//				     PangoAttrIterator *cached_iter);
