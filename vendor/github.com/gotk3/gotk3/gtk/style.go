// Same copyright and license as the rest of the files in this project
// This file contains style related functions and structures

package gtk

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"unsafe"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
)

type StyleProviderPriority int

const (
	STYLE_PROVIDER_PRIORITY_FALLBACK    StyleProviderPriority = C.GTK_STYLE_PROVIDER_PRIORITY_FALLBACK
	STYLE_PROVIDER_PRIORITY_THEME                             = C.GTK_STYLE_PROVIDER_PRIORITY_THEME
	STYLE_PROVIDER_PRIORITY_SETTINGS                          = C.GTK_STYLE_PROVIDER_PRIORITY_SETTINGS
	STYLE_PROVIDER_PRIORITY_APPLICATION                       = C.GTK_STYLE_PROVIDER_PRIORITY_APPLICATION
	STYLE_PROVIDER_PRIORITY_USER                              = C.GTK_STYLE_PROVIDER_PRIORITY_USER
)

/*
 * GtkStyleContext
 */

// StyleContext is a representation of GTK's GtkStyleContext.
type StyleContext struct {
	*glib.Object
}

// native returns a pointer to the underlying GtkStyleContext.
func (v *StyleContext) native() *C.GtkStyleContext {
	if v == nil || v.Object == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkStyleContext(p)
}

func wrapStyleContext(obj *glib.Object) *StyleContext {
	return &StyleContext{obj}
}

func (v *StyleContext) AddClass(class_name string) {
	cstr := C.CString(class_name)
	defer C.free(unsafe.Pointer(cstr))

	C.gtk_style_context_add_class(v.native(), (*C.gchar)(cstr))
}

func (v *StyleContext) RemoveClass(class_name string) {
	cstr := C.CString(class_name)
	defer C.free(unsafe.Pointer(cstr))

	C.gtk_style_context_remove_class(v.native(), (*C.gchar)(cstr))
}

func fromNativeStyleContext(c *C.GtkStyleContext) (*StyleContext, error) {
	if c == nil {
		return nil, nilPtrErr
	}

	obj := glib.Take(unsafe.Pointer(c))
	return wrapStyleContext(obj), nil
}

// GetStyleContext is a wrapper around gtk_widget_get_style_context().
func (v *Widget) GetStyleContext() (*StyleContext, error) {
	return fromNativeStyleContext(C.gtk_widget_get_style_context(v.native()))
}

// GetParent is a wrapper around gtk_style_context_get_parent().
func (v *StyleContext) GetParent() (*StyleContext, error) {
	return fromNativeStyleContext(C.gtk_style_context_get_parent(v.native()))
}

// GetProperty is a wrapper around gtk_style_context_get_property().
func (v *StyleContext) GetProperty(property string, state StateFlags) (interface{}, error) {
	cstr := (*C.gchar)(C.CString(property))
	defer C.free(unsafe.Pointer(cstr))

	var gval C.GValue
	C.gtk_style_context_get_property(v.native(), cstr, C.GtkStateFlags(state), &gval)
	val := glib.ValueFromNative(unsafe.Pointer(&gval))
	return val.GoValue()
}

// GetStyleProperty is a wrapper around gtk_style_context_get_style_property().
func (v *StyleContext) GetStyleProperty(property string) (interface{}, error) {
	cstr := (*C.gchar)(C.CString(property))
	defer C.free(unsafe.Pointer(cstr))

	var gval C.GValue
	C.gtk_style_context_get_style_property(v.native(), cstr, &gval)
	val := glib.ValueFromNative(unsafe.Pointer(&gval))
	return val.GoValue()
}

