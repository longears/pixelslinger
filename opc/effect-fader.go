package opc

// Fader effect
//   Listen to a midi knob and fade the entire pattern to black.
//   Fade the even pixels to black first, then the odd pixels.

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	"time"
)

func MakeEffectFader(locations []float64) ByteThread {

	var (
		flashDuration = 3.0 / 40.0 // seconds
		eyelidBlend   = 0.25       // size of eyelid gradient relative to bounding box

	)

	// get bounding box
	n_pixels := len(locations) / 3
	var max_coord_x, max_coord_y, max_coord_z float64
	var min_coord_x, min_coord_y, min_coord_z float64
	for ii := 0; ii < n_pixels; ii++ {
		x := locations[ii*3+0]
		y := locations[ii*3+1]
		z := locations[ii*3+2]
        if ii == 0 || x > max_coord_x { max_coord_x = x }
        if ii == 0 || y > max_coord_y { max_coord_y = y }
        if ii == 0 || z > max_coord_z { max_coord_z = z }
        if ii == 0 || x < min_coord_x { min_coord_x = x }
        if ii == 0 || y < min_coord_y { min_coord_y = y }
        if ii == 0 || z < min_coord_z { min_coord_z = z }
	}

	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {

		lastFlashTime := 0.0
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8

			// flash white when pad is down
			pad1 := midiState.KeyVolumes[midi.LPD8_PAD1]
			if pad1 > 0 {
				lastFlashTime = t
			}
			flashAmt := colorutils.Clamp(colorutils.Remap(t-lastFlashTime, 0, flashDuration, 1, 0), 0, 1)

			// gain knob
			knob1 := float64(midiState.ControllerValues[midi.LPD8_KNOB1]) / 127.0
			gain0 := colorutils.Clamp(colorutils.Remap(knob1, 0.75, 0.95, 0, 1), 0, 1)
			gain1 := colorutils.Clamp(colorutils.Remap(knob1, 0.40, 0.50, 0, 1), 0, 1)
			gain2 := colorutils.Clamp(colorutils.Remap(knob1, 0.05, 0.25, 0, 1), 0, 1)

			// eyelid knob
			knob2 := float64(midiState.ControllerValues[midi.LPD8_KNOB2]) / 127.0
			knob2 = colorutils.Clamp(colorutils.Remap(knob2, 0.05, 0.95, 0, 1), 0, 1)

			// fill in bytes array
			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------
				//pct := float64(ii) / float64(n_pixels)

				r := float64(bytes[ii*3+0]) / 255
				g := float64(bytes[ii*3+1]) / 255
				b := float64(bytes[ii*3+2]) / 255

				//x := locations[ii*3+0]
				//y := locations[ii*3+1]
				z := locations[ii*3+2]

				// zp ranges from 0 to 1 in the bounding box
				zp := colorutils.Remap(z, min_coord_z, max_coord_z, 0, 1)

				// gain knob
				gain := 1.0
				if ii%4 == 0 || ii%4 == 2 {
					gain = gain0
				} else if ii%4 == 3 {
					gain = gain1
				} else {
					gain = gain2
				}
				r *= gain
				g *= gain
				b *= gain

				// eyelid
				eyelid := colorutils.Clamp(colorutils.Remap(zp,
					knob2-(1-zp)*eyelidBlend/2,
					knob2+zp*eyelidBlend/2, 1, 0), 0, 1)
				r *= eyelid
				g *= eyelid
				b *= eyelid

				// flash
				r = r*(1-flashAmt) + 1*flashAmt
				g = g*(1-flashAmt) + 1*flashAmt
				b = b*(1-flashAmt) + 1*flashAmt

				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
