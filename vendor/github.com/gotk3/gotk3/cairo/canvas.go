package cairo

// #cgo pkg-config: cairo cairo-gobject
// #include <stdlib.h>
// #include <cairo.h>
// #include <cairo-gobject.h>
import "C"

import (
	"reflect"
	"runtime"
	"unsafe"
)

// Context is a representation of Cairo's cairo_t.
type Context struct {
	context *C.cairo_t
}

// native returns a pointer to the underlying cairo_t.
func (v *Context) native() *C.cairo_t {
	if v == nil {
		return nil
	}
	return v.context
}

func (v *Context) GetCContext() *C.cairo_t {
	return v.native()
}

// Native returns a pointer to the underlying cairo_t.
func (v *Context) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalContext(p uintptr) (interface{}, error) {
	c := C.g_value_get_boxed((*C.GValue)(unsafe.Pointer(p)))
	context := (*C.cairo_t)(unsafe.Pointer(c))
	return wrapContext(context), nil
}

func wrapContext(context *C.cairo_t) *Context {
	return &Context{context}
}

func WrapContext(p uintptr) *Context {
	context := (*C.cairo_t)(unsafe.Pointer(p))
	return wrapContext(context)
}

// Create is a wrapper around cairo_create().
func Create(target *Surface) *Context {
	c := C.cairo_create(target.native())
	ctx := wrapContext(c)
	runtime.SetFinalizer(ctx, (*Context).destroy)
	return ctx
}

// reference is a wrapper around cairo_reference().
func (v *Context) reference() {
	v.context = C.cairo_reference(v.native())
}

// destroy is a wrapper around cairo_destroy().
func (v *Context) destroy() {
	C.cairo_destroy(v.native())
}

// Status is a wrapper around cairo_status().
func (v *Context) Status() Status {
	c := C.cairo_status(v.native())
	return Status(c)
}

// Save is a wrapper around cairo_save().
func (v *Context) Save() {
	C.cairo_save(v.native())
}

// Restore is a wrapper around cairo_restore().
func (v *Context) Restore() {
	C.cairo_restore(v.native())
}

// GetTarget is a wrapper around cairo_get_target().
func (v *Context) GetTarget() *Surface {
	c := C.cairo_get_target(v.native())
	s := wrapSurface(c)
	s.reference()
	runtime.SetFinalizer(s, (*Surface).destroy)
	return s
}

// PushGroup is a wrapper around cairo_push_group().
func (v *Context) PushGroup() {
	C.cairo_push_group(v.native())
}

// PushGroupWithContent is a wrapper around cairo_push_group_with_content().
func (v *Context) PushGroupWithContent(content Content) {
	C.cairo_push_group_with_content(v.native(), C.cairo_content_t(content))
}

// TODO(jrick) PopGroup (depends on Pattern)

// PopGroupToSource is a wrapper around cairo_pop_group_to_source().
func (v *Context) PopGroupToSource() {
	C.cairo_pop_group_to_source(v.native())
}

// GetGroupTarget is a wrapper around cairo_get_group_target().
func (v *Context) GetGroupTarget() *Surface {
	c := C.cairo_get_group_target(v.native())
	s := wrapSurface(c)
	s.reference()
	runtime.SetFinalizer(s, (*Surface).destroy)
	return s
}

// SetSourceRGB is a wrapper around cairo_set_source_rgb().
func (v *Context) SetSourceRGB(red, green, blue float64) {
	C.cairo_set_source_rgb(v.native(), C.double(red), C.double(green),
		C.double(blue))
}

// SetSourceRGBA is a wrapper around cairo_set_source_rgba().
func (v *Context) SetSourceRGBA(red, green, blue, alpha float64) {
	C.cairo_set_source_rgba(v.native(), C.double(red), C.double(green),
		C.double(blue), C.double(alpha))
}

// TODO(jrick) SetSource (depends on Pattern)

// SetSourceSurface is a wrapper around cairo_set_source_surface().
func (v *Context) SetSourceSurface(surface *Surface, x, y float64) {
	C.cairo_set_source_surface(v.native(), surface.native(), C.double(x),
		C.double(y))
}

// TODO(jrick) GetSource (depends on Pattern)

// SetAntialias is a wrapper around cairo_set_antialias().
func (v *Context) SetAntialias(antialias Antialias) {
	C.cairo_set_antialias(v.native(), C.cairo_antialias_t(antialias))
}

// GetAntialias is a wrapper around cairo_get_antialias().
func (v *Context) GetAntialias() Antialias {
	c := C.cairo_get_antialias(v.native())
	return Antialias(c)
}

// SetDash is a wrapper around cairo_set_dash().
func (v *Context) SetDash(dashes []float64, offset float64) {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&dashes))
	cdashes := (*C.double)(unsafe.Pointer(header.Data))
	C.cairo_set_dash(v.native(), cdashes, C.int(header.Len),
		C.double(offset))
}

