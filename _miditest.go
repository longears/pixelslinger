package main

import (
	"fmt"
	"github.com/longears/pixelslinger/midi"
)

func main() {
	fmt.Println("-------------------------------------------------------")
	midiMessageChan := midi.GetMidiMessageStream("/dev/midi1")
	for midiMessage := range midiMessageChan {
		fmt.Println(midiMessage)
	}
}
