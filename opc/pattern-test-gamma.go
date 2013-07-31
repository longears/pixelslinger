package opc

// Test Gamma
//   A rainbowy pattern which makes it easy to see if the gamma is set correctly.
//   This pattern should look saturated, not pastel with cyan-yellow-mageta overtones.

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	"time"
)

func MakePatternTestGamma(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8

			// fill in bytes array
			var r, g, b float64
			for ii := 0; ii < n_pixels; ii++ {

				// make moving rgb sawtooth waves
				r = colorutils.PosMod2(float64(ii)-t*2.84, 16) / 15.0
				g = colorutils.PosMod2(float64(ii)-t*4.00, 19) / 18.0
				b = colorutils.PosMod2(float64(ii)-t*5.37, 27) / 26.0
				// convert sawtooth waves into triangle waves
				r = r*2 - 1
				g = g*2 - 1
				b = b*2 - 1
				if r < 0 {
					r *= -1
				}
				if g < 0 {
					g *= -1
				}
				if b < 0 {
					b *= -1
				}
				// monochrome region in first 32 LEDs
				if ii < 32 {
					g = r
					b = r
				}

				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
