package opc

// Raver plaid
//   A rainbowy pattern with moving diagonal black stripes

import (
	"bitbucket.org/davidwallace/go-metal/colorutils"
	"bitbucket.org/davidwallace/go-metal/midi"
	"time"
)

func MakePatternTestRGB(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiMessageChan chan *midi.MidiMessage) {
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