// GetDashCount is a wrapper around cairo_get_dash_count().
func (v *Context) GetDashCount() int {
	c := C.cairo_get_dash_count(v.native())
	return int(c)
}

// GetDash is a wrapper around cairo_get_dash().
func (v *Context) GetDash() (dashes []float64, offset float64) {
	dashCount := v.GetDashCount()
	cdashes := (*C.double)(C.calloc(8, C.size_t(dashCount)))
	var coffset C.double
	C.cairo_get_dash(v.native(), cdashes, &coffset)
	header := (*reflect.SliceHeader)((unsafe.Pointer(&dashes)))
	header.Data = uintptr(unsafe.Pointer(cdashes))
	header.Len = dashCount
	header.Cap = dashCount
	return dashes, float64(coffset)
}

// SetFillRule is a wrapper around cairo_set_fill_rule().
func (v *Context) SetFillRule(fillRule FillRule) {
	C.cairo_set_fill_rule(v.native(), C.cairo_fill_rule_t(fillRule))
}

// GetFillRule is a wrapper around cairo_get_fill_rule().
func (v *Context) GetFillRule() FillRule {
	c := C.cairo_get_fill_rule(v.native())
	return FillRule(c)
}

// SetLineCap is a wrapper around cairo_set_line_cap().
func (v *Context) SetLineCap(lineCap LineCap) {
	C.cairo_set_line_cap(v.native(), C.cairo_line_cap_t(lineCap))
}

// GetLineCap is a wrapper around cairo_get_line_cap().
func (v *Context) GetLineCap() LineCap {
	c := C.cairo_get_line_cap(v.native())
	return LineCap(c)
}

// SetLineJoin is a wrapper around cairo_set_line_join().
func (v *Context) SetLineJoin(lineJoin LineJoin) {
	C.cairo_set_line_join(v.native(), C.cairo_line_join_t(lineJoin))
}

// GetLineJoin is a wrapper around cairo_get_line_join().
func (v *Context) GetLineJoin() LineJoin {
	c := C.cairo_get_line_join(v.native())
	return LineJoin(c)
}

// SetLineWidth is a wrapper around cairo_set_line_width().
func (v *Context) SetLineWidth(width float64) {
	C.cairo_set_line_width(v.native(), C.double(width))
}

// GetLineWidth is a wrapper cairo_get_line_width().
func (v *Context) GetLineWidth() float64 {
	c := C.cairo_get_line_width(v.native())
	return float64(c)
}

// SetMiterLimit is a wrapper around cairo_set_miter_limit().
func (v *Context) SetMiterLimit(limit float64) {
	C.cairo_set_miter_limit(v.native(), C.double(limit))
}

// GetMiterLimit is a wrapper around cairo_get_miter_limit().
func (v *Context) GetMiterLimit() float64 {
	c := C.cairo_get_miter_limit(v.native())
	return float64(c)
}

// SetOperator is a wrapper around cairo_set_operator().
func (v *Context) SetOperator(op Operator) {
	C.cairo_set_operator(v.native(), C.cairo_operator_t(op))
}

// GetOperator is a wrapper around cairo_get_operator().
func (v *Context) GetOperator() Operator {
	c := C.cairo_get_operator(v.native())
	return Operator(c)
}

// SetTolerance is a wrapper around cairo_set_tolerance().
func (v *Context) SetTolerance(tolerance float64) {
	C.cairo_set_tolerance(v.native(), C.double(tolerance))
}

// GetTolerance is a wrapper around cairo_get_tolerance().
func (v *Context) GetTolerance() float64 {
	c := C.cairo_get_tolerance(v.native())
	return float64(c)
}

// Clip is a wrapper around cairo_clip().
func (v *Context) Clip() {
	C.cairo_clip(v.native())
}

// ClipPreserve is a wrapper around cairo_clip_preserve().
func (v *Context) ClipPreserve() {
	C.cairo_clip_preserve(v.native())
}

// ClipExtents is a wrapper around cairo_clip_extents().
func (v *Context) ClipExtents() (x1, y1, x2, y2 float64) {
	var cx1, cy1, cx2, cy2 C.double
	C.cairo_clip_extents(v.native(), &cx1, &cy1, &cx2, &cy2)
	return float64(cx1), float64(cy1), float64(cx2), float64(cy2)
}

// InClip is a wrapper around cairo_in_clip().
func (v *Context) InClip(x, y float64) bool {
	c := C.cairo_in_clip(v.native(), C.double(x), C.double(y))
	return gobool(c)
}

// ResetClip is a wrapper around cairo_reset_clip().
func (v *Context) ResetClip() {
	C.cairo_reset_clip(v.native())
}

