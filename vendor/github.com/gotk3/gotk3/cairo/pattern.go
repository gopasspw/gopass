package cairo

// #cgo pkg-config: cairo cairo-gobject
// #include <stdlib.h>
// #include <cairo.h>
// #include <cairo-gobject.h>
import "C"

import (
	"runtime"
	"unsafe"
)

/*
 * cairo_pattern_t
 */

// Pattern is a representation of Cairo's cairo_pattern_t.
type Pattern struct {
	pattern *C.cairo_pattern_t
}

func NewPatternFromRGB(red, green, blue float64) (*Pattern, error) {
	patternNative := C.cairo_pattern_create_rgb(C.double(red), C.double(green), C.double(blue))
	status := Status(C.cairo_pattern_status(patternNative))
	if status != STATUS_SUCCESS {
		return nil, ErrorStatus(status)
	}
	pattern := wrapPattern(patternNative)
	runtime.SetFinalizer(pattern, (*Pattern).destroy)
	return pattern, nil
}

func NewPatternFromRGBA(red, green, blue, alpha float64) (*Pattern, error) {
	patternNative := C.cairo_pattern_create_rgba(C.double(red), C.double(green), C.double(blue), C.double(alpha))
	status := Status(C.cairo_pattern_status(patternNative))
	if status != STATUS_SUCCESS {
		return nil, ErrorStatus(status)
	}
	pattern := wrapPattern(patternNative)
	runtime.SetFinalizer(pattern, (*Pattern).destroy)
	return pattern, nil
}

func NewPatternForSurface(s *Surface) (*Pattern, error) {
	patternNative := C.cairo_pattern_create_for_surface(s.native())
	status := Status(C.cairo_pattern_status(patternNative))
	if status != STATUS_SUCCESS {
		return nil, ErrorStatus(status)
	}
	pattern := wrapPattern(patternNative)
	runtime.SetFinalizer(pattern, (*Pattern).destroy)
	return pattern, nil
}

func NewPatternLinear(x0, y0, x1, y1 float64) (*Pattern, error) {
	patternNative := C.cairo_pattern_create_linear(C.double(x0), C.double(y0), C.double(x1), C.double(y1))
	status := Status(C.cairo_pattern_status(patternNative))
	if status != STATUS_SUCCESS {
		return nil, ErrorStatus(status)
	}
	pattern := wrapPattern(patternNative)
	runtime.SetFinalizer(pattern, (*Pattern).destroy)
	return pattern, nil
}

func NewPatternRadial(x0, y0, r0, x1, y1, r1 float64) (*Pattern, error) {
	patternNative := C.cairo_pattern_create_radial(C.double(x0), C.double(y0), C.double(r0),
		C.double(x1), C.double(y1), C.double(r1))
	status := Status(C.cairo_pattern_status(patternNative))
	if status != STATUS_SUCCESS {
		return nil, ErrorStatus(status)
	}
	pattern := wrapPattern(patternNative)
	runtime.SetFinalizer(pattern, (*Pattern).destroy)
	return pattern, nil
}

func (p *Pattern) AddColorStopRGB(offset, red, green, blue float64) error {
	C.cairo_pattern_add_color_stop_rgb(p.native(), C.double(offset),
		C.double(red), C.double(green), C.double(blue))
	status := Status(C.cairo_pattern_status(p.native()))
	if status != STATUS_SUCCESS {
		return ErrorStatus(status)
	}
	return nil
}

func (p *Pattern) AddColorStopRGBA(offset, red, green, blue, alpha float64) error {
	C.cairo_pattern_add_color_stop_rgba(p.native(), C.double(offset),
		C.double(red), C.double(green), C.double(blue), C.double(alpha))
	status := Status(C.cairo_pattern_status(p.native()))
	if status != STATUS_SUCCESS {
		return ErrorStatus(status)
	}
	return nil
}

func (v *Context) SetSource(p *Pattern) {
	C.cairo_set_source(v.native(), p.native())
}

// native returns a pointer to the underlying cairo_pattern_t.
func (v *Pattern) native() *C.cairo_pattern_t {
	if v == nil {
		return nil
	}
	return v.pattern
}

// Native returns a pointer to the underlying cairo_pattern_t.
func (v *Pattern) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalPattern(p uintptr) (interface{}, error) {
	c := C.g_value_get_boxed((*C.GValue)(unsafe.Pointer(p)))
	pattern := (*C.cairo_pattern_t)(unsafe.Pointer(c))
	return wrapPattern(pattern), nil
}

func wrapPattern(pattern *C.cairo_pattern_t) *Pattern {
	return &Pattern{pattern}
}

// reference is a wrapper around cairo_pattern_reference().
func (v *Pattern) reference() {
	v.pattern = C.cairo_pattern_reference(v.native())
}

// destroy is a wrapper around cairo_pattern_destroy().
func (v *Pattern) destroy() {
	C.cairo_pattern_destroy(v.native())
}

// Status is a wrapper around cairo_pattern_status().
func (v *Pattern) Status() Status {
	c := C.cairo_pattern_status(v.native())
	return Status(c)
}
