package utils

import (
	"encoding/json"
	"math"
	"personal_schedule_service/global"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	_ "time/tzdata"
)

func RoundToTwoDecimal(val float64) float64 {
	return math.Round(val*100) / 100
}

func Difference[T comparable](a, b []T) []T {
	m := make(map[T]struct{}, len(b))
	for _, item := range b {
		m[item] = struct{}{}
	}

	var diff []T
	for _, item := range a {
		if _, found := m[item]; !found {
			diff = append(diff, item)
		}
	}
	return diff
}

func ToJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}

func ToBoolPointer(b bool) *bool {
	return &b
}

func ToStringPointer(s string) *string {
	return &s
}

func StringToInt32(s string) int32 {
	var result int32
	err := json.Unmarshal([]byte(s), &result)
	if err != nil {
		return 0
	}
	return result
}

func ToIint64Pointer(i int64) *int64 {
	return &i
}

func Ternary[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}
func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func SafeInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func SafeInt64(i *int64) int64 {
	if i == nil {
		return 0
	}

	return *i
}

func RemoveAccent(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return strings.ToLower(result)
}

func TruncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func VietNameLocalDateRangeUTC(localDate string) (time.Time, time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", localDate, global.HCMTimeLocation)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	startLocal := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, global.HCMTimeLocation)
	endLocal := startLocal.Add(24 * time.Hour)

	return startLocal.UTC(), endLocal.UTC(), nil
}

func ToCompactJSON(data any) (string, error) {
	if data == nil {
		return "null", nil
	}

	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ParseLocalTimePtrToUTC(value string, layout string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	t, err := time.ParseInLocation(layout, value, global.HCMTimeLocation)
	if err != nil {
		return nil, err
	}

	utc := t.UTC()
	return &utc, nil
}

func ParseLocalTimeToUTC(value string, layout string) (time.Time, error) {
	t, err := time.ParseInLocation(layout, value, global.HCMTimeLocation)
	if err != nil {
		return time.Time{}, err
	}

	return t.UTC(), nil
}
