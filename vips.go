// Package vips provides minimal purego bindings for libvips image processing.
// This package is NOT safe for concurrent use of the same Image instance.
// Callers must synchronize access if sharing Images across goroutines.
package vips

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/adambenhassen/libvips-purego/internal/ffi"
	"github.com/ebitengine/purego"
)

// Config holds initialization options for libvips.
type Config struct {
	MaxCacheSize int // Max operations to cache (0 = disable cache)
	MaxCacheMem  int // Max memory for cache in bytes (0 = disable cache memory limit)
}

var (
	initialized atomic.Bool
	initMu      sync.Mutex
	vipsLib     uintptr
	glibLib     uintptr
	gobjectLib  uintptr

	// Default cache values (set during Startup, used by ClearCache)
	defaultMaxCacheSize int
	defaultMaxCacheMem  uint64
)

// isInitialized returns true if vips has been initialized.
// Thread-safe via atomic.Bool.
func isInitialized() bool {
	return initialized.Load()
}

// Startup initializes libvips with the given configuration.
// Must be called before any image operations.
func Startup(config *Config) error {
	initMu.Lock()
	defer initMu.Unlock()

	if initialized.Load() {
		return nil
	}

	// Load libraries
	libPath, checkedPaths := getLibraryPath()
	if libPath == "" {
		return fmt.Errorf("libvips not found. Checked paths: %v", checkedPaths)
	}

	var err error
	vipsLib, err = purego.Dlopen(libPath, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("failed to load libvips from %s: %w", libPath, err)
	}

	glibPath, checkedPaths := getGLibPath()
	if glibPath == "" {
		return fmt.Errorf("glib not found. Checked paths: %v", checkedPaths)
	}
	glibLib, err = purego.Dlopen(glibPath, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("failed to load glib from %s: %w", glibPath, err)
	}

	gobjectPath, checkedPaths := getGObjectPath()
	if gobjectPath == "" {
		return fmt.Errorf("gobject not found. Checked paths: %v", checkedPaths)
	}
	gobjectLib, err = purego.Dlopen(gobjectPath, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("failed to load gobject from %s: %w", gobjectPath, err)
	}

	// Register FFI functions
	ffi.Register(vipsLib, glibLib, gobjectLib)

	// Initialize libvips with program name to suppress GLib warnings
	if result := ffi.VipsInit(cstrProgName); result != 0 {
		return fmt.Errorf("vips_init failed: %s", getError())
	}

	// Apply configuration
	if config != nil {
		if config.MaxCacheSize > 0 {
			defaultMaxCacheSize = config.MaxCacheSize
			ffi.VipsCacheSetMax(config.MaxCacheSize)
		} else if config.MaxCacheSize == 0 {
			// Explicitly disable cache
			ffi.VipsCacheSetMax(0)
		}

		if config.MaxCacheMem > 0 {
			defaultMaxCacheMem = uint64(config.MaxCacheMem)
			ffi.VipsCacheSetMaxMem(defaultMaxCacheMem)
		} else if config.MaxCacheMem == 0 {
			// Explicitly disable cache memory
			ffi.VipsCacheSetMaxMem(0)
		}
	}

	initialized.Store(true)
	return nil
}

// Shutdown releases all libvips resources.
// Should be called before program exit.
func Shutdown() {
	initMu.Lock()
	defer initMu.Unlock()

	if !initialized.Load() {
		return
	}

	ffi.VipsShutdown()
	initialized.Store(false)
}

// Version returns the libvips version string (e.g., "8.15.0").
// Returns empty string if vips is not initialized or version is unavailable.
// Call Startup() before calling this function.
func Version() string {
	if !isInitialized() {
		return ""
	}
	ptr := ffi.VipsVersionString()
	if ptr == 0 {
		return ""
	}
	return goString(ptr)
}

// ClearCache clears the libvips operation cache.
// Sets cache to 0, then restores to configured defaults.
// Does nothing if vips is not initialized.
func ClearCache() {
	if !isInitialized() {
		return
	}
	// Clear by setting to 0
	ffi.VipsCacheSetMax(0)
	ffi.VipsCacheSetMaxMem(0)

	// Restore defaults if they were set
	if defaultMaxCacheSize > 0 {
		ffi.VipsCacheSetMax(defaultMaxCacheSize)
	}
	if defaultMaxCacheMem > 0 {
		ffi.VipsCacheSetMaxMem(defaultMaxCacheMem)
	}
}

// getError retrieves the current error message from libvips and clears it.
func getError() error {
	ptr := ffi.VipsErrorBuffer()
	ffi.VipsErrorClear()
	if ptr == 0 {
		return errors.New("unknown vips error")
	}
	msg := goString(ptr)
	if msg == "" {
		return errors.New("unknown vips error")
	}
	return errors.New(msg)
}

// goString converts a C string pointer to a Go string.
// Limits reading to maxCStringLen bytes to prevent runaway reads on corruption.
func goString(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	// Read bytes until null terminator, with safety limit
	const maxLen = 64 * 1024 // 64KB - more than enough for any error message
	var buf []byte
	for i := 0; i < maxLen; i++ {
		b := *(*byte)(unsafe.Pointer(ptr))
		if b == 0 {
			break
		}
		buf = append(buf, b)
		ptr++
	}
	return string(buf)
}
