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
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {

		flashDuration := 2.0 / 40.0 // seconds

		var r, g, b, gain, flash, lastFlashTime float64
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8

			// flash white when pad is down
			pad1 := midiState.KeyVolumes[midi.LPD8_PAD1]
			if pad1 > 0 {
				lastFlashTime = t
			}
			flash = colorutils.Clamp(colorutils.Remap(t-lastFlashTime, 0, flashDuration, 1, 0), 0, 1)

			// gain knob
			knob1 := float64(midiState.ControllerValues[midi.LPD8_KNOB1]) / 127.0
			gain0 := colorutils.Clamp(colorutils.Remap(knob1, 0.75, 0.95, 0, 1), 0, 1)
			gain1 := colorutils.Clamp(colorutils.Remap(knob1, 0.40, 0.50, 0, 1), 0, 1)
			gain2 := colorutils.Clamp(colorutils.Remap(knob1, 0.05, 0.25, 0, 1), 0, 1)

			// fill in bytes array
			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------
				//pct := float64(ii) / float64(n_pixels)

				r = float64(bytes[ii*3+0]) / 255
				g = float64(bytes[ii*3+1]) / 255
				b = float64(bytes[ii*3+2]) / 255

				// gain knob
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

				// flash
				r = r*(1-flash) + 1*flash
				g = g*(1-flash) + 1*flash
				b = b*(1-flash) + 1*flash

				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
