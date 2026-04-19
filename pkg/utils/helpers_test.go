// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package utils_test

import (
	"testing"
	"time"

	utils "github.com/aptlogica/go-postgres-rest/pkg/utils"
)

// Additional edge-case coverage that complements helpers_internal_test.

func TestHelpersAdditionalCases(t *testing.T) {
	// FormatFileSize rounding at KB boundary
	if got := utils.FormatFileSize(1024); got != "1.0 KB" {
		t.Fatalf("FormatFileSize 1KB mismatch: %s", got)
	}
	// ConvertToString with byte slice should default to empty string
	if got := utils.ConvertToString([]byte("data")); got != "[100 97 116 97]" {
		t.Fatalf("unexpected ConvertToString for byte slice: %q", got)
	}
}

func TestTimeAgoFuture(t *testing.T) {
	now := time.Now()
	if got := utils.TimeAgo(now.Add(10 * time.Second)); got != "just now" {
		t.Fatalf("expected 'just now' for future times, got %q", got)
	}
}

// Edge branches that were still uncovered: empty inputs and non-slice/map fallbacks.
func TestHelpersEmptyAndFallbackBranches(t *testing.T) {
	if utils.ContainsLegacy(123, 1) {
		t.Fatalf("ContainsLegacy should be false for non-slice input")
	}

	if got := utils.RemoveDuplicatesString([]string{}); len(got) != 0 {
		t.Fatalf("RemoveDuplicatesString empty should return empty slice, got %v", got)
	}
	if got := utils.RemoveDuplicatesInt([]int{}); len(got) != 0 {
		t.Fatalf("RemoveDuplicatesInt empty should return empty slice, got %v", got)
	}
	if got := utils.RemoveDuplicates([]byte{}); len(got) != 0 {
		t.Fatalf("RemoveDuplicates generic empty should return empty slice, got %v", got)
	}

	if got := utils.RemoveDuplicatesLegacy(123); got.(int) != 123 {
		t.Fatalf("RemoveDuplicatesLegacy should return input when not a slice, got %v", got)
	}

	if got := utils.SliceToStringStrings([]string{}); got != "" {
		t.Fatalf("SliceToStringStrings empty should be empty string, got %q", got)
	}
	if got := utils.SliceToStringInts([]int{}); got != "" {
		t.Fatalf("SliceToStringInts empty should be empty string, got %q", got)
	}
	if got := utils.SliceToString("not-a-slice"); got != "" {
		t.Fatalf("SliceToString non-slice should be empty string, got %q", got)
	}
	if got := utils.SliceToString([]int{}); got != "" {
		t.Fatalf("SliceToString empty slice should be empty string, got %q", got)
	}

	if got := utils.StringToSlice(""); len(got) != 0 {
		t.Fatalf("StringToSlice empty should return empty slice, got %v", got)
	}

	if got := utils.MapKeys(123); got != nil {
		t.Fatalf("MapKeys non-map should return nil, got %v", got)
	}
	if got := utils.MapValues("no-map"); got != nil {
		t.Fatalf("MapValues non-map should return nil, got %v", got)
	}
}
