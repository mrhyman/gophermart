package util

import "strconv"

func RoundToTwoDecimals(value float64) float64 {
	str := strconv.FormatFloat(value, 'f', 2, 64)
	result, _ := strconv.ParseFloat(str, 64)
	return result
}
