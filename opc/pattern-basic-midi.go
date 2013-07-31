package opc

// Basic Midi
//   Turns on one LED for each of the 128 MIDI pitches when that node is being played.
//   LEDs are colored in rainbow order according to the circle of fifths.

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	"time"
)

const MIDI_VOLUME_GAIN = 1.5    // multiply incoming midi volumes by this much
const MIDI_BRIGHTNESS_MIN = 0.3 // midi volume 1/127, after MIDI_VOLUME_GAIN, is remapped to this
const MIDI_BRIGHTNESS_MAX = 1   // midi volume 127, after MIDI_VOLUME_GAIN, is remapped to this
const SECONDS_TO_FADE = 0.3     // how long it takes to fade to black after key is lifted
const FADING_GAIN = 0.3         // fading pixels start at their normal value * this amount
const COLOR_BLEEDING_RAD = 3    // radius of glow effect around pressed keys.  set to 0 for no bleeding
const COLOR_BLEEDING_GAIN = 0.2 // brightness of glow effect (range 0-1)
const MIN_VISIBLE_COLOR = 0.04  // min pixel brightness which is actually visible (range 0-1)
const SUSTAIN = true            // leave lights on while keys are held down

func pitchToRGB(pitch int) (float64, float64, float64) {
	var r, g, b float64
	pitchClass := pitch % 12
	// circle of fifths
	pitchClass = (pitchClass*5 + 1) % 12
	switch pitchClass {
	case 0:
		r = 1
		g = 0
		b = 0
	case 1:
		r = 0.9
		g = 0.4
		b = 0
	case 2:
		r = 0.8
		g = 0.8
		b = 0
	case 3:
		r = 0.4
		g = 0.9
		b = 0
	case 4:
		r = 0
		g = 1
		b = 0
	case 5:
		r = 0
		g = 0.9
		b = 0.4
	case 6:
		r = 0
		g = 0.8
		b = 0.8
	case 7:
		r = 0
		g = 0.4
		b = 0.9
	case 8:
		r = 0
		g = 0
		b = 1
	case 9:
		r = 0.4
		g = 0
		b = 0.9
	case 10:
		r = 0.8
		g = 0
		b = 0.8
	case 11:
		r = 0.9
		g = 0
		b = 0.4
	}
	return r, g, b
}

func MakePatternBasicMidi(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {

		// the current volume of each key, from 0 to 1, after applying MIDI_* adjustments
		keyVolumes := make([]float64, 128)
		// smoothed value: like keyVolumes, but fades away slowly when key is off
		smoothedVolumes := make([]float64, 128)

		last_t := float64(0)
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8
			tDiff := colorutils.Clamp(t-last_t, 0, 5) // limit to max of 5 second to avoid pathological value at startup

			// update keyVolumes from MidiState
			if !SUSTAIN {
				// if not sustaining, reset volumes to zero each frame
				// causing it to be on for 1 frame only when the key is initially pressed
				for ii, _ := range keyVolumes {
					keyVolumes[ii] = 0
				}
			}
			// read midiState into keyVolumes and apply gain adjustments
			for ii, v := range midiState.KeyVolumes {
				if v == 0 {
					keyVolumes[ii] = 0
				} else {
					keyVolumes[ii] = colorutils.Clamp(colorutils.Remap(float64(v)/127*MIDI_VOLUME_GAIN, 0, 1, MIDI_BRIGHTNESS_MIN, MIDI_BRIGHTNESS_MAX), 0, 1)
				}
			}

			// update smoothedVolumes
			for ii, v := range smoothedVolumes {
				// fade old values
				smoothedVolumes[ii] = colorutils.Clamp(v-tDiff/SECONDS_TO_FADE, 0, 1)
				// re-apply current values if greater
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

					r := float64(0)
					g := float64(0)
					b := float64(0)

					// get color based on pitch
					pr, pg, pb := pitchToRGB(ii)

					// apply brightness from keyVolumes or smoothedVolumes
					if k > 0 {
						// key is currently down
						r = pr * k
						g = pg * k
						b = pb * k

					} else {
						// key not currently down.  use smoothed value which is fading away over time
						r = pr * s
						g = pg * s
						b = pb * s
					}

					// color bleeding
					if COLOR_BLEEDING_RAD > 0 {
						for offset := -COLOR_BLEEDING_RAD; offset <= COLOR_BLEEDING_RAD; offset++ {
							if ii == 0 || ii+offset < 0 || ii+offset >= len(keyVolumes) {
								continue
							}
							if keyVolumes[ii+offset] > 0 {
								brightness := float64(offset) / (float64(COLOR_BLEEDING_RAD) + 1)
								if brightness < 0 {
									brightness = -brightness
								}
								brightness = 1 - brightness
								brightness *= COLOR_BLEEDING_GAIN
								pr2, pg2, pb2 := pitchToRGB(ii + offset)
								r += pr2 * keyVolumes[ii+offset] * brightness
								g += pg2 * keyVolumes[ii+offset] * brightness
								b += pb2 * keyVolumes[ii+offset] * brightness
							}
						}
					}

					// avoid black clipping
					if r > 0 {
						r = colorutils.Remap(r, 0, 1, MIN_VISIBLE_COLOR, 1)
					}
					if g > 0 {
						g = colorutils.Remap(g, 0, 1, MIN_VISIBLE_COLOR, 1)
					}
					if b > 0 {
						b = colorutils.Remap(b, 0, 1, MIN_VISIBLE_COLOR, 1)
					}

					bytes[ii*3+0] = colorutils.FloatToByte(r)
					bytes[ii*3+1] = colorutils.FloatToByte(g)
					bytes[ii*3+2] = colorutils.FloatToByte(b)
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
