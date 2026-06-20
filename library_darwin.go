//go:build darwin

package vips

import "os"

// getLibraryPath returns the path to libvips and the list of checked paths.
// Returns empty string if not found.
func getLibraryPath() (string, []string) {
	paths := []string{
		"/opt/homebrew/lib/libvips.dylib", // arm64 (M1/M2)
		"/usr/local/lib/libvips.dylib",    // amd64 Intel
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, paths
		}
	}
	return "", paths
}

// getGLibPath returns the path to glib and the list of checked paths.
func getGLibPath() (string, []string) {
	paths := []string{
		"/opt/homebrew/lib/libglib-2.0.dylib",
		"/usr/local/lib/libglib-2.0.dylib",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, paths
		}
	}
	return "", paths
}

// getGObjectPath returns the path to gobject and the list of checked paths.
func getGObjectPath() (string, []string) {
	paths := []string{
		"/opt/homebrew/lib/libgobject-2.0.dylib",
		"/usr/local/lib/libgobject-2.0.dylib",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, paths
		}
	}
	return "", paths
}
