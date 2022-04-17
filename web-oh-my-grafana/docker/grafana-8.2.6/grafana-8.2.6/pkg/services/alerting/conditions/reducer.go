package conditions

import (
	"math"

	"sort"

	"github.com/grafana/grafana/pkg/components/null"
	"github.com/grafana/grafana/pkg/plugins"
)

// queryReducer reduces a timeseries to a nullable float
type queryReducer struct {

	// Type is how the timeseries should be reduced.
	// Ex avg, sum, max, min, count
	Type string
}

//nolint: gocyclo
//nolint: staticcheck // plugins.DataTimeSeries deprecated
func (s *queryReducer) Reduce(series plugins.DataTimeSeries) null.Float {
	if len(series.Points) == 0 {
		return null.FloatFromPtr(nil)
	}

	value := float64(0)
	allNull := true

	switch s.Type {
	case "avg":
		validPointsCount := 0
		for _, point := range series.Points {
			if isValid(point[0]) {
				value += point[0].Float64
				validPointsCount++
				allNull = false
			}
		}
		if validPointsCount > 0 {
			value /= float64(validPointsCount)
		}
	case "sum":
		for _, point := range series.Points {
			if isValid(point[0]) {
				value += point[0].Float64
				allNull = false
			}
		}
	case "min":
		value = math.MaxFloat64
		for _, point := range series.Points {
			if isValid(point[0]) {
				allNull = false
				if value > point[0].Float64 {
					value = point[0].Float64
				}
			}
		}
	case "max":
		value = -math.MaxFloat64
		for _, point := range series.Points {
			if isValid(point[0]) {
				allNull = false
				if value < point[0].Float64 {
					value = point[0].Float64
				}
			}
		}
	case "count":
		value = float64(len(series.Points))
		allNull = false
	case "last":
		points := series.Points
		for i := len(points) - 1; i >= 0; i-- {
			if isValid(points[i][0]) {
				value = points[i][0].Float64
				allNull = false
				break
			}
		}
	case "median":
		var values []float64
		for _, v := range series.Points {
			if isValid(v[0]) {
				allNull = false
				values = append(values, v[0].Float64)
			}
		}
		if len(values) >= 1 {
			sort.Float64s(values)
			length := len(values)
			if length%2 == 1 {
				value = values[(length-1)/2]
			} else {
				value = (values[(length/2)-1] + values[length/2]) / 2
			}
		}
	case "diff":
		allNull, value = calculateDiff(series, allNull, value, diff)
	case "diff_abs":
		allNull, value = calculateDiff(series, allNull, value, diffAbs)
	case "percent_diff":
		allNull, value = calculateDiff(series, allNull, value, percentDiff)
	case "percent_diff_abs":
		allNull, value = calculateDiff(series, allNull, value, percentDiffAbs)
	case "count_non_null":
		for _, v := range series.Points {
			if isValid(v[0]) {
				value++
			}
		}

		if value > 0 {
			allNull = false
		}
	}

	if allNull {
		return null.FloatFromPtr(nil)
	}

	return null.FloatFrom(value)
}

func newSimpleReducer(t string) *queryReducer {
	return &queryReducer{Type: t}
}

//nolint: staticcheck // plugins.* deprecated
func calculateDiff(series plugins.DataTimeSeries, allNull bool, value float64, fn func(float64, float64) float64) (bool, float64) {
	var (
		points = series.Points
		first  float64
		i      int
	)
	// get the newest point
	for i = len(points) - 1; i >= 0; i-- {
		if isValid(points[i][0]) {
			allNull = false
			first = points[i][0].Float64
			break
		}
	}
	if i >= 1 {
		// get the oldest point
		points = points[0:i]
		for i := 0; i < len(points); i++ {
			if isValid(points[i][0]) {
				allNull = false
				value = fn(first, points[i][0].Float64)
				break
			}
		}
	}
	return allNull, value
}

func isValid(f null.Float) bool {
	return f.Valid && !math.IsNaN(f.Float64)
}

var diff = func(newest, oldest float64) float64 {
	return newest - oldest
}

var diffAbs = func(newest, oldest float64) float64 {
	return math.Abs(newest - oldest)
}

var percentDiff = func(newest, oldest float64) float64 {
	return (newest - oldest) / math.Abs(oldest) * 100
}

var percentDiffAbs = func(newest, oldest float64) float64 {
	return math.Abs((newest - oldest) / oldest * 100)
}
