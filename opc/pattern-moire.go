package opc

// Moire

import (
	"bitbucket.org/davidwallace/go-metal/colorutils"
	"bitbucket.org/davidwallace/go-metal/midi"
	"time"
)

const MIN_PERIOD = 0.3
const MAX_PERIOD = 10

func MakePatternMoire(locations []float64) ByteThread {

	n_pixels := len(locations) / 3

	// get bounding box
	var max_coord_x, max_coord_y, max_coord_z float64
	var min_coord_x, min_coord_y, min_coord_z float64
	for ii := 0; ii < n_pixels; ii++ {
		x := locations[ii*3+0]
		y := locations[ii*3+1]
		z := locations[ii*3+2]
		if ii == 0 || x > max_coord_x {
			max_coord_x = x
		}
		if ii == 0 || y > max_coord_y {
			max_coord_y = y
		}
		if ii == 0 || z > max_coord_z {
			max_coord_z = z
		}
		if ii == 0 || x < min_coord_x {
			min_coord_x = x
		}
		if ii == 0 || y < min_coord_y {
			min_coord_y = y
		}
		if ii == 0 || z < min_coord_z {
			min_coord_z = z
		}
	}

	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
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

				xPeriod := colorutils.Remap(x, min_coord_x, max_coord_x, MIN_PERIOD, MAX_PERIOD)
				yPeriod := colorutils.Remap(y, min_coord_y, max_coord_y, MIN_PERIOD, MAX_PERIOD)
				zPeriod := colorutils.Remap(z, min_coord_z, max_coord_z, MIN_PERIOD, MAX_PERIOD)

				// offset, period, minn, maxx
				r := colorutils.Cos2(t, 0, xPeriod, 0, 1)
				g := colorutils.Cos2(t, 0, yPeriod, 0, 1)
				b := colorutils.Cos2(t, 0, zPeriod, 0, 1)
				_ = r
				_ = g

				bytes[ii*3+0] = colorutils.FloatToByte(b)
				bytes[ii*3+1] = colorutils.FloatToByte(b)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
