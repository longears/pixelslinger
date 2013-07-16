package colorutils

import "math"

const TABLE_SIZE = 2048

var COS_LOOKUP = make([]float64, TABLE_SIZE)

func init() {
	// build cos lookup table
	for ii := 0; ii < TABLE_SIZE; ii++ {
		x := float64(ii) / float64(TABLE_SIZE) * (math.Pi * 2)
		COS_LOOKUP[ii] = math.Cos(x)
	}
}

func CosTable(x float64) float64 {
	pct := x / (math.Pi * 2)
	ii := int64(pct*TABLE_SIZE + 0.5)
	ii = ii % TABLE_SIZE
	if ii < 0 {
		ii += TABLE_SIZE
	}
	return COS_LOOKUP[ii]
}

// Faster version of math.Mod(a,b) based on math.Modf
// Slightly less accurate, especially if b is very large or small
func Mod2(a, b float64) float64 {
	_, f := math.Modf(a / b)
	return f * b
}

// Given a float in the range 0-1, return a byte from 0 to 255.
func FloatToByte(x float64) byte {
	if x >= 1 {
		return 255
	} else if x <= 0 {
		return 0
	} else {
		return byte(x * 256)
	}
}

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
	//return math.Max(minn, math.Min(maxx, x))

	// this is much faster than using math.Max
	if x <= minn {
		return minn
	} else if x >= maxx {
		return maxx
	} else {
		return x
	}
}

//A cosine curve scaled to fit in a 0-1 range and 0-1 domain by default.
//offset: how much to slide the curve across the domain (should be 0-1)
//period: the length of one wave
//minn, maxx: the output range
func Cos(x, offset, period, minn, maxx float64) float64 {
	var value = math.Cos((x/period-offset)*math.Pi*2)/2 + 0.5
	return value*(maxx-minn) + minn
}
func Cos2(x, offset, period, minn, maxx float64) float64 {
	var value = CosTable((x/period-offset)*math.Pi*2)/2 + 0.5
	return value*(maxx-minn) + minn
}

//Expand the color values by a factor of mult around the pivot value of center.
//color: an (r, g, b) tuple
//center: a float -- the fixed point
//mult: a float -- expand or contract the values around the center point
func Contrast(x, center, mult float64) float64 {
	return (x-center)*mult + center
}
func RGBContrast(r, g, b, center, mult float64) (r2 float64, g2 float64, b2 float64) {
	r2 = (r-center)*mult + center
	g2 = (g-center)*mult + center
	b2 = (b-center)*mult + center
	return
}

// If x is less than threshold, return 0.  Otherwise, return x.
func ClipBlack(x, threshold float64) float64 {
	if x < threshold {
		return 0
	} else {
		return x
	}
}

// TODO: RBGClipBlackByChannel
// TODO: RBGClipBlackByLuminance

//Return the distance between floats a and b, modulo n.
//The result is always non-negative.
//For example, thinking of a clock:
//mod_dist(11, 1, 12) == 2 because you can "wrap around".
func ModDist(a, b, n float64) float64 {
	return math.Min(math.Mod(a-b+n, n), math.Mod(b-a+n, n))
}
func ModDist2(a, b, n float64) float64 {
	return math.Min(Mod2(a-b+n, n), Mod2(b-a+n, n))
}

// Apply a gamma exponent to x.  If x is negative, use 0 intead.
func Gamma(x, gamma float64) float64 {
	if x <= 0 {
		return 0
	} else {
		return math.Pow(x, gamma)
	}
}
func RGBGamma(r, g, b, gamma float64) (float64, float64, float64) {
	if r <= 0 {
		r = 0
	} else {
		r = math.Pow(r, gamma)
	}
	if g <= 0 {
		g = 0
	} else {
		g = math.Pow(g, gamma)
	}
	if b <= 0 {
		b = 0
	} else {
		b = math.Pow(b, gamma)
	}
	return r, g, b
}
