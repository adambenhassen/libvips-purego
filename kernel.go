package vips

import "fmt"

// Kernel defines the resampling kernel for resize operations.
type Kernel int

const (
	KernelNearest  Kernel = 0 // Nearest neighbor - fastest, pixelated
	KernelLinear   Kernel = 1 // Bilinear interpolation
	KernelCubic    Kernel = 2 // Bicubic interpolation
	KernelMitchell Kernel = 3 // Mitchell-Netravali - good for photos
	KernelLanczos2 Kernel = 4 // Lanczos with a=2
	KernelLanczos3 Kernel = 5 // Lanczos with a=3 - sharpest, recommended
)

// String returns the kernel name for debugging.
func (k Kernel) String() string {
	switch k {
	case KernelNearest:
		return "nearest"
	case KernelLinear:
		return "linear"
	case KernelCubic:
		return "cubic"
	case KernelMitchell:
		return "mitchell"
	case KernelLanczos2:
		return "lanczos2"
	case KernelLanczos3:
		return "lanczos3"
	default:
		return fmt.Sprintf("Kernel(%d)", k)
	}
}
