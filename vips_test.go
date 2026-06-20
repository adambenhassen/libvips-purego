package vips

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Initialize vips for all tests
	if err := Startup(&Config{
		MaxCacheSize: 0,
		MaxCacheMem:  0,
	}); err != nil {
		panic("failed to start vips: " + err.Error())
	}
	defer Shutdown()

	os.Exit(m.Run())
}

func TestStartupShutdown(t *testing.T) {
	// Already started in TestMain, just verify it's initialized
	if Version() == "" {
		t.Error("expected non-empty version string after Startup")
	}
}

func TestVersion(t *testing.T) {
	v := Version()
	if v == "" {
		t.Error("expected non-empty version string")
		return
	}
	t.Logf("libvips version: %s", v)
}

func TestClearCache(t *testing.T) {
	// Should not panic
	ClearCache()
}
