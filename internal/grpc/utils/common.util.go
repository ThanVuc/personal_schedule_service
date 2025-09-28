package utils

import (
	"encoding/json"
	"math"
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
