package opc

// Shield
//   Creates a shimmering electric blue / purple pattern.

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/config"
	"github.com/longears/pixelslinger/midi"
	"time"
)

func MakePatternShield(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		last_t := 0.0
		t := 0.0
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3

			// time and speed knob bookkeeping
			this_t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8
			speedKnob := float64(midiState.ControllerValues[config.SPEED_KNOB]) / 127.0
			if speedKnob < 0.5 {
				speedKnob = colorutils.RemapAndClamp(speedKnob, 0, 0.4, 0, 1)
			} else {
				speedKnob = colorutils.RemapAndClamp(speedKnob, 0.6, 1, 1, 4)
			}
			if midiState.KeyVolumes[config.SLOWMO_PAD] > 0 {
				speedKnob *= 0.25
			}
			if last_t != 0 {
				t += (this_t - last_t) * speedKnob
			}
			last_t = this_t

			// fill in bytes slice
			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------

				var r, g, b float64
				x := locations[ii*3+0]
				y := locations[ii*3+1]
				z := locations[ii*3+2]

				//x, y, z = z, x, y
				//z2 := z + colorutils.Cos(x+y, t/18, 5, 0, 1.5)

				// warp coordinates up and down to give the horizontal stripes some wiggle
				z2 := z + colorutils.Cos(x+y, t/18, 5, 0, 0.5)

				// make large sine wave traveling upward slowly
				// it slowly goes back and forth between contrasty and not contrasty
				bigWaveMin := colorutils.Cos(t*0.05, 0, 1, -1.8, 0.3) // was 0.3 before sinewave'd
				bigWave := colorutils.Cos2(z2, t/4, 1.5, bigWaveMin, 1)

				// make small wave very quickly moving upward
				smallWave := colorutils.Cos2(z2, t*4, 0.3, 0.7, 1)
				// small wave comes in pulses controlled by smallWaveAmt
				smallWaveAmtPeriod := 0.5 // try 0.2, 2, 5
				smallWaveAmt := colorutils.Clamp(colorutils.Cos2(t*0.3-0.2*(x+y+z), 0, smallWaveAmtPeriod, 0, 1), 0, 1)
				smallWaveAmt *= smallWaveAmt // square it for toe falloff smoothing
				// apply pulses to small wave by using smallWaveAmt to crossfade with a constant value
				smallWave = 0.7*(1-smallWaveAmt) + smallWave*smallWaveAmt

				// combine big and small waves
				wave := bigWave * smallWave

				// crossfade between two color schemes
				// 0 is purple and blue, 1 is green and blue
				purpleGreenMix := colorutils.Clamp(colorutils.Cos(x+t*0.23, 0, 3, 0, 1), 0, 1)

				// green scheme
				r += purpleGreenMix * (wave - 0.6) * 0.2
				g += purpleGreenMix * (wave - 0.4) * 2.0
				b += purpleGreenMix * wave * 0.8

				// purple scheme
				r += (1 - purpleGreenMix) * (wave - 0.4) * 1.6
				g += (1 - purpleGreenMix) * (wave - 0.3) * 0.4
				b += (1 - purpleGreenMix) * wave * 1

				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
