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
		{glib.Type(C.pango_attr_type_get_type()), marshalAttrType},
		{glib.Type(C.pango_underline_get_type()), marshalUnderline},
	}
	glib.RegisterGValueMarshalers(tm)
}

/* PangoColor */

// Color is a representation of PangoColor.
type Color struct {
	pangoColor *C.PangoColor
}

// Native returns a pointer to the underlying PangoColor.
func (v *Color) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *Color) native() *C.PangoColor {
	return (*C.PangoColor)(unsafe.Pointer(v.pangoColor))
}

func (v *Color) Set(red, green, blue uint16) {
	v.native().red = C.guint16(red)
	v.native().green = C.guint16(green)
	v.native().blue = C.guint16(blue)
}

func (v *Color) Get() (red, green, blue uint16) {
	return uint16(v.native().red), uint16(v.native().green), uint16(v.native().blue)
}

//PangoColor *pango_color_copy     (const PangoColor *src);
func (v *Color) Copy(c *Color) *Color {
	w := new(Color)
	w.pangoColor = C.pango_color_copy(v.native())
	return w
}

//void        pango_color_free     (PangoColor       *color);
func (v *Color) Free() {
	C.pango_color_free(v.native())
}

//gboolean    pango_color_parse    (PangoColor       *color,
//			  const char       *spec);
func (v *Color) Parse(spec string) bool {
	cstr := C.CString(spec)
	defer C.free(unsafe.Pointer(cstr))
	c := C.pango_color_parse(v.native(), (*C.char)(cstr))
	return gobool(c)
}

//gchar      *pango_color_to_string(const PangoColor *color);
func (v *Color) ToString() string {
	c := C.pango_color_to_string(v.native())
	return C.GoString((*C.char)(c))
}

/* ---  ---  --- Attributes ---  ---  ---  */

// AttrList is a representation of PangoAttrList.
type AttrList struct {
	pangoAttrList *C.PangoAttrList
}

// Native returns a pointer to the underlying PangoLayout.
func (v *AttrList) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *AttrList) native() *C.PangoAttrList {
	return (*C.PangoAttrList)(unsafe.Pointer(v.pangoAttrList))
}

// AttrType is a representation of Pango's PangoAttrType.
type AttrType int

const (
	ATTR_INVALID             AttrType = C.PANGO_ATTR_INVALID             /* 0 is an invalid attribute type */
	ATTR_LANGUAGE            AttrType = C.PANGO_ATTR_LANGUAGE            /* PangoAttrLanguage */
	ATTR_FAMILY              AttrType = C.PANGO_ATTR_FAMILY              /* PangoAttrString */
	ATTR_STYLE               AttrType = C.PANGO_ATTR_STYLE               /* PangoAttrInt */
	ATTR_WEIGHT              AttrType = C.PANGO_ATTR_WEIGHT              /* PangoAttrInt */
	ATTR_VARIANT             AttrType = C.PANGO_ATTR_VARIANT             /* PangoAttrInt */
	ATTR_STRETCH             AttrType = C.PANGO_ATTR_STRETCH             /* PangoAttrInt */
	ATTR_SIZE                AttrType = C.PANGO_ATTR_SIZE                /* PangoAttrSize */
	ATTR_FONT_DESC           AttrType = C.PANGO_ATTR_FONT_DESC           /* PangoAttrFontDesc */
	ATTR_FOREGROUND          AttrType = C.PANGO_ATTR_FOREGROUND          /* PangoAttrColor */
	ATTR_BACKGROUND          AttrType = C.PANGO_ATTR_BACKGROUND          /* PangoAttrColor */
	ATTR_UNDERLINE           AttrType = C.PANGO_ATTR_UNDERLINE           /* PangoAttrInt */
	ATTR_STRIKETHROUGH       AttrType = C.PANGO_ATTR_STRIKETHROUGH       /* PangoAttrInt */
	ATTR_RISE                AttrType = C.PANGO_ATTR_RISE                /* PangoAttrInt */
	ATTR_SHAPE               AttrType = C.PANGO_ATTR_SHAPE               /* PangoAttrShape */
	ATTR_SCALE               AttrType = C.PANGO_ATTR_SCALE               /* PangoAttrFloat */
	ATTR_FALLBACK            AttrType = C.PANGO_ATTR_FALLBACK            /* PangoAttrInt */
	ATTR_LETTER_SPACING      AttrType = C.PANGO_ATTR_LETTER_SPACING      /* PangoAttrInt */
	ATTR_UNDERLINE_COLOR     AttrType = C.PANGO_ATTR_UNDERLINE_COLOR     /* PangoAttrColor */
	ATTR_STRIKETHROUGH_COLOR AttrType = C.PANGO_ATTR_STRIKETHROUGH_COLOR /* PangoAttrColor */
	ATTR_ABSOLUTE_SIZE       AttrType = C.PANGO_ATTR_ABSOLUTE_SIZE       /* PangoAttrSize */
	ATTR_GRAVITY             AttrType = C.PANGO_ATTR_GRAVITY             /* PangoAttrInt */
	ATTR_GRAVITY_HINT        AttrType = C.PANGO_ATTR_GRAVITY_HINT        /* PangoAttrInt */

)

