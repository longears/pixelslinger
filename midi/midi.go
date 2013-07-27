package midi

import (
    "fmt"
)

const NOTE_OFF = 0x80
const NOTE_ON = 0x90
const AFTERTOUCH = 0xa0
const CONTROLLER = 0xb0
const PROGRAM_CHANGE = 0xc0
const CHANNEL_PRESSURE = 0xd0
const PITCH_BEND = 0xe0
const SYSTEM = 0xf0

type MidiMessage struct {
    kind byte // one of the constants above
    key byte  // key, controller, instrument
    value byte // velocity, touch, controller value, channel pressure
}

//================================================================================

func MidiThread(inCh chan byte, outCh chan *MidiMessage) {
    fmt.Println("[midi] starting thread")
    message := new(MidiMessage) // pointer
    ii := 0
    messageIsGood := false
    for b := range inCh {
        fmt.Println("[midi] getting byte", b)
        if b >= 128 {
            fmt.Println("[midi]     this is a control byte")
            // we got a control byte, so send the last message (if good)...
            if messageIsGood {
                outCh <- message
            }
            // and begin a new message.
            message = new(MidiMessage)
            message.kind = b
            messageIsGood = true
            ii = 0
        } else {
            fmt.Println("[midi]     this is a data byte")
            ii += 1
            if ii == 1 {
                message.key = b
            } else if ii == 2 {
                message.value = b
            } else {
                // this message was too long.  let's ignore it.
                messageIsGood = false
            }
        }
    }
    fmt.Println("[midi] thread is done")
}





