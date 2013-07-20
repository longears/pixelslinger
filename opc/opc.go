package opc

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

//--------------------------------------------------------------------------------
// NET-RELATED CONSTANTS

// Times are in milliseconds
const CONNECTION_TRIES = 1
const WAIT_TO_RETRY = 1000
const WAIT_BETWEEN_RETRIES = 1

//--------------------------------------------------------------------------------
// OPC LAYOUT FORMAT

// Read locations from JSON file into a slice of floats
func ReadLocations(fn string) []float64 {
	locations := make([]float64, 0)
	var file *os.File
	var err error
	if file, err = os.Open(fn); err != nil {
		panic(fmt.Sprintf("[opc.ReadLocations] could not open layout file: %s", fn))
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '[' || line[0] == ']' {
			continue
		}
		line = strings.Split(line, "[")[1]
		line = strings.Split(line, "]")[0]
		coordStrings := strings.Split(line, ", ")
		var x, y, z float64
		x, err = strconv.ParseFloat(coordStrings[0], 64)
		y, err = strconv.ParseFloat(coordStrings[1], 64)
		z, err = strconv.ParseFloat(coordStrings[2], 64)
		locations = append(locations, x, y, z)
	}
	fmt.Printf("[opc.ReadLocations] Read %v pixel locations from %s\n", len(locations), fn)
	return locations
}

//--------------------------------------------------------------------------------
// NET HELPERS

// Try to connect.  Retry several times in a row if needed.
// On failure, return nil.
func getConnection(ipPort string) net.Conn {
	fmt.Printf("[opc.getConnection] connecting to %v...\n", ipPort)
	triesLeft := CONNECTION_TRIES
	var conn net.Conn
	var err error
	for {
		conn, err = net.Dial("tcp", ipPort)
		if err == nil {
			// success
			fmt.Println("[opc.getConnection]    connected")
			return conn
		}
		fmt.Println("[opc.getConnection]", triesLeft, err)
		time.Sleep(WAIT_BETWEEN_RETRIES * time.Millisecond)
		triesLeft -= 1
		if triesLeft == 0 {
			// failure
			return nil
		}
	}
}

//--------------------------------------------------------------------------------
// SENDING GOROUTINES

func SendToDevNullThread(bytesIn chan []byte, bytesOut chan []byte) {
	fmt.Println("[opc.SendToDevNullThread] starting up")
	for bytes := range bytesIn {
		bytesOut <- bytes
	}
}

// Recieve byte slices over the pixelsToSend channel.
// When we get one, write it to the SPI file descriptor and toggle one of the Beaglebone's onboard LEDs.
// After sending the frame, send 1 over the bytesOut channel.
// The byte slice should hold values from 0 to 255 in [r g b  r g b  r g b  ... ] order.
// Loops until the input channel is closed.
func SendToLPD8806Thread(bytesIn chan []byte, bytesOut chan []byte, spiFn string) {
	fmt.Println("[opc.SendToLPD8806Thread] starting up")

	// open output file and keep the file descriptor around
	spiFile, err := os.Create(spiFn)
	if err != nil {
		fmt.Println("[opc.SendToLPD8806Thread] Error opening SPI file:")
		fmt.Println(err)
		os.Exit(1)
	}
	// close spiFile on exit and check for its returned error
	defer func() {
		if err := spiFile.Close(); err != nil {
			panic(err)
		}
	}()

	// as we get byte slices over the channel...
	for bytes := range bytesIn {
		fmt.Println("[opc.SendToLPD8806Thread] starting to send", len(bytes), "bytes")

		// build a new slice of bytes in the format the LED strand wants
		// TODO: avoid allocating these bytes over and over
		bytes := make([]byte, 0)

		// leading zeros to begin a new frame of bytes
		numZeroes := (len(bytes) / 32) + 2
		for ii := 0; ii < numZeroes*5; ii++ {
			bytes = append(bytes, 0)
		}

		// bytes
		for _, v := range bytes {
			// high bit must be always on, remaining seven bits are data
			v2 := 128 | (v >> 1)
			bytes = append(bytes, v2)
		}

		// final zero to latch the last pixel
		bytes = append(bytes, 0)

		// actually send bytes over the wire
		if _, err := spiFile.Write(bytes); err != nil {
			panic(err)
		}

		bytesOut <- bytes
	}
}

// Initiate and Maintain a connection to ipPort.
// When a slice comes in through bytesIn, send it with an OPC header.
// Loops until the input channel is closed.
func SendToOpcThread(bytesIn chan []byte, bytesOut chan []byte, ipPort string) {
	fmt.Println("[opc.SendToOpcThread] starting up")

	var conn net.Conn
	var err error

	for bytes := range bytesIn {
		// if the connection has gone bad, make a new one
		if conn == nil {
			conn = getConnection(ipPort)
		}
		// if that didn't work, wait a second and restart the loop
		if conn == nil {
			bytesOut <- bytes
			fmt.Println("[opc.SendToOpcThread] waiting to retry")
			time.Sleep(WAIT_TO_RETRY * time.Millisecond)
			continue
		}

		// ok, at this point the connection is good

		// make and send OPC header
		channel := byte(0)
		command := byte(0)
		lenLowByte := byte(len(bytes) % 256)
		lenHighByte := byte(len(bytes) / 256)
		header := []byte{channel, command, lenHighByte, lenLowByte}
		_, err = conn.Write(header)
		if err != nil {
			// net error -- set conn to nil so we can try to make a new one
			fmt.Println("[opc.SendToOpcThread]", err)
			conn = nil
			bytesOut <- bytes
			continue
		}

		// send actual pixel values
		_, err = conn.Write(bytes)
		if err != nil {
			// net error -- set conn to nil so we can try to make a new one
			fmt.Println("[opc.SendToOpcThread]", err)
			conn = nil
			bytesOut <- bytes
			continue
		}
		bytesOut <- bytes
	}
}
