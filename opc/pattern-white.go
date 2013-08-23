package opc

// White
//   Set all pixels to white.

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/config"
	"github.com/longears/pixelslinger/midi"
)

func MakePatternWhite(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		for bytes := range bytesIn {
			H := float64(midiState.ControllerValues[config.HUE_KNOB]) / 127.0
			FADE_TO_WHITE := float64(midiState.ControllerValues[config.MORPH_KNOB]) / 127.0

			r, g, b := colorutils.HslToRgb(H, 1.0, 0.5)
			r = r*(1-FADE_TO_WHITE) + 1*FADE_TO_WHITE
			g = g*(1-FADE_TO_WHITE) + 1*FADE_TO_WHITE
			b = b*(1-FADE_TO_WHITE) + 1*FADE_TO_WHITE

			n_pixels := len(bytes) / 3
			for ii := 0; ii < n_pixels; ii++ {
				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)
			}
			bytesOut <- bytes
		}
	}
}
