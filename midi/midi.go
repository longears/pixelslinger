package midi

import (
    "fmt"
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

// special channel numbers for system messages
const CLOCK = byte(8)
const START = byte(10)
const STOP = byte(12)

//================================================================================
// MIDIMESSAGE TYPE

type MidiMessage struct {
    kind byte // one of the constants above
    channel byte 
    key byte  // key, controller, instrument
    value byte // velocity, touch, controller value, channel pressure
}

func debug(s string) {
    //fmt.Println("    [midi]", s)
}

func (m *MidiMessage) String() string {
    kindStr := "other"
    switch m.kind {
    case NOTE_OFF:
        kindStr = "note-off"
    case NOTE_ON:
        kindStr = "note-on"
    case AFTERTOUCH:
        kindStr = "aftertouch"
    case CONTROLLER:
        kindStr = "controller"
    }
    if kindStr == "other" {
        return fmt.Sprintf("(%#x ch=%v key=%v val=%v)", m.kind, m.channel, m.key, m.value)
    } else {
        return fmt.Sprintf("(%s ch=%v key=%v val=%v)", kindStr, m.channel, m.key, m.value)
    }
}

//================================================================================

func MidiThread(inCh chan byte, outCh chan *MidiMessage) {
    debug("starting thread")
    message := new(MidiMessage) // pointer
    ii := 0
    for b := range inCh {
        debug("")
        debug(fmt.Sprintf("ii = %v, got byte %v", ii, b))
        debug(fmt.Sprintf("current message = %v", message))

        if b >= 128 {
            debug("     this is a control byte")
            message = new(MidiMessage)
            message.kind = b & 0xf0
            message.channel = b & 0x0f
            ii = 1
            if message.kind == SYSTEM && (message.channel == CLOCK || message.channel == START || message.channel == STOP) {
                debug("sending")
                outCh <- message
                ii = 0
            }
            continue
        }

        debug("     this is a data byte")
        if ii == 0 {
            // ignore
            continue
        } else if ii == 1 {
            debug("     1")
            message.key = b
            if message.kind == PROGRAM_CHANGE || message.kind == CHANNEL_PRESSURE {
                debug("sending")
                outCh <- message
                ii = 0
                continue
            }
        } else if ii == 2 {
            debug("     2")
            message.value = b
            if message.kind == NOTE_OFF || message.kind == NOTE_ON || message.kind == AFTERTOUCH || message.kind == CONTROLLER || message.kind == PITCH_BEND {
                debug("sending")
                outCh <- message
                ii = 0
                continue
            }
        } else {
            // do nothing, wait for another control byte
        }
        ii += 1
    }
    debug(" thread is done")
    close(outCh)
}





