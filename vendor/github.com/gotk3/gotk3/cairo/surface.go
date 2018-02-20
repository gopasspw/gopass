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

// TODO(jrick) SetUserData (depends on UserDataKey and DestroyFunc)

// TODO(jrick) GetUserData (depends on UserDataKey)

/*
 * cairo_surface_t
 */

// Surface is a representation of Cairo's cairo_surface_t.
type Surface struct {
	surface *C.cairo_surface_t
}

func NewSurfaceFromPNG(fileName string) (*Surface, error) {

	cstr := C.CString(fileName)
	defer C.free(unsafe.Pointer(cstr))

	surfaceNative := C.cairo_image_surface_create_from_png(cstr)

	status := Status(C.cairo_surface_status(surfaceNative))
	if status != STATUS_SUCCESS {
		return nil, ErrorStatus(status)
	}

	return &Surface{surfaceNative}, nil
}

// CreateImageSurface is a wrapper around cairo_image_surface_create().
func CreateImageSurface(format Format, width, height int) *Surface {
	c := C.cairo_image_surface_create(C.cairo_format_t(format),
		C.int(width), C.int(height))
	s := wrapSurface(c)
	runtime.SetFinalizer(s, (*Surface).destroy)
	return s
}

// native returns a pointer to the underlying cairo_surface_t.
func (v *Surface) native() *C.cairo_surface_t {
	if v == nil {
		return nil
	}
	return v.surface
}

// Native returns a pointer to the underlying cairo_surface_t.
func (v *Surface) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalSurface(p uintptr) (interface{}, error) {
	c := C.g_value_get_boxed((*C.GValue)(unsafe.Pointer(p)))
	surface := (*C.cairo_surface_t)(unsafe.Pointer(c))
	return wrapSurface(surface), nil
}

func wrapSurface(surface *C.cairo_surface_t) *Surface {
	return &Surface{surface}
}

// NewSurface creates a gotk3 cairo Surface from a pointer to a
// C cairo_surface_t.  This is primarily designed for use with other
// gotk3 packages and should be avoided by applications.
func NewSurface(s uintptr, needsRef bool) *Surface {
	ptr := (*C.cairo_surface_t)(unsafe.Pointer(s))
	surface := wrapSurface(ptr)
	if needsRef {
		surface.reference()
	}
	runtime.SetFinalizer(surface, (*Surface).destroy)
	return surface
}

// CreateSimilar is a wrapper around cairo_surface_create_similar().
func (v *Surface) CreateSimilar(content Content, width, height int) *Surface {
	c := C.cairo_surface_create_similar(v.native(),
		C.cairo_content_t(content), C.int(width), C.int(height))
	s := wrapSurface(c)
	runtime.SetFinalizer(s, (*Surface).destroy)
	return s
}

// TODO cairo_surface_create_similar_image (since 1.12)

// CreateForRectangle is a wrapper around cairo_surface_create_for_rectangle().
func (v *Surface) CreateForRectangle(x, y, width, height float64) *Surface {
	c := C.cairo_surface_create_for_rectangle(v.native(), C.double(x),
		C.double(y), C.double(width), C.double(height))
	s := wrapSurface(c)
	runtime.SetFinalizer(s, (*Surface).destroy)
	return s
}

// reference is a wrapper around cairo_surface_reference().
func (v *Surface) reference() {
	v.surface = C.cairo_surface_reference(v.native())
}

// destroy is a wrapper around cairo_surface_destroy().
func (v *Surface) destroy() {
	C.cairo_surface_destroy(v.native())
}

// Status is a wrapper around cairo_surface_status().
func (v *Surface) Status() Status {
	c := C.cairo_surface_status(v.native())
	return Status(c)
}

// Flush is a wrapper around cairo_surface_flush().
func (v *Surface) Flush() {
	C.cairo_surface_flush(v.native())
}

// TODO(jrick) GetDevice (requires Device bindings)

// TODO(jrick) GetFontOptions (require FontOptions bindings)

// TODO(jrick) GetContent (requires Content bindings)

// MarkDirty is a wrapper around cairo_surface_mark_dirty().
func (v *Surface) MarkDirty() {
	C.cairo_surface_mark_dirty(v.native())
}

// MarkDirtyRectangle is a wrapper around cairo_surface_mark_dirty_rectangle().
func (v *Surface) MarkDirtyRectangle(x, y, width, height int) {
	C.cairo_surface_mark_dirty_rectangle(v.native(), C.int(x), C.int(y),
		C.int(width), C.int(height))
}

// SetDeviceOffset is a wrapper around cairo_surface_set_device_offset().
func (v *Surface) SetDeviceOffset(x, y float64) {
	C.cairo_surface_set_device_offset(v.native(), C.double(x), C.double(y))
}

// GetDeviceOffset is a wrapper around cairo_surface_get_device_offset().
func (v *Surface) GetDeviceOffset() (x, y float64) {
	var xOffset, yOffset C.double
	C.cairo_surface_get_device_offset(v.native(), &xOffset, &yOffset)
	return float64(xOffset), float64(yOffset)
}

// SetFallbackResolution is a wrapper around
// cairo_surface_set_fallback_resolution().
func (v *Surface) SetFallbackResolution(xPPI, yPPI float64) {
	C.cairo_surface_set_fallback_resolution(v.native(), C.double(xPPI),
		C.double(yPPI))
}

// GetFallbackResolution is a wrapper around
// cairo_surface_get_fallback_resolution().
func (v *Surface) GetFallbackResolution() (xPPI, yPPI float64) {
	var x, y C.double
	C.cairo_surface_get_fallback_resolution(v.native(), &x, &y)
	return float64(x), float64(y)
}

// GetType is a wrapper around cairo_surface_get_type().
func (v *Surface) GetType() SurfaceType {
	c := C.cairo_surface_get_type(v.native())
	return SurfaceType(c)
}

// TODO(jrick) SetUserData (depends on UserDataKey and DestroyFunc)

// TODO(jrick) GetUserData (depends on UserDataKey)

// CopyPage is a wrapper around cairo_surface_copy_page().
func (v *Surface) CopyPage() {
	C.cairo_surface_copy_page(v.native())
}

// ShowPage is a wrapper around cairo_surface_show_page().
func (v *Surface) ShowPage() {
	C.cairo_surface_show_page(v.native())
}

// HasShowTextGlyphs is a wrapper around cairo_surface_has_show_text_glyphs().
func (v *Surface) HasShowTextGlyphs() bool {
	c := C.cairo_surface_has_show_text_glyphs(v.native())
	return gobool(c)
}

// TODO(jrick) SetMimeData (depends on DestroyFunc)

// GetMimeData is a wrapper around cairo_surface_get_mime_data().  The
// returned mimetype data is returned as a Go byte slice.
func (v *Surface) GetMimeData(mimeType MimeType) []byte {
	cstr := C.CString(string(mimeType))
	defer C.free(unsafe.Pointer(cstr))
	var data *C.uchar
	var length C.ulong
	C.cairo_surface_get_mime_data(v.native(), cstr, &data, &length)
	return C.GoBytes(unsafe.Pointer(data), C.int(length))
}

// TODO(jrick) SupportsMimeType (since 1.12)

// TODO(jrick) MapToImage (since 1.12)

// TODO(jrick) UnmapImage (since 1.12)
