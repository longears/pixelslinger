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
// TYPES

// Most of this library is built from ByteThread functions which you can
// string together via channels.
// These are meant to be run as goroutines.  They read in one
// byte slice at a time through the input channel, do something to it,
// then return it over the output channel when done.
// These are used both for sources and destinations of pixel data.
// They should loop forever until the input channel is closed, then return.
// The byte slice should hold values from 0 to 255 in [r g b  r g b  r g b  ... ] order
// so its total length is 3 times the number of pixels in the LED strip.
type ByteThread func(chan []byte, chan []byte)

//--------------------------------------------------------------------------------
// CONSTANTS

const SPI_CHUNK_SIZE = 2048

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

// Returns a ByteThread which passes byte slices from the input to
// the output channels without doing anything.
func MakeSendToDevNullThread() ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte) {
		fmt.Println("[opc.SendToDevNullThread] starting up")
		for bytes := range bytesIn {
			bytesOut <- bytes
		}
	}
}

// Return a ByteThread which prints the bytes to the screen.
func MakeSendToScreenThread() ByteThread {
	const MAX_LEN = 19
	result := make([]string, MAX_LEN)
	return func(bytesIn chan []byte, bytesOut chan []byte) {
		fmt.Println("[opc.SendToDevNullThread] starting up")
		for bytes := range bytesIn {
			for ii := 0; ii < len(bytes) && ii < MAX_LEN; ii++ {
				if ii%4 == 3 {
					result[ii] = "|"
					ii += 1
				}
				result[ii] = fmt.Sprintf("%3d", bytes[ii])
			}
			fmt.Printf("[ %s ...] %v px \n", strings.Join(result, " "), len(bytes)/3)
			bytesOut <- bytes
		}
	}
}

// Return a ByteThread which writes bytes to SPI via the given filename (probably "/dev/spidev1.0").
// Format the outgoing bytes for LED strips which use the LPD8806 chipset.
// If the SPI device can't be opened, exit the whole program with exit status 1.
func MakeSendToLPD8806Thread(spiFn string) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte) {
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
			// build a new slice of bytes in the format the LED strand wants
			// TODO: avoid allocating these bytes over and over
			spiBytes := make([]byte, 0)

			// leading zeros to begin a new frame of bytes
			numZeroes := (len(bytes) + 31)/32 + 1;
			for ii := 0; ii < numZeroes*5; ii++ {
				spiBytes = append(spiBytes, 0)
			}

			// bytes
			for _, v := range bytes {
				// high bit must be always on, remaining seven bits are data
				v2 := 128 | (v >> 1)
				spiBytes = append(spiBytes, v2)
			}

			// final zero to latch the last pixel
			spiBytes = append(spiBytes, 0)

			// write spiBytes to the wire in chunks
            //fmt.Println("sending", len(bytes), " + ", numZeroes, " zeroes = ", len(spiBytes), "bytes")
            bytesSent := 0
            for ii := 0; ii < len(spiBytes); ii += SPI_CHUNK_SIZE {
                endIndex := ii + SPI_CHUNK_SIZE
                if endIndex > len(spiBytes) {
                    endIndex = len(spiBytes)
                }
                thisChunk := spiBytes[ii:endIndex]
                bytesSent += len(thisChunk)
                if _, err := spiFile.Write(thisChunk); err != nil {
                    panic(err)
                }
            }
            //fmt.Println(bytesSent,len(spiBytes))

			bytesOut <- bytes
		}
	}
}

// Return a ByteThread which sends the bytes out as OPC messages to the given ipPort.
// Initiates and maintains a long-lived connection to ipPort.  If the connection is bad at any point
// (or was never good to begin with), keep trying to reconnect whenever new bytes come in.
// Sometimes sleeps during reconnection attempts; this blocks the input channel.
// Silently drop bytes if it's not possible to send them.
// Creates OPC headers for each byte slice it sends.
func MakeSendToOpcThread(ipPort string) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte) {
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
}
