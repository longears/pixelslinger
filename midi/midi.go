package midi

import (
	"errors"
	"fmt"
	"os"
)

//================================================================================
// CONSTANTS

// kinds
const NOTE_OFF = byte(0x80)
const NOTE_ON = byte(0x90)
const AFTERTOUCH = byte(0xa0)
const CONTROLLER = byte(0xb0)
const PROGRAM_CHANGE = byte(0xc0)
const CHANNEL_PRESSURE = byte(0xd0)
const PITCH_BEND = byte(0xe0)
const SYSTEM = byte(0xf0)

// special channel numbers for SYSTEM messages
const CLOCK = byte(8)
const START = byte(10)
const STOP = byte(12)

//================================================================================
// MIDIMESSAGE TYPE

type MidiMessage struct {
	Kind    byte // one of the constants above
	Channel byte // either a channel number of one of the special channel constants
	Key     byte // key, controller, instrument
	Value   byte // velocity, touch, controller value, channel pressure
}

func debug(s string) {
	//fmt.Println("    [midi]", s)
}

func (m *MidiMessage) String() string {
	kindStr := "other"
	switch m.Kind {
	case NOTE_OFF:
		kindStr = "NOTE_OFF"
	case NOTE_ON:
		kindStr = "NOTE_ON"
	case AFTERTOUCH:
		kindStr = "AFTERTOUCH"
	case CONTROLLER:
		kindStr = "CONTROLLER"
	case SYSTEM:
		kindStr = "SYSTEM"
	}
	if kindStr == "other" {
		return fmt.Sprintf("(%#x ch=%v key=%v val=%v)", m.Kind, m.Channel, m.Key, m.Value)
	} else {
		return fmt.Sprintf("(%s ch=%v key=%v val=%v)", kindStr, m.Channel, m.Key, m.Value)
	}
}

//================================================================================
// PARSE MIDI BYTES INTO MESSAGE OBJECTS

// Read a stream of raw MIDI bytes on inCh, parse them into *MidiMessage structs,
// and send over outCh.
// It accepts CLOCK, START, and STOP system messages but other system messages
// are ignored.
func MidiStreamParserThread(inCh chan byte, outCh chan *MidiMessage) {
	debug("starting thread")
	message := new(MidiMessage) // pointer
	ii := 0
	for b := range inCh {
		debug("")
		debug(fmt.Sprintf("ii = %v, got byte %v", ii, b))
		debug(fmt.Sprintf("current message = %v", message))

		// special handling for control bytes which mark the start of messages
		if b >= 128 {
			debug("     this is a control byte")
			message = new(MidiMessage)
			message.Kind = b & 0xf0
			message.Channel = b & 0x0f
			ii = 1
			if message.Kind == SYSTEM && (message.Channel == CLOCK || message.Channel == START || message.Channel == STOP) {
				debug("sending")
				outCh <- message
				ii = 0
			}
			continue
		}

		// handle data bytes differently depending on what number they are in a message
		debug("     this is a data byte")
		if ii == 0 {
			// we should never get a data byte at ii = 0.
			// if it happens it means we're not understanding this part of the stream,
			// so drop the byte and do nothing until we get another control byte.
			continue
		} else if ii == 1 {
			debug("     1")
			message.Key = b
			// these kinds expect 1 data byte, so we might be done:
			if message.Kind == PROGRAM_CHANGE || message.Kind == CHANNEL_PRESSURE {
				debug("sending")
				outCh <- message
				ii = 0
				continue
			}
		} else if ii == 2 {
			debug("     2")
			message.Value = b
			// these kinds expect 2 data bytes, so we might be done:
			if message.Kind == NOTE_OFF || message.Kind == NOTE_ON || message.Kind == AFTERTOUCH || message.Kind == CONTROLLER || message.Kind == PITCH_BEND {
				debug("sending")
				outCh <- message
				ii = 0
				continue
			}
		} else {
			// we're ignoring messages with more than 2 data bytes.
			// do nothing, wait for another control byte
		}

		ii += 1
	}

	// if we get here, inCh has been closed
	debug(" thread is done")
	close(outCh)
}

//================================================================================
// HELPERS

// Start some threads which will read and parse incoming MIDI messages in the background.
// Return a channel which emits *MidiMessage objects.
// "path" should be the path to the midi device, e.g. "/dev/midi1".
// If there's an error opening the midi device file, return an error (otherwise nil) and
// return an empty and closed channel.
func GetMidiMessageStream(path string) (chan *MidiMessage, error) {
	midiByteChan := make(chan byte, 1000)
	midiMessageChan := make(chan *MidiMessage, 100)
	errorChan := make(chan error)

	go FileByteStreamerThread(path, midiByteChan, errorChan)

	// wait until the thread has tried to open the file
	err := <-errorChan
	if err != nil {
		close(midiMessageChan)
		return midiMessageChan, err
	}

	go MidiStreamParserThread(midiByteChan, midiMessageChan)

	return midiMessageChan, nil
}

// Stream the bytes from the given path, one byte at a time
// Assumes the file is a special device file which will never hit EOF.
// If the file can't be opened, send an error over errorChan.
// Once the file is successfully open, send a nil over errorChan.
func FileByteStreamerThread(path string, outCh chan byte, errorChan chan error) {
	file, err := os.Open(path)
	if err != nil {
		errorChan <- errors.New(fmt.Sprintf("Couldn't open file: %v", path))
		return
	}
	defer file.Close()
	errorChan <- nil

	buf := make([]byte, 1)
	for {
		count, err := file.Read(buf)
		if err != nil {
			panic(err)
		}
		if count != 1 {
			panic(fmt.Sprintf("count was %s", count))
		}
		outCh <- buf[0]
	}
}
