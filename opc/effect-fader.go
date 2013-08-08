package opc

// Fire
//   Make a burning fire pattern.
//   This pattern is scaled to fit the layout from top to bottom (z).

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	//"time"
)

func MakeEffectFader(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			//t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8

			// fill in bytes array
			var gain float64
			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------

				//pct := float64(ii) / float64(n_pixels)
				//gain := colorutils.Cos(pct, t, 0.2, 0, 1)

				knob1 := float64(midiState.ControllerValues[1]) / 127.0

				if ii%2 == 0 {
					gain = colorutils.Clamp(colorutils.Remap(knob1, 0, 0.5, 0, 1), 0, 1)
				} else {
					gain = colorutils.Clamp(colorutils.Remap(knob1, 0.5, 1, 0, 1), 0, 1)
				}

				bytes[ii*3+0] = byte(float64(bytes[ii*3+0]) * gain)
				bytes[ii*3+1] = byte(float64(bytes[ii*3+1]) * gain)
				bytes[ii*3+2] = byte(float64(bytes[ii*3+2]) * gain)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
