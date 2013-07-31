package main

import (
	"fmt"
	"github.com/longears/pixelslinger/midi"
)

func main() {
	fmt.Println("-------------------------------------------------------")
	midiMessageChan, err := midi.GetMidiMessageStream("/dev/midi1")
	if err != nil {
		fmt.Println(err)
	}
	for midiMessage := range midiMessageChan {
		fmt.Println("    ", midiMessage)
	}
}
