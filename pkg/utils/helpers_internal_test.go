// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package utils_test

import (
	"bytes"
	"crypto/rand"
	"reflect"
	"testing"
	"time"

	utils "github.com/aptlogica/go-postgres-rest/pkg/utils"
)

// These tests live alongside the utils package to ensure coverage is reported for core helpers.

func TestGenerateIDDeterministic(t *testing.T) {
	original := rand.Reader
	t.Cleanup(func() { rand.Reader = original })
	rand.Reader = bytes.NewReader([]byte{0xAA, 0xBB, 0xCC, 0xDD})

	got := utils.GenerateID(4)
	if got != "aabbccdd" {
		t.Fatalf("expected deterministic hex, got %s", got)
	}
	if len(got) != 8 {
		t.Fatalf("expected length 8, got %d", len(got))
	}
}

func TestConvertHelpers(t *testing.T) {
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		name string
		in   interface{}
		want string
	}{
		{"string", "x", "x"},
		{"int", int64(5), "5"},
		{"uint", uint(7), "7"},
		{"float", 3.14, "3.14"},
		{"bool", true, "true"},
		{"time", now, now.Format(time.RFC3339)},
		{"nil", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.ConvertToString(tt.in); got != tt.want {
				t.Fatalf("ConvertToString mismatch: got %q want %q", got, tt.want)
			}
		})
	}

	if v, err := utils.ConvertToInt("10"); err != nil || v != 10 {
		t.Fatalf("ConvertToInt failed: %v %d", err, v)
	}
	if _, err := utils.ConvertToInt("bad"); err == nil {
		t.Fatalf("expected ConvertToInt error")
	}
	if v, err := utils.ConvertToFloat("1.5"); err != nil || v != 1.5 {
		t.Fatalf("ConvertToFloat failed: %v %f", err, v)
	}
	if _, err := utils.ConvertToFloat("bad"); err == nil {
		t.Fatalf("expected ConvertToFloat error")
	}
	if v, err := utils.ConvertToBool("true"); err != nil || v != true {
		t.Fatalf("ConvertToBool failed: %v %v", err, v)
	}
	if _, err := utils.ConvertToBool("bad"); err == nil {
		t.Fatalf("expected ConvertToBool error")
	}
}

func TestEmptyChecks(t *testing.T) {
	if !utils.IsEmptyString("") || utils.IsEmptyString("a") {
		t.Fatalf("IsEmptyString failed")
	}
	if !utils.IsEmptySlice([]int{}) || utils.IsEmptySlice([]int{1}) {
		t.Fatalf("IsEmptySlice failed")
	}
	if !utils.IsEmptyMap(map[string]int{}) || utils.IsEmptyMap(map[string]int{"a": 1}) {
		t.Fatalf("IsEmptyMap failed")
	}
	if !utils.IsEmpty[int](0) || utils.IsEmpty[int](1) {
		t.Fatalf("IsEmpty generic failed")
	}
	if !utils.IsEmptyLegacy(nil) || utils.IsEmptyLegacy(1) {
		t.Fatalf("IsEmptyLegacy failed")
	}
	if !utils.IsEmptyLegacy([]int{}) || utils.IsEmptyLegacy([]int{1}) {
		t.Fatalf("IsEmptyLegacy slice branch failed")
	}
	if !utils.IsEmptyLegacy((*int)(nil)) {
		t.Fatalf("IsEmptyLegacy should treat nil pointer as empty")
	}
	if !utils.IsEmptyLegacy((chan int)(nil)) {
		t.Fatalf("IsEmptyLegacy should treat nil channel as empty")
	}
	if utils.IsEmptyLegacy(true) {
		t.Fatalf("IsEmptyLegacy bool true should be non-empty")
	}
	if !utils.IsEmptyLegacy(false) {
		t.Fatalf("IsEmptyLegacy bool false should be empty")
	}
	if utils.IsEmptyLegacy(struct{ X int }{X: 1}) {
		t.Fatalf("IsEmptyLegacy default branch should be non-empty")
	}
}

func TestContainsHelpers(t *testing.T) {
	if !utils.ContainsString([]string{"a", "b"}, "b") || utils.ContainsString([]string{"a"}, "b") {
		t.Fatalf("ContainsString failed")
	}
	if !utils.ContainsInt([]int{1, 2}, 2) || utils.ContainsInt([]int{1}, 2) {
		t.Fatalf("ContainsInt failed")
	}
	if !utils.ContainsInt64([]int64{1, 2}, 2) || utils.ContainsInt64([]int64{1}, 2) {
		t.Fatalf("ContainsInt64 failed")
	}
	if !utils.Contains([]byte{1, 2}, 2) || utils.Contains([]byte{1}, 2) {
		t.Fatalf("Contains generic failed")
	}
	if !utils.ContainsLegacy([]int{1, 2}, 2) || utils.ContainsLegacy([]int{1}, 2) {
		t.Fatalf("ContainsLegacy failed")
	}
}

