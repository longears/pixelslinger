package midi 

import (
    "testing"
    "fmt"
)

//================================================================================

func TestMidiThread(t *testing.T) {
    fmt.Println("[test] beginning test")

    // prepare bytes to send
    seq := []byte{0x80, 60, 0}

    // make channels and start thread
    inCh := make(chan byte, 1000)
    outCh := make(chan *MidiMessage, 1000)
    fmt.Println("[test] starting thread...")
    go MidiThread(inCh, outCh)
    fmt.Println("[test]     done")

    // send bytes, close channel, get results back
    fmt.Println("[test] sending bytes...")
    for _,v := range seq {
        inCh <- v
    }
    fmt.Println("[test]     done")

    fmt.Println("[test] closing channel...")
    close(inCh)
    fmt.Println("[test]     done")

    fmt.Println("[test] getting results...")
    for midiMessage := range outCh {
        fmt.Printf("[test]    %v\n",midiMessage)
    }
    fmt.Println("[test]     done")

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

