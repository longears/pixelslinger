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

func MakePatternChevron(locations []float64) ByteThread {

	var (

        MORPH = 0.0 // 0 to 1.  0 is large blend, 1 is tiny blend

		SIDE_SCALE = 1.0 // Horizontal scale (x and y).  Smaller numbers compress things horizontally.

        DISPERSAL = 0.2

		WHITE_WAVE_PERIOD = 0.4
		WHITE_WAVE_SPEED  = 0.58
		WHITE_WAVE_THRESH = 0.9

		RED_WAVE_PERIOD = 0.4
		RED_WAVE_SPEED  = 0.2 // positive is down
		RED_WAVE_THRESH = 0.9

		BLEND_PERIOD = 0.3
		BLEND_SPEED  = -0.33 // positive is down
		BLEND_THRESH = 0.5 * (1-MORPH) + 0.99 * MORPH  // 1 is red, 0 is white
        BLEND_THRESH_AMT = 2.0 * (1-MORPH) + 20.0 * MORPH // contrast amount
	)

	// get bounding box
	n_pixels := len(locations) / 3
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

				// scale the height (z) of the layout to fit in the range 0-1
				// and scale x and y accordingly
				z_scale := max_coord_z - min_coord_z
				if z_scale == 0 { // avoid divide by zero
					z_scale = 0.05
				}
				xp := x / z_scale / SIDE_SCALE
				yp := y / z_scale / SIDE_SCALE
				zp := (z - min_coord_z) / z_scale

				_ = xp
				_ = yp
				_ = zp

				// bend space so that things seem to accelerate upwards
				zp1 := math.Pow(zp+0.02, 2 - DISPERSAL)
				zp2 := math.Pow(zp+0.02, 2)
				zp3 := math.Pow(zp+0.02, 2 + DISPERSAL)

				if xp < 0 {
					xp = -xp
				}

				// cos: offset, period, min, max

                // white wave
				rA := 0.8 * colorutils.Clamp(colorutils.Contrast(colorutils.Cos(xp-zp1, t*WHITE_WAVE_SPEED, WHITE_WAVE_PERIOD, 0, 1), WHITE_WAVE_THRESH, 2), 0, 1)
				gA := 1.0 * colorutils.Clamp(colorutils.Contrast(colorutils.Cos(xp-zp2, t*WHITE_WAVE_SPEED, WHITE_WAVE_PERIOD, 0, 1), WHITE_WAVE_THRESH, 2), 0, 1)
				bA := 1.0 * colorutils.Clamp(colorutils.Contrast(colorutils.Cos(xp-zp3, t*WHITE_WAVE_SPEED, WHITE_WAVE_PERIOD, 0, 1), WHITE_WAVE_THRESH, 2), 0, 1)

                // red wave
				rB := 1.0 * colorutils.Clamp(colorutils.Contrast(colorutils.Cos(xp-zp3, t*RED_WAVE_SPEED, RED_WAVE_PERIOD, 0, 1), RED_WAVE_THRESH, 2), 0, 1)
				gB := 0.5 * colorutils.Clamp(colorutils.Contrast(colorutils.Cos(xp-zp2, t*RED_WAVE_SPEED, RED_WAVE_PERIOD, 0, 1), RED_WAVE_THRESH, 2), 0, 1)
				bB := 0.5 * colorutils.Clamp(colorutils.Contrast(colorutils.Cos(xp-zp1, t*RED_WAVE_SPEED, RED_WAVE_PERIOD, 0, 1), RED_WAVE_THRESH, 2), 0, 1)

				// // accent color
				// rB = 0.3 //+ colorutils.Cos(t, 0, 7.30, -0.1, 0.3)
				// gB = 0.4 //+ colorutils.Cos(t, 0, 7.37, -0.1, 0.3)
				// bB = 0.5 //+ colorutils.Cos(t, 0, 7.43, -0.1, 0.3)

                blendOffset := t * BLEND_SPEED
                //blendOffset := colorutils.Cos(t, 0, 6, -0.8, 0.8)
				blend := colorutils.Clamp(colorutils.Contrast(colorutils.Cos(xp/3-zp, blendOffset, BLEND_PERIOD, 0, 1), BLEND_THRESH, BLEND_THRESH_AMT), 0, 1)

				bytes[ii*3+0] = colorutils.FloatToByte(rA*blend + rB*(1-blend))
				bytes[ii*3+1] = colorutils.FloatToByte(gA*blend + gB*(1-blend))
				bytes[ii*3+2] = colorutils.FloatToByte(bA*blend + bB*(1-blend))

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
