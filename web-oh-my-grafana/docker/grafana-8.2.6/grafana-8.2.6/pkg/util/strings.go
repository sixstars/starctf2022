package util

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

// StringsFallback2 returns the first of two not empty strings.
func StringsFallback2(val1 string, val2 string) string {
	return stringsFallback(val1, val2)
}

// StringsFallback3 returns the first of three not empty strings.
func StringsFallback3(val1 string, val2 string, val3 string) string {
	return stringsFallback(val1, val2, val3)
}

func stringsFallback(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// SplitString splits a string by commas or empty spaces.
func SplitString(str string) []string {
	if len(str) == 0 {
		return []string{}
	}

	return regexp.MustCompile("[, ]+").Split(str, -1)
}

// GetAgeString returns a string representing certain time from years to minutes.
func GetAgeString(t time.Time) string {
	if t.IsZero() {
		return "?"
	}

	sinceNow := time.Since(t)
	minutes := sinceNow.Minutes()
	years := int(math.Floor(minutes / 525600))
	months := int(math.Floor(minutes / 43800))
	days := int(math.Floor(minutes / 1440))
	hours := int(math.Floor(minutes / 60))
	var amount string
	if years > 0 {
		if years == 1 {
			amount = "year"
		} else {
			amount = "years"
		}
		return fmt.Sprintf("%d %s", years, amount)
	}
	if months > 0 {
		if months == 1 {
			amount = "month"
		} else {
			amount = "months"
		}
		return fmt.Sprintf("%d %s", months, amount)
	}
	if days > 0 {
		if days == 1 {
			amount = "day"
		} else {
			amount = "days"
		}
		return fmt.Sprintf("%d %s", days, amount)
	}
	if hours > 0 {
		if hours == 1 {
			amount = "hour"
		} else {
			amount = "hours"
		}
		return fmt.Sprintf("%d %s", hours, amount)
	}
	if int(minutes) > 0 {
		if int(minutes) == 1 {
			amount = "minute"
		} else {
			amount = "minutes"
		}
		return fmt.Sprintf("%d %s", int(minutes), amount)
	}

	return "< 1 minute"
}

// ToCamelCase changes kebab case, snake case or mixed strings to camel case. See unit test for examples.
func ToCamelCase(str string) string {
	var finalParts []string
	parts := strings.Split(str, "_")

	for _, part := range parts {
		finalParts = append(finalParts, strings.Split(part, "-")...)
	}

	for index, part := range finalParts[1:] {
		finalParts[index+1] = strings.Title(part)
	}

	return strings.Join(finalParts, "")
}
