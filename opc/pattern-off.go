package opc

// Solid black

import (
	"bitbucket.org/davidwallace/go-metal/midi"
)

func MakePatternOff(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiMessageChan chan *midi.MidiMessage) {
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
