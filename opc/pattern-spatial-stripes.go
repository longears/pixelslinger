package opc

// Spatial Stripes
//   Creates spatial sine wave stripes: x in the red channel, y--green, z--blue
//   Also makes a white dot which moves down the strip non-spatially in the order
//   that the LEDs are indexed.

import (
	"bitbucket.org/davidwallace/go-metal/colorutils"
	"bitbucket.org/davidwallace/go-metal/midi"
	"math"
	"time"
)

func MakePatternSpatialStripes(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiMessageChan chan *midi.MidiMessage) {
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8
			// fill in bytes slice
			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------

				// make moving stripes for x, y, and z
				x := locations[ii*3+0]
				y := locations[ii*3+1]
				z := locations[ii*3+2]
				r := colorutils.Cos(x, t/4, 1, 0, 0.7) // offset, period, minn, max
				g := colorutils.Cos(y, t/4, 1, 0, 0.7)
				b := colorutils.Cos(z, t/4, 1, 0, 0.7)
				r, g, b = colorutils.RGBContrast(r, g, b, 0.5, 2)

				// make a moving white dot showing the order of the pixels in the layout file
				spark_ii := colorutils.PosMod2(t*80, float64(n_pixels))
				spark_rad := float64(8)
				spark_val := math.Max(0, (spark_rad-colorutils.ModDist2(float64(ii), float64(spark_ii), float64(n_pixels)))/spark_rad)
				spark_val = math.Min(1, spark_val*2)
				r += spark_val
				g += spark_val
				b += spark_val

				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
