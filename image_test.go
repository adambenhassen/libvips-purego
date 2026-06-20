package vips

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func loadTestImage(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("failed to read test image %s: %v", name, err)
	}
	return data
}

func TestNewImageFromBuffer(t *testing.T) {
	data := loadTestImage(t, "test.jpg")

	img, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	defer img.Close()

	if !img.IsValid() {
		t.Error("expected valid image")
	}
}

func TestNewImageFromBufferPNG(t *testing.T) {
	data := loadTestImage(t, "test.png")

	img, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	defer img.Close()

	if !img.IsValid() {
		t.Error("expected valid image")
	}
}

func TestNewImageFromBufferEmpty(t *testing.T) {
	_, err := NewImageFromBuffer(nil)
	if err == nil {
		t.Error("expected error for nil data")
	}

	_, err = NewImageFromBuffer([]byte{})
	if err == nil {
		t.Error("expected error for empty data")
	}
}

func TestNewImageFromBufferInvalidData(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{"random bytes", []byte{0x01, 0x02, 0x03, 0x04, 0x05}},
		{"truncated JPEG header", []byte{0xFF, 0xD8, 0xFF, 0xE0}},
		{"text content", []byte("this is not an image")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			img, err := NewImageFromBuffer(tc.data)
			if err == nil {
				img.Close()
				t.Error("expected error for invalid image data")
			}
		})
	}
}

func TestImageDimensions(t *testing.T) {
	data := loadTestImage(t, "test.jpg")

	img, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	defer img.Close()

	w, h := img.Width(), img.Height()
	t.Logf("Image dimensions: %dx%d", w, h)

	if w != 272 {
		t.Errorf("expected width 272, got %d", w)
	}
	if h != 92 {
		t.Errorf("expected height 92, got %d", h)
	}
}

func TestResize(t *testing.T) {
	data := loadTestImage(t, "test.jpg")

	img, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	defer img.Close()

	origW, origH := img.Width(), img.Height()

	// Resize to 50%
	err = img.Resize(0.5, KernelLanczos3)
	if err != nil {
		t.Fatalf("Resize failed: %v", err)
	}

	newW, newH := img.Width(), img.Height()
	t.Logf("Resized from %dx%d to %dx%d", origW, origH, newW, newH)

	// Check new dimensions (should be approximately half)
	expectedW := origW / 2
	expectedH := origH / 2

	// Allow 1 pixel tolerance for rounding
	if newW < expectedW-1 || newW > expectedW+1 {
		t.Errorf("expected width ~%d, got %d", expectedW, newW)
	}
	if newH < expectedH-1 || newH > expectedH+1 {
		t.Errorf("expected height ~%d, got %d", expectedH, newH)
	}
}

func TestResizeInvalidScale(t *testing.T) {
	data := loadTestImage(t, "test.jpg")

	testCases := []struct {
		name  string
		scale float64
	}{
		{"zero scale", 0},
		{"negative scale", -0.5},
		{"very negative", -1000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			img, err := NewImageFromBuffer(data)
			if err != nil {
				t.Fatalf("NewImageFromBuffer failed: %v", err)
			}
			defer img.Close()

			err = img.Resize(tc.scale, KernelLanczos3)
			if err == nil {
				t.Errorf("expected error for scale %v", tc.scale)
			}
		})
	}
}

func TestResizeInvalidKernel(t *testing.T) {
	data := loadTestImage(t, "test.jpg")

	img, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	defer img.Close()

	err = img.Resize(0.5, Kernel(99))
	if err == nil {
		t.Error("expected error for invalid kernel")
	}
}

func TestResizeAllKernels(t *testing.T) {
	data := loadTestImage(t, "test.jpg")
	kernels := []Kernel{
		KernelNearest, KernelLinear, KernelCubic,
		KernelMitchell, KernelLanczos2, KernelLanczos3,
	}

	for _, k := range kernels {
		t.Run(fmt.Sprintf("kernel_%d", k), func(t *testing.T) {
			img, err := NewImageFromBuffer(data)
			if err != nil {
				t.Fatalf("NewImageFromBuffer failed: %v", err)
			}
			defer img.Close()

			err = img.Resize(0.5, k)
			if err != nil {
				t.Errorf("Resize with kernel %d failed: %v", k, err)
			}
		})
	}
}

