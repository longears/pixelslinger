package opc

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	"time"
)

func MakePatternEye(locations []float64) ByteThread {

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
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8
            _ = t

            const EYELID_SOFTNESS = 0.2

            pupilTheta := colorutils.PosMod2(t/4.0, 1) * 360  // in degrees
            // offset, period, minn, maxx
            topEyelidClose := colorutils.Cos(t, 0, 5, 0, 1)  // 0 is totally open, 1 is totally closed

			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------

				// How far along the strip are we?
				pct := float64(ii) / float64(n_pixels)
                pixelTheta := pct * 360.0

                //x := locations[ii*3 + 0]
                //y := locations[ii*3 + 1]
                z := locations[ii*3 + 2]

                zp := colorutils.Remap(z, min_coord_z, max_coord_z, 0, 1)

                pupil := colorutils.ModDist(pupilTheta, pixelTheta, 360) / 180.0
                pupil = colorutils.RemapAndClamp(pupil, 0.9, 0.93, 0, 1)

                topEyelidHere := colorutils.RemapAndClamp(1-zp, topEyelidClose-EYELID_SOFTNESS*(1-topEyelidClose), topEyelidClose+EYELID_SOFTNESS*topEyelidClose, 0, 1)

                r := 0.3 * pupil + 0.5 * (1-pupil)
                g := 0.6 * pupil + 0.5 * (1-pupil)
                b := 1.0 * pupil + 0.5 * (1-pupil)

                r *= topEyelidHere
                g *= topEyelidHere
                b *= topEyelidHere

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