func marshalAttrType(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return AttrType(c), nil
}

// Underline is a representation of Pango's PangoUnderline.
type Underline int

const (
	UNDERLINE_NONE   Underline = C.PANGO_UNDERLINE_NONE
	UNDERLINE_SINGLE Underline = C.PANGO_UNDERLINE_SINGLE
	UNDERLINE_DOUBLE Underline = C.PANGO_UNDERLINE_DOUBLE
	UNDERLINE_LOW    Underline = C.PANGO_UNDERLINE_LOW
	UNDERLINE_ERROR  Underline = C.PANGO_UNDERLINE_ERROR
)

func marshalUnderline(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return Underline(c), nil
}

const (
	ATTR_INDEX_FROM_TEXT_BEGINNING uint = 0
	ATTR_INDEX_TO_TEXT_END         uint = C.G_MAXUINT
)

// Attribute is a representation of Pango's PangoAttribute.
type Attribute struct {
	pangoAttribute *C.PangoAttribute
	//start_index, end_index uint
}

// Native returns a pointer to the underlying PangoColor.
func (v *Attribute) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *Attribute) native() *C.PangoAttribute {
	return (*C.PangoAttribute)(unsafe.Pointer(v.pangoAttribute))
}

/*
//typedef gboolean (*PangoAttrFilterFunc) (PangoAttribute *attribute,
//					 gpointer        user_data);
func (v *Attribute) AttrFilterFunc(user_data uintptr) bool {
	c := C.PangoAttrFilterFunc(Attribute.native(), C.gpointer(user_data))
	return gobool(c)
}

//typedef gpointer (*PangoAttrDataCopyFunc) (gconstpointer user_data);
func AttrDataCopyFunc(user_data uintptr) uintptr {
	c := C.PangoAttrDataCopyFunc(C.gpointer(user_data))
	return uintptr(c)
}
*/

// AttrClass is a representation of Pango's PangoAttrClass.
type AttrClass struct {
	//PangoAttrType type;
}

// AttrString is a representation of Pango's PangoAttrString.
type AttrString struct {
	Attribute
	//char *value;
}

// AttrLanguage is a representation of Pango's PangoAttrLanguage.
type AttrLanguage struct {
	Attribute
	//PangoLanguage *value;
}

// AttrInt is a representation of Pango's PangoAttrInt.
type AttrInt struct {
	Attribute
	//int value;
}

// AttrFloat is a representation of Pango's PangoAttrFloat.
type AttrFloat struct {
	Attribute
	//double value;
}

// AttrColor is a representation of Pango's AttrColor.
type AttrColor struct {
	Attribute
	Color
}

// AttrSize is a representation of Pango's PangoAttrSize.
type AttrSize struct {
	Attribute
	//int size;
	//guint absolute : 1;
}

// AttrShape is a representation of Pango's PangoAttrShape.
type AttrShape struct {
	Attribute
	//PangoRectangle ink_rect;
	//PangoRectangle logical_rect;

	//gpointer              data;
	//PangoAttrDataCopyFunc copy_func;
	//GDestroyNotify        destroy_func;
}

// AttrFontDesc is a representation of Pango's PangoAttrFontDesc.
type AttrFontDesc struct {
	Attribute
	//PangoFontDescription *desc;
}