func TestExtractArea(t *testing.T) {
	data := loadTestImage(t, "test.jpg")

	img, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	defer img.Close()

	// Extract center 50x50 region
	err = img.ExtractArea(10, 10, 50, 50)
	if err != nil {
		t.Fatalf("ExtractArea failed: %v", err)
	}

	w, h := img.Width(), img.Height()
	if w != 50 {
		t.Errorf("expected width 50, got %d", w)
	}
	if h != 50 {
		t.Errorf("expected height 50, got %d", h)
	}
}

func TestExtractAreaInvalidParams(t *testing.T) {
	data := loadTestImage(t, "test.jpg") // 272x92 image

	testCases := []struct {
		name            string
		left, top, w, h int
	}{
		{"zero width", 0, 0, 0, 50},
		{"zero height", 0, 0, 50, 0},
		{"negative width", 0, 0, -10, 50},
		{"negative left", -1, 0, 50, 50},
		{"negative top", 0, -1, 50, 50},
		{"exceeds width", 250, 0, 50, 50},  // 250 + 50 > 272
		{"exceeds height", 0, 80, 50, 50},  // 80 + 50 > 92
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			img, err := NewImageFromBuffer(data)
			if err != nil {
				t.Fatalf("NewImageFromBuffer failed: %v", err)
			}
			defer img.Close()

			err = img.ExtractArea(tc.left, tc.top, tc.w, tc.h)
			if err == nil {
				t.Errorf("expected error for %s", tc.name)
			}
		})
	}
}

func TestSharpen(t *testing.T) {
	data := loadTestImage(t, "test.jpg")

	img, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	defer img.Close()

	// Apply sharpen - should not return error
	err = img.Sharpen(0.5, 1, 2)
	if err != nil {
		t.Fatalf("Sharpen failed: %v", err)
	}

	// Verify image is still valid
	if img.Width() == 0 || img.Height() == 0 {
		t.Error("image has zero dimensions after sharpen")
	}
}

func TestImageClose(t *testing.T) {
	data := loadTestImage(t, "test.jpg")

	img, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}

	// Close should not panic
	img.Close()

	// Double close should not panic
	img.Close()

	// Operations on closed image should fail gracefully
	if img.Width() != 0 {
		t.Error("expected Width to return 0 for closed image")
	}
	if img.Height() != 0 {
		t.Error("expected Height to return 0 for closed image")
	}
	if img.IsValid() {
		t.Error("expected IsValid to return false for closed image")
	}
}

func TestOperationsOnClosedImage(t *testing.T) {
	data := loadTestImage(t, "test.jpg")

	img, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	img.Close()

	// Resize should fail gracefully
	err = img.Resize(0.5, KernelLanczos3)
	if err == nil || err.Error() != "image is closed" {
		t.Errorf("Resize on closed image: expected 'image is closed' error, got %v", err)
	}

	// ExtractArea should fail gracefully
	err = img.ExtractArea(0, 0, 10, 10)
	if err == nil || err.Error() != "image is closed" {
		t.Errorf("ExtractArea on closed image: expected 'image is closed' error, got %v", err)
	}

	// Sharpen should fail gracefully
	err = img.Sharpen(0.5, 1, 2)
	if err == nil || err.Error() != "image is closed" {
		t.Errorf("Sharpen on closed image: expected 'image is closed' error, got %v", err)
	}
}

func TestOperationChain(t *testing.T) {
	data := loadTestImage(t, "test.jpg")

	img, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	defer img.Close()

	// Simulate real processing pipeline
	if err := img.Resize(0.5, KernelLanczos3); err != nil {
		t.Fatalf("resize failed: %v", err)
	}
	if err := img.Sharpen(0.5, 1, 2); err != nil {
		t.Fatalf("sharpen failed: %v", err)
	}
	if err := img.ExtractArea(0, 0, 50, 30); err != nil {
		t.Fatalf("extract failed: %v", err)
	}

	webp, _, err := img.ExportWebp(&WebpExportParams{Quality: 80})
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if len(webp) == 0 {
		t.Error("expected non-empty output")
	}

	// Verify WebP format
	if len(webp) < 12 || string(webp[0:4]) != "RIFF" {
		t.Error("output is not valid WebP")
	}
}

func TestIsValid(t *testing.T) {
	data := loadTestImage(t, "test.jpg")

	img, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}

	if !img.IsValid() {
		t.Error("expected new image to be valid")
	}

	img.Close()

	if img.IsValid() {
		t.Error("expected closed image to be invalid")
	}
}
