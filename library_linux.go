//go:build linux

package vips

import (
	"os"
	"path/filepath"
)

// getLibraryPath returns the path to libvips and the list of checked paths.
// Returns empty string and all checked paths if not found at any absolute path.
// For bare library names (dlopen fallback), returns the name to try.
func getLibraryPath() (string, []string) {
	paths := []string{
		"/usr/lib/x86_64-linux-gnu/libvips.so.42",  // Debian amd64
		"/usr/lib/aarch64-linux-gnu/libvips.so.42", // Debian arm64
		"libvips.so.42", // LD_LIBRARY_PATH fallback
	}
	for _, p := range paths {
		// For absolute paths, check existence; for bare names, always try via dlopen
		if !filepath.IsAbs(p) {
			return p, paths
		}
		if _, err := os.Stat(p); err == nil {
			return p, paths
		}
	}
	return "", paths
}

// getGLibPath returns the path to glib and the list of checked paths.
func getGLibPath() (string, []string) {
	paths := []string{
		"/usr/lib/x86_64-linux-gnu/libglib-2.0.so.0",
		"/usr/lib/aarch64-linux-gnu/libglib-2.0.so.0",
		"libglib-2.0.so.0", // LD_LIBRARY_PATH fallback
	}
	for _, p := range paths {
		if !filepath.IsAbs(p) {
			return p, paths
		}
		if _, err := os.Stat(p); err == nil {
			return p, paths
		}
	}
	return "", paths
}

// getGObjectPath returns the path to gobject and the list of checked paths.
func getGObjectPath() (string, []string) {
	paths := []string{
		"/usr/lib/x86_64-linux-gnu/libgobject-2.0.so.0",
		"/usr/lib/aarch64-linux-gnu/libgobject-2.0.so.0",
		"libgobject-2.0.so.0", // LD_LIBRARY_PATH fallback
	}
	for _, p := range paths {
		if !filepath.IsAbs(p) {
			return p, paths
		}
		if _, err := os.Stat(p); err == nil {
			return p, paths
		}
	}
	return "", paths
}
