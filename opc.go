package main

import (
	"bitbucket.org/davidwallace/go-opc/colorutils"
	"bufio"
	"fmt"
	"github.com/davecheney/profile"
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
const PIXEL_SLEEP_PER_FRAME = 1

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

// Connect to an ip:port and send the values array with an OPC header in front of it.
func networkThread(sendThisSlice chan []byte, sliceIsSent chan int, ipPort string) {
	var conn net.Conn
	var err error

	// loop forever, getting slices to send
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

		// ok, connection is good

		// make and send header
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

func pixelThread(fillThisSlice chan []byte, sliceIsFilled chan int) {
	var (
		// how many sine wave cycles are squeezed into our n_pixels
		// 24 happens to create nice diagonal stripes on the wall layout
		freq_r float64 = 24
		freq_g float64 = 24
		freq_b float64 = 24

		// how many seconds the color sine waves take to shift through a complete cycle
		speed_r float64 = 7
		speed_g float64 = -13
		speed_b float64 = 19
	)

	for {
		// wait for slice to fill
		values := <-fillThisSlice
		n_pixels := len(values) / 3

		t := float64(time.Now().UnixNano()) / 1.0e9

		// fill in values array
		for ii := 0; ii < n_pixels; ii++ {
			pct := float64(ii) / float64(n_pixels)

			// diagonal black stripes
			pct_jittered := math.Mod((pct*77)+37, 37)
			blackstripes := colorutils.Cos(pct_jittered, t*0.05, 1, -1.5, 1.5) // offset, period, minn, maxx
			blackstripes_offset := colorutils.Cos(t, 0.9, 60, -0.5, 3)
			blackstripes = colorutils.Clamp(blackstripes+blackstripes_offset, 0, 1)

			// 3 sine waves for r, g, b which are out of sync with each other
			r := blackstripes * colorutils.Remap(math.Cos((t/speed_r+pct*freq_r)*math.Pi*2), -1, 1, 0, 1)
			g := blackstripes * colorutils.Remap(math.Cos((t/speed_g+pct*freq_g)*math.Pi*2), -1, 1, 0, 1)
			b := blackstripes * colorutils.Remap(math.Cos((t/speed_b+pct*freq_b)*math.Pi*2), -1, 1, 0, 1)

			//values[ii*3+0] = colorutils.FloatToByte(r)
			//values[ii*3+1] = colorutils.FloatToByte(g)
			//values[ii*3+2] = colorutils.FloatToByte(b)
			saveToSlice(values, ii, r, g, b)
		}

		// done
		time.Sleep(PIXEL_SLEEP_PER_FRAME * time.Millisecond)
		sliceIsFilled <- 1
	}
}

func saveToSlice(slice []byte, ii int, r, g, b float64) {
	slice[ii*3+0] = colorutils.FloatToByte(r)
	slice[ii*3+1] = colorutils.FloatToByte(g)
	slice[ii*3+2] = colorutils.FloatToByte(b)
}

func main() {
	defer profile.Start(profile.CPUProfile).Stop()

	path := "layouts/freespace.json"
	ipPort := "127.0.0.1:7890"
	//ipPort := "192.168.11.11:7890"

	LOCATIONS := readLocations(path)
	N_PIXELS := len(LOCATIONS) / 3
	VALUES := make([][]byte, 2)
	VALUES[0] = make([]byte, N_PIXELS*3)
	VALUES[1] = make([]byte, N_PIXELS*3)

	filling, sending := 0, 1

	fillThisSlice := make(chan []byte, 0)
	sliceIsFilled := make(chan int, 0)
	sendThisSlice := make(chan []byte, 0)
	sliceIsSent := make(chan int, 0)

	// start threads
	go networkThread(sendThisSlice, sliceIsSent, ipPort)
	go pixelThread(fillThisSlice, sliceIsFilled)

	// main loop
	startTime := float64(time.Now().UnixNano()) / 1.0e9
	lastPrintTime := startTime
	framesSinceLastPrint := int(0)
	var t float64
	for {
		// fps bookkeeping
		t = float64(time.Now().UnixNano()) / 1.0e9
		framesSinceLastPrint += 1
		if t > lastPrintTime+1 {
			lastPrintTime = t
			fmt.Printf("[main] %f ms (%d fps)\n", 1000.0/float64(framesSinceLastPrint), framesSinceLastPrint)
			framesSinceLastPrint = 0
		}

        // quit after 10 seconds, for profiling purposes
		if t > startTime+10 {
			return
		}

		// start filling and sending
		fillThisSlice <- VALUES[filling]
		sendThisSlice <- VALUES[sending]

		// wait until both are ready
		<-sliceIsFilled
		<-sliceIsSent

		// swap
		filling, sending = sending, filling
	}
}
