package gtk

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"errors"
	"runtime"
	"sync"
	"unsafe"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/pango"
)

func init() {
	tm := []glib.TypeMarshaler{
		// Enums
		{glib.Type(C.gtk_page_orientation_get_type()), marshalPageOrientation},
		{glib.Type(C.gtk_print_error_get_type()), marshalPrintError},
		{glib.Type(C.gtk_print_operation_action_get_type()), marshalPrintOperationAction},
		{glib.Type(C.gtk_print_operation_result_get_type()), marshalPrintOperationResult},
		{glib.Type(C.gtk_print_status_get_type()), marshalPrintStatus},
		{glib.Type(C.gtk_unit_get_type()), marshalUnit},

		// Objects/Interfaces
		{glib.Type(C.gtk_number_up_layout_get_type()), marshalNumberUpLayout},
		{glib.Type(C.gtk_page_orientation_get_type()), marshalPageOrientation},
		{glib.Type(C.gtk_page_set_get_type()), marshalPageSet},
		{glib.Type(C.gtk_page_setup_get_type()), marshalPageSetup},
		{glib.Type(C.gtk_print_context_get_type()), marshalPrintContext},
		{glib.Type(C.gtk_print_duplex_get_type()), marshalPrintDuplex},
		{glib.Type(C.gtk_print_operation_get_type()), marshalPrintOperation},
		{glib.Type(C.gtk_print_operation_preview_get_type()), marshalPrintOperationPreview},
		{glib.Type(C.gtk_print_pages_get_type()), marshalPrintPages},
		{glib.Type(C.gtk_print_quality_get_type()), marshalPrintQuality},
		{glib.Type(C.gtk_print_settings_get_type()), marshalPrintSettings},

		// Boxed
		{glib.Type(C.gtk_paper_size_get_type()), marshalPaperSize},
	}

	glib.RegisterGValueMarshalers(tm)

	WrapMap["GtkPageSetup"] = wrapPageSetup
	WrapMap["GtkPrintContext"] = wrapPrintContext
	WrapMap["GtkPrintOperation"] = wrapPrintOperation
	WrapMap["GtkPrintOperationPreview"] = wrapPrintOperationPreview
	WrapMap["GtkPrintSettings"] = wrapPrintSettings
}

/*
 * Constants
 */

// NumberUpLayout is a representation of GTK's GtkNumberUpLayout.
type NumberUpLayout int

const (
	NUMBER_UP_LAYOUT_LEFT_TO_RIGHT_TOP_TO_BOTTOM NumberUpLayout = C.GTK_NUMBER_UP_LAYOUT_LEFT_TO_RIGHT_TOP_TO_BOTTOM
	NUMBER_UP_LAYOUT_LEFT_TO_RIGHT_BOTTOM_TO_TOP NumberUpLayout = C.GTK_NUMBER_UP_LAYOUT_LEFT_TO_RIGHT_BOTTOM_TO_TOP
	NUMBER_UP_LAYOUT_RIGHT_TO_LEFT_TOP_TO_BOTTOM NumberUpLayout = C.GTK_NUMBER_UP_LAYOUT_RIGHT_TO_LEFT_TOP_TO_BOTTOM
	NUMBER_UP_LAYOUT_RIGHT_TO_LEFT_BOTTOM_TO_TOP NumberUpLayout = C.GTK_NUMBER_UP_LAYOUT_RIGHT_TO_LEFT_BOTTOM_TO_TOP
	NUMBER_UP_LAYOUT_TOP_TO_BOTTOM_LEFT_TO_RIGHT NumberUpLayout = C.GTK_NUMBER_UP_LAYOUT_TOP_TO_BOTTOM_LEFT_TO_RIGHT
	NUMBER_UP_LAYOUT_TOP_TO_BOTTOM_RIGHT_TO_LEFT NumberUpLayout = C.GTK_NUMBER_UP_LAYOUT_TOP_TO_BOTTOM_RIGHT_TO_LEFT
	NUMBER_UP_LAYOUT_BOTTOM_TO_TOP_LEFT_TO_RIGHT NumberUpLayout = C.GTK_NUMBER_UP_LAYOUT_BOTTOM_TO_TOP_LEFT_TO_RIGHT
	NUMBER_UP_LAYOUT_BOTTOM_TO_TOP_RIGHT_TO_LEFT NumberUpLayout = C.GTK_NUMBER_UP_LAYOUT_BOTTOM_TO_TOP_RIGHT_TO_LEFT
)

func marshalNumberUpLayout(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return NumberUpLayout(c), nil
}

// PageOrientation is a representation of GTK's GtkPageOrientation.
type PageOrientation int

const (
	PAGE_ORIENTATION_PORTRAIT          PageOrientation = C.GTK_PAGE_ORIENTATION_PORTRAIT
	PAGE_ORIENTATION_LANDSCAPE         PageOrientation = C.GTK_PAGE_ORIENTATION_LANDSCAPE
	PAGE_ORIENTATION_REVERSE_PORTRAIT  PageOrientation = C.GTK_PAGE_ORIENTATION_REVERSE_PORTRAIT
	PAGE_ORIENTATION_REVERSE_LANDSCAPE PageOrientation = C.GTK_PAGE_ORIENTATION_REVERSE_LANDSCAPE
)

func marshalPageOrientation(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return PageOrientation(c), nil
}

// PrintDuplex is a representation of GTK's GtkPrintDuplex.
type PrintDuplex int

const (
	PRINT_DUPLEX_SIMPLEX    PrintDuplex = C.GTK_PRINT_DUPLEX_SIMPLEX
	PRINT_DUPLEX_HORIZONTAL PrintDuplex = C.GTK_PRINT_DUPLEX_HORIZONTAL
	PRINT_DUPLEX_VERTICAL   PrintDuplex = C.GTK_PRINT_DUPLEX_VERTICAL
)

func marshalPrintDuplex(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return PrintDuplex(c), nil
}

// PrintPages is a representation of GTK's GtkPrintPages.
type PrintPages int

const (
	PRINT_PAGES_ALL       PrintPages = C.GTK_PRINT_PAGES_ALL
	PRINT_PAGES_CURRENT   PrintPages = C.GTK_PRINT_PAGES_CURRENT
	PRINT_PAGES_RANGES    PrintPages = C.GTK_PRINT_PAGES_RANGES
	PRINT_PAGES_SELECTION PrintPages = C.GTK_PRINT_PAGES_SELECTION
)

func marshalPrintPages(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return PrintPages(c), nil
}

// PageSet is a representation of GTK's GtkPageSet.
type PageSet int

const (
	PAGE_SET_ALL  PageSet = C.GTK_PAGE_SET_ALL
	PAGE_SET_EVEN PageSet = C.GTK_PAGE_SET_EVEN
	PAGE_SET_ODD  PageSet = C.GTK_PAGE_SET_ODD
)

func marshalPageSet(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return PageSet(c), nil
}

// PrintOperationAction is a representation of GTK's GtkPrintError.
type PrintError int

const (
	PRINT_ERROR_GENERAL        PrintError = C.GTK_PRINT_ERROR_GENERAL
	PRINT_ERROR_INTERNAL_ERROR PrintError = C.GTK_PRINT_ERROR_INTERNAL_ERROR
	PRINT_ERROR_NOMEM          PrintError = C.GTK_PRINT_ERROR_NOMEM
	PRINT_ERROR_INVALID_FILE   PrintError = C.GTK_PRINT_ERROR_INVALID_FILE
)

func marshalPrintError(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return PrintError(c), nil
}

// PrintOperationAction is a representation of GTK's GtkPrintOperationAction.
type PrintOperationAction int

const (
	PRINT_OPERATION_ACTION_PRINT_DIALOG PrintOperationAction = C.GTK_PRINT_OPERATION_ACTION_PRINT_DIALOG
	PRINT_OPERATION_ACTION_PRINT        PrintOperationAction = C.GTK_PRINT_OPERATION_ACTION_PRINT
	PRINT_OPERATION_ACTION_PREVIEW      PrintOperationAction = C.GTK_PRINT_OPERATION_ACTION_PREVIEW
	PRINT_OPERATION_ACTION_EXPORT       PrintOperationAction = C.GTK_PRINT_OPERATION_ACTION_EXPORT
)

func marshalPrintOperationAction(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return PrintOperationAction(c), nil
}

