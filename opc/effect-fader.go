package opc

// Fader effect
//   Listen to a midi knob and fade the entire pattern to black.
//   Fade the even pixels to black first, then the odd pixels.

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/config"
	"github.com/longears/pixelslinger/midi"
	"math"
	"math/rand"
	"time"
)

func MakeEffectFader(locations []float64) ByteThread {

	const (
		FLASH_DURATION_MIN = 2.0 / 40.0  // in seconds
		FLASH_DURATION_MAX = 10.0 / 40.0 // in seconds
		FLASH_R            = 0.6
		FLASH_G            = 0.84
		FLASH_B            = 1.00
		FLASH_DURATION_EXP = 5.0 // exponent for random duration

		MAX_TWINKLE_DENSITY = 0.3
		TWINKLE_DURATION    = 8.0 / 40.0

		EYELID_BLEND = 0.25 // size of eyelid gradient relative to entire bounding box

		FADE_TO_BLACK_TIME = 15.0 / 40.0 // in seconds
	)

	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {

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

		// make persistant random values
		rng := rand.New(rand.NewSource(99))
		randomValues := make([]float64, len(locations)/3)
		for ii := range randomValues {
			randomValues[ii] = math.Pow(rng.Float64(), FLASH_DURATION_EXP)
		}

		lastFlashTime := 0.0
		lastTwinkleTime := 0.0
		lastTwinklePad := 0.0
		lastFadeToBlackPad := 0.0
		fadeToBlackBeginTime := 0.0
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8

			// lightning flash pad
			flashPad := midiState.KeyVolumes[config.FLASH_PAD]
			if flashPad > 0 {
				lastFlashTime = t
			}

			// twinkle strobe pad
			twinklePad := float64(midiState.KeyVolumes[config.TWINKLE_PAD]) / 127.0
			if twinklePad > 0 {
				lastTwinklePad = twinklePad
				lastTwinkleTime = t
			}

			// blink regions
			blinkCirclePad := float64(midiState.KeyVolumes[config.BLINK_CIRCLE_PAD]) / 127.0
			blinkArchPad := float64(midiState.KeyVolumes[config.BLINK_ARCH_PAD]) / 127.0
			blinkBackPad := float64(midiState.KeyVolumes[config.BLINK_BACK_PAD]) / 127.0

			// gain knob
			gainKnob := float64(midiState.ControllerValues[config.GAIN_KNOB]) / 127.0
			gain0 := colorutils.Clamp(colorutils.Remap(gainKnob, 0.75, 0.95, 0, 1), 0, 1)
			gain1 := colorutils.Clamp(colorutils.Remap(gainKnob, 0.40, 0.50, 0, 1), 0, 1)
			gain2 := colorutils.Clamp(colorutils.Remap(gainKnob, 0.05, 0.25, 0, 1), 0, 1)

			// eyelid knob
			eyelidKnob := float64(midiState.ControllerValues[config.EYELID_KNOB]) / 127.0
			eyelidKnob = colorutils.Clamp(colorutils.Remap(eyelidKnob, 0.05, 0.95, 0, 1), 0, 1)

			// saturation knob
			desatKnob := float64(midiState.ControllerValues[config.DESAT_KNOB]) / 127.0

			// fade to black pad
			fadeToBlackPad := float64(midiState.KeyVolumes[config.FADE_TO_BLACK_PAD]) / 127.0
			if fadeToBlackPad > 0 && lastFadeToBlackPad == 0 {
				// pad has just gone down
				fadeToBlackBeginTime = t
			}
			fadeToBlackAmount := 1.0
			if fadeToBlackPad > 0 {
				fadeToBlackAmount = 1 - colorutils.Clamp((t-fadeToBlackBeginTime)/FADE_TO_BLACK_TIME, 0, 1)
			}
			lastFadeToBlackPad = fadeToBlackPad

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
					eyelidKnob-(1-zp)*EYELID_BLEND/2,
					eyelidKnob+zp*EYELID_BLEND/2, 1, 0), 0, 1)
				r *= eyelid
				g *= eyelid
				b *= eyelid

				// lightning flash
				flashAmt := colorutils.Clamp(colorutils.Remap(t-lastFlashTime, 0, FLASH_DURATION_MIN+randomValues[ii]*(FLASH_DURATION_MAX-FLASH_DURATION_MIN), 1, 0), 0, 1)
				r = r*(1-flashAmt) + flashAmt*FLASH_R
				g = g*(1-flashAmt) + flashAmt*FLASH_G
				b = b*(1-flashAmt) + flashAmt*FLASH_B

				// twinkle strobe
				twinkleAmt := colorutils.Clamp(colorutils.Remap(t-lastTwinkleTime, 0, TWINKLE_DURATION, 1, 0), 0, 1)
				if twinkleAmt > 0 {
					thisTwinkle := rand.Float64()
					if thisTwinkle < lastTwinklePad*MAX_TWINKLE_DENSITY {
						thisTwinkle = twinkleAmt
					} else {
						thisTwinkle = 0
					}
					r += thisTwinkle
					g += thisTwinkle
					b += thisTwinkle
				}

				// blink regions
				if blinkCirclePad > 0 && ii < 160*1 {
					r = 1
					g = 1
					b = 1
				}
				if blinkArchPad > 0 && 160*1 <= ii && ii < 160*3 {
					r = 1
					g = 1
					b = 1
				}
				if blinkBackPad > 0 && 160*3 <= ii && ii < 160*5 {
					r = 1
					g = 1
					b = 1
				}

				// desaturation
				if desatKnob != 0 {
					gray := (r + g + b) / 3.0 * 1.3 // boost it a little bit
					r = r*(1-desatKnob) + gray*desatKnob
					g = g*(1-desatKnob) + gray*desatKnob
					b = b*(1-desatKnob) + gray*desatKnob
				}

				// fade to black
				if fadeToBlackPad > 0 {
					r *= fadeToBlackAmount
					g *= fadeToBlackAmount
					b *= fadeToBlackAmount
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
