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
    channel byte 
    key byte  // key, controller, instrument
    value byte // velocity, touch, controller value, channel pressure
}

//================================================================================

func MidiThread(inCh chan byte, outCh chan *MidiMessage) {
    fmt.Println("    [midi] starting thread")
    message := new(MidiMessage) // pointer
    ii := 0
    for b := range inCh {
        fmt.Printf("\n    [midi] ii = %v, got byte %v\n", ii, b)
        fmt.Printf("    [midi] current message = %v\n", message)

        if b >= 128 {
            fmt.Println("    [midi]     this is a control byte")
            message = new(MidiMessage)
            message.kind = b & 0xf0
            message.channel = b & 0x0f
            ii = 1
            continue
        }

        fmt.Println("    [midi]     this is a data byte")
        if ii == 1 {
            fmt.Println("    [midi]     1")
            message.key = b
            if message.kind == PROGRAM_CHANGE || message.kind == CHANNEL_PRESSURE {
                fmt.Println("    [midi] sending")
                outCh <- message
                ii = 0
                fmt.Println(message)
                continue
            }
        } else if ii == 2 {
            fmt.Println("    [midi]     2")
            message.value = b
            if message.kind == NOTE_OFF || message.kind == NOTE_ON || message.kind == AFTERTOUCH || message.kind == CONTROLLER || message.kind == PITCH_BEND {
                fmt.Println("    [midi] sending")
                outCh <- message
                ii = 0
                continue
            }
        } else {
            // do nothing, wait for another control byte
        }
        ii += 1
    }
    fmt.Println("    [midi] thread is done")
    close(outCh)
}





