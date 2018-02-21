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
	//	"github.com/andre-hub/gotk3/glib"
	//	"github.com/andre-hub/gotk3/cairo"
	"unsafe"
)

// GlyphGeometry is a representation of PangoGlyphGeometry.
type GlyphGeometry struct {
	pangoGlyphGeometry *C.PangoGlyphGeometry
}

// Native returns a pointer to the underlying PangoLayout.
func (v *GlyphGeometry) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *GlyphGeometry) native() *C.PangoGlyphGeometry {
	return (*C.PangoGlyphGeometry)(unsafe.Pointer(v.pangoGlyphGeometry))
}

// GlyphVisAttr is a representation of PangoGlyphVisAttr.
type GlyphVisAttr struct {
	pangoGlyphVisAttr *C.PangoGlyphGeometry
}

// Native returns a pointer to the underlying PangoGlyphVisAttr.
func (v *GlyphVisAttr) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *GlyphVisAttr) native() *C.PangoGlyphVisAttr {
	return (*C.PangoGlyphVisAttr)(unsafe.Pointer(v.pangoGlyphVisAttr))
}

// GlyphInfo is a representation of PangoGlyphInfo.
type GlyphInfo struct {
	pangoGlyphInfo *C.PangoGlyphInfo
}

// Native returns a pointer to the underlying PangoGlyphInfo.
func (v *GlyphInfo) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *GlyphInfo) native() *C.PangoGlyphInfo {
	return (*C.PangoGlyphInfo)(unsafe.Pointer(v.pangoGlyphInfo))
}

// GlyphGeometry is a representation of PangoGlyphString.
type GlyphString struct {
	pangoGlyphString *C.PangoGlyphString
}

// Native returns a pointer to the underlying PangoGlyphString.
func (v *GlyphString) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func (v *GlyphString) native() *C.PangoGlyphString {
	return (*C.PangoGlyphString)(unsafe.Pointer(v.pangoGlyphString))
}
