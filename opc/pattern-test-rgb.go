package opc

// Test RGB
//   Pattern for testing the RGB order of LED strips.
//   Emits these colors to the first 6 LEDs:
//      red, green, blue, black, white, black
//   For the rest of the pixels it makes a slowly moving red and black sine wave.

import (
	"bitbucket.org/davidwallace/pixelslinger/colorutils"
	"bitbucket.org/davidwallace/pixelslinger/midi"
	"time"
)

func MakePatternTestRGB(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8
			_ = t

			// fill in bytes array
			var r, g, b float64
			for ii := 0; ii < n_pixels; ii++ {

				switch ii {
				case 0:
					r = 1
					g = 0
					b = 0
				case 1:
					r = 0
					g = 1
					b = 0
				case 2:
					r = 0
					g = 0
					b = 1
				case 3:
					r = 0
					g = 0
					b = 0
				case 4:
					r = 1
					g = 1
					b = 1
				case 5:
					r = 0
					g = 0
					b = 0
				default:
					// x, offset, period, minn, maxx
					r = colorutils.Cos(float64(ii), t/4, 32, 0, 1)
					g = 0
					b = 0
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
