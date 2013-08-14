package config

import (
	"github.com/longears/pixelslinger/midi"
)

// midi pads
const (
	FLASH_PAD = midi.LPD8_PAD1
	TWINKLE_PAD = midi.LPD8_PAD2
)

// midi knobs
const (
	GAIN_KNOB   = midi.LPD8_KNOB1
	EYELID_KNOB = midi.LPD8_KNOB2
	SPEED_KNOB  = midi.LPD8_KNOB3
)

// knob starting values before they have been moved
//  (because the midi hardware only sends us values when the knobs move)
var DEFAULT_KNOB_VALUES map[byte]byte

func init() {
	DEFAULT_KNOB_VALUES = map[byte]byte{
		GAIN_KNOB:   127,
		EYELID_KNOB: 127,
		SPEED_KNOB:  63,
	}
}
