package utils_test

import (
	"bytes"
	"crypto/rand"
	"reflect"
	"testing"
	"time"

	"go-postgres-rest/pkg/utils"
)

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
	// Edge cases
	if utils.ContainsString([]string{}, "a") {
		t.Fatalf("ContainsString empty slice should return false")
	}
	if utils.ContainsInt([]int{}, 1) {
		t.Fatalf("ContainsInt empty slice should return false")
	}
	if !utils.ContainsLegacy([]int{1, 2}, 2) || utils.ContainsLegacy([]int{1}, 2) {
		t.Fatalf("ContainsLegacy failed")
	}
	if utils.ContainsLegacy([]int{}, 1) {
		t.Fatalf("ContainsLegacy empty slice should return false")
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
		{-30 * time.Second, "just now"},
		{-2 * time.Minute, "2 minutes ago"},
		{-90 * time.Minute, "1 hour ago"},
		{-25 * time.Hour, "1 day ago"},
		{-8 * 24 * time.Hour, "1 week ago"},
		{-40 * 24 * time.Hour, "1 month ago"},
		{-400 * 24 * time.Hour, "1 year ago"},
	}
	for _, c := range cases {
		got := utils.TimeAgo(now.Add(c.dur))
		if got != c.want {
			t.Fatalf("TimeAgo for %v got %q want %q", c.dur, got, c.want)
		}
	}
}

func TestSliceToStringValidation(t *testing.T) {
	if got := utils.SliceToString(123); got != "" {
		t.Fatalf("SliceToString should return empty for non-slice, got %q", got)
	}
}

func TestPathHelpersMapKeysValuesTypes(t *testing.T) {
	type custom struct{ A int }
	m := map[custom]string{{A: 1}: "x"}
	keys := utils.MapKeys(m)
	if len(keys) != 1 || !reflect.DeepEqual(keys[0], custom{A: 1}) {
		t.Fatalf("MapKeys with custom key failed: %#v", keys)
	}
	vals := utils.MapValues(m)
	if len(vals) != 1 || vals[0] != "x" {
		t.Fatalf("MapValues with custom key failed: %#v", vals)
	}
}

func TestStringToSliceEmpty(t *testing.T) {
	if got := utils.StringToSlice(""); len(got) != 0 {
		t.Fatalf("expected empty slice for empty string")
	}
}

func TestLegacyHelpersEdgeCases(t *testing.T) {
	// IsEmptyLegacy with various kinds
	if !utils.IsEmptyLegacy(false) || utils.IsEmptyLegacy(true) {
		t.Fatalf("IsEmptyLegacy bool handling failed")
	}
	if !utils.IsEmptyLegacy([]int{}) || utils.IsEmptyLegacy([]int{1}) {
		t.Fatalf("IsEmptyLegacy slice handling failed")
	}
	if !utils.IsEmptyLegacy("") || utils.IsEmptyLegacy("a") {
		t.Fatalf("IsEmptyLegacy string handling failed")
	}
	if !utils.IsEmptyLegacy(map[string]int{}) || utils.IsEmptyLegacy(map[string]int{"a": 1}) {
		t.Fatalf("IsEmptyLegacy map handling failed")
	}
	if !utils.IsEmptyLegacy([0]int{}) || utils.IsEmptyLegacy([1]int{1}) {
		t.Fatalf("IsEmptyLegacy array handling failed")
	}
	ch := make(chan int, 1)
	if !utils.IsEmptyLegacy(ch) {
		t.Fatalf("IsEmptyLegacy empty chan should be empty")
	}
	ch <- 1
	if utils.IsEmptyLegacy(ch) {
		t.Fatalf("IsEmptyLegacy non-empty chan should be non-empty")
	}
	<-ch
	if !utils.IsEmptyLegacy(int(0)) || utils.IsEmptyLegacy(int(1)) {
		t.Fatalf("IsEmptyLegacy int handling failed")
	}
	if !utils.IsEmptyLegacy(uint(0)) || utils.IsEmptyLegacy(uint(1)) {
		t.Fatalf("IsEmptyLegacy uint handling failed")
	}
	if !utils.IsEmptyLegacy(float32(0)) || utils.IsEmptyLegacy(float32(1)) {
		t.Fatalf("IsEmptyLegacy float32 handling failed")
	}
	if !utils.IsEmptyLegacy(float64(0)) || utils.IsEmptyLegacy(float64(1)) {
		t.Fatalf("IsEmptyLegacy float64 handling failed")
	}
	var ptr *int
	if !utils.IsEmptyLegacy(ptr) {
		t.Fatalf("IsEmptyLegacy nil ptr should be empty")
	}
	val := 5
	ptr = &val
	if utils.IsEmptyLegacy(ptr) {
		t.Fatalf("IsEmptyLegacy non-nil ptr should be non-empty")
	}
	var iface interface{}
	if !utils.IsEmptyLegacy(iface) {
		t.Fatalf("IsEmptyLegacy nil interface should be empty")
	}
	iface = 1
	if utils.IsEmptyLegacy(iface) {
		t.Fatalf("IsEmptyLegacy non-nil non-zero interface should be non-empty")
	}
	// Unsupported type should return false
	if utils.IsEmptyLegacy(struct{}{}) {
		t.Fatalf("IsEmptyLegacy unsupported type should return false")
	}

	// ContainsLegacy non-slice returns false
	if utils.ContainsLegacy(123, 1) {
		t.Fatalf("ContainsLegacy should be false for non-slice input")
	}

	// RemoveDuplicatesLegacy with non-slice should echo input
	if got := utils.RemoveDuplicatesLegacy(123); got != 123 {
		t.Fatalf("RemoveDuplicatesLegacy non-slice should return same value")
	}

	// MapKeys/MapValues non-map inputs return nil
	if utils.MapKeys(123) != nil || utils.MapValues("notmap") != nil {
		t.Fatalf("MapKeys/MapValues should return nil for non-map inputs")
	}
}

// Benchmarks for performance-critical functions
func BenchmarkContainsString(b *testing.B) {
	slice := []string{"a", "b", "c", "d", "e"}
	for i := 0; i < b.N; i++ {
		utils.ContainsString(slice, "c")
	}
}

func BenchmarkContainsInt(b *testing.B) {
	slice := []int{1, 2, 3, 4, 5}
	for i := 0; i < b.N; i++ {
		utils.ContainsInt(slice, 3)
	}
}

func BenchmarkRemoveDuplicatesString(b *testing.B) {
	slice := []string{"a", "b", "a", "c", "b"}
	for i := 0; i < b.N; i++ {
		utils.RemoveDuplicatesString(slice)
	}
}

func BenchmarkRemoveDuplicatesInt(b *testing.B) {
	slice := []int{1, 2, 1, 3, 2}
	for i := 0; i < b.N; i++ {
		utils.RemoveDuplicatesInt(slice)
	}
}

func BenchmarkReverseStrings(b *testing.B) {
	slice := []string{"a", "b", "c", "d", "e"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.ReverseStrings(slice)
		utils.ReverseStrings(slice) // reverse back
	}
}

func BenchmarkSliceToStringStrings(b *testing.B) {
	slice := []string{"a", "b", "c", "d", "e"}
	for i := 0; i < b.N; i++ {
		utils.SliceToStringStrings(slice)
	}
}

func BenchmarkSliceToStringInts(b *testing.B) {
	slice := []int{1, 2, 3, 4, 5}
	for i := 0; i < b.N; i++ {
		utils.SliceToStringInts(slice)
	}
}
