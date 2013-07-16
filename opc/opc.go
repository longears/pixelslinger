package opc

import (
	"bufio"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const CONNECTION_TRIES = 1

// times in ms
const WAIT_TO_RETRY = 1000
const WAIT_BETWEEN_RETRIES = 1

// Read locations from JSON file into a slice of floats
func readLocations(fn string) []float64 {
	locations := make([]float64, 0)
	var file *os.File
	var err error
	if file, err = os.Open(fn); err != nil {
		panic(fmt.Sprintf("could not open layout file: %s", fn))
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
	fmt.Printf("[readLocations] Read %v pixel locations from %s\n", len(locations), fn)
	return locations
}

// Try to connect a couple of times
// If we fail after several tries, return nil
func getConnection(ipPort string) net.Conn {
	fmt.Printf("[getConnection] connecting to %v...\n", ipPort)
	triesLeft := CONNECTION_TRIES
	var conn net.Conn
	var err error
	for {
		conn, err = net.Dial("tcp", ipPort)
		if err == nil {
			break
		}
		fmt.Println("[getConnection", triesLeft, err)
		time.Sleep(WAIT_BETWEEN_RETRIES * time.Millisecond)
		triesLeft -= 1
		if triesLeft == 0 {
			return nil
		}
	}
	fmt.Println("[getConnection]    connected")
	return conn
}

// Initiate and Maintain a connection to ipPort.
// When a slice comes in through sendThisSlice, send it with an OPC header.
// Loop forever.
func networkThread(sendThisSlice chan []byte, sliceIsSent chan int, ipPort string) {
	var conn net.Conn
	var err error

	for {
		// wait to get a slice to send
		values := <-sendThisSlice

		// if the connection has gone bad, make a new one
		if conn == nil {
			conn = getConnection(ipPort)
		}
		// if that didn't work, wait a second and restart the loop
		if conn == nil {
			sliceIsSent <- 1
			time.Sleep(WAIT_TO_RETRY * time.Millisecond)
			continue
		}

		// ok, at this point the connection is good

		// make and send OPC header
		channel := byte(0)
		command := byte(0)
		lenLowByte := byte(len(values) % 256)
		lenHighByte := byte(len(values) / 256)
		header := []byte{channel, command, lenHighByte, lenLowByte}
		_, err = conn.Write(header)
		if err != nil {
			// net error -- set conn to nil so we can try to make a new one
			fmt.Println("[net]", err)
			conn = nil
			sliceIsSent <- 1
			continue
		}

		// send actual pixel values
		_, err = conn.Write(values)
		if err != nil {
			// net error -- set conn to nil so we can try to make a new one
			fmt.Println("[net]", err)
			conn = nil
			sliceIsSent <- 1
			continue
		}
		sliceIsSent <- 1
	}
}

// Launch the pixelThread and suck pixels out of it
// Also launch the networkThread and feed the pixels to it
// Run until timeToRun seconds have passed
// Set timeToRun to 0 to run forever
// Set timeToRun to a negative to benchmark your pixelThread function by itself.
func MainLoop(layoutPath, ipPort string, pixelThread func(chan []byte, chan int, []float64), timeToRun float64) {
	// load location and build initial slices
	locations := readLocations(layoutPath)
	n_pixels := len(locations) / 3
	values := make([][]byte, 2)
	values[0] = make([]byte, n_pixels*3)
	values[1] = make([]byte, n_pixels*3)

	filling, sending := 0, 1

	fillThisSlice := make(chan []byte, 0)
	sliceIsFilled := make(chan int, 0)
	sendThisSlice := make(chan []byte, 0)
	sliceIsSent := make(chan int, 0)

	// start threads
	go networkThread(sendThisSlice, sliceIsSent, ipPort)
	go pixelThread(fillThisSlice, sliceIsFilled, locations)

	// main loop
	startTime := float64(time.Now().UnixNano()) / 1.0e9
	lastPrintTime := startTime
	framesSinceLastPrint := int(0)
	var t float64
	for {
		// fps reporting and bookkeeping
		t = float64(time.Now().UnixNano()) / 1.0e9
		framesSinceLastPrint += 1
		if t > lastPrintTime+1 {
			lastPrintTime = t
			fmt.Printf("[main] %f ms (%d fps)\n", 1000.0/float64(framesSinceLastPrint), framesSinceLastPrint)
			framesSinceLastPrint = 0
		}

		// quit after a while, for profiling purposes
		if timeToRun != 0 && t > startTime+math.Abs(timeToRun) {
			return
		}

		// start filling and sending
		fillThisSlice <- values[filling]
		if timeToRun >= 0 {
			sendThisSlice <- values[sending]
		}

		// wait until both are ready
		<-sliceIsFilled
		if timeToRun >= 0 {
			<-sliceIsSent
		}

		// swap
		filling, sending = sending, filling
	}
}
