package cairo

// #cgo pkg-config: cairo cairo-gobject
// #include <stdlib.h>
// #include <cairo.h>
// #include <cairo-gobject.h>
import "C"

import (
	"unsafe"
)

// FontSlant is a representation of Cairo's cairo_font_slant_t
type FontSlant int

const (
	FONT_SLANT_NORMAL  FontSlant = C.CAIRO_FONT_SLANT_NORMAL
	FONT_SLANT_ITALIC  FontSlant = C.CAIRO_FONT_SLANT_ITALIC
	FONT_SLANT_OBLIQUE FontSlant = C.CAIRO_FONT_SLANT_OBLIQUE
)

// FontWeight is a representation of Cairo's cairo_font_weight_t
type FontWeight int

const (
	FONT_WEIGHT_NORMAL FontWeight = C.CAIRO_FONT_WEIGHT_NORMAL
	FONT_WEIGHT_BOLD   FontWeight = C.CAIRO_FONT_WEIGHT_BOLD
)

func (v *Context) SelectFontFace(family string, slant FontSlant, weight FontWeight) {
	cstr := C.CString(family)
	defer C.free(unsafe.Pointer(cstr))
	C.cairo_select_font_face(v.native(), (*C.char)(cstr), C.cairo_font_slant_t(slant), C.cairo_font_weight_t(weight))
}

func (v *Context) SetFontSize(size float64) {
	C.cairo_set_font_size(v.native(), C.double(size))
}

// TODO: cairo_set_font_matrix

// TODO: cairo_get_font_matrix

// TODO: cairo_set_font_options

// TODO: cairo_get_font_options

// TODO: cairo_set_font_face

// TODO: cairo_get_font_face

// TODO: cairo_set_scaled_font

// TODO: cairo_get_scaled_font

func (v *Context) ShowText(utf8 string) {
	cstr := C.CString(utf8)
	defer C.free(unsafe.Pointer(cstr))
	C.cairo_show_text(v.native(), (*C.char)(cstr))
}

// TODO: cairo_show_glyphs

// TODO: cairo_show_text_glyphs

type FontExtents struct {
	Ascent      float64
	Descent     float64
	Height      float64
	MaxXAdvance float64
	MaxYAdvance float64
}

func (v *Context) FontExtents() FontExtents {
	var extents C.cairo_font_extents_t
	C.cairo_font_extents(v.native(), &extents)
	return FontExtents{
		Ascent:      float64(extents.ascent),
		Descent:     float64(extents.descent),
		Height:      float64(extents.height),
		MaxXAdvance: float64(extents.max_x_advance),
		MaxYAdvance: float64(extents.max_y_advance),
	}
}

type TextExtents struct {
	XBearing float64
	YBearing float64
	Width    float64
	Height   float64
	XAdvance float64
	YAdvance float64
}

func (v *Context) TextExtents(utf8 string) TextExtents {
	cstr := C.CString(utf8)
	defer C.free(unsafe.Pointer(cstr))
	var extents C.cairo_text_extents_t
	C.cairo_text_extents(v.native(), (*C.char)(cstr), &extents)
	return TextExtents{
		XBearing: float64(extents.x_bearing),
		YBearing: float64(extents.y_bearing),
		Width:    float64(extents.width),
		Height:   float64(extents.height),
		XAdvance: float64(extents.x_advance),
		YAdvance: float64(extents.y_advance),
	}
}

// TODO: cairo_glyph_extents

// TODO: cairo_toy_font_face_create

// TODO: cairo_toy_font_face_get_family

// TODO: cairo_toy_font_face_get_slant

// TODO: cairo_toy_font_face_get_weight

// TODO: cairo_glyph_allocate

// TODO: cairo_glyph_free

// TODO: cairo_text_cluster_allocate

// TODO: cairo_text_cluster_free