// PrintOperationResult is a representation of GTK's GtkPrintOperationResult.
type PrintOperationResult int

const (
	PRINT_OPERATION_RESULT_ERROR       PrintOperationResult = C.GTK_PRINT_OPERATION_RESULT_ERROR
	PRINT_OPERATION_RESULT_APPLY       PrintOperationResult = C.GTK_PRINT_OPERATION_RESULT_APPLY
	PRINT_OPERATION_RESULT_CANCEL      PrintOperationResult = C.GTK_PRINT_OPERATION_RESULT_CANCEL
	PRINT_OPERATION_RESULT_IN_PROGRESS PrintOperationResult = C.GTK_PRINT_OPERATION_RESULT_IN_PROGRESS
)

func marshalPrintOperationResult(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return PrintOperationResult(c), nil
}

// PrintStatus is a representation of GTK's GtkPrintStatus.
type PrintStatus int

const (
	PRINT_STATUS_INITIAL          PrintStatus = C.GTK_PRINT_STATUS_INITIAL
	PRINT_STATUS_PREPARING        PrintStatus = C.GTK_PRINT_STATUS_PREPARING
	PRINT_STATUS_GENERATING_DATA  PrintStatus = C.GTK_PRINT_STATUS_GENERATING_DATA
	PRINT_STATUS_SENDING_DATA     PrintStatus = C.GTK_PRINT_STATUS_SENDING_DATA
	PRINT_STATUS_PENDING          PrintStatus = C.GTK_PRINT_STATUS_PENDING
	PRINT_STATUS_PENDING_ISSUE    PrintStatus = C.GTK_PRINT_STATUS_PENDING_ISSUE
	PRINT_STATUS_PRINTING         PrintStatus = C.GTK_PRINT_STATUS_PRINTING
	PRINT_STATUS_FINISHED         PrintStatus = C.GTK_PRINT_STATUS_FINISHED
	PRINT_STATUS_FINISHED_ABORTED PrintStatus = C.GTK_PRINT_STATUS_FINISHED_ABORTED
)

func marshalPrintStatus(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return PrintStatus(c), nil
}

// PrintQuality is a representation of GTK's GtkPrintQuality.
type PrintQuality int

const (
	PRINT_QUALITY_LOW    PrintQuality = C.GTK_PRINT_QUALITY_LOW
	PRINT_QUALITY_NORMAL PrintQuality = C.GTK_PRINT_QUALITY_NORMAL
	PRINT_QUALITY_HIGH   PrintQuality = C.GTK_PRINT_QUALITY_HIGH
	PRINT_QUALITY_DRAFT  PrintQuality = C.GTK_PRINT_QUALITY_DRAFT
)

func marshalPrintQuality(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return PrintQuality(c), nil
}

// Unit is a representation of GTK's GtkUnit.
type Unit int

const (
	GTK_UNIT_NONE   Unit = C.GTK_UNIT_NONE
	GTK_UNIT_POINTS Unit = C.GTK_UNIT_POINTS
	GTK_UNIT_INCH   Unit = C.GTK_UNIT_INCH
	GTK_UNIT_MM     Unit = C.GTK_UNIT_MM
)

func marshalUnit(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return Unit(c), nil
}

/*
 * GtkPageSetup
 */
type PageSetup struct {
	*glib.Object
}

func (ps *PageSetup) native() *C.GtkPageSetup {
	if ps == nil || ps.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(ps.GObject)
	return C.toGtkPageSetup(p)
}

func marshalPageSetup(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapPageSetup(obj), nil
}

func wrapPageSetup(obj *glib.Object) *PageSetup {
	return &PageSetup{obj}
}

// PageSetupNew() is a wrapper around gtk_page_setup_new().
func PageSetupNew() (*PageSetup, error) {
	c := C.gtk_page_setup_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := glib.Take(unsafe.Pointer(c))
	return wrapPageSetup(obj), nil
}

// Copy() is a wrapper around gtk_page_setup_copy().
func (ps *PageSetup) Copy() (*PageSetup, error) {
	c := C.gtk_page_setup_copy(ps.native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := glib.Take(unsafe.Pointer(c))
	return wrapPageSetup(obj), nil
}

// GetOrientation() is a wrapper around gtk_page_setup_get_orientation().
func (ps *PageSetup) GetOrientation() PageOrientation {
	c := C.gtk_page_setup_get_orientation(ps.native())
	return PageOrientation(c)
}

// SetOrientation() is a wrapper around gtk_page_setup_set_orientation().
func (ps *PageSetup) SetOrientation(orientation PageOrientation) {
	C.gtk_page_setup_set_orientation(ps.native(), C.GtkPageOrientation(orientation))
}

// GetPaperSize() is a wrapper around gtk_page_setup_get_paper_size().
func (ps *PageSetup) GetPaperSize() *PaperSize {
	c := C.gtk_page_setup_get_paper_size(ps.native())
	p := &PaperSize{c}
	runtime.SetFinalizer(p, (*PaperSize).free)
	return p
}

// SetPaperSize() is a wrapper around gtk_page_setup_set_paper_size().
func (ps *PageSetup) SetPaperSize(size *PaperSize) {
	C.gtk_page_setup_set_paper_size(ps.native(), size.native())
}

// GetTopMargin() is a wrapper around gtk_page_setup_get_top_margin().
func (ps *PageSetup) GetTopMargin(unit Unit) float64 {
	c := C.gtk_page_setup_get_top_margin(ps.native(), C.GtkUnit(unit))
	return float64(c)
}

// SetTopMargin() is a wrapper around gtk_page_setup_set_top_margin().
func (ps *PageSetup) SetTopMargin(margin float64, unit Unit) {
	C.gtk_page_setup_set_top_margin(ps.native(), C.gdouble(margin), C.GtkUnit(unit))
}

// GetBottomMargin() is a wrapper around gtk_page_setup_get_bottom_margin().
func (ps *PageSetup) GetBottomMargin(unit Unit) float64 {
	c := C.gtk_page_setup_get_bottom_margin(ps.native(), C.GtkUnit(unit))
	return float64(c)
}

// SetBottomMargin() is a wrapper around gtk_page_setup_set_bottom_margin().
func (ps *PageSetup) SetBottomMargin(margin float64, unit Unit) {
	C.gtk_page_setup_set_bottom_margin(ps.native(), C.gdouble(margin), C.GtkUnit(unit))
}

// GetLeftMargin() is a wrapper around gtk_page_setup_get_left_margin().
func (ps *PageSetup) GetLeftMargin(unit Unit) float64 {
	c := C.gtk_page_setup_get_left_margin(ps.native(), C.GtkUnit(unit))
	return float64(c)
}

// SetLeftMargin() is a wrapper around gtk_page_setup_set_left_margin().
func (ps *PageSetup) SetLeftMargin(margin float64, unit Unit) {
	C.gtk_page_setup_set_left_margin(ps.native(), C.gdouble(margin), C.GtkUnit(unit))
}

// GetRightMargin() is a wrapper around gtk_page_setup_get_right_margin().
func (ps *PageSetup) GetRightMargin(unit Unit) float64 {
	c := C.gtk_page_setup_get_right_margin(ps.native(), C.GtkUnit(unit))
	return float64(c)
}

// SetRightMargin() is a wrapper around gtk_page_setup_set_right_margin().
func (ps *PageSetup) SetRightMargin(margin float64, unit Unit) {
	C.gtk_page_setup_set_right_margin(ps.native(), C.gdouble(margin), C.GtkUnit(unit))
}

// SetPaperSizeAndDefaultMargins() is a wrapper around gtk_page_setup_set_paper_size_and_default_margins().
func (ps *PageSetup) SetPaperSizeAndDefaultMargins(size *PaperSize) {
	C.gtk_page_setup_set_paper_size_and_default_margins(ps.native(), size.native())
}

// GetPaperWidth() is a wrapper around gtk_page_setup_get_paper_width().
func (ps *PageSetup) GetPaperWidth(unit Unit) float64 {
	c := C.gtk_page_setup_get_paper_width(ps.native(), C.GtkUnit(unit))
	return float64(c)
}

// GetPaperHeight() is a wrapper around gtk_page_setup_get_paper_height().
func (ps *PageSetup) GetPaperHeight(unit Unit) float64 {
	c := C.gtk_page_setup_get_paper_height(ps.native(), C.GtkUnit(unit))
	return float64(c)
}

// GetPageWidth() is a wrapper around gtk_page_setup_get_page_width().
func (ps *PageSetup) GetPageWidth(unit Unit) float64 {
	c := C.gtk_page_setup_get_page_width(ps.native(), C.GtkUnit(unit))
	return float64(c)
}

// GetPageHeight() is a wrapper around gtk_page_setup_get_page_height().
func (ps *PageSetup) GetPageHeight(unit Unit) float64 {
	c := C.gtk_page_setup_get_page_height(ps.native(), C.GtkUnit(unit))
	return float64(c)
}

// PageSetupNewFromFile() is a wrapper around gtk_page_setup_new_from_file().
func PageSetupNewFromFile(fileName string) (*PageSetup, error) {
	cstr := C.CString(fileName)
	defer C.free(unsafe.Pointer(cstr))
	var err *C.GError = nil
	c := C.gtk_page_setup_new_from_file((*C.gchar)(cstr), &err)
	if c == nil {
		defer C.g_error_free(err)
		return nil, errors.New(C.GoString((*C.char)(err.message)))
	}
	obj := glib.Take(unsafe.Pointer(c))
	return &PageSetup{obj}, nil

}

// PageSetupNewFromKeyFile() is a wrapper around gtk_page_setup_new_from_key_file().

// PageSetupLoadFile() is a wrapper around gtk_page_setup_load_file().
func (ps *PageSetup) PageSetupLoadFile(name string) error {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	var err *C.GError = nil
	res := C.gtk_page_setup_load_file(ps.native(), cstr, &err)
	if !gobool(res) {
		defer C.g_error_free(err)
		return errors.New(C.GoString((*C.char)(err.message)))
	}
	return nil
}

// PageSetupLoadKeyFile() is a wrapper around gtk_page_setup_load_key_file().

// PageSetupToFile() is a wrapper around gtk_page_setup_to_file().
func (ps *PageSetup) PageSetupToFile(name string) error {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	var err *C.GError = nil
	res := C.gtk_page_setup_to_file(ps.native(), cstr, &err)
	if !gobool(res) {
		defer C.g_error_free(err)
		return errors.New(C.GoString((*C.char)(err.message)))
	}
	return nil
}

// PageSetupToKeyFile() is a wrapper around gtk_page_setup_to_key_file().

/*
 * GtkPaperSize
 */

// PaperSize is a representation of GTK's GtkPaperSize
type PaperSize struct {
	GtkPaperSize *C.GtkPaperSize
}

// native returns a pointer to the underlying GtkPaperSize.
func (ps *PaperSize) native() *C.GtkPaperSize {
	if ps == nil {
		return nil
	}
	return ps.GtkPaperSize
}

func marshalPaperSize(p uintptr) (interface{}, error) {
	c := C.g_value_get_boxed((*C.GValue)(unsafe.Pointer(p)))
	return &PaperSize{(*C.GtkPaperSize)(unsafe.Pointer(c))}, nil
}

const (
	UNIT_PIXEL           int    = C.GTK_UNIT_PIXEL
	PAPER_NAME_A3        string = C.GTK_PAPER_NAME_A3
	PAPER_NAME_A4        string = C.GTK_PAPER_NAME_A4
	PAPER_NAME_A5        string = C.GTK_PAPER_NAME_A5
	PAPER_NAME_B5        string = C.GTK_PAPER_NAME_B5
	PAPER_NAME_LETTER    string = C.GTK_PAPER_NAME_LETTER
	PAPER_NAME_EXECUTIVE string = C.GTK_PAPER_NAME_EXECUTIVE
	PAPER_NAME_LEGAL     string = C.GTK_PAPER_NAME_LEGAL
)

// PaperSizeNew() is a wrapper around gtk_paper_size_new().
func PaperSizeNew(name string) (*PaperSize, error) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	var gName *C.gchar

	if name == "" {
		gName = nil
	} else {
		gName = (*C.gchar)(cstr)
	}

	c := C.gtk_paper_size_new(gName)
	if c == nil {
		return nil, nilPtrErr
	}

	t := &PaperSize{c}
	runtime.SetFinalizer(t, (*PaperSize).free)
	return t, nil
}

