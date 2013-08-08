package opc

// White
//   Set all pixels to white.

import (
	"github.com/longears/pixelslinger/midi"
)

func MakePatternWhite(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			for ii := 0; ii < n_pixels; ii++ {
				bytes[ii*3+0] = 255
				bytes[ii*3+1] = 255
				bytes[ii*3+2] = 255
			}
			bytesOut <- bytes
		}
	}
}
