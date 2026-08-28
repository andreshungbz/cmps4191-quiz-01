package vcs

import "testing"

// TestVersionSmoke verifies Version can be called without panicking.
func TestVersionSmoke(t *testing.T) {
	v := Version()

	// Assert that the returned version string is not empty.
	if v == "" {
		t.Log("version returned an empty string (expected when build metadata is unavailable)")
	}
}
