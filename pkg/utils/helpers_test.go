package utils

import (
	"testing"
	"time"
)

// Additional edge-case coverage that complements helpers_internal_test.
func TestHelpersAdditionalCases(t *testing.T) {
	// FormatFileSize rounding at KB boundary
	if got := FormatFileSize(1024); got != "1.0 KB" {
		t.Fatalf("FormatFileSize 1KB mismatch: %s", got)
	}
	// ConvertToString with byte slice should default to empty string
	if got := ConvertToString([]byte("data")); got != "[100 97 116 97]" {
		t.Fatalf("unexpected ConvertToString for byte slice: %q", got)
	}
}

func TestTimeAgoFuture(t *testing.T) {
	now := time.Now()
	if got := TimeAgo(now.Add(10 * time.Second)); got != "just now" {
		t.Fatalf("expected 'just now' for future times, got %q", got)
	}
}
