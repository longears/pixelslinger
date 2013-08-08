package opc

// Fader effect
//   Listen to a midi knob and fade the entire pattern to black.
//   Fade the even pixels to black first, then the odd pixels.

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	//"time"
)

func MakeEffectFader(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		var r, g, b, gain, flash float64
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			//t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8

			// flash white when pad is down
			pad1 := midiState.KeyVolumes[midi.LPD8_PAD1]
			flash *= 0.6
			if pad1 > 0 {
				flash = 1
			}

			// knob fade to black
			knob1 := float64(midiState.ControllerValues[midi.LPD8_KNOB1]) / 127.0

			// fill in bytes array
			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------
				//pct := float64(ii) / float64(n_pixels)

				r = float64(bytes[ii*3+0]) / 255
				g = float64(bytes[ii*3+1]) / 255
				b = float64(bytes[ii*3+2]) / 255

				r += flash
				g += flash * 1.1
				b += flash * 1.2

				// knob fade to black
				if ii%2 == 0 {
					gain = colorutils.Clamp(colorutils.Remap(knob1, 0, 0.4, 0, 1), 0, 1)
				} else {
					gain = colorutils.Clamp(colorutils.Remap(knob1, 0.6, 1, 0, 1), 0, 1)
				}
				r *= gain
				g *= gain
				b *= gain

				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
