package opc

// Eye
//   An eye-of-Sauron effect which only works on vertical circles (e.g. not in the X-Y plane).
//   It limits itself to the first 160 pixels; the rest will be black.

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	"math"
	"math/rand"
	"time"
)

func MakePatternEye(locations []float64) ByteThread {

    // get bounding box
    n_pixels := len(locations) / 3
    if n_pixels > 160 {
        n_pixels = 160
    }
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

		const (
			PUPIL_SIZE            = 0.1
			PUPIL_SOFTNESS        = 0.05
			EYELID_SOFTNESS       = 0.2  // fuzziness of edge of eyelid
			MAX_BIG_MOVE_THETA    = 200  // max degrees to move during a big move
			MIN_BIG_MOVE_THETA    = 30   // min degrees to move during a big move
			SMALL_MOVE_THETA      = 10   // max degrees to move during a small move
			FORBIDDEN_UPPER_THETA = 55   // pupil won't go higher than +/- this value
			BIG_MOVE_PROB         = 0.3  // chance that a move is a big move
			TOP_EYELID_MAX_OPEN   = 0.12 // 0 is open, 1 closed
			TOP_EYELID_MAX_CLOSE  = 0.55 // 0 is open, 1 is closed
		)

		lastPupilTheta := 90.0
		nextPupilTheta := 90.0
		lastPupilTime := 0.0
		nextPupilTime := 0.0
		holdingStill := 0 // is this a real movement or a holding-still movement?

		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			n_actual_pixels := len(bytes) / 3
			if n_pixels > 160 {
				n_pixels = 160
			}
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8

			// if the current move is over, figure out the next move
			var moveDuration float64
			if t > nextPupilTime {
				holdingStill = 1 - holdingStill

				// is this a big move or a small move?
				bigMove := rand.Float64() < BIG_MOVE_PROB

				lastPupilTheta = nextPupilTheta
				if holdingStill == 0 {
					lastPupilTheta = colorutils.PosMod(lastPupilTheta, 360)
					if bigMove {
						// big move
						randomSign := 1.0
						if rand.Float64() < 0.5 {
							randomSign = -1.0
						}
						nextPupilTheta = lastPupilTheta + colorutils.Remap(rand.Float64(), 0, 1, MIN_BIG_MOVE_THETA, MAX_BIG_MOVE_THETA)*randomSign
					} else {
						// small move
						nextPupilTheta = lastPupilTheta + (rand.Float64()*2-1)*SMALL_MOVE_THETA
					}
					nextPupilTheta = colorutils.Clamp(nextPupilTheta, 1, 359)
					//nextPupilTheta = colorutils.PosMod(nextPupilTheta, 360)
				}

				if holdingStill == 1 {
					moveDuration = rand.Float64()*0.3 + 0.1
				} else {
					moveDuration = colorutils.Abs(nextPupilTheta-lastPupilTheta)/180*0.2 + 0.05
				}
				lastPupilTime = nextPupilTime
				nextPupilTime = t + moveDuration

			}

			// compute pupilTheta, interpolating for the current move
			pupilTheta := colorutils.EaseRemapAndClamp(t, lastPupilTime, nextPupilTime, lastPupilTheta, nextPupilTheta)
			// push away from the forbidden theta zone at the top
			pupilTheta = colorutils.Remap(pupilTheta, 0, 360, FORBIDDEN_UPPER_THETA, 360-FORBIDDEN_UPPER_THETA)
			// rotate to the bottom of the eye
			pupilTheta += 180

			// eyelid open/closed-ness depends on pupil position
			topEyelidClose := colorutils.RemapAndClamp(math.Cos(pupilTheta*math.Pi/180), -1, 1, TOP_EYELID_MAX_OPEN, TOP_EYELID_MAX_CLOSE)

			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------

				// How far along the strip are we?
				pct := float64(ii) / float64(n_pixels)
				pixelTheta := pct * 360.0

				z := locations[ii*3+2]
				zp := colorutils.Remap(z, min_coord_z, max_coord_z, 0, 1)

				// flip everything upside down
				// might be needed depending on which layout you're using
				pixelTheta = colorutils.PosMod(pixelTheta+180, 360)
				//zp = 1 - zp

				pupilHere := colorutils.ModDist(pupilTheta, pixelTheta, 360) / 180.0
				pupilHereThreshold := colorutils.RemapAndClamp(pupilHere, 1-PUPIL_SIZE-PUPIL_SOFTNESS, 1-PUPIL_SIZE, 0, 1)
				pupilHereInPupil := colorutils.RemapAndClamp(pupilHere, 1-PUPIL_SIZE, 1-PUPIL_SIZE/2, 0, 1)

				topEyelidHere := colorutils.RemapAndClamp(1-zp, topEyelidClose-EYELID_SOFTNESS*(1-topEyelidClose), topEyelidClose+EYELID_SOFTNESS*topEyelidClose, 0, 1)

				zp2 := 1 - (1-zp)*(1-zp)
				//           top ...... bottom
				eyeWhiteR := 0.6*zp2 + 0.7*(1-zp2)
				eyeWhiteG := 0.0*zp2 + 0.0*(1-zp2)
				eyeWhiteB := 0.0*zp2 + 0.1*(1-zp2)

				//        inner      ......        outer
				pupilR := 2.0*pupilHereInPupil + 0.5*(1-pupilHereInPupil)
				pupilG := 0.9*pupilHereInPupil + 0.4*(1-pupilHereInPupil)
				pupilB := 0.4*pupilHereInPupil + 0.2*(1-pupilHereInPupil)

				r := pupilR*pupilHereThreshold + eyeWhiteR*(1-pupilHereThreshold)
				g := pupilG*pupilHereThreshold + eyeWhiteG*(1-pupilHereThreshold)
				b := pupilB*pupilHereThreshold + eyeWhiteB*(1-pupilHereThreshold)

				r *= topEyelidHere
				g *= topEyelidHere
				b *= topEyelidHere

				// Write into the byte slice
				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}
            
            // the rest of the unused pixels should be black
			for ii := n_pixels; ii < n_actual_pixels; ii++ {
				bytes[ii*3+0] = 0;
				bytes[ii*3+1] = 0;
				bytes[ii*3+2] = 0;
            }

			// Send our completed byte slice over the output channel
			bytesOut <- bytes
		}
	}
}
