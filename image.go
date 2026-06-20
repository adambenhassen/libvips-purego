package vips

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"

	"github.com/adambenhassen/libvips-purego/internal/ffi"
)

// Pre-allocated C strings for libvips option names.
// These are allocated once at package init and never freed,
// avoiding the dangling pointer issue with dynamic allocation.
var (
	cstrKernel   = mustCString("kernel")
	cstrSigma    = mustCString("sigma")
	cstrX1       = mustCString("x1")
	cstrM2       = mustCString("m2")
	cstrQ        = mustCString("Q")
	cstrStrip    = mustCString("strip")
	cstrProgName = mustCString("vips") // Program name for vips_init
)

// mustCString allocates a permanent C string that will never be freed.
// Used for static option name strings passed to libvips.
func mustCString(s string) uintptr {
	b := make([]byte, len(s)+1)
	copy(b, s)
	b[len(s)] = 0
	// Keep a reference to prevent GC
	cstringKeepAlive = append(cstringKeepAlive, b)
	return uintptr(unsafe.Pointer(&b[0]))
}

// cstringKeepAlive holds references to C strings to prevent GC.
var cstringKeepAlive [][]byte

// Image represents a libvips image.
// Image is NOT safe for concurrent use from multiple goroutines.
// The caller must synchronize access if sharing an Image across goroutines.
type Image struct {
	ptr uintptr
}

// imageFinalizer releases the image if Close() was not called.
func imageFinalizer(img *Image) {
	if img.ptr != 0 {
		ffi.GObjectUnref(img.ptr)
		img.ptr = 0
	}
}

// NewImageFromBuffer loads an image from a byte buffer.
// Supports JPEG, PNG, WebP, GIF, and other formats that libvips can decode.
func NewImageFromBuffer(data []byte) (*Image, error) {
	if !isInitialized() {
		return nil, errors.New("vips not initialized: call Startup first")
	}
	if len(data) == 0 {
		return nil, errors.New("empty image data")
	}

	// Pin the data to prevent GC from moving it during the C call
	pinner := runtime.Pinner{}
	pinner.Pin(&data[0])
	defer pinner.Unpin()

	// Call vips_image_new_from_buffer(buf, len, NULL, NULL)
	// NULL option string, NULL terminator for varargs
	ptr := ffi.VipsImageNewFromBuffer(
		unsafe.Pointer(&data[0]),
		uint64(len(data)),
		uintptr(0), // NULL option string
	)

	if ptr == 0 {
		return nil, fmt.Errorf("failed to load image: %w", getError())
	}

	img := &Image{ptr: ptr}
	runtime.SetFinalizer(img, imageFinalizer)
	return img, nil
}

// IsValid returns true if the image has not been closed.
func (img *Image) IsValid() bool {
	return img.ptr != 0
}

// Width returns the image width in pixels.
// Returns 0 if the image is closed.
func (img *Image) Width() int {
	if img.ptr == 0 {
		return 0
	}
	return ffi.VipsImageGetWidth(img.ptr)
}

// Height returns the image height in pixels.
// Returns 0 if the image is closed.
func (img *Image) Height() int {
	if img.ptr == 0 {
		return 0
	}
	return ffi.VipsImageGetHeight(img.ptr)
}

// Resize scales the image by the given factor using the specified kernel.
// A scale of 0.5 halves the size, 2.0 doubles it.
// The operation modifies the image in place.
func (img *Image) Resize(scale float64, kernel Kernel) error {
	if img.ptr == 0 {
		return errors.New("image is closed")
	}
	if scale <= 0 {
		return errors.New("scale must be positive")
	}
	if kernel < KernelNearest || kernel > KernelLanczos3 {
		return fmt.Errorf("invalid kernel: %d", kernel)
	}

	var outPtr uintptr

	// vips_resize(in, out, scale, "kernel", kernel_value, NULL)
	result := ffi.VipsResize(
		img.ptr,
		&outPtr,
		scale,
		cstrKernel, int(kernel),
		uintptr(0), // NULL terminator
	)

	if result != 0 {
		return fmt.Errorf("resize failed: %w", getError())
	}

	// Unref old image, replace with new
	ffi.GObjectUnref(img.ptr)
	img.ptr = outPtr

	return nil
}

// ExtractArea extracts a rectangular region from the image.
// The operation modifies the image in place.
func (img *Image) ExtractArea(left, top, width, height int) error {
	if img.ptr == 0 {
		return errors.New("image is closed")
	}
	if width <= 0 || height <= 0 {
		return errors.New("width and height must be positive")
	}
	if left < 0 || top < 0 {
		return errors.New("left and top must be non-negative")
	}

	// Bounds check: ensure extraction region fits within image
	imgWidth, imgHeight := img.Width(), img.Height()
	if left+width > imgWidth {
		return fmt.Errorf("extraction region exceeds image width: left(%d) + width(%d) > %d", left, width, imgWidth)
	}
	if top+height > imgHeight {
		return fmt.Errorf("extraction region exceeds image height: top(%d) + height(%d) > %d", top, height, imgHeight)
	}

	var outPtr uintptr

	result := ffi.VipsExtractArea(
		img.ptr,
		&outPtr,
		left, top, width, height,
	)

	if result != 0 {
		return fmt.Errorf("extract area failed: %w", getError())
	}

	// Unref old image, replace with new
	ffi.GObjectUnref(img.ptr)
	img.ptr = outPtr

	return nil
}

// Sharpen applies unsharp masking to the image.
// Parameters:
//   - sigma: gaussian sigma value (typical: 0.5-1.0)
//   - x1: flat area threshold (typical: 1.0-2.0)
//   - m2: maximum brightening amount (typical: 2.0-4.0)
//
// The operation modifies the image in place.
func (img *Image) Sharpen(sigma, x1, m2 float64) error {
	if img.ptr == 0 {
		return errors.New("image is closed")
	}

	var outPtr uintptr

	// vips_sharpen(in, out, "sigma", sigma, "x1", x1, "m2", m2, NULL)
	result := ffi.VipsSharpen(
		img.ptr,
		&outPtr,
		cstrSigma, sigma,
		cstrX1, x1,
		cstrM2, m2,
		uintptr(0), // NULL terminator
	)

	if result != 0 {
		return fmt.Errorf("sharpen failed: %w", getError())
	}

	// Unref old image, replace with new
	ffi.GObjectUnref(img.ptr)
	img.ptr = outPtr

	return nil
}

// Close releases the image resources.
// The image must not be used after calling Close.
// Close is idempotent and safe to call multiple times.
func (img *Image) Close() {
	if img.ptr != 0 {
		ffi.GObjectUnref(img.ptr)
		img.ptr = 0
		runtime.SetFinalizer(img, nil) // Clear finalizer since we manually closed
	}
}
