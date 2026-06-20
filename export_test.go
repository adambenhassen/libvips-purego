package vips

import (
	"testing"
)

func TestExportWebp(t *testing.T) {
	data := loadTestImage(t, "test.jpg")

	img, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	defer img.Close()

	// Export with default params
	webpData, _, err := img.ExportWebp(nil)
	if err != nil {
		t.Fatalf("ExportWebp failed: %v", err)
	}

	// Check WebP magic bytes: "RIFF" + 4 bytes size + "WEBP"
	if len(webpData) < 12 {
		t.Fatal("WebP data too short")
	}

	if string(webpData[0:4]) != "RIFF" {
		t.Errorf("expected RIFF header, got %q", string(webpData[0:4]))
	}
	if string(webpData[8:12]) != "WEBP" {
		t.Errorf("expected WEBP signature, got %q", string(webpData[8:12]))
	}

	t.Logf("Exported WebP: %d bytes", len(webpData))
}

func TestExportWebpWithParams(t *testing.T) {
	data := loadTestImage(t, "test.jpg")

	img, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	defer img.Close()

	params := &WebpExportParams{
		Quality:       80,
		StripMetadata: true,
	}

	webpData, _, err := img.ExportWebp(params)
	if err != nil {
		t.Fatalf("ExportWebp failed: %v", err)
	}

	if len(webpData) == 0 {
		t.Error("expected non-empty WebP data")
	}
}

func TestExportWebpQuality(t *testing.T) {
	data := loadTestImage(t, "test.jpg")

	// Export at quality 20
	img1, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	defer img1.Close()

	lowQuality, _, err := img1.ExportWebp(&WebpExportParams{Quality: 20})
	if err != nil {
		t.Fatalf("ExportWebp (Q=20) failed: %v", err)
	}

	// Export at quality 95
	img2, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	defer img2.Close()

	highQuality, _, err := img2.ExportWebp(&WebpExportParams{Quality: 95})
	if err != nil {
		t.Fatalf("ExportWebp (Q=95) failed: %v", err)
	}

	t.Logf("Low quality (Q=20): %d bytes", len(lowQuality))
	t.Logf("High quality (Q=95): %d bytes", len(highQuality))

	// Verify both are valid WebP
	if len(lowQuality) < 12 || string(lowQuality[0:4]) != "RIFF" {
		t.Error("low quality output is not valid WebP")
	}
	if len(highQuality) < 12 || string(highQuality[0:4]) != "RIFF" {
		t.Error("high quality output is not valid WebP")
	}
}

func TestExportWebpFromPNG(t *testing.T) {
	data := loadTestImage(t, "test.png")

	img, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	defer img.Close()

	webpData, _, err := img.ExportWebp(nil)
	if err != nil {
		t.Fatalf("ExportWebp failed: %v", err)
	}

	// Verify magic bytes
	if len(webpData) < 12 {
		t.Fatal("WebP data too short")
	}
	if string(webpData[0:4]) != "RIFF" || string(webpData[8:12]) != "WEBP" {
		t.Error("invalid WebP format")
	}

	t.Logf("PNG to WebP: %d bytes (original PNG: %d bytes)", len(webpData), len(data))
}

func TestNewWebpExportParams(t *testing.T) {
	params := NewWebpExportParams()
	if params.Quality != 75 {
		t.Errorf("expected default quality 75, got %d", params.Quality)
	}
	if params.StripMetadata != false {
		t.Error("expected default StripMetadata to be false")
	}
}

func TestExportWebpQualityClamping(t *testing.T) {
	data := loadTestImage(t, "test.jpg")

	// Test negative quality (should clamp to 0)
	img1, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	defer img1.Close()

	_, _, err = img1.ExportWebp(&WebpExportParams{Quality: -10})
	if err != nil {
		t.Errorf("negative quality should be clamped, got error: %v", err)
	}

	// Test quality > 100 (should clamp to 100)
	img2, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	defer img2.Close()

	_, _, err = img2.ExportWebp(&WebpExportParams{Quality: 150})
	if err != nil {
		t.Errorf("quality > 100 should be clamped, got error: %v", err)
	}
}

func TestExportWebpClosedImage(t *testing.T) {
	data := loadTestImage(t, "test.jpg")

	img, err := NewImageFromBuffer(data)
	if err != nil {
		t.Fatalf("NewImageFromBuffer failed: %v", err)
	}
	img.Close()

	_, _, err = img.ExportWebp(nil)
	if err == nil {
		t.Error("expected error when exporting closed image")
	}
	if err.Error() != "image is closed" {
		t.Errorf("expected 'image is closed' error, got: %v", err)
	}
}
