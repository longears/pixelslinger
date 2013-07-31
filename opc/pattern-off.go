package opc

// Off
//   Set all pixels to black.

import (
	"github.com/longears/pixelslinger/midi"
)

func MakePatternOff(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			for ii := 0; ii < n_pixels; ii++ {
				bytes[ii*3+0] = 0
				bytes[ii*3+1] = 0
				bytes[ii*3+2] = 0
			}
			bytesOut <- bytes
		}
	}
}
