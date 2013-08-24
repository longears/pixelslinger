package opc

// Raver plaid
//   A rainbowy pattern with moving diagonal black stripes

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/config"
	"github.com/longears/pixelslinger/midi"
	"math"
	"time"
)

func MakePatternRaverPlaid(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {

		// Variables we'll want to tweak later to adjust the pattern artistically
		var (
			// how many sine wave cycles are squeezed into our n_pixels
			// 24 happens to create nice diagonal stripes on the wall layout
			freq_r float64 = 24
			freq_g float64 = 24
			freq_b float64 = 24

			// how many seconds the color sine waves take to shift through a complete cycle
			speed_r float64 = 7
			speed_g float64 = -13
			speed_b float64 = 19
		)

		// This code is running in its own thread.  It is recieving byte slices over
		// the bytesIn channel, filling them with pixel colors, and then sending them
		// back over the bytesOut channel.
		// A "slice" is Go's version of a list or array, loosely speaking.
		// The byte slice is in the following format:
		//    [r, g, b,  r, g, b,  ...]
		// where each value is a byte in the range 0-255.

		// This pattern doesn't care about the pixels' locations.  If it did, it would
		// be using the locations slice which looks like this:
		//    [x, y, z,  x, y, z,  ...]
		// The "spatial-stripes" pattern is a good example of that.

		// Wait for the next incoming byte slice
		last_t := 0.0
		t := 0.0
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3

			// Get the current time in Unix seconds.
			// This requires some time and speed knob bookkeeping
			this_t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8
			speedKnob := float64(midiState.ControllerValues[config.SPEED_KNOB]) / 127.0
			if speedKnob < 0.5 {
				speedKnob = colorutils.RemapAndClamp(speedKnob, 0, 0.4, 0, 1)
			} else {
				speedKnob = colorutils.RemapAndClamp(speedKnob, 0.6, 1, 1, 4)
			}
			if midiState.KeyVolumes[config.SLOWMO_PAD] > 0 {
				speedKnob *= 0.25
			}
			if last_t != 0 {
				t += (this_t - last_t) * speedKnob
			}
			last_t = this_t

			// For each pixel...
			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------

				// How far along the strip are we?
				pct := float64(ii) / float64(n_pixels)

				// Replicate a quirk in the original python version of this pattern
				pct /= 2

				// Make diagonal black stripes using a slowly shifting sine wave
				// For more details on the "colorutils" package:
				//    http://godoc.org/github.com/longears/pixelslinger/colorutils
				pct_jittered := colorutils.PosMod2((pct * 77), 37)
				blackstripes := colorutils.Cos(pct_jittered, t*0.05, 1, -1.5, 1.5) // offset, period, minn, maxx
				blackstripes_offset := colorutils.Cos(t, 0.9, 60, -0.5, 3)         // slowly change the width of the stripes over a minute
				blackstripes = colorutils.Clamp(blackstripes+blackstripes_offset, 0, 1)

				// 3 sine waves for r, g, b which are out of sync with each other
				r := blackstripes * colorutils.Remap(math.Cos((t/speed_r+pct*freq_r)*math.Pi*2), -1, 1, 0, 1)
				g := blackstripes * colorutils.Remap(math.Cos((t/speed_g+pct*freq_g)*math.Pi*2), -1, 1, 0, 1)
				b := blackstripes * colorutils.Remap(math.Cos((t/speed_b+pct*freq_b)*math.Pi*2), -1, 1, 0, 1)

				// Write into the byte slice
				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}

			// Send our completed byte slice over the output channel
			bytesOut <- bytes
		}
	}
}
