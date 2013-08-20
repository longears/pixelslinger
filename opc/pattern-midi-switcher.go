package opc

// Basic Midi
//   Turns on one LED for each of the 128 MIDI pitches when that node is being played.
//   LEDs are colored in rainbow order according to the circle of fifths.

import (
	"github.com/longears/pixelslinger/midi"
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/config"
	"time"
)

func MakePatternMidiSwitcher(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {

        // The patterns that our MIDI knob will switch between
        PATTERN_LIST := []string{
            "diamond",
            "fire",
            "raver-plaid",
            "shield",
            "spatial-stripes",
            "eye",
        }

        // channels for communication with subpattern
        chanToPattern := make(chan []byte, 0)
        chanFromPattern := make(chan []byte, 0)

        var patternName, lastPatternName string
		for bytes := range bytesIn {
            t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8

            // decide which subpattern we want for this frame

            // // VERSION A for testing
            // switchKnob := colorutils.PosMod2(t, 1)
            // _ = config.SWITCH_KNOB

            // VERSION B for production
            _ = t
            _ = colorutils.PosMod
			switchKnob := float64(midiState.ControllerValues[config.SWITCH_KNOB]) / 127.0

            // assume switchKnob is between 0 and 1
            ii := int(switchKnob * float64(len(PATTERN_LIST)) * 0.99999)
            patternName = PATTERN_LIST[ii]

            // Subpattern has changed.  Close old one and start new one.
            // This is not ideal because it has to re-init each pattern every
            // time they switch.  This makes the patterns always start in the same
            // place depending on how their time calculations work.
            // On the other hand, if we kept all the patterns around all the time,
            // some might freak out with weird time calculations when their frames
            // are sometimes a long time apart while other patterns were running.
            if patternName != lastPatternName {
                // If patterns are properly written using "for byte := range bytesIn", this
                // should terminate the previous pattern thread
                close(chanToPattern)
                // Make a new to-pattern channel.  We will re-use the from-pattern channel.
                chanToPattern = make(chan []byte, 0)
                sourceThreadMaker := PATTERN_REGISTRY[patternName]
                sourceThread := sourceThreadMaker(locations)
                go sourceThread(chanToPattern, chanFromPattern, midiState)
            }
            lastPatternName = patternName

            // send byte slice to subpattern
            chanToPattern <- bytes
            // get result back from subpattern
            bytes = <-chanFromPattern

            // send our result back to our parent
			bytesOut <- bytes
		}
	}
}

