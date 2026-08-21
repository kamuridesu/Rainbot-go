package rucoy

import (
	"strconv"
	"strings"
)

func absRucoyNumber(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}

func formatRucoyNumber(value int64) string {
	raw := strconv.FormatInt(value, 10)
	parts := make([]string, 0, len(raw)/3+1)

	for len(raw) > 3 {
		parts = append([]string{raw[len(raw)-3:]}, parts...)
		raw = raw[:len(raw)-3]
	}

	parts = append([]string{raw}, parts...)
	return strings.Join(parts, ".")
}
