package opc

// Raver plaid
//   A rainbowy pattern with moving diagonal black stripes

import (
	"bitbucket.org/davidwallace/go-metal/colorutils"
	"bitbucket.org/davidwallace/go-metal/midi"
	"fmt"
	"time"
)

func getAvailableMidiMessages(midiMessageChan chan *midi.MidiMessage) []*midi.MidiMessage {
	result := make([]*midi.MidiMessage, 0)
	for {
		if len(midiMessageChan) == 0 {
			break
		}
		result = append(result, <-midiMessageChan)
	}
	return result
}

const MIDI_VOLUME_GAIN = 1.5    // multiply incoming midi volumes by this much
const MIDI_BRIGHTNESS_MIN = 0.5 // midi volume 1/127, after MIDI_VOLUME_GAIN, -> this much
const MIDI_BRIGHTNESS_MAX = 1   // midi volume 127, after MIDI_VOLUME_GAIN, -> this much
const SECONDS_TO_FADE = 1

func MakePatternBasicMidi(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiMessageChan chan *midi.MidiMessage) {

		// the current volume of each key, from 0 to 1, after applying MIDI_* adjustments
		keyVolumes := make([]float64, 128)
		// smoothed value: like keyVolumes, but fades away slowly when key is off
		smoothedVolumes := make([]float64, 128)

		last_t := float64(0)
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8
			tDiff := colorutils.Clamp(t-last_t, 0, 5) // limit to max of 5 second to avoid pathological value at startup

			// fetch midi messages and maintain state of keyVolumes
			midiMessages := getAvailableMidiMessages(midiMessageChan)
			for _, m := range midiMessages {
				fmt.Println("        ", m)
				if m.Kind == midi.NOTE_ON {
					if m.Value == 0 {
						keyVolumes[m.Key] = 0
					} else {
						keyVolumes[m.Key] = colorutils.Remap(float64(m.Value)/127*MIDI_VOLUME_GAIN, 0, 1, MIDI_BRIGHTNESS_MIN, MIDI_BRIGHTNESS_MAX)
					}
				}
			}

			// update smoothedVolumes: fade old values and re-apply current states
			for ii, v := range smoothedVolumes {
				smoothedVolumes[ii] = colorutils.Clamp(v-tDiff/SECONDS_TO_FADE, 0, 1)
				if keyVolumes[ii] > smoothedVolumes[ii] {
					smoothedVolumes[ii] = keyVolumes[ii]
				}
			}

			// fill in bytes array
			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------

				if ii < len(keyVolumes) {
					k := keyVolumes[ii]
					s := smoothedVolumes[ii]
					if k > 0 {
						// key is currently down
						bytes[ii*3+0] = colorutils.FloatToByte(k)
						bytes[ii*3+1] = colorutils.FloatToByte(k)
						bytes[ii*3+2] = colorutils.FloatToByte(k)
					} else {
						// key not currently down.  use smoothed value which is fading away over time
						bytes[ii*3+0] = colorutils.FloatToByte(s)
						bytes[ii*3+1] = colorutils.FloatToByte(0)
						bytes[ii*3+2] = colorutils.FloatToByte(0)
					}
				} else {
					// if we have more LEDs than MIDI keys
					bytes[ii*3+0] = 0
					bytes[ii*3+1] = 0
					bytes[ii*3+2] = 0
				}

				//--------------------------------------------------------------------------------
			}

			last_t = t
			bytesOut <- bytes
		}
	}
}
