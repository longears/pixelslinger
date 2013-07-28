package midi

import (
    "time"
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
// If the path can't be opened, it will keep retrying once a second forever until it succeeds.
func GetMidiMessageStream(path string) chan *MidiMessage {
	midiByteChan := make(chan byte, 1000)
	midiMessageChan := make(chan *MidiMessage, 100)

	go TenaciousFileByteStreamerThread(path, midiByteChan)
	go MidiStreamParserThread(midiByteChan, midiMessageChan)

	return midiMessageChan
}

// Stream the bytes from the given path, one byte at a time
// Assumes the file is a special device file which will never hit EOF.
// If the file can't be opened, it will keep trying once a second forver until it succeeds.
func TenaciousFileByteStreamerThread(path string, outCh chan byte) {
    for {
        file, err := os.Open(path)
        if err != nil {
            time.Sleep(time.Duration(1 * time.Second))
            fmt.Println("[midi] couldn't open midi device:", path, " ... waiting and trying again")
            continue
        }
        defer file.Close()
        fmt.Println("[midi] successfully opened midi device", path)

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
}

//================================================================================
// MIDISTATE TYPE

// Keeps track of the current state of the keys and controllers.
type MidiState struct {
    KeyVolumes [128]byte
    ControllerValues [128]byte
    RecentMidiMessages []*MidiMessage
}

// Pull all the available MidiMessages out of the channel.  Requires a channel
// with a buffer length greater than zero.
func getAvailableMidiMessages(midiMessageChan chan *MidiMessage) []*MidiMessage {
	result := make([]*MidiMessage, 0)
	for {
		if len(midiMessageChan) == 0 {
			break
		}
		result = append(result, <-midiMessageChan)
	}
	return result
}

// Given a channel of MidiMessages, read all the messages that are available right now
// and update the MidiState object.
// Requires a channel that has a buffer length greater than zero.
func (midiState *MidiState) UpdateStateFromChannel(midiMessageChan chan *MidiMessage) {
    midiState.UpdateStateFromSlice(getAvailableMidiMessages(midiMessageChan))
}

// Given a slice of MidiMessages, update the MidiState object.
// Note that MidiState.RecentMidiMessages will become a pointer to the
// slice provided to this function; it's not a copy of the slice.
func (midiState *MidiState) UpdateStateFromSlice(midiMessages []*MidiMessage) {
    midiState.RecentMidiMessages = midiMessages
    for _, m := range(midiState.RecentMidiMessages) {
        switch m.Kind {
        case NOTE_OFF:
            midiState.KeyVolumes[m.Key] = 0
        case NOTE_ON:
            midiState.KeyVolumes[m.Key] = m.Value
        case CONTROLLER:
            midiState.ControllerValues[m.Key] = m.Value
        }
    }
}

