package utils

import "math"

const (
	KB int = 1024
	MB     = KB * 1024
	GB     = MB * 1024
)

func Precision(f float64, prec int, round bool) float64 {
	pow10N := math.Pow10(prec)
	if round {
		return (math.Trunc(f+0.5/pow10N) * pow10N) / pow10N
	}
	return math.Trunc((f)*pow10N) / pow10N
}
