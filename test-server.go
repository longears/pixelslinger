package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func setLED(ledNum int, val int) {
	ledFn := fmt.Sprintf("/sys/class/leds/beaglebone:green:usr%d/brightness", ledNum)
	fmt.Println(ledFn)

	// open output file
	ledFile, err := os.Create(ledFn)
	if err != nil {
		panic(err)
	}
	// close ledFile on exit and check for its returned error
	defer func() {
		if err := ledFile.Close(); err != nil {
			panic(err)
		}
	}()

	if _, err := ledFile.WriteString(strconv.Itoa(val)); err != nil {
		panic(err)
	}
}

func sendBytes(fd *os.File, bytes []byte) {
	fmt.Println("[sendBytes]", bytes)
	if _, err := fd.Write(bytes); err != nil {
		panic(err)
	}
}

func spiThread(pixelsToSend chan []byte, sendingIsDone chan int) {

	spiFn := "/dev/spidev1.0"

	// open output file
	spiFile, err := os.Create(spiFn)
	if err != nil {
		panic(err)
	}
	// close spiFile on exit and check for its returned error
	defer func() {
		if err := spiFile.Close(); err != nil {
			panic(err)
		}
	}()

	flipper := 0
	for pixels := range pixelsToSend {
		fmt.Println("[send] starting to send", len(pixels), "values")

		setLED(0, flipper)
		flipper = 1 - flipper

		bytes := make([]byte, 0)

		// zeros
		numZeroes := (len(pixels) / 32) + 2
		for ii := 0; ii < numZeroes*5; ii++ {
			bytes = append(bytes, 0)
		}

		// pixels
		for _, v := range pixels {
			// high bit is always on, remaining seven bits are data
			v2 := 128 | (v >> 1)
			bytes = append(bytes, v2)
		}

		// final zero
		bytes = append(bytes, 0)
		sendBytes(spiFile, bytes)

		sendingIsDone <- 1
	}
}

func main() {
	fmt.Println("--------------------------------------------------------------------------------\\")

	pixelsToSend := make(chan []byte, 0)
	sendingIsDone := make(chan int, 0)
	go spiThread(pixelsToSend, sendingIsDone)

	for ii := 0; true; ii = (ii + 1) % 256 {
		pixels := []byte{255, 0, 0, 0, 255, 0, 0, 0, 255, 0, 0, 0}
		pixels[9] = byte((ii * 9) % 256)
		pixels[10] = byte((ii * 9) % 256)
		pixels[11] = byte((ii * 9) % 256)
		fmt.Println("[main] pixels =", pixels)
		pixelsToSend <- pixels
		<-sendingIsDone
		time.Sleep(300 * time.Millisecond)
	}

	fmt.Println("--------------------------------------------------------------------------------/")
}
