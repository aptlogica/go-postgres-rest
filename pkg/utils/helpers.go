// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// GenerateID generates a random ID string
func GenerateID(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ConvertToString converts various types to string
func ConvertToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ConvertToInt converts string to int
func ConvertToInt(value string) (int, error) { return strconv.Atoi(value) }

// ConvertToFloat converts string to float64
func ConvertToFloat(value string) (float64, error) { return strconv.ParseFloat(value, 64) }

// ConvertToBool converts string to bool
func ConvertToBool(value string) (bool, error) { return strconv.ParseBool(value) }

// ============================================================================
// IsEmpty variants - NO REFLECTION for common types
// ============================================================================

// IsEmptyString checks if a string is empty - O(1)
func IsEmptyString(s string) bool { return len(s) == 0 }

// IsEmptySlice checks if a slice is empty - O(1)
func IsEmptySlice[T any](s []T) bool { return len(s) == 0 }

// IsEmptyMap checks if a map is empty - O(1)
func IsEmptyMap[K comparable, V any](m map[K]V) bool { return len(m) == 0 }

// IsEmpty checks if a comparable value equals its zero value
func IsEmpty[T comparable](v T) bool {
	var zero T
	return v == zero
}

// IsEmptyLegacy checks if a value is empty (fallback for non-comparable types)
// This is ONLY for types that cannot be handled by IsEmpty[T]
// DEPRECATED: Use typed variants (IsEmptyString, IsEmptySlice[T], etc.) when possible
func IsEmptyLegacy(value interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	return false
}

// ============================================================================
// Contains variants - NO REFLECTION for common types
// ============================================================================

// Contains checks if a slice contains an element - O(n)
func Contains[T comparable](slice []T, element T) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

// Type-specific wrappers for Contains (for backward compatibility)
func ContainsString(slice []string, element string) bool { return Contains(slice, element) }
func ContainsInt(slice []int, element int) bool          { return Contains(slice, element) }
func ContainsInt64(slice []int64, element int64) bool    { return Contains(slice, element) }

// validateSlice checks if the value is a slice and returns its reflect.Value
func validateSlice(slice interface{}) (reflect.Value, bool) {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return s, false
	}
	return s, true
}

// iterateSlice applies a function to each element of a reflected slice
func iterateSlice(s reflect.Value, fn func(i int, elem interface{}) bool) bool {
	for i := 0; i < s.Len(); i++ {
		if fn(i, s.Index(i).Interface()) {
			return true
		}
	}
	return false
}

// ContainsLegacy checks if a slice contains a specific element (fallback)
// This uses reflection for non-comparable types
// DEPRECATED: Use typed variants (ContainsString, ContainsInt, Contains[T]) when possible
func ContainsLegacy(slice interface{}, element interface{}) bool {
	s, ok := validateSlice(slice)
	if !ok {
		return false
	}
	return iterateSlice(s, func(_ int, elem interface{}) bool {
		return reflect.DeepEqual(elem, element)
	})
}

// ============================================================================
// RemoveDuplicates variants - NO REFLECTION for common types
// ============================================================================

