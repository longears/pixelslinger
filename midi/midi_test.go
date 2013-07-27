package midi 

import (
    "testing"
    "fmt"
)

//================================================================================


func midiBytesToMessages(bytes []byte) []*MidiMessage {
    //fmt.Println("[subtest] beginning test")

    midiMessages := make([]*MidiMessage,0)

    // make channels and start thread
    inCh := make(chan byte, 1000)
    outCh := make(chan *MidiMessage, 1000)
    //fmt.Println("[subtest] starting thread...")
    go MidiThread(inCh, outCh)
    //fmt.Println("[subtest]     done")

    // send bytes, close channel, get results back
    //fmt.Println("[subtest] sending bytes...")
    for _,v := range bytes {
        inCh <- v
    }
    //fmt.Println("[subtest]     done")

    //fmt.Println("[subtest] closing channel...")
    close(inCh)
    //fmt.Println("[subtest]     done")

    //fmt.Println("[subtest] getting results...")
    for midiMessage := range outCh {
        //fmt.Printf("[subtest]    %v\n",midiMessage)
        midiMessages = append(midiMessages, midiMessage)
    }
    //fmt.Println("[subtest]     done")
    return midiMessages
}


func midiTest(t *testing.T, bytes []byte, expectedMessageKinds []byte) {
    midiMessages := midiBytesToMessages(bytes)
    fmt.Println("[test] ", bytes, "-->", midiMessages)
    if len(midiMessages) != len(expectedMessageKinds) {
        t.Errorf("incorrect number of response messages")
        return
    }
    for ii := range(midiMessages) {
        if midiMessages[ii].kind != expectedMessageKinds[ii] {
            t.Errorf("incorrect message kind")
        }
    }
}


func TestMidiStreamParser(t *testing.T) {
    midiTest(t, []byte{0x90, 60, 0}, []byte{NOTE_ON})
    midiTest(t, []byte{7, 0x90, 60, 0}, []byte{NOTE_ON})
    midiTest(t, []byte{0x90, 60, 0, 7}, []byte{NOTE_ON})
    midiTest(t, []byte{0x9f, 60, 0}, []byte{NOTE_ON})
    midiTest(t, []byte{0x90, 31, 127, 0x90, 31, 0}, []byte{NOTE_ON, NOTE_ON})
    midiTest(t, []byte{0x90, 31, 127, 7, 0x90, 31, 0}, []byte{NOTE_ON, NOTE_ON})
    midiTest(t, []byte{0x90, 31, 127, 7, 7, 7, 7, 7, 0x90, 31, 0}, []byte{NOTE_ON, NOTE_ON})
    midiTest(t, []byte{0xb0, 64, 127, 0x90, 60, 0}, []byte{CONTROLLER, NOTE_ON})
    midiTest(t, []byte{0x90, 31, 127, 0xf0+CLOCK, 0x90, 31, 0}, []byte{NOTE_ON, SYSTEM, NOTE_ON})
    midiTest(t, []byte{0x90, 31, 127, 0xf0+START, 0x90, 31, 0}, []byte{NOTE_ON, SYSTEM, NOTE_ON})
    midiTest(t, []byte{0x90, 31, 127, 0xf0+STOP, 0x90, 31, 0}, []byte{NOTE_ON, SYSTEM, NOTE_ON})
    midiTest(t, []byte{0x90, 31, 127, 0xf0, 0x90, 31, 0}, []byte{NOTE_ON, NOTE_ON})
}


//================================================================================
/*
func TestCosTable(t *testing.T) {
	var correct, approx, diff float64
	for x := -30.0; x < 30; x += 0.1387 {
		correct = math.Cos(x)
		approx = CosTable(x)
		diff = math.Abs(correct - approx)
		if diff > 0.1 {
			t.Errorf("Cos != CosTable: %v - %v = %v", correct, approx, diff)
		}
	}
	var bigOffset float64 = 1373963358.2 * 2 * 3.14159
	for x := bigOffset - 30; x < bigOffset+30; x += 0.1387 {
		correct = math.Cos(x)
		approx = CosTable(x)
		diff = math.Abs(correct - approx)
		if diff > 0.1 {
			t.Errorf("Cos != CosTable: %v - %v = %v", correct, approx, diff)
		}
	}
}
*/

