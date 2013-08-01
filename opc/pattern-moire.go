package opc

// Moire

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	"time"
)

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

				//// make moving stripes for x, y, and z
				//x := locations[ii*3+0]
				//y := locations[ii*3+1]
				//z := locations[ii*3+2]

				//MIN_PERIOD := 0.3
				//MAX_PERIOD := 10
				//period := colorutils.Remap(x, min_coord_x, max_coord_x, MIN_PERIOD, MAX_PERIOD)
				rPeriod := 1 / float64(ii+10)
				gPeriod := 1 / float64((ii+160/3)%160+10)
				bPeriod := 1 / float64((ii+2*160/3)%160+10)

				// offset, period, minn, maxx
				r := colorutils.Cos(t*0.1, 0, rPeriod, 0, 1)
				g := colorutils.Cos(t*0.1, 0.11, gPeriod, 0, 1)
				b := colorutils.Cos(t*0.1, 0.37, bPeriod, 0, 1)

				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