// GetScreen is a wrapper around gtk_style_context_get_screen().
func (v *StyleContext) GetScreen() (*gdk.Screen, error) {
	c := C.gtk_style_context_get_screen(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	d := &gdk.Screen{glib.Take(unsafe.Pointer(c))}
	return d, nil
}

// GetState is a wrapper around gtk_style_context_get_state().
func (v *StyleContext) GetState() StateFlags {
	return StateFlags(C.gtk_style_context_get_state(v.native()))
}

// GetColor is a wrapper around gtk_style_context_get_color().
func (v *StyleContext) GetColor(state StateFlags) *gdk.RGBA {
	gdkColor := gdk.NewRGBA()
	C.gtk_style_context_get_color(v.native(), C.GtkStateFlags(state), (*C.GdkRGBA)(unsafe.Pointer(gdkColor.Native())))
	return gdkColor
}

// LookupColor is a wrapper around gtk_style_context_lookup_color().
func (v *StyleContext) LookupColor(colorName string) (*gdk.RGBA, bool) {
	cstr := (*C.gchar)(C.CString(colorName))
	defer C.free(unsafe.Pointer(cstr))
	gdkColor := gdk.NewRGBA()
	ret := C.gtk_style_context_lookup_color(v.native(), cstr, (*C.GdkRGBA)(unsafe.Pointer(gdkColor.Native())))
	return gdkColor, gobool(ret)
}

// StyleContextResetWidgets is a wrapper around gtk_style_context_reset_widgets().
func StyleContextResetWidgets(v *gdk.Screen) {
	C.gtk_style_context_reset_widgets((*C.GdkScreen)(unsafe.Pointer(v.Native())))
}

// Restore is a wrapper around gtk_style_context_restore().
func (v *StyleContext) Restore() {
	C.gtk_style_context_restore(v.native())
}

// Save is a wrapper around gtk_style_context_save().
func (v *StyleContext) Save() {
	C.gtk_style_context_save(v.native())
}

// SetParent is a wrapper around gtk_style_context_set_parent().
func (v *StyleContext) SetParent(p *StyleContext) {
	C.gtk_style_context_set_parent(v.native(), p.native())
}

// HasClass is a wrapper around gtk_style_context_has_class().
func (v *StyleContext) HasClass(className string) bool {
	cstr := C.CString(className)
	defer C.free(unsafe.Pointer(cstr))

	return gobool(C.gtk_style_context_has_class(v.native(), (*C.gchar)(cstr)))
}

// SetScreen is a wrapper around gtk_style_context_set_screen().
func (v *StyleContext) SetScreen(s *gdk.Screen) {
	C.gtk_style_context_set_screen(v.native(), (*C.GdkScreen)(unsafe.Pointer(s.Native())))
}

// SetState is a wrapper around gtk_style_context_set_state().
func (v *StyleContext) SetState(state StateFlags) {
	C.gtk_style_context_set_state(v.native(), C.GtkStateFlags(state))
}

type IStyleProvider interface {
	toStyleProvider() *C.GtkStyleProvider
}

// AddProvider is a wrapper around gtk_style_context_add_provider().
func (v *StyleContext) AddProvider(provider IStyleProvider, prio uint) {
	C.gtk_style_context_add_provider(v.native(), provider.toStyleProvider(), C.guint(prio))
}

// AddProviderForScreen is a wrapper around gtk_style_context_add_provider_for_screen().
func AddProviderForScreen(s *gdk.Screen, provider IStyleProvider, prio uint) {
	C.gtk_style_context_add_provider_for_screen((*C.GdkScreen)(unsafe.Pointer(s.Native())), provider.toStyleProvider(), C.guint(prio))
}

// RemoveProvider is a wrapper around gtk_style_context_remove_provider().
func (v *StyleContext) RemoveProvider(provider IStyleProvider) {
	C.gtk_style_context_remove_provider(v.native(), provider.toStyleProvider())
}

// RemoveProviderForScreen is a wrapper around gtk_style_context_remove_provider_for_screen().
func RemoveProviderForScreen(s *gdk.Screen, provider IStyleProvider) {
	C.gtk_style_context_remove_provider_for_screen((*C.GdkScreen)(unsafe.Pointer(s.Native())), provider.toStyleProvider())
}

// GtkStyleContext * 	gtk_style_context_new ()
// void 	gtk_style_context_get ()
// GtkTextDirection 	gtk_style_context_get_direction ()
// GtkJunctionSides 	gtk_style_context_get_junction_sides ()
// const GtkWidgetPath * 	gtk_style_context_get_path ()
// GdkFrameClock * 	gtk_style_context_get_frame_clock ()
// void 	gtk_style_context_get_style ()
// void 	gtk_style_context_get_style_valist ()
// void 	gtk_style_context_get_valist ()
// GtkCssSection * 	gtk_style_context_get_section ()
// void 	gtk_style_context_get_background_color ()
// void 	gtk_style_context_get_border_color ()
// void 	gtk_style_context_get_border ()
// void 	gtk_style_context_get_padding ()
// void 	gtk_style_context_get_margin ()
// const PangoFontDescription * 	gtk_style_context_get_font ()
// void 	gtk_style_context_invalidate ()
// gboolean 	gtk_style_context_state_is_running ()
// GtkIconSet * 	gtk_style_context_lookup_icon_set ()
// void 	gtk_style_context_cancel_animations ()
// void 	gtk_style_context_scroll_animations ()
// void 	gtk_style_context_notify_state_change ()
// void 	gtk_style_context_pop_animatable_region ()
// void 	gtk_style_context_push_animatable_region ()
// void 	gtk_style_context_set_background ()
// void 	gtk_style_context_set_direction ()
// void 	gtk_style_context_set_junction_sides ()
// void 	gtk_style_context_set_path ()
// void 	gtk_style_context_add_region ()
// void 	gtk_style_context_remove_region ()
// gboolean 	gtk_style_context_has_region ()
// GList * 	gtk_style_context_list_regions ()
// void 	gtk_style_context_set_frame_clock ()
// void 	gtk_style_context_set_scale ()
// gint 	gtk_style_context_get_scale ()
// GList * 	gtk_style_context_list_classes ()
