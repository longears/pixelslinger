package opc

// Raver plaid
//   A rainbowy pattern with moving diagonal black stripes

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	"time"
)

func MakePatternDitherTest(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8

			// fill in bytes array
			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------

				pct := float64(ii) / float64(n_pixels)

                v := colorutils.Cos2(pct, t*0.2, 0.5, 0, 0.2)

				bytes[ii*3+0] = colorutils.FloatToByte(v)
				bytes[ii*3+1] = colorutils.FloatToByte(v)
				bytes[ii*3+2] = colorutils.FloatToByte(v)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
