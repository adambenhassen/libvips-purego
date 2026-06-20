package vips

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/adambenhassen/libvips-purego/internal/ffi"
)

// WebpExportParams configures WebP encoding options.
type WebpExportParams struct {
	Quality       int  // Quality factor (0-100, default 75). Values outside range are clamped.
	StripMetadata bool // Strip all metadata
}

// NewWebpExportParams returns WebpExportParams with default values.
func NewWebpExportParams() *WebpExportParams {
	return &WebpExportParams{
		Quality:       75,
		StripMetadata: false,
	}
}

// ExportWebp exports the image as WebP format.
// Returns the encoded bytes and the WebP metadata.
// The second return value is for API compatibility (always nil).
func (img *Image) ExportWebp(params *WebpExportParams) ([]byte, *ImageMetadata, error) {
	if img.ptr == 0 {
		return nil, nil, errors.New("image is closed")
	}

	if params == nil {
		params = NewWebpExportParams()
	}

	// Clamp quality to valid range
	quality := params.Quality
	if quality < 0 {
		quality = 0
	}
	if quality > 100 {
		quality = 100
	}

	var bufPtr uintptr
	var bufLen uint64

	// Convert strip to int for C (0 or 1)
	strip := 0
	if params.StripMetadata {
		strip = 1
	}

	// vips_webpsave_buffer(in, &buf, &len, "Q", quality, "strip", strip, NULL)
	result := ffi.VipsWebpsaveBuffer(
		img.ptr,
		&bufPtr,
		&bufLen,
		cstrQ, quality,
		cstrStrip, strip,
		uintptr(0), // NULL terminator
	)

	if result != 0 {
		return nil, nil, fmt.Errorf("WebP export failed: %w", getError())
	}

	if bufPtr == 0 || bufLen == 0 {
		return nil, nil, errors.New("WebP export produced no data")
	}

	// Copy the buffer to Go memory
	data := make([]byte, bufLen)
	src := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), bufLen)
	copy(data, src)

	// Free the C buffer
	ffi.GFree(bufPtr)

	return data, nil, nil
}

// ImageMetadata is a placeholder for API compatibility.
// Currently not implemented.
type ImageMetadata struct{}
