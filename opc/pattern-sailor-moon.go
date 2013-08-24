package opc

// Sailor Moon
//   Waves of magenta and cyan sparkles.

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	"math"
	"math/rand"
	"time"
)

func MakePatternSailorMoon(locations []float64) ByteThread {

	var (
		TWINKLE_SPEED   = 0.27
		TWINKLE_DENSITY = 0.3
		WAVE_SPEED      = 0.75
		WAVE_PERIOD     = 4.0
		WAVE_EXPONENT   = 20.0
	)

	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		rng := rand.New(rand.NewSource(99))

		// make persistant random values
		randomValues := make([]float64, len(locations)/3)
		for ii := range randomValues {
			randomValues[ii] = rng.Float64()
		}

		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8

			// fill in bytes array
			var r, g, b float64
			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------

				random := randomValues[ii]
				switch {
				case random < 0.5:
					// magenta
					r = 1
					g = 0.3
					b = 0.8
				case random < 0.85:
					// cyan
					r = 0.4
					g = 0.7
					b = 1
				case random < 0.90:
					// white
					r = 1
					g = 1
					b = 1
				default:
					// super bright magenta
					r = 2
					g = 0.6
					b = 1.6
				}

				// twinkle occasional LEDs
				twinkle := colorutils.PosMod(random*7+t*TWINKLE_SPEED, 1)
				twinkle = math.Abs(twinkle*2 - 1)
				twinkle = colorutils.Remap(twinkle, 0, 1, -1/TWINKLE_DENSITY, 1.1)
				twinkle = colorutils.Clamp(twinkle, -0.5, 1.1)
				twinkle = math.Pow(twinkle, 5)
				// offset, period, min, max
				twinkle *= math.Pow(colorutils.Cos2(t*WAVE_SPEED-float64(ii)/float64(n_pixels), 0, WAVE_PERIOD, 0.1, 1.0), WAVE_EXPONENT)

				bytes[ii*3+0] = colorutils.FloatToByte(r * twinkle)
				bytes[ii*3+1] = colorutils.FloatToByte(g * twinkle)
				bytes[ii*3+2] = colorutils.FloatToByte(b * twinkle)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
