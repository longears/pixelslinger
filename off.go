package main

// TODO
//  command line parameters (layout file, ip:port)
//  figure out how to have multiple executable files in a directory
//  make lookup tables (Cos, Gamma, ...?)
//  write a pattern that relies on location

import (
	"bitbucket.org/davidwallace/go-tower/opc"
	"fmt"
)

func pixelThread(fillThisSlice chan []byte, sliceIsFilled chan int, locations []float64) {

	for {
		// wait for slice to fill
		values := <-fillThisSlice
		n_pixels := len(values) / 3
		// fill in values array
		for ii := 0; ii < n_pixels; ii++ {
			//--------------------------------------------------------------------------------

			values[ii*3+0] = 0
			values[ii*3+1] = 0
			values[ii*3+2] = 0

			//--------------------------------------------------------------------------------
		}
		sliceIsFilled <- 1
	}
	panic("SUCCESS")
}

func main() {
	fmt.Println("--------------------------------------------------------------------------------\\")
	defer fmt.Println("--------------------------------------------------------------------------------/")

	layoutPath, ipPort, _, _ := opc.ParseFlags()

	// load location and build initial slices
	locations := opc.ReadLocations(layoutPath)
	n_pixels := len(locations) / 3
	values := make([]byte, n_pixels*3)

	sendThisSlice := make(chan []byte, 0)
	sliceIsSent := make(chan int, 0)

	// start threads
	go opc.NetworkThread(sendThisSlice, sliceIsSent, ipPort)

	fmt.Println("OFF!")
	sendThisSlice <- values
	<-sliceIsSent

}
