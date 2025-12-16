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

// ConvertToInt converts string to int with error handling
func ConvertToInt(value string) (int, error) {
	return strconv.Atoi(value)
}

// ConvertToFloat converts string to float64 with error handling
func ConvertToFloat(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

// ConvertToBool converts string to bool with error handling
func ConvertToBool(value string) (bool, error) {
	return strconv.ParseBool(value)
}

// IsEmpty checks if a value is empty
func IsEmpty(value interface{}) bool {
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

// Contains checks if a slice contains a specific element
func Contains(slice interface{}, element interface{}) bool {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return false
	}

	for i := 0; i < s.Len(); i++ {
		if reflect.DeepEqual(s.Index(i).Interface(), element) {
			return true
		}
	}

	return false
}

// RemoveDuplicates removes duplicate elements from a slice
func RemoveDuplicates(slice interface{}) interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return slice
	}

	seen := make(map[interface{}]bool)
	result := reflect.MakeSlice(s.Type(), 0, s.Len())

	for i := 0; i < s.Len(); i++ {
		val := s.Index(i).Interface()
		if !seen[val] {
			seen[val] = true
			result = reflect.Append(result, s.Index(i))
		}
	}

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

// SliceToString converts a slice of any type to a comma-separated string
func SliceToString(slice interface{}) string {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return ""
	}

	var parts []string
	for i := 0; i < s.Len(); i++ {
		parts = append(parts, ConvertToString(s.Index(i).Interface()))
	}

	return strings.Join(parts, ", ")
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

// MapKeys returns the keys of a map as a slice
func MapKeys(m interface{}) []interface{} {
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		return nil
	}

	keys := v.MapKeys()
	result := make([]interface{}, len(keys))
	for i, key := range keys {
		result[i] = key.Interface()
	}

	return result
}

// MapValues returns the values of a map as a slice
func MapValues(m interface{}) []interface{} {
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		return nil
	}

	keys := v.MapKeys()
	result := make([]interface{}, len(keys))
	for i, key := range keys {
		result[i] = v.MapIndex(key).Interface()
	}

	return result
}

// Reverse reverses a slice in place
func Reverse(slice interface{}) {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return
	}

	for i, j := 0, s.Len()-1; i < j; i, j = i+1, j-1 {
		vi, vj := s.Index(i), s.Index(j)
		temp := vi.Interface()
		vi.Set(vj)
		vj.Set(reflect.ValueOf(temp))
	}
}

// TimeAgo returns a human-readable time difference
func TimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case diff < 365*24*time.Hour:
		months := int(diff.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(diff.Hours() / (24 * 365))
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}
