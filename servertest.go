package main

import (
	"fmt"
	"github.com/longears/pixelslinger/opc"
)

func main() {
	ch := opc.LaunchOpcServer(":7890")
	for opcMessage := range ch {
		fmt.Printf("[servertest] Got OPC message. channel %v, command %v, length %v\n", opcMessage.Channel, opcMessage.Command, len(opcMessage.Bytes))
	}
}
