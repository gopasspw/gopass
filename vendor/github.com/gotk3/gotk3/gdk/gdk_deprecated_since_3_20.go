//+build gtk_3_6 gtk_3_8 gtk_3_10 gtk_3_12 gtk_3_14 gtk_3_16 gtk_3_18

package gdk

// #cgo pkg-config: gdk-3.0
// #include <gdk/gdk.h>
import "C"
import (
	"runtime"
	"unsafe"

	"github.com/gotk3/gotk3/glib"
)

// Grab() is a wrapper around gdk_device_grab().
func (v *Device) Grab(w *Window, ownership GrabOwnership, owner_events bool, event_mask EventMask, cursor *Cursor, time uint32) GrabStatus {
	ret := C.gdk_device_grab(
		v.native(),
		w.native(),
		C.GdkGrabOwnership(ownership),
		gbool(owner_events),
		C.GdkEventMask(event_mask),
		cursor.native(),
		C.guint32(time),
	)
	return GrabStatus(ret)
}

// GetClientPointer() is a wrapper around gdk_device_manager_get_client_pointer().
func (v *DeviceManager) GetClientPointer() (*Device, error) {
	c := C.gdk_device_manager_get_client_pointer(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	return &Device{glib.Take(unsafe.Pointer(c))}, nil
}

// ListDevices() is a wrapper around gdk_device_manager_list_devices().
func (v *DeviceManager) ListDevices(tp DeviceType) *glib.List {
	clist := C.gdk_device_manager_list_devices(v.native(), C.GdkDeviceType(tp))
	if clist == nil {
		return nil
	}

	//TODO: WrapList should set the finalizer
	glist := glib.WrapList(uintptr(unsafe.Pointer(clist)))
	glist.DataWrapper(func(ptr unsafe.Pointer) interface{} {
		return &Device{&glib.Object{glib.ToGObject(ptr)}}
	})
	runtime.SetFinalizer(glist, func(glist *glib.List) {
		glist.Free()
	})
	return glist
}

// Ungrab() is a wrapper around gdk_device_ungrab().
func (v *Device) Ungrab(time uint32) {
	C.gdk_device_ungrab(v.native(), C.guint32(time))
}

// GetDeviceManager() is a wrapper around gdk_display_get_device_manager().
func (v *Display) GetDeviceManager() (*DeviceManager, error) {
	c := C.gdk_display_get_device_manager(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	return &DeviceManager{glib.Take(unsafe.Pointer(c))}, nil
}

// GetScreen() is a wrapper around gdk_display_get_screen().
func (v *Display) GetScreen(screenNum int) (*Screen, error) {
	c := C.gdk_display_get_screen(v.native(), C.gint(screenNum))
	if c == nil {
		return nil, nilPtrErr
	}

	return &Screen{glib.Take(unsafe.Pointer(c))}, nil
}