func TestRemoveDuplicates(t *testing.T) {
	if got := utils.RemoveDuplicatesString([]string{"a", "b", "a"}); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("RemoveDuplicatesString mismatch: %v", got)
	}
	if got := utils.RemoveDuplicatesInt([]int{1, 2, 1}); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("RemoveDuplicatesInt mismatch: %v", got)
	}
	if got := utils.RemoveDuplicates([]byte{1, 1, 2}); !reflect.DeepEqual(got, []byte{1, 2}) {
		t.Fatalf("RemoveDuplicates generic mismatch: %v", got)
	}
	if got, ok := utils.RemoveDuplicatesLegacy([]interface{}{1, 1, 2}).([]interface{}); !ok || len(got) != 2 {
		t.Fatalf("RemoveDuplicatesLegacy mismatch: %#v", got)
	}
}

func TestStringHelpers(t *testing.T) {
	if utils.TruncateString("hello", 2) != "he" {
		t.Fatalf("TruncateString short failed")
	}
	if utils.TruncateString("hello", 5) != "hello" {
		t.Fatalf("TruncateString equal failed")
	}
	if utils.TruncateString("hello", 4) != "h..." {
		t.Fatalf("TruncateString ellipsis failed")
	}

	if got := utils.FormatFileSize(512); got != "512 B" {
		t.Fatalf("FormatFileSize small mismatch: %s", got)
	}
	if got := utils.FormatFileSize(5*1024*1024 + 100); got != "5.0 MB" {
		t.Fatalf("FormatFileSize large mismatch: %s", got)
	}

	if got := utils.SliceToStringStrings([]string{"a", "b"}); got != "a, b" {
		t.Fatalf("SliceToStringStrings mismatch: %s", got)
	}
	if got := utils.SliceToStringInts([]int{1, 2}); got != "1, 2" {
		t.Fatalf("SliceToStringInts mismatch: %s", got)
	}
	if got := utils.SliceToString([]int{1, 2}); got != "1, 2" {
		t.Fatalf("SliceToString mismatch: %s", got)
	}
	if got := utils.StringToSlice("a, b"); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("StringToSlice mismatch: %v", got)
	}
}

func TestMapHelpers(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	keys := utils.MapKeys(m)
	if len(keys) != 2 {
		t.Fatalf("MapKeys length mismatch: %v", keys)
	}
	vals := utils.MapValues(m)
	if len(vals) != 2 {
		t.Fatalf("MapValues length mismatch: %v", vals)
	}
}

func TestReverseHelpers(t *testing.T) {
	s1 := []string{"a", "b"}
	utils.ReverseStrings(s1)
	if !reflect.DeepEqual(s1, []string{"b", "a"}) {
		t.Fatalf("ReverseStrings failed: %v", s1)
	}
	s2 := []int{1, 2}
	utils.ReverseInts(s2)
	if !reflect.DeepEqual(s2, []int{2, 1}) {
		t.Fatalf("ReverseInts failed: %v", s2)
	}
	s3 := []int64{1, 2}
	utils.ReverseInt64s(s3)
	if !reflect.DeepEqual(s3, []int64{2, 1}) {
		t.Fatalf("ReverseInt64s failed: %v", s3)
	}
	s4 := []byte{1, 2, 3}
	utils.Reverse(s4)
	if !reflect.DeepEqual(s4, []byte{3, 2, 1}) {
		t.Fatalf("Reverse generic failed: %v", s4)
	}

	// ReverseLegacy covers reflection-based swap and non-slice no-op
	s5 := []interface{}{1, "a", 3}
	utils.ReverseLegacy(s5)
	if !reflect.DeepEqual(s5, []interface{}{3, "a", 1}) {
		t.Fatalf("ReverseLegacy failed: %v", s5)
	}
	var notSlice = 123
	utils.ReverseLegacy(notSlice) // should not panic
}

func TestTimeAgo(t *testing.T) {
	now := time.Now()
	cases := []struct {
		dur  time.Duration
		want string
	}{
		{30 * time.Second, "just now"},
		{2 * time.Minute, "2 minutes ago"},
		{2 * time.Hour, "2 hours ago"},
		{48 * time.Hour, "2 days ago"},
		{14 * 24 * time.Hour, "2 weeks ago"},
		{60 * 24 * time.Hour, "2 months ago"},
		{2 * 365 * 24 * time.Hour, "2 years ago"},
	}

	for _, tc := range cases {
		got := utils.TimeAgo(now.Add(-tc.dur))
		if got != tc.want {
			t.Fatalf("TimeAgo(%s) = %s, want %s", tc.dur, got, tc.want)
		}
	}

	_ = utils.TimeAgo(time.Now())
}

func TestTimeAgoSingularBoundaries(t *testing.T) {
	now := time.Now()
	singulars := []struct {
		dur  time.Duration
		want string
	}{
		{1 * time.Minute, "1 minute ago"},
		{1 * time.Hour, "1 hour ago"},
		{24 * time.Hour, "1 day ago"},
		{7 * 24 * time.Hour, "1 week ago"},
		{30 * 24 * time.Hour, "1 month ago"},
		{365 * 24 * time.Hour, "1 year ago"},
	}

	for _, tc := range singulars {
		got := utils.TimeAgo(now.Add(-tc.dur))
		if got != tc.want {
			t.Fatalf("TimeAgo singular %s = %s, want %s", tc.dur, got, tc.want)
		}
	}
}
