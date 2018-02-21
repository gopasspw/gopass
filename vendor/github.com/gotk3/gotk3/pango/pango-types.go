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
// #include <stdlib.h>
import "C"
import (
	"unsafe"
)

// LogAttr is a representation of PangoLogAttr.
type LogAttr struct {
	pangoLogAttr *C.PangoLogAttr
}

// Native returns a pointer to the underlying PangoLayout.
func (v *LogAttr) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *LogAttr) native() *C.PangoLogAttr {
	return (*C.PangoLogAttr)(unsafe.Pointer(v.pangoLogAttr))
}

// EngineLang is a representation of PangoEngineLang.
type EngineLang struct {
	pangoEngineLang *C.PangoEngineLang
}

// Native returns a pointer to the underlying PangoLayout.
func (v *EngineLang) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *EngineLang) native() *C.PangoEngineLang {
	return (*C.PangoEngineLang)(unsafe.Pointer(v.pangoEngineLang))
}

// EngineShape is a representation of PangoEngineShape.
type EngineShape struct {
	pangoEngineShape *C.PangoEngineShape
}

// Native returns a pointer to the underlying PangoLayout.
func (v *EngineShape) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *EngineShape) native() *C.PangoEngineShape {
	return (*C.PangoEngineShape)(unsafe.Pointer(v.pangoEngineShape))
}

// Font is a representation of PangoFont.
type Font struct {
	pangoFont *C.PangoFont
}

// Native returns a pointer to the underlying PangoLayout.
func (v *Font) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *Font) native() *C.PangoFont {
	return (*C.PangoFont)(unsafe.Pointer(v.pangoFont))
}

// FontMap is a representation of PangoFontMap.
type FontMap struct {
	pangoFontMap *C.PangoFontMap
}

// Native returns a pointer to the underlying PangoLayout.
func (v *FontMap) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *FontMap) native() *C.PangoFontMap {
	return (*C.PangoFontMap)(unsafe.Pointer(v.pangoFontMap))
}

func wrapFontMap(fontMap *C.PangoFontMap) *FontMap {
	return &FontMap{fontMap}
}

func WrapFontMap(p uintptr) *FontMap {
	fontMap := (*C.PangoFontMap)(unsafe.Pointer(p))
	return wrapFontMap(fontMap)
}

// Rectangle is a representation of PangoRectangle.
type Rectangle struct {
	pangoRectangle *C.PangoRectangle
}

// Native returns a pointer to the underlying PangoLayout.
func (v *Rectangle) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *Rectangle) native() *C.PangoRectangle {
	return (*C.PangoRectangle)(unsafe.Pointer(v.pangoRectangle))
}

// Glyph is a representation of PangoGlyph
type Glyph uint32

//void pango_extents_to_pixels (PangoRectangle *inclusive,
//			      PangoRectangle *nearest);
func (inclusive *Rectangle) ExtentsToPixels(nearest *Rectangle) {
	C.pango_extents_to_pixels(inclusive.native(), nearest.native())
}

func RectangleNew(x, y, width, height int) *Rectangle {
	r := new(Rectangle)
	r.pangoRectangle = C.createPangoRectangle((C.int)(x), (C.int)(y), (C.int)(width), (C.int)(height))
	return r
}