// PaperSizeNewFromPPD() is a wrapper around gtk_paper_size_new_from_ppd().
func PaperSizeNewFromPPD(name, displayName string, width, height float64) (*PaperSize, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cDisplayName := C.CString(displayName)
	defer C.free(unsafe.Pointer(cDisplayName))
	c := C.gtk_paper_size_new_from_ppd((*C.gchar)(cName), (*C.gchar)(cDisplayName),
		C.gdouble(width), C.gdouble(height))
	if c == nil {
		return nil, nilPtrErr
	}
	t := &PaperSize{c}
	runtime.SetFinalizer(t, (*PaperSize).free)
	return t, nil
}

// PaperSizeNewCustom() is a wrapper around gtk_paper_size_new_custom().
func PaperSizeNewCustom(name, displayName string, width, height float64, unit Unit) (*PaperSize, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cDisplayName := C.CString(displayName)
	defer C.free(unsafe.Pointer(cDisplayName))
	c := C.gtk_paper_size_new_custom((*C.gchar)(cName), (*C.gchar)(cDisplayName),
		C.gdouble(width), C.gdouble(height), C.GtkUnit(unit))
	if c == nil {
		return nil, nilPtrErr
	}
	t := &PaperSize{c}
	runtime.SetFinalizer(t, (*PaperSize).free)
	return t, nil
}

// Copy() is a wrapper around gtk_paper_size_copy().
func (ps *PaperSize) Copy() (*PaperSize, error) {
	c := C.gtk_paper_size_copy(ps.native())
	if c == nil {
		return nil, nilPtrErr
	}
	t := &PaperSize{c}
	runtime.SetFinalizer(t, (*PaperSize).free)
	return t, nil
}

// free() is a wrapper around gtk_paper_size_free().
func (ps *PaperSize) free() {
	C.gtk_paper_size_free(ps.native())
}

// IsEqual() is a wrapper around gtk_paper_size_is_equal().
func (ps *PaperSize) IsEqual(other *PaperSize) bool {
	c := C.gtk_paper_size_is_equal(ps.native(), other.native())
	return gobool(c)
}

// PaperSizeGetPaperSizes() is a wrapper around gtk_paper_size_get_paper_sizes().
func PaperSizeGetPaperSizes(includeCustom bool) *glib.List {
	clist := C.gtk_paper_size_get_paper_sizes(gbool(includeCustom))
	if clist == nil {
		return nil
	}

	glist := glib.WrapList(uintptr(unsafe.Pointer(clist)))
	glist.DataWrapper(func(ptr unsafe.Pointer) interface{} {
		return &PaperSize{(*C.GtkPaperSize)(ptr)}
	})

	runtime.SetFinalizer(glist, func(glist *glib.List) {
		glist.FreeFull(func(item interface{}) {
			ps := item.(*PaperSize)
			C.gtk_paper_size_free(ps.GtkPaperSize)
		})
	})

	return glist
}

// GetName() is a wrapper around gtk_paper_size_get_name().
func (ps *PaperSize) GetName() string {
	c := C.gtk_paper_size_get_name(ps.native())
	return C.GoString((*C.char)(c))
}

// GetDisplayName() is a wrapper around gtk_paper_size_get_display_name().
func (ps *PaperSize) GetDisplayName() string {
	c := C.gtk_paper_size_get_display_name(ps.native())
	return C.GoString((*C.char)(c))
}

