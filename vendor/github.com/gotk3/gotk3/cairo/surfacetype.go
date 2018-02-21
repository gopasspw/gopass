package cairo

// #cgo pkg-config: cairo cairo-gobject
// #include <stdlib.h>
// #include <cairo.h>
// #include <cairo-gobject.h>
import "C"
import (
	"unsafe"
)

// SurfaceType is a representation of Cairo's cairo_surface_type_t.
type SurfaceType int

const (
	SURFACE_TYPE_IMAGE          SurfaceType = C.CAIRO_SURFACE_TYPE_IMAGE
	SURFACE_TYPE_PDF            SurfaceType = C.CAIRO_SURFACE_TYPE_PDF
	SURFACE_TYPE_PS             SurfaceType = C.CAIRO_SURFACE_TYPE_PS
	SURFACE_TYPE_XLIB           SurfaceType = C.CAIRO_SURFACE_TYPE_XLIB
	SURFACE_TYPE_XCB            SurfaceType = C.CAIRO_SURFACE_TYPE_XCB
	SURFACE_TYPE_GLITZ          SurfaceType = C.CAIRO_SURFACE_TYPE_GLITZ
	SURFACE_TYPE_QUARTZ         SurfaceType = C.CAIRO_SURFACE_TYPE_QUARTZ
	SURFACE_TYPE_WIN32          SurfaceType = C.CAIRO_SURFACE_TYPE_WIN32
	SURFACE_TYPE_BEOS           SurfaceType = C.CAIRO_SURFACE_TYPE_BEOS
	SURFACE_TYPE_DIRECTFB       SurfaceType = C.CAIRO_SURFACE_TYPE_DIRECTFB
	SURFACE_TYPE_SVG            SurfaceType = C.CAIRO_SURFACE_TYPE_SVG
	SURFACE_TYPE_OS2            SurfaceType = C.CAIRO_SURFACE_TYPE_OS2
	SURFACE_TYPE_WIN32_PRINTING SurfaceType = C.CAIRO_SURFACE_TYPE_WIN32_PRINTING
	SURFACE_TYPE_QUARTZ_IMAGE   SurfaceType = C.CAIRO_SURFACE_TYPE_QUARTZ_IMAGE
	SURFACE_TYPE_SCRIPT         SurfaceType = C.CAIRO_SURFACE_TYPE_SCRIPT
	SURFACE_TYPE_QT             SurfaceType = C.CAIRO_SURFACE_TYPE_QT
	SURFACE_TYPE_RECORDING      SurfaceType = C.CAIRO_SURFACE_TYPE_RECORDING
	SURFACE_TYPE_VG             SurfaceType = C.CAIRO_SURFACE_TYPE_VG
	SURFACE_TYPE_GL             SurfaceType = C.CAIRO_SURFACE_TYPE_GL
	SURFACE_TYPE_DRM            SurfaceType = C.CAIRO_SURFACE_TYPE_DRM
	SURFACE_TYPE_TEE            SurfaceType = C.CAIRO_SURFACE_TYPE_TEE
	SURFACE_TYPE_XML            SurfaceType = C.CAIRO_SURFACE_TYPE_XML
	SURFACE_TYPE_SKIA           SurfaceType = C.CAIRO_SURFACE_TYPE_SKIA
	SURFACE_TYPE_SUBSURFACE     SurfaceType = C.CAIRO_SURFACE_TYPE_SUBSURFACE
	// SURFACE_TYPE_COGL           SurfaceType = C.CAIRO_SURFACE_TYPE_COGL (since 1.12)
)

func marshalSurfaceType(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return SurfaceType(c), nil
}
