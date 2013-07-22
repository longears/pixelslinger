package opc

// Raver plaid
//   A rainbowy pattern with moving diagonal black stripes

import (
	"bitbucket.org/davidwallace/go-metal/colorutils"
	"math"
	"time"
)

func MakePatternTestGamma(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte) {
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8
			_ = t

			// fill in bytes array
			var r, g, b float64
			for ii := 0; ii < n_pixels; ii++ {

				// make moving rgb sawtooth waves
				r = colorutils.PosMod(float64(ii)-t*2.84, 16) / 15.0
				g = colorutils.PosMod(float64(ii)-t*4.00, 19) / 18.0
				b = colorutils.PosMod(float64(ii)-t*5.37, 27) / 26.0
				// convert sawtooth waves into triangle waves
				r = math.Abs(r*2 - 1)
				g = math.Abs(g*2 - 1)
				b = math.Abs(b*2 - 1)
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