/*
PangoAttrType         pango_attr_type_register (const gchar        *name);
const char *          pango_attr_type_get_name (PangoAttrType       type) G_GNUC_CONST;

void             pango_attribute_init        (PangoAttribute       *attr,
					      const PangoAttrClass *klass);
PangoAttribute * pango_attribute_copy        (const PangoAttribute *attr);
void             pango_attribute_destroy     (PangoAttribute       *attr);
gboolean         pango_attribute_equal       (const PangoAttribute *attr1,
					      const PangoAttribute *attr2) G_GNUC_PURE;

PangoAttribute *pango_attr_language_new      (PangoLanguage              *language);
PangoAttribute *pango_attr_family_new        (const char                 *family);
PangoAttribute *pango_attr_foreground_new    (guint16                     red,
					      guint16                     green,
					      guint16                     blue);
PangoAttribute *pango_attr_background_new    (guint16                     red,
					      guint16                     green,
					      guint16                     blue);
PangoAttribute *pango_attr_size_new          (int                         size);
PangoAttribute *pango_attr_size_new_absolute (int                         size);
PangoAttribute *pango_attr_style_new         (PangoStyle                  style);
PangoAttribute *pango_attr_weight_new        (PangoWeight                 weight);
PangoAttribute *pango_attr_variant_new       (PangoVariant                variant);
PangoAttribute *pango_attr_stretch_new       (PangoStretch                stretch);
PangoAttribute *pango_attr_font_desc_new     (const PangoFontDescription *desc);

PangoAttribute *pango_attr_underline_new           (PangoUnderline underline);
PangoAttribute *pango_attr_underline_color_new     (guint16        red,
						    guint16        green,
						    guint16        blue);
PangoAttribute *pango_attr_strikethrough_new       (gboolean       strikethrough);
PangoAttribute *pango_attr_strikethrough_color_new (guint16        red,
						    guint16        green,
						    guint16        blue);

PangoAttribute *pango_attr_rise_new          (int                         rise);
PangoAttribute *pango_attr_scale_new         (double                      scale_factor);
PangoAttribute *pango_attr_fallback_new      (gboolean                    enable_fallback);
PangoAttribute *pango_attr_letter_spacing_new (int                        letter_spacing);

PangoAttribute *pango_attr_shape_new           (const PangoRectangle       *ink_rect,
						const PangoRectangle       *logical_rect);
PangoAttribute *pango_attr_shape_new_with_data (const PangoRectangle       *ink_rect,
						const PangoRectangle       *logical_rect,
						gpointer                    data,
						PangoAttrDataCopyFunc       copy_func,
						GDestroyNotify              destroy_func);

PangoAttribute *pango_attr_gravity_new      (PangoGravity     gravity);
PangoAttribute *pango_attr_gravity_hint_new (PangoGravityHint hint);

GType              pango_attr_list_get_type      (void) G_GNUC_CONST;
PangoAttrList *    pango_attr_list_new           (void);
PangoAttrList *    pango_attr_list_ref           (PangoAttrList  *list);
void               pango_attr_list_unref         (PangoAttrList  *list);
PangoAttrList *    pango_attr_list_copy          (PangoAttrList  *list);
void               pango_attr_list_insert        (PangoAttrList  *list,
						  PangoAttribute *attr);
void               pango_attr_list_insert_before (PangoAttrList  *list,
						  PangoAttribute *attr);
void               pango_attr_list_change        (PangoAttrList  *list,
						  PangoAttribute *attr);
void               pango_attr_list_splice        (PangoAttrList  *list,
						  PangoAttrList  *other,
						  gint            pos,
						  gint            len);

PangoAttrList *pango_attr_list_filter (PangoAttrList       *list,
				       PangoAttrFilterFunc  func,
				       gpointer             data);

PangoAttrIterator *pango_attr_list_get_iterator  (PangoAttrList  *list);

void               pango_attr_iterator_range    (PangoAttrIterator     *iterator,
						 gint                  *start,
						 gint                  *end);
gboolean           pango_attr_iterator_next     (PangoAttrIterator     *iterator);
PangoAttrIterator *pango_attr_iterator_copy     (PangoAttrIterator     *iterator);
void               pango_attr_iterator_destroy  (PangoAttrIterator     *iterator);
PangoAttribute *   pango_attr_iterator_get      (PangoAttrIterator     *iterator,
						 PangoAttrType          type);
void               pango_attr_iterator_get_font (PangoAttrIterator     *iterator,
						 PangoFontDescription  *desc,
						 PangoLanguage        **language,
						 GSList               **extra_attrs);
GSList *          pango_attr_iterator_get_attrs (PangoAttrIterator     *iterator);


gboolean pango_parse_markup (const char                 *markup_text,
			     int                         length,
			     gunichar                    accel_marker,
			     PangoAttrList             **attr_list,
			     char                      **text,
			     gunichar                   *accel_char,
			     GError                    **error);

GMarkupParseContext * pango_markup_parser_new (gunichar               accel_marker);
gboolean              pango_markup_parser_finish (GMarkupParseContext   *context,
                                                  PangoAttrList        **attr_list,
                                                  char                 **text,
                                                  gunichar              *accel_char,
                                                  GError               **error);
*/
