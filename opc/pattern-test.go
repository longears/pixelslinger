package opc

// Test
//   General purpose test pattern.
//   Cycles between three modes:
//      0. set the entire strip to the same random color each frame
//      1. repeat these 4 colors down the strand: red green blue white
//      2. ruler mode for counting LEDs
//   Ruler mode:
//      The first LED of each chunk of 32 is red
//      The last of each chunk is green
//      Every 8th LED is dark blue

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	"math/rand"
	"time"
)

func MakePatternTest(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		rng := rand.New(rand.NewSource(99))
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8

			// fill in bytes array
			var r, g, b float64
			// mode 0: random solid color across whole strip
			if int(t/3)%3 == 0 {
				r = colorutils.Remap(rng.Float64(), 0, 1, 0.1, 1)
				g = colorutils.Remap(rng.Float64(), 0, 1, 0.1, 1)
				b = colorutils.Remap(rng.Float64(), 0, 1, 0.1, 1)
			}
			for ii := 0; ii < n_pixels; ii++ {

				// mode 1: r g b white r g b white ....
				if int(t/3)%3 == 1 {
					if ii%4 == 0 {
						r = 1.0
						g = 0.0
						b = 0.0
					} else if ii%4 == 1 {
						r = 0.0
						g = 1.0
						b = 0.0
					} else if ii%4 == 2 {
						r = 0.0
						g = 0.0
						b = 1.0
					} else if ii%4 == 3 {
						r = 1.0
						g = 1.0
						b = 1.0
					}
				}

				// mode 2: count the leds
				// the first of each chunk of 32 leds is red
				// the last of each chunk is green
				// every 8th led is dark blue
				if int(t/3)%3 == 2 {
					r = 0
					g = 0
					b = 0
					if ii%32 == 0 {
						r = 1
					}
					if ii%32 == 31 {
						g = 1
					}
					if ii%8 == 0 {
						b = 0.3
					}
				}

				twinkle := colorutils.Remap(rng.Float64(), 0, 1, 0.75, 1)
				bytes[ii*3+0] = colorutils.FloatToByte(r * twinkle)
				bytes[ii*3+1] = colorutils.FloatToByte(g * twinkle)
				bytes[ii*3+2] = colorutils.FloatToByte(b * twinkle)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