// GetPPDName() is a wrapper around gtk_paper_size_get_ppd_name().
func (ps *PaperSize) GetPPDName() (string, error) {
	c := C.gtk_paper_size_get_ppd_name(ps.native())
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

// GetWidth() is a wrapper around gtk_paper_size_get_width().
func (ps *PaperSize) GetWidth(unit Unit) float64 {
	c := C.gtk_paper_size_get_width(ps.native(), C.GtkUnit(unit))
	return float64(c)
}

// GetHeight() is a wrapper around gtk_paper_size_get_height().
func (ps *PaperSize) GetHeight(unit Unit) float64 {
	c := C.gtk_paper_size_get_width(ps.native(), C.GtkUnit(unit))
	return float64(c)
}

// IsCustom() is a wrapper around gtk_paper_size_is_custom().
func (ps *PaperSize) IsCustom() bool {
	c := C.gtk_paper_size_is_custom(ps.native())
	return gobool(c)
}

// SetSize() is a wrapper around gtk_paper_size_set_size().
func (ps *PaperSize) SetSize(width, height float64, unit Unit) {
	C.gtk_paper_size_set_size(ps.native(), C.gdouble(width), C.gdouble(height), C.GtkUnit(unit))
}

// GetDefaultTopMargin() is a wrapper around gtk_paper_size_get_default_top_margin().
func (ps *PaperSize) GetDefaultTopMargin(unit Unit) float64 {
	c := C.gtk_paper_size_get_default_top_margin(ps.native(), C.GtkUnit(unit))
	return float64(c)
}

// GetDefaultBottomMargin() is a wrapper around gtk_paper_size_get_default_bottom_margin().
func (ps *PaperSize) GetDefaultBottomMargin(unit Unit) float64 {
	c := C.gtk_paper_size_get_default_bottom_margin(ps.native(), C.GtkUnit(unit))
	return float64(c)
}

// GetDefaultLeftMargin() is a wrapper around gtk_paper_size_get_default_left_margin().
func (ps *PaperSize) GetDefaultLeftMargin(unit Unit) float64 {
	c := C.gtk_paper_size_get_default_left_margin(ps.native(), C.GtkUnit(unit))
	return float64(c)
}

// GetDefaultRightMargin() is a wrapper around gtk_paper_size_get_default_right_margin().
func (ps *PaperSize) GetDefaultRightMargin(unit Unit) float64 {
	c := C.gtk_paper_size_get_default_right_margin(ps.native(), C.GtkUnit(unit))
	return float64(c)
}

// PaperSizeGetDefault() is a wrapper around gtk_paper_size_get_default().
func PaperSizeGetDefaultRightMargin(unit Unit) string {
	c := C.gtk_paper_size_get_default()
	return C.GoString((*C.char)(c))
}

// PaperSizeNewFromKeyFile() is a wrapper around gtk_paper_size_new_from_key_file().
// PaperSizeToKeyFile() is a wrapper around gtk_paper_size_to_key_file().

/*
 * GtkPrintContext
 */

// PrintContext is a representation of GTK's GtkPrintContext.
type PrintContext struct {
	*glib.Object
}

// native() returns a pointer to the underlying GtkPrintContext.
func (pc *PrintContext) native() *C.GtkPrintContext {
	if pc == nil || pc.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(pc.GObject)
	return C.toGtkPrintContext(p)
}

func marshalPrintContext(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapPrintContext(obj), nil
}

func wrapPrintContext(obj *glib.Object) *PrintContext {
	return &PrintContext{obj}
}

// GetCairoContext() is a wrapper around gtk_print_context_get_cairo_context().
func (pc *PrintContext) GetCairoContext() *cairo.Context {
	c := C.gtk_print_context_get_cairo_context(pc.native())
	return cairo.WrapContext(uintptr(unsafe.Pointer(c)))
}

// SetCairoContext() is a wrapper around gtk_print_context_set_cairo_context().
func (pc *PrintContext) SetCairoContext(cr *cairo.Context, dpiX, dpiY float64) {
	C.gtk_print_context_set_cairo_context(pc.native(),
		(*C.cairo_t)(unsafe.Pointer(cr.Native())),
		C.double(dpiX), C.double(dpiY))
}

// GetPageSetup() is a wrapper around gtk_print_context_get_page_setup().
func (pc *PrintContext) GetPageSetup() *PageSetup {
	c := C.gtk_print_context_get_page_setup(pc.native())
	obj := glib.Take(unsafe.Pointer(c))
	return wrapPageSetup(obj)
}

// GetWidth() is a wrapper around gtk_print_context_get_width().
func (pc *PrintContext) GetWidth() float64 {
	c := C.gtk_print_context_get_width(pc.native())
	return float64(c)
}

// GetHeight() is a wrapper around gtk_print_context_get_height().
func (pc *PrintContext) GetHeight() float64 {
	c := C.gtk_print_context_get_height(pc.native())
	return float64(c)
}

// GetDpiX() is a wrapper around gtk_print_context_get_dpi_x().
func (pc *PrintContext) GetDpiX() float64 {
	c := C.gtk_print_context_get_dpi_x(pc.native())
	return float64(c)
}

// GetDpiY() is a wrapper around gtk_print_context_get_dpi_y().
func (pc *PrintContext) GetDpiY() float64 {
	c := C.gtk_print_context_get_dpi_y(pc.native())
	return float64(c)
}

// GetPangoFontMap() is a wrapper around gtk_print_context_get_pango_fontmap().
func (pc *PrintContext) GetPangoFontMap() *pango.FontMap {
	c := C.gtk_print_context_get_pango_fontmap(pc.native())
	return pango.WrapFontMap(uintptr(unsafe.Pointer(c)))
}

// CreatePangoContext() is a wrapper around gtk_print_context_create_pango_context().
func (pc *PrintContext) CreatePangoContext() *pango.Context {
	c := C.gtk_print_context_create_pango_context(pc.native())
	return pango.WrapContext(uintptr(unsafe.Pointer(c)))
}

// CreatePangoLayout() is a wrapper around gtk_print_context_create_pango_layout().
func (pc *PrintContext) CreatePangoLayout() *pango.Layout {
	c := C.gtk_print_context_create_pango_layout(pc.native())
	return pango.WrapLayout(uintptr(unsafe.Pointer(c)))
}

// GetHardMargins() is a wrapper around gtk_print_context_get_hard_margins().
func (pc *PrintContext) GetHardMargins() (float64, float64, float64, float64, error) {
	var top, bottom, left, right C.gdouble
	c := C.gtk_print_context_get_hard_margins(pc.native(), &top, &bottom, &left, &right)
	if gobool(c) == false {
		return 0.0, 0.0, 0.0, 0.0, errors.New("unable to retrieve hard margins")
	}
	return float64(top), float64(bottom), float64(left), float64(right), nil
}

/*
 * GtkPrintOperation
 */
type PrintOperation struct {
	*glib.Object

	// Interfaces
	PrintOperationPreview
}

func (po *PrintOperation) native() *C.GtkPrintOperation {
	if po == nil || po.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(po.GObject)
	return C.toGtkPrintOperation(p)
}

func (v *PrintOperation) toPrintOperationPreview() *C.GtkPrintOperationPreview {
	if v == nil {
		return nil
	}
	return C.toGtkPrintOperationPreview(unsafe.Pointer(v.GObject))
}

func marshalPrintOperation(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapPrintOperation(obj), nil
}

func wrapPrintOperation(obj *glib.Object) *PrintOperation {
	pop := wrapPrintOperationPreview(obj)
	return &PrintOperation{obj, *pop}
}

// PrintOperationNew() is a wrapper around gtk_print_operation_new().
func PrintOperationNew() (*PrintOperation, error) {
	c := C.gtk_print_operation_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := glib.Take(unsafe.Pointer(c))
	return wrapPrintOperation(obj), nil
}

// SetAllowAsync() is a wrapper around gtk_print_operation_set_allow_async().
func (po *PrintOperation) PrintOperationSetAllowAsync(allowSync bool) {
	C.gtk_print_operation_set_allow_async(po.native(), gbool(allowSync))
}

// GetError() is a wrapper around gtk_print_operation_get_error().
func (po *PrintOperation) PrintOperationGetError() error {
	var err *C.GError = nil
	C.gtk_print_operation_get_error(po.native(), &err)
	defer C.g_error_free(err)
	return errors.New(C.GoString((*C.char)(err.message)))
}

// SetDefaultPageSetup() is a wrapper around gtk_print_operation_set_default_page_setup().
func (po *PrintOperation) SetDefaultPageSetup(ps *PageSetup) {
	C.gtk_print_operation_set_default_page_setup(po.native(), ps.native())
}

// GetDefaultPageSetup() is a wrapper around gtk_print_operation_get_default_page_setup().
func (po *PrintOperation) GetDefaultPageSetup() (*PageSetup, error) {
	c := C.gtk_print_operation_get_default_page_setup(po.native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := glib.Take(unsafe.Pointer(c))
	return wrapPageSetup(obj), nil
}

// SetPrintSettings() is a wrapper around gtk_print_operation_set_print_settings().
func (po *PrintOperation) SetPrintSettings(ps *PrintSettings) {
	C.gtk_print_operation_set_print_settings(po.native(), ps.native())
}

// GetPrintSettings() is a wrapper around gtk_print_operation_get_print_settings().
func (po *PrintOperation) GetPrintSettings(ps *PageSetup) (*PrintSettings, error) {
	c := C.gtk_print_operation_get_print_settings(po.native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := glib.Take(unsafe.Pointer(c))
	return wrapPrintSettings(obj), nil
}

// SetJobName() is a wrapper around gtk_print_operation_set_job_name().
func (po *PrintOperation) SetJobName(name string) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_print_operation_set_job_name(po.native(), (*C.gchar)(cstr))
}

// SetNPages() is a wrapper around gtk_print_operation_set_n_pages().
func (po *PrintOperation) SetNPages(pages int) {
	C.gtk_print_operation_set_n_pages(po.native(), C.gint(pages))
}

// GetNPagesToPrint() is a wrapper around gtk_print_operation_get_n_pages_to_print().
func (po *PrintOperation) GetNPagesToPrint() int {
	c := C.gtk_print_operation_get_n_pages_to_print(po.native())
	return int(c)
}

// SetCurrentPage() is a wrapper around gtk_print_operation_set_current_page().
func (po *PrintOperation) SetCurrentPage(page int) {
	C.gtk_print_operation_set_current_page(po.native(), C.gint(page))
}

// SetUseFullPage() is a wrapper around gtk_print_operation_set_use_full_page().
func (po *PrintOperation) SetUseFullPage(full bool) {
	C.gtk_print_operation_set_use_full_page(po.native(), gbool(full))
}

// SetUnit() is a wrapper around gtk_print_operation_set_unit().
func (po *PrintOperation) SetUnit(unit Unit) {
	C.gtk_print_operation_set_unit(po.native(), C.GtkUnit(unit))
}

// SetExportFilename() is a wrapper around gtk_print_operation_set_export_filename().
func (po *PrintOperation) SetExportFilename(name string) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_print_operation_set_export_filename(po.native(), (*C.gchar)(cstr))
}

// SetShowProgress() is a wrapper around gtk_print_operation_set_show_progress().
func (po *PrintOperation) SetShowProgress(show bool) {
	C.gtk_print_operation_set_show_progress(po.native(), gbool(show))
}

// SetTrackPrintStatus() is a wrapper around gtk_print_operation_set_track_print_status().
func (po *PrintOperation) SetTrackPrintStatus(progress bool) {
	C.gtk_print_operation_set_track_print_status(po.native(), gbool(progress))
}

// SetCustomTabLabel() is a wrapper around gtk_print_operation_set_custom_tab_label().
func (po *PrintOperation) SetCustomTabLabel(label string) {
	cstr := C.CString(label)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_print_operation_set_custom_tab_label(po.native(), (*C.gchar)(cstr))
}

// Run() is a wrapper around gtk_print_operation_run().
func (po *PrintOperation) Run(action PrintOperationAction, parent *Window) (PrintOperationResult, error) {
	var err *C.GError = nil
	c := C.gtk_print_operation_run(po.native(), C.GtkPrintOperationAction(action), parent.native(), &err)
	res := PrintOperationResult(c)
	if res == PRINT_OPERATION_RESULT_ERROR {
		defer C.g_error_free(err)
		return res, errors.New(C.GoString((*C.char)(err.message)))
	}
	return res, nil
}

// Cancel() is a wrapper around gtk_print_operation_cancel().
func (po *PrintOperation) Cancel() {
	C.gtk_print_operation_cancel(po.native())
}

// DrawPageFinish() is a wrapper around gtk_print_operation_draw_page_finish().
func (po *PrintOperation) DrawPageFinish() {
	C.gtk_print_operation_draw_page_finish(po.native())
}

// SetDeferDrawing() is a wrapper around gtk_print_operation_set_defer_drawing().
func (po *PrintOperation) SetDeferDrawing() {
	C.gtk_print_operation_set_defer_drawing(po.native())
}

// GetStatus() is a wrapper around gtk_print_operation_get_status().
func (po *PrintOperation) GetStatus() PrintStatus {
	c := C.gtk_print_operation_get_status(po.native())
	return PrintStatus(c)
}

// GetStatusString() is a wrapper around gtk_print_operation_get_status_string().
func (po *PrintOperation) GetStatusString() string {
	c := C.gtk_print_operation_get_status_string(po.native())
	return C.GoString((*C.char)(c))
}

// IsFinished() is a wrapper around gtk_print_operation_is_finished().
func (po *PrintOperation) IsFinished() bool {
	c := C.gtk_print_operation_is_finished(po.native())
	return gobool(c)
}

// SetSupportSelection() is a wrapper around gtk_print_operation_set_support_selection().
func (po *PrintOperation) SetSupportSelection(selection bool) {
	C.gtk_print_operation_set_support_selection(po.native(), gbool(selection))
}

// GetSupportSelection() is a wrapper around gtk_print_operation_get_support_selection().
func (po *PrintOperation) GetSupportSelection() bool {
	c := C.gtk_print_operation_get_support_selection(po.native())
	return gobool(c)
}

// SetHasSelection() is a wrapper around gtk_print_operation_set_has_selection().
func (po *PrintOperation) SetHasSelection(selection bool) {
	C.gtk_print_operation_set_has_selection(po.native(), gbool(selection))
}

// GetHasSelection() is a wrapper around gtk_print_operation_get_has_selection().
func (po *PrintOperation) GetHasSelection() bool {
	c := C.gtk_print_operation_get_has_selection(po.native())
	return gobool(c)
}

// SetEmbedPageSetup() is a wrapper around gtk_print_operation_set_embed_page_setup().
func (po *PrintOperation) SetEmbedPageSetup(embed bool) {
	C.gtk_print_operation_set_embed_page_setup(po.native(), gbool(embed))
}

// GetEmbedPageSetup() is a wrapper around gtk_print_operation_get_embed_page_setup().
func (po *PrintOperation) GetEmbedPageSetup() bool {
	c := C.gtk_print_operation_get_embed_page_setup(po.native())
	return gobool(c)
}

// PrintRunPageSetupDialog() is a wrapper around gtk_print_run_page_setup_dialog().
func PrintRunPageSetupDialog(parent *Window, pageSetup *PageSetup, settings *PrintSettings) *PageSetup {
	c := C.gtk_print_run_page_setup_dialog(parent.native(), pageSetup.native(), settings.native())
	obj := glib.Take(unsafe.Pointer(c))
	return wrapPageSetup(obj)
}

type PageSetupDoneCallback func(setup *PageSetup, userData uintptr)

type pageSetupDoneCallbackData struct {
	fn   PageSetupDoneCallback
	data uintptr
}

var (
	pageSetupDoneCallbackRegistry = struct {
		sync.RWMutex
		next int
		m    map[int]pageSetupDoneCallbackData
	}{
		next: 1,
		m:    make(map[int]pageSetupDoneCallbackData),
	}
)

// PrintRunPageSetupDialogAsync() is a wrapper around gtk_print_run_page_setup_dialog_async().
func PrintRunPageSetupDialogAsync(parent *Window, setup *PageSetup,
	settings *PrintSettings, cb PageSetupDoneCallback, data uintptr) {

	pageSetupDoneCallbackRegistry.Lock()
	id := pageSetupDoneCallbackRegistry.next
	pageSetupDoneCallbackRegistry.next++
	pageSetupDoneCallbackRegistry.m[id] =
		pageSetupDoneCallbackData{fn: cb, data: data}
	pageSetupDoneCallbackRegistry.Unlock()

	C._gtk_print_run_page_setup_dialog_async(parent.native(), setup.native(),
		settings.native(), C.gpointer(uintptr(id)))
}

/*
 * GtkPrintOperationPreview
 */

// PrintOperationPreview is a representation of GTK's GtkPrintOperationPreview GInterface.
type PrintOperationPreview struct {
	*glib.Object
}

// IPrintOperationPreview is an interface type implemented by all structs
// embedding a PrintOperationPreview.  It is meant to be used as an argument type
// for wrapper functions that wrap around a C GTK function taking a
// GtkPrintOperationPreview.
type IPrintOperationPreview interface {
	toPrintOperationPreview() *C.GtkPrintOperationPreview
}

// native() returns a pointer to the underlying GObject as a GtkPrintOperationPreview.
func (v *PrintOperationPreview) native() *C.GtkPrintOperationPreview {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkPrintOperationPreview(p)
}

func marshalPrintOperationPreview(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapPrintOperationPreview(obj), nil
}

func wrapPrintOperationPreview(obj *glib.Object) *PrintOperationPreview {
	return &PrintOperationPreview{obj}
}

func (v *PrintOperationPreview) toPrintOperationPreview() *C.GtkPrintOperationPreview {
	if v == nil {
		return nil
	}
	return v.native()
}

// RenderPage()() is a wrapper around gtk_print_operation_preview_render_page().
func (pop *PrintOperationPreview) RenderPage(page int) {
	C.gtk_print_operation_preview_render_page(pop.native(), C.gint(page))
}

// EndPreview()() is a wrapper around gtk_print_operation_preview_end_preview().
func (pop *PrintOperationPreview) EndPreview() {
	C.gtk_print_operation_preview_end_preview(pop.native())
}

// IsSelected()() is a wrapper around gtk_print_operation_preview_is_selected().
func (pop *PrintOperationPreview) IsSelected(page int) bool {
	c := C.gtk_print_operation_preview_is_selected(pop.native(), C.gint(page))
	return gobool(c)
}

/*
 * GtkPrintSettings
 */

type PrintSettings struct {
	*glib.Object
}

func (ps *PrintSettings) native() *C.GtkPrintSettings {
	if ps == nil || ps.GObject == nil {
		return nil
	}

	p := unsafe.Pointer(ps.GObject)
	return C.toGtkPrintSettings(p)
}

func marshalPrintSettings(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return wrapPrintSettings(glib.Take(unsafe.Pointer(c))), nil
}

func wrapPrintSettings(obj *glib.Object) *PrintSettings {
	return &PrintSettings{obj}
}

const (
	PRINT_SETTINGS_PRINTER              string = C.GTK_PRINT_SETTINGS_PRINTER
	PRINT_SETTINGS_ORIENTATION          string = C.GTK_PRINT_SETTINGS_ORIENTATION
	PRINT_SETTINGS_PAPER_FORMAT         string = C.GTK_PRINT_SETTINGS_PAPER_FORMAT
	PRINT_SETTINGS_PAPER_WIDTH          string = C.GTK_PRINT_SETTINGS_PAPER_WIDTH
	PRINT_SETTINGS_PAPER_HEIGHT         string = C.GTK_PRINT_SETTINGS_PAPER_HEIGHT
	PRINT_SETTINGS_USE_COLOR            string = C.GTK_PRINT_SETTINGS_USE_COLOR
	PRINT_SETTINGS_COLLATE              string = C.GTK_PRINT_SETTINGS_COLLATE
	PRINT_SETTINGS_REVERSE              string = C.GTK_PRINT_SETTINGS_REVERSE
	PRINT_SETTINGS_DUPLEX               string = C.GTK_PRINT_SETTINGS_DUPLEX
	PRINT_SETTINGS_QUALITY              string = C.GTK_PRINT_SETTINGS_QUALITY
	PRINT_SETTINGS_N_COPIES             string = C.GTK_PRINT_SETTINGS_N_COPIES
	PRINT_SETTINGS_NUMBER_UP            string = C.GTK_PRINT_SETTINGS_NUMBER_UP
	PRINT_SETTINGS_NUMBER_UP_LAYOUT     string = C.GTK_PRINT_SETTINGS_NUMBER_UP_LAYOUT
	PRINT_SETTINGS_RESOLUTION           string = C.GTK_PRINT_SETTINGS_RESOLUTION
	PRINT_SETTINGS_RESOLUTION_X         string = C.GTK_PRINT_SETTINGS_RESOLUTION_X
	PRINT_SETTINGS_RESOLUTION_Y         string = C.GTK_PRINT_SETTINGS_RESOLUTION_Y
	PRINT_SETTINGS_PRINTER_LPI          string = C.GTK_PRINT_SETTINGS_PRINTER_LPI
	PRINT_SETTINGS_SCALE                string = C.GTK_PRINT_SETTINGS_SCALE
	PRINT_SETTINGS_PRINT_PAGES          string = C.GTK_PRINT_SETTINGS_PRINT_PAGES
	PRINT_SETTINGS_PAGE_RANGES          string = C.GTK_PRINT_SETTINGS_PAGE_RANGES
	PRINT_SETTINGS_PAGE_SET             string = C.GTK_PRINT_SETTINGS_PAGE_SET
	PRINT_SETTINGS_DEFAULT_SOURCE       string = C.GTK_PRINT_SETTINGS_DEFAULT_SOURCE
	PRINT_SETTINGS_MEDIA_TYPE           string = C.GTK_PRINT_SETTINGS_MEDIA_TYPE
	PRINT_SETTINGS_DITHER               string = C.GTK_PRINT_SETTINGS_DITHER
	PRINT_SETTINGS_FINISHINGS           string = C.GTK_PRINT_SETTINGS_FINISHINGS
	PRINT_SETTINGS_OUTPUT_BIN           string = C.GTK_PRINT_SETTINGS_OUTPUT_BIN
	PRINT_SETTINGS_OUTPUT_DIR           string = C.GTK_PRINT_SETTINGS_OUTPUT_DIR
	PRINT_SETTINGS_OUTPUT_BASENAME      string = C.GTK_PRINT_SETTINGS_OUTPUT_BASENAME
	PRINT_SETTINGS_OUTPUT_FILE_FORMAT   string = C.GTK_PRINT_SETTINGS_OUTPUT_FILE_FORMAT
	PRINT_SETTINGS_OUTPUT_URI           string = C.GTK_PRINT_SETTINGS_OUTPUT_URI
	PRINT_SETTINGS_WIN32_DRIVER_EXTRA   string = C.GTK_PRINT_SETTINGS_WIN32_DRIVER_EXTRA
	PRINT_SETTINGS_WIN32_DRIVER_VERSION string = C.GTK_PRINT_SETTINGS_WIN32_DRIVER_VERSION
)

// PrintSettingsNew() is a wrapper around gtk_print_settings_new().
func PrintSettingsNew() (*PrintSettings, error) {
	c := C.gtk_print_settings_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := glib.Take(unsafe.Pointer(c))
	return wrapPrintSettings(obj), nil
}

// Copy() is a wrapper around gtk_print_settings_copy().
func (ps *PrintSettings) Copy() (*PrintSettings, error) {
	c := C.gtk_print_settings_copy(ps.native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := glib.Take(unsafe.Pointer(c))
	return wrapPrintSettings(obj), nil
}

// HasKey() is a wrapper around gtk_print_settings_has_key().
func (ps *PrintSettings) HasKey(key string) bool {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_print_settings_has_key(ps.native(), (*C.gchar)(cstr))
	return gobool(c)
}

// Get() is a wrapper around gtk_print_settings_get().
func (ps *PrintSettings) Get(key string) string {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_print_settings_get(ps.native(), (*C.gchar)(cstr))
	return C.GoString((*C.char)(c))
}

// Set() is a wrapper around gtk_print_settings_set().
// TODO: Since value can't be nil, we can't unset values here.
func (ps *PrintSettings) Set(key, value string) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	C.gtk_print_settings_set(ps.native(), (*C.gchar)(cKey), (*C.gchar)(cValue))
}

// Unset() is a wrapper around gtk_print_settings_unset().
func (ps *PrintSettings) Unset(key string) {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_print_settings_unset(ps.native(), (*C.gchar)(cstr))
}

type PrintSettingsCallback func(key, value string, userData uintptr)

type printSettingsCallbackData struct {
	fn       PrintSettingsCallback
	userData uintptr
}

var (
	printSettingsCallbackRegistry = struct {
		sync.RWMutex
		next int
		m    map[int]printSettingsCallbackData
	}{
		next: 1,
		m:    make(map[int]printSettingsCallbackData),
	}
)

// Foreach() is a wrapper around gtk_print_settings_foreach().
func (ps *PrintSettings) ForEach(cb PrintSettingsCallback, userData uintptr) {
	printSettingsCallbackRegistry.Lock()
	id := printSettingsCallbackRegistry.next
	printSettingsCallbackRegistry.next++
	printSettingsCallbackRegistry.m[id] =
		printSettingsCallbackData{fn: cb, userData: userData}
	printSettingsCallbackRegistry.Unlock()

	C._gtk_print_settings_foreach(ps.native(), C.gpointer(uintptr(id)))
}

// GetBool() is a wrapper around gtk_print_settings_get_bool().
func (ps *PrintSettings) GetBool(key string) bool {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_print_settings_get_bool(ps.native(), (*C.gchar)(cstr))
	return gobool(c)
}

// SetBool() is a wrapper around gtk_print_settings_set_bool().
func (ps *PrintSettings) SetBool(key string, value bool) {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_print_settings_set_bool(ps.native(), (*C.gchar)(cstr), gbool(value))
}

// GetDouble() is a wrapper around gtk_print_settings_get_double().
func (ps *PrintSettings) GetDouble(key string) float64 {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_print_settings_get_double(ps.native(), (*C.gchar)(cstr))
	return float64(c)
}

// GetDoubleWithDefault() is a wrapper around gtk_print_settings_get_double_with_default().
func (ps *PrintSettings) GetDoubleWithDefault(key string, def float64) float64 {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_print_settings_get_double_with_default(ps.native(),
		(*C.gchar)(cstr), C.gdouble(def))
	return float64(c)
}

// SetDouble() is a wrapper around gtk_print_settings_set_double().
func (ps *PrintSettings) SetDouble(key string, value float64) {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_print_settings_set_double(ps.native(), (*C.gchar)(cstr), C.gdouble(value))
}

// GetLength() is a wrapper around gtk_print_settings_get_length().
func (ps *PrintSettings) GetLength(key string, unit Unit) float64 {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_print_settings_get_length(ps.native(), (*C.gchar)(cstr), C.GtkUnit(unit))
	return float64(c)
}

// SetLength() is a wrapper around gtk_print_settings_set_length().
func (ps *PrintSettings) SetLength(key string, value float64, unit Unit) {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_print_settings_set_length(ps.native(), (*C.gchar)(cstr), C.gdouble(value), C.GtkUnit(unit))
}

// GetInt() is a wrapper around gtk_print_settings_get_int().
func (ps *PrintSettings) GetInt(key string) int {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_print_settings_get_int(ps.native(), (*C.gchar)(cstr))
	return int(c)
}

// GetIntWithDefault() is a wrapper around gtk_print_settings_get_int_with_default().
func (ps *PrintSettings) GetIntWithDefault(key string, def int) int {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gtk_print_settings_get_int_with_default(ps.native(), (*C.gchar)(cstr), C.gint(def))
	return int(c)
}

// SetInt() is a wrapper around gtk_print_settings_set_int().
func (ps *PrintSettings) SetInt(key string, value int) {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_print_settings_set_int(ps.native(), (*C.gchar)(cstr), C.gint(value))
}

// GetPrinter() is a wrapper around gtk_print_settings_get_printer().
func (ps *PrintSettings) GetPrinter() string {
	c := C.gtk_print_settings_get_printer(ps.native())
	return C.GoString((*C.char)(c))
}

// SetPrinter() is a wrapper around gtk_print_settings_set_printer().
func (ps *PrintSettings) SetPrinter(printer string) {
	cstr := C.CString(printer)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_print_settings_set_printer(ps.native(), (*C.gchar)(cstr))
}

// GetOrientation() is a wrapper around gtk_print_settings_get_orientation().
func (ps *PrintSettings) GetOrientation() PageOrientation {
	c := C.gtk_print_settings_get_orientation(ps.native())
	return PageOrientation(c)
}

// SetOrientation() is a wrapper around gtk_print_settings_set_orientation().
func (ps *PrintSettings) SetOrientation(orientation PageOrientation) {
	C.gtk_print_settings_set_orientation(ps.native(), C.GtkPageOrientation(orientation))
}

// GetPaperSize() is a wrapper around gtk_print_settings_get_paper_size().
func (ps *PrintSettings) GetPaperSize() (*PaperSize, error) {
	c := C.gtk_print_settings_get_paper_size(ps.native())
	if c == nil {
		return nil, nilPtrErr
	}
	p := &PaperSize{c}
	runtime.SetFinalizer(p, (*PaperSize).free)
	return p, nil
}

// SetPaperSize() is a wrapper around gtk_print_settings_set_paper_size().
func (ps *PrintSettings) SetPaperSize(size *PaperSize) {
	C.gtk_print_settings_set_paper_size(ps.native(), size.native())
}

// GetPaperWidth() is a wrapper around gtk_print_settings_get_paper_width().
func (ps *PrintSettings) GetPaperWidth(unit Unit) float64 {
	c := C.gtk_print_settings_get_paper_width(ps.native(), C.GtkUnit(unit))
	return float64(c)
}

// SetPaperWidth() is a wrapper around gtk_print_settings_set_paper_width().
func (ps *PrintSettings) SetPaperWidth(width float64, unit Unit) {
	C.gtk_print_settings_set_paper_width(ps.native(), C.gdouble(width), C.GtkUnit(unit))
}

// GetPaperHeight() is a wrapper around gtk_print_settings_get_paper_height().
func (ps *PrintSettings) GetPaperHeight(unit Unit) float64 {
	c := C.gtk_print_settings_get_paper_height(ps.native(), C.GtkUnit(unit))
	return float64(c)
}

// SetPaperHeight() is a wrapper around gtk_print_settings_set_paper_height().
func (ps *PrintSettings) SetPaperHeight(width float64, unit Unit) {
	C.gtk_print_settings_set_paper_height(ps.native(), C.gdouble(width), C.GtkUnit(unit))
}

// GetUseColor() is a wrapper around gtk_print_settings_get_use_color().
func (ps *PrintSettings) GetUseColor() bool {
	c := C.gtk_print_settings_get_use_color(ps.native())
	return gobool(c)
}

// SetUseColor() is a wrapper around gtk_print_settings_set_use_color().
func (ps *PrintSettings) SetUseColor(color bool) {
	C.gtk_print_settings_set_use_color(ps.native(), gbool(color))
}

// GetCollate() is a wrapper around gtk_print_settings_get_collate().
func (ps *PrintSettings) GetCollate() bool {
	c := C.gtk_print_settings_get_collate(ps.native())
	return gobool(c)
}

// SetCollate() is a wrapper around gtk_print_settings_set_collate().
func (ps *PrintSettings) SetCollate(collate bool) {
	C.gtk_print_settings_set_collate(ps.native(), gbool(collate))
}

// GetReverse() is a wrapper around gtk_print_settings_get_reverse().
func (ps *PrintSettings) GetReverse() bool {
	c := C.gtk_print_settings_get_reverse(ps.native())
	return gobool(c)
}

// SetReverse() is a wrapper around gtk_print_settings_set_reverse().
func (ps *PrintSettings) SetReverse(reverse bool) {
	C.gtk_print_settings_set_reverse(ps.native(), gbool(reverse))
}

// GetDuplex() is a wrapper around gtk_print_settings_get_duplex().
func (ps *PrintSettings) GetDuplex() PrintDuplex {
	c := C.gtk_print_settings_get_duplex(ps.native())
	return PrintDuplex(c)
}

// SetDuplex() is a wrapper around gtk_print_settings_set_duplex().
func (ps *PrintSettings) SetDuplex(duplex PrintDuplex) {
	C.gtk_print_settings_set_duplex(ps.native(), C.GtkPrintDuplex(duplex))
}

// GetQuality() is a wrapper around gtk_print_settings_get_quality().
func (ps *PrintSettings) GetQuality() PrintQuality {
	c := C.gtk_print_settings_get_quality(ps.native())
	return PrintQuality(c)
}

// SetQuality() is a wrapper around gtk_print_settings_set_quality().
func (ps *PrintSettings) SetQuality(quality PrintQuality) {
	C.gtk_print_settings_set_quality(ps.native(), C.GtkPrintQuality(quality))
}

// GetNCopies() is a wrapper around gtk_print_settings_get_n_copies().
func (ps *PrintSettings) GetNCopies() int {
	c := C.gtk_print_settings_get_n_copies(ps.native())
	return int(c)
}

// SetNCopies() is a wrapper around gtk_print_settings_set_n_copies().
func (ps *PrintSettings) SetNCopies(copies int) {
	C.gtk_print_settings_set_n_copies(ps.native(), C.gint(copies))
}

// GetNmberUp() is a wrapper around gtk_print_settings_get_number_up().
func (ps *PrintSettings) GetNmberUp() int {
	c := C.gtk_print_settings_get_number_up(ps.native())
	return int(c)
}

// SetNumberUp() is a wrapper around gtk_print_settings_set_number_up().
func (ps *PrintSettings) SetNumberUp(numberUp int) {
	C.gtk_print_settings_set_number_up(ps.native(), C.gint(numberUp))
}

// GetNumberUpLayout() is a wrapper around gtk_print_settings_get_number_up_layout().
func (ps *PrintSettings) GetNumberUpLayout() NumberUpLayout {
	c := C.gtk_print_settings_get_number_up_layout(ps.native())
	return NumberUpLayout(c)
}

// SetNumberUpLayout() is a wrapper around gtk_print_settings_set_number_up_layout().
func (ps *PrintSettings) SetNumberUpLayout(numberUpLayout NumberUpLayout) {
	C.gtk_print_settings_set_number_up_layout(ps.native(), C.GtkNumberUpLayout(numberUpLayout))
}

// GetResolution() is a wrapper around gtk_print_settings_get_resolution().
func (ps *PrintSettings) GetResolution() int {
	c := C.gtk_print_settings_get_resolution(ps.native())
	return int(c)
}

// SetResolution() is a wrapper around gtk_print_settings_set_resolution().
func (ps *PrintSettings) SetResolution(resolution int) {
	C.gtk_print_settings_set_resolution(ps.native(), C.gint(resolution))
}

// SetResolutionXY() is a wrapper around gtk_print_settings_set_resolution_xy().
func (ps *PrintSettings) SetResolutionXY(resolutionX, resolutionY int) {
	C.gtk_print_settings_set_resolution_xy(ps.native(), C.gint(resolutionX), C.gint(resolutionY))
}

// GetResolutionX() is a wrapper around gtk_print_settings_get_resolution_x().
func (ps *PrintSettings) GetResolutionX() int {
	c := C.gtk_print_settings_get_resolution_x(ps.native())
	return int(c)
}

// GetResolutionY() is a wrapper around gtk_print_settings_get_resolution_y().
func (ps *PrintSettings) GetResolutionY() int {
	c := C.gtk_print_settings_get_resolution_y(ps.native())
	return int(c)
}

// GetPrinterLpi() is a wrapper around gtk_print_settings_get_printer_lpi().
func (ps *PrintSettings) GetPrinterLpi() float64 {
	c := C.gtk_print_settings_get_printer_lpi(ps.native())
	return float64(c)
}

// SetPrinterLpi() is a wrapper around gtk_print_settings_set_printer_lpi().
func (ps *PrintSettings) SetPrinterLpi(lpi float64) {
	C.gtk_print_settings_set_printer_lpi(ps.native(), C.gdouble(lpi))
}

// GetScale() is a wrapper around gtk_print_settings_get_scale().
func (ps *PrintSettings) GetScale() float64 {
	c := C.gtk_print_settings_get_scale(ps.native())
	return float64(c)
}

// SetScale() is a wrapper around gtk_print_settings_set_scale().
func (ps *PrintSettings) SetScale(scale float64) {
	C.gtk_print_settings_set_scale(ps.native(), C.gdouble(scale))
}

// GetPrintPages() is a wrapper around gtk_print_settings_get_print_pages().
func (ps *PrintSettings) GetPrintPages() PrintPages {
	c := C.gtk_print_settings_get_print_pages(ps.native())
	return PrintPages(c)
}

// SetPrintPages() is a wrapper around gtk_print_settings_set_print_pages().
func (ps *PrintSettings) SetPrintPages(pages PrintPages) {
	C.gtk_print_settings_set_print_pages(ps.native(), C.GtkPrintPages(pages))
}

// GetPageSet() is a wrapper around gtk_print_settings_get_page_set().
func (ps *PrintSettings) GetPageSet(pages PrintPages) PageSet {
	c := C.gtk_print_settings_get_page_set(ps.native())
	return PageSet(c)
}

// SetPageSet() is a wrapper around gtk_print_settings_set_page_set().
func (ps *PrintSettings) SetPageSet(pageSet PageSet) {
	C.gtk_print_settings_set_page_set(ps.native(), C.GtkPageSet(pageSet))
}

// GetDefaultSource() is a wrapper around gtk_print_settings_get_default_source().
func (ps *PrintSettings) GetDefaultSource() string {
	c := C.gtk_print_settings_get_default_source(ps.native())
	return C.GoString((*C.char)(c))
}

// SetSefaultSource() is a wrapper around gtk_print_settings_set_default_source().
func (ps *PrintSettings) SetSefaultSource(defaultSource string) {
	cstr := C.CString(defaultSource)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_print_settings_set_default_source(ps.native(), (*C.gchar)(cstr))
}

// GetMediaType() is a wrapper around gtk_print_settings_get_media_type().
func (ps *PrintSettings) GetMediaType() string {
	c := C.gtk_print_settings_get_media_type(ps.native())
	return C.GoString((*C.char)(c))
}

// SetMediaType() is a wrapper around gtk_print_settings_set_media_type().
func (ps *PrintSettings) SetMediaType(mediaType string) {
	cstr := C.CString(mediaType)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_print_settings_set_media_type(ps.native(), (*C.gchar)(cstr))
}

// GetDither() is a wrapper around gtk_print_settings_get_dither().
func (ps *PrintSettings) GetDither() string {
	c := C.gtk_print_settings_get_dither(ps.native())
	return C.GoString((*C.char)(c))
}

// SetDither() is a wrapper around gtk_print_settings_set_dither().
func (ps *PrintSettings) SetDither(dither string) {
	cstr := C.CString(dither)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_print_settings_set_dither(ps.native(), (*C.gchar)(cstr))
}

// GetFinishings() is a wrapper around gtk_print_settings_get_finishings().
func (ps *PrintSettings) GetFinishings() string {
	c := C.gtk_print_settings_get_finishings(ps.native())
	return C.GoString((*C.char)(c))
}

// SetFinishings() is a wrapper around gtk_print_settings_set_finishings().
func (ps *PrintSettings) SetFinishings(dither string) {
	cstr := C.CString(dither)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_print_settings_set_finishings(ps.native(), (*C.gchar)(cstr))
}

// GetOutputBin() is a wrapper around gtk_print_settings_get_output_bin().
func (ps *PrintSettings) GetOutputBin() string {
	c := C.gtk_print_settings_get_output_bin(ps.native())
	return C.GoString((*C.char)(c))
}

// SetOutputBin() is a wrapper around gtk_print_settings_set_output_bin().
func (ps *PrintSettings) SetOutputBin(bin string) {
	cstr := C.CString(bin)
	defer C.free(unsafe.Pointer(cstr))
	C.gtk_print_settings_set_output_bin(ps.native(), (*C.gchar)(cstr))
}

// PrintSettingsNewFromFile() is a wrapper around gtk_print_settings_new_from_file().
func PrintSettingsNewFromFile(name string) (*PrintSettings, error) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	var err *C.GError = nil
	c := C.gtk_print_settings_new_from_file((*C.gchar)(cstr), &err)
	if c == nil {
		defer C.g_error_free(err)
		return nil, errors.New(C.GoString((*C.char)(err.message)))
	}
	obj := glib.Take(unsafe.Pointer(c))
	return wrapPrintSettings(obj), nil
}

// PrintSettingsNewFromKeyFile() is a wrapper around gtk_print_settings_new_from_key_file().

// LoadFile() is a wrapper around gtk_print_settings_load_file().
func (ps *PrintSettings) LoadFile(name string) error {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	var err *C.GError = nil
	c := C.gtk_print_settings_load_file(ps.native(), (*C.gchar)(cstr), &err)
	if gobool(c) == false {
		defer C.g_error_free(err)
		return errors.New(C.GoString((*C.char)(err.message)))
	}
	return nil
}

// LoadKeyFile() is a wrapper around gtk_print_settings_load_key_file().

// ToFile() is a wrapper around gtk_print_settings_to_file().
func (ps *PrintSettings) ToFile(name string) error {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	var err *C.GError = nil
	c := C.gtk_print_settings_to_file(ps.native(), (*C.gchar)(cstr), &err)
	if gobool(c) == false {
		return errors.New(C.GoString((*C.char)(err.message)))
	}
	return nil
}

// ToKeyFile() is a wrapper around gtk_print_settings_to_key_file().
