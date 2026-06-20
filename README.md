<p align="center">
  <img src="assets/banner.svg" alt="libvips-purego" width="100%">
</p>

<p align="center">
  Minimal, <strong>cgo-free</strong> Go bindings for <a href="https://www.libvips.org/">libvips</a>, built on <a href="https://github.com/ebitengine/purego">purego</a>.
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/adambenhassen/libvips-purego"><img src="https://pkg.go.dev/badge/github.com/adambenhassen/libvips-purego.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/adambenhassen/libvips-purego"><img src="https://goreportcard.com/badge/github.com/adambenhassen/libvips-purego" alt="Go Report Card"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>
</p>

libvips is loaded at runtime via `dlopen`, so you get fast, low-memory image processing without a C toolchain, `CGO_ENABLED=1`, or cross-compilation headaches.

## Why

- **No cgo** — builds with `CGO_ENABLED=0`; cross-compile freely.
- **Fast & low-memory** — libvips streams images and uses a fraction of the memory of most decoders.
- **Tiny surface** — a focused API for the common resize / crop / sharpen / encode pipeline.

## Requirements

- Go 1.25+
- libvips installed on the host (the shared library is loaded at runtime)

### Installing libvips

```sh
# macOS
brew install vips

# Debian / Ubuntu
apt-get install libvips42
```

The loader searches standard install locations:

| Platform | Paths searched |
| --- | --- |
| macOS | `/opt/homebrew/lib` (Apple Silicon), `/usr/local/lib` (Intel) |
| Linux | `/usr/lib/x86_64-linux-gnu`, `/usr/lib/aarch64-linux-gnu`, then `LD_LIBRARY_PATH` |

## Install

```sh
go get github.com/adambenhassen/libvips-purego
```

## Quick start

```go
package main

import (
	"log"
	"os"

	vips "github.com/adambenhassen/libvips-purego"
)

func main() {
	if err := vips.Startup(&vips.Config{MaxCacheSize: 100}); err != nil {
		log.Fatal(err)
	}
	defer vips.Shutdown()

	data, err := os.ReadFile("input.jpg")
	if err != nil {
		log.Fatal(err)
	}

	img, err := vips.NewImageFromBuffer(data)
	if err != nil {
		log.Fatal(err)
	}
	defer img.Close()

	// Resize to 50% and sharpen.
	if err := img.Resize(0.5, vips.KernelLanczos3); err != nil {
		log.Fatal(err)
	}
	if err := img.Sharpen(0.5, 1.5, 2.0); err != nil {
		log.Fatal(err)
	}

	out, _, err := img.ExportWebp(&vips.WebpExportParams{Quality: 80, StripMetadata: true})
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile("output.webp", out, 0o644); err != nil {
		log.Fatal(err)
	}
}
```

## Usage

### Lifecycle

`Startup` must be called once before any image work, and `Shutdown` once before exit.

```go
vips.Startup(&vips.Config{
	MaxCacheSize: 100,       // max cached operations (0 disables)
	MaxCacheMem:  50 << 20,  // max cache memory in bytes (0 disables)
})
defer vips.Shutdown()

vips.Version()    // e.g. "8.15.0"
vips.ClearCache() // drop cached operations, keep configured limits
```

### Working with images

```go
img, err := vips.NewImageFromBuffer(data) // JPEG, PNG, WebP, GIF, ...
defer img.Close()                         // idempotent; also run by a finalizer

img.Width()
img.Height()

img.Resize(2.0, vips.KernelLanczos3)      // scale factor + resampling kernel
img.ExtractArea(left, top, w, h)          // crop
img.Sharpen(sigma, x1, m2)                // unsharp mask
```

Operations mutate the image **in place**, so they compose naturally in a pipeline.

### Encoding

```go
out, _, err := img.ExportWebp(&vips.WebpExportParams{
	Quality:       80,   // 0–100, clamped
	StripMetadata: true,
})
```

### Resampling kernels

| Kernel | Notes |
| --- | --- |
| `KernelNearest` | Fastest, pixelated |
| `KernelLinear` | Bilinear |
| `KernelCubic` | Bicubic |
| `KernelMitchell` | Mitchell–Netravali, good for photos |
| `KernelLanczos2` | Lanczos, a=2 |
| `KernelLanczos3` | Sharpest — recommended default |

## Concurrency

An `Image` is **not** safe for concurrent use. Each goroutine should create and own its own `Image`, or callers must synchronize access. `Startup`, `Shutdown`, and the cache helpers are safe to call from multiple goroutines.

## Scope

This is a deliberately small binding covering load → resize / crop / sharpen → WebP. It is not a full libvips wrapper. PRs that extend the operation and codec coverage are welcome.

## License

MIT — see [LICENSE](LICENSE).