// Rectangle is a wrapper around cairo_rectangle().
func (v *Context) Rectangle(x, y, w, h float64) {
	C.cairo_rectangle(v.native(), C.double(x), C.double(y), C.double(w), C.double(h))
}

// Arc is a wrapper around cairo_arc().
func (v *Context) Arc(xc, yc, radius, angle1, angle2 float64) {
	C.cairo_arc(v.native(), C.double(xc), C.double(yc), C.double(radius), C.double(angle1), C.double(angle2))
}

// ArcNegative is a wrapper around cairo_arc_negative().
func (v *Context) ArcNegative(xc, yc, radius, angle1, angle2 float64) {
	C.cairo_arc_negative(v.native(), C.double(xc), C.double(yc), C.double(radius), C.double(angle1), C.double(angle2))
}

// LineTo is a wrapper around cairo_line_to().
func (v *Context) LineTo(x, y float64) {
	C.cairo_line_to(v.native(), C.double(x), C.double(y))
}

// CurveTo is a wrapper around cairo_curve_to().
func (v *Context) CurveTo(x1, y1, x2, y2, x3, y3 float64) {
	C.cairo_curve_to(v.native(), C.double(x1), C.double(y1), C.double(x2), C.double(y2), C.double(x3), C.double(y3))
}

// MoveTo is a wrapper around cairo_move_to().
func (v *Context) MoveTo(x, y float64) {
	C.cairo_move_to(v.native(), C.double(x), C.double(y))
}

// TODO(jrick) CopyRectangleList (depends on RectangleList)

// Fill is a wrapper around cairo_fill().
func (v *Context) Fill() {
	C.cairo_fill(v.native())
}

// ClosePath is a wrapper around cairo_close_path().
func (v *Context) ClosePath() {
	C.cairo_close_path(v.native())
}

// NewPath is a wrapper around cairo_new_path().
func (v *Context) NewPath() {
	C.cairo_new_path(v.native())
}

// GetCurrentPoint is a wrapper around cairo_get_current_point().
func (v *Context) GetCurrentPoint() (x, y float64) {
	C.cairo_get_current_point(v.native(), (*C.double)(&x), (*C.double)(&y))
	return
}

// FillPreserve is a wrapper around cairo_fill_preserve().
func (v *Context) FillPreserve() {
	C.cairo_fill_preserve(v.native())
}

// FillExtents is a wrapper around cairo_fill_extents().
func (v *Context) FillExtents() (x1, y1, x2, y2 float64) {
	var cx1, cy1, cx2, cy2 C.double
	C.cairo_fill_extents(v.native(), &cx1, &cy1, &cx2, &cy2)
	return float64(cx1), float64(cy1), float64(cx2), float64(cy2)
}

// InFill is a wrapper around cairo_in_fill().
func (v *Context) InFill(x, y float64) bool {
	c := C.cairo_in_fill(v.native(), C.double(x), C.double(y))
	return gobool(c)
}

// TODO(jrick) Mask (depends on Pattern)

// MaskSurface is a wrapper around cairo_mask_surface().
func (v *Context) MaskSurface(surface *Surface, surfaceX, surfaceY float64) {
	C.cairo_mask_surface(v.native(), surface.native(), C.double(surfaceX),
		C.double(surfaceY))
}

// Paint is a wrapper around cairo_paint().
func (v *Context) Paint() {
	C.cairo_paint(v.native())
}

// PaintWithAlpha is a wrapper around cairo_paint_with_alpha().
func (v *Context) PaintWithAlpha(alpha float64) {
	C.cairo_paint_with_alpha(v.native(), C.double(alpha))
}

// Stroke is a wrapper around cairo_stroke().
func (v *Context) Stroke() {
	C.cairo_stroke(v.native())
}

// StrokePreserve is a wrapper around cairo_stroke_preserve().
func (v *Context) StrokePreserve() {
	C.cairo_stroke_preserve(v.native())
}

// StrokeExtents is a wrapper around cairo_stroke_extents().
func (v *Context) StrokeExtents() (x1, y1, x2, y2 float64) {
	var cx1, cy1, cx2, cy2 C.double
	C.cairo_stroke_extents(v.native(), &cx1, &cy1, &cx2, &cy2)
	return float64(cx1), float64(cy1), float64(cx2), float64(cy2)
}

// InStroke is a wrapper around cairo_in_stroke().
func (v *Context) InStroke(x, y float64) bool {
	c := C.cairo_in_stroke(v.native(), C.double(x), C.double(y))
	return gobool(c)
}

// CopyPage is a wrapper around cairo_copy_page().
func (v *Context) CopyPage() {
	C.cairo_copy_page(v.native())
}

// ShowPage is a wrapper around cairo_show_page().
func (v *Context) ShowPage() {
	C.cairo_show_page(v.native())
}