// RemoveDuplicates removes duplicate elements from a slice
func RemoveDuplicates[T comparable](slice []T) []T {
	if len(slice) == 0 {
		return slice
	}
	seen := make(map[T]bool)
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// Type-specific wrappers for RemoveDuplicates (for backward compatibility)
func RemoveDuplicatesString(slice []string) []string { return RemoveDuplicates(slice) }
func RemoveDuplicatesInt(slice []int) []int          { return RemoveDuplicates(slice) }

// RemoveDuplicatesLegacy removes duplicate elements from a slice (fallback with reflection)
// DEPRECATED: Use typed variants (RemoveDuplicatesString, RemoveDuplicatesInt, RemoveDuplicates[T]) when possible
func RemoveDuplicatesLegacy(slice interface{}) interface{} {
	s, ok := validateSlice(slice)
	if !ok {
		return slice
	}

	seen := make(map[interface{}]bool)
	result := reflect.MakeSlice(s.Type(), 0, s.Len())

	iterateSlice(s, func(i int, val interface{}) bool {
		if !seen[val] {
			seen[val] = true
			result = reflect.Append(result, s.Index(i))
		}
		return false
	})

	return result.Interface()
}

// TruncateString truncates a string to a specified length
func TruncateString(str string, length int) string {
	if len(str) <= length {
		return str
	}

	if length <= 3 {
		return str[:length]
	}

	return str[:length-3] + "..."
}

// FormatFileSize formats a file size in bytes to human readable format
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ============================================================================
// SliceToString variants - OPTIMIZED with pre-allocation
// ============================================================================

// sliceToStringHelper converts slices to comma-separated strings
func sliceToStringHelper[T any](slice []T, converter func(T) string) string {
	if len(slice) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, v := range slice {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(converter(v))
	}
	return sb.String()
}

// Type-specific wrappers for SliceToString
func SliceToStringStrings(slice []string) string {
	return sliceToStringHelper(slice, func(s string) string { return s })
}
func SliceToStringInts(slice []int) string {
	return sliceToStringHelper(slice, func(i int) string { return fmt.Sprintf("%d", i) })
}

// SliceToString converts a slice of any type to a comma-separated string
// Uses reflection only as fallback - prefer typed variants when possible
func SliceToString(slice interface{}) string {
	s, ok := validateSlice(slice)
	if !ok || s.Len() == 0 {
		return ""
	}

	var sb strings.Builder
	iterateSlice(s, func(i int, elem interface{}) bool {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(ConvertToString(elem))
		return false
	})

	return sb.String()
}

// StringToSlice converts a comma-separated string to a slice of strings
func StringToSlice(str string) []string {
	if str == "" {
		return []string{}
	}

	parts := strings.Split(str, ",")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}

	return parts
}

// ============================================================================
// Map functions - keep reflection but improve
// ============================================================================

// validateMap checks if the value is a map and returns its reflect.Value and keys
func validateMap(m interface{}) (reflect.Value, []reflect.Value, bool) {
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		return v, nil, false
	}
	return v, v.MapKeys(), true
}

// MapKeys returns the keys of a map as a slice
func MapKeys(m interface{}) []interface{} {
	_, keys, ok := validateMap(m)
	if !ok {
		return nil
	}
	result := make([]interface{}, len(keys))
	for i, key := range keys {
		result[i] = key.Interface()
	}
	return result
}

// MapValues returns the values of a map as a slice
func MapValues(m interface{}) []interface{} {
	v, keys, ok := validateMap(m)
	if !ok {
		return nil
	}
	result := make([]interface{}, len(keys))
	for i, key := range keys {
		result[i] = v.MapIndex(key).Interface()
	}
	return result
}

// ============================================================================
// Reverse variants - NO REFLECTION for common types
// ============================================================================

// Reverse reverses a slice in place
func Reverse[T any](slice []T) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// Type-specific wrappers for Reverse (for backward compatibility)
func ReverseStrings(slice []string) { Reverse(slice) }
func ReverseInts(slice []int)       { Reverse(slice) }
func ReverseInt64s(slice []int64)   { Reverse(slice) }

// ReverseLegacy reverses a slice in place (fallback with reflection)
// DEPRECATED: Use typed variants (ReverseStrings, ReverseInts, Reverse[T]) when possible
func ReverseLegacy(slice interface{}) {
	s, ok := validateSlice(slice)
	if !ok {
		return
	}
	for i, j := 0, s.Len()-1; i < j; i, j = i+1, j-1 {
		vi, vj := s.Index(i), s.Index(j)
		temp := reflect.ValueOf(vi.Interface())
		vi.Set(vj)
		vj.Set(temp)
	}
}

// formatTimeUnit is a helper function to format time units with singular/plural handling
func formatTimeUnit(count int, unit string) string {
	if count == 1 {
		return fmt.Sprintf("1 %s ago", unit)
	}
	return fmt.Sprintf("%d %ss ago", count, unit)
}

// TimeAgo returns a human-readable time difference
func TimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		return formatTimeUnit(int(diff.Minutes()), "minute")
	case diff < 24*time.Hour:
		return formatTimeUnit(int(diff.Hours()), "hour")
	case diff < 7*24*time.Hour:
		return formatTimeUnit(int(diff.Hours()/24), "day")
	case diff < 30*24*time.Hour:
		return formatTimeUnit(int(diff.Hours()/(24*7)), "week")
	case diff < 365*24*time.Hour:
		return formatTimeUnit(int(diff.Hours()/(24*30)), "month")
	default:
		return formatTimeUnit(int(diff.Hours()/(24*365)), "year")
	}
}
