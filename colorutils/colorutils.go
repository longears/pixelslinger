package colorutils

import "math"

// Remap the float x from the range oldmin-oldmax to the range newmin-newmax
// Does not clamp values that exceed min or max.
// For example, to make a sine wave that goes between 0 and 256:
//     remap(math.sin(time.time()), -1, 1, 0, 256)
func Remap(x, oldmin, oldmax, newmin, newmax float64) float64 {
	var zero_to_one float64
	if oldmax == oldmin {
		zero_to_one = 0.5
	} else {
		zero_to_one = (x - oldmin) / (oldmax - oldmin)
	}
	return zero_to_one*(newmax-newmin) + newmin
}

// Restrict the float x to the range minn-maxx."""
func Clamp(x, minn, maxx float64) float64 {
	return math.Max(minn, math.Min(maxx, x))
}
