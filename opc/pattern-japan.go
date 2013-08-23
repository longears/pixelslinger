package opc

// Spatial Stripes
//   Creates spatial sine wave stripes: x in the red channel, y--green, z--blue
//   Also makes a white dot which moves down the strip non-spatially in the order
//   that the LEDs are indexed.

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	"math"
	"time"
)

func MakePatternJapan(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8

			var (
				NUM_BEAMS  = 5.0
				BEAM_SPEED = 0.33

				NUM_WAVES  = 0.1
				WAVE_SPEED = 0.3
			)

			// fill in bytes slice
			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------

				// make moving stripes for x, y, and z
				x := locations[ii*3+0]
				y := locations[ii*3+1]
				z := locations[ii*3+2]

				d := math.Sqrt(x*x + z*z)
				theta := math.Atan2(x, -z-0*y) + math.Pi // 0 is up, positive is clockwise, 0 to 2*pi
				theta2 := theta / (2 * math.Pi)

				beam := colorutils.Cos(theta2, t*BEAM_SPEED, 1.0/NUM_BEAMS, 0, 1)
				beam1 := colorutils.Contrast(beam, 0.8, 4) * 0.8
				beam2 := colorutils.Contrast(beam, 1.0, 6) * 0.6
				wave := colorutils.Contrast(colorutils.Cos(d, t*WAVE_SPEED, 1.0/NUM_WAVES, 0, 1), 1.0, 9)

				beam1 = colorutils.Clamp(beam1, 0, 1)
				beam2 = colorutils.Clamp(beam2, 0, 1)
				wave = colorutils.Clamp(wave, 0, 1)

				r := wave + beam1
				g := wave
				b := wave + beam2

				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
