package ffi

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

// C function pointers registered via purego
var (
	// Lifecycle
	VipsInit          func(argv0 uintptr) int
	VipsShutdown      func()
	VipsVersionString func() uintptr

	// Cache management
	VipsCacheSetMax    func(max int)
	VipsCacheSetMaxMem func(maxMem uint64)

	// Image loading
	VipsImageNewFromBuffer func(buf unsafe.Pointer, length uint64, optionString uintptr, args ...any) uintptr

	// Image properties
	VipsImageGetWidth  func(image uintptr) int
	VipsImageGetHeight func(image uintptr) int

	// Operations
	VipsResize      func(in uintptr, out *uintptr, scale float64, args ...any) int
	VipsExtractArea func(in uintptr, out *uintptr, left, top, width, height int) int
	VipsSharpen     func(in uintptr, out *uintptr, args ...any) int

	// Export
	VipsWebpsaveBuffer func(in uintptr, buf *uintptr, length *uint64, args ...any) int

	// Error handling
	VipsErrorBuffer func() uintptr
	VipsErrorClear  func()

	// GLib memory management
	GObjectUnref func(object uintptr)
	GFree        func(mem uintptr)
)

// Register loads the libraries and registers all function pointers.
func Register(vipsLib, glibLib, gobjectLib uintptr) {
	// Lifecycle
	purego.RegisterLibFunc(&VipsInit, vipsLib, "vips_init")
	purego.RegisterLibFunc(&VipsShutdown, vipsLib, "vips_shutdown")
	purego.RegisterLibFunc(&VipsVersionString, vipsLib, "vips_version_string")

	// Cache management
	purego.RegisterLibFunc(&VipsCacheSetMax, vipsLib, "vips_cache_set_max")
	purego.RegisterLibFunc(&VipsCacheSetMaxMem, vipsLib, "vips_cache_set_max_mem")

	// Image loading
	purego.RegisterLibFunc(&VipsImageNewFromBuffer, vipsLib, "vips_image_new_from_buffer")

	// Image properties
	purego.RegisterLibFunc(&VipsImageGetWidth, vipsLib, "vips_image_get_width")
	purego.RegisterLibFunc(&VipsImageGetHeight, vipsLib, "vips_image_get_height")

	// Operations
	purego.RegisterLibFunc(&VipsResize, vipsLib, "vips_resize")
	purego.RegisterLibFunc(&VipsExtractArea, vipsLib, "vips_extract_area")
	purego.RegisterLibFunc(&VipsSharpen, vipsLib, "vips_sharpen")

	// Export
	purego.RegisterLibFunc(&VipsWebpsaveBuffer, vipsLib, "vips_webpsave_buffer")

	// Error handling
	purego.RegisterLibFunc(&VipsErrorBuffer, vipsLib, "vips_error_buffer")
	purego.RegisterLibFunc(&VipsErrorClear, vipsLib, "vips_error_clear")

	// GLib/GObject functions
	purego.RegisterLibFunc(&GObjectUnref, gobjectLib, "g_object_unref")
	purego.RegisterLibFunc(&GFree, glibLib, "g_free")
}
