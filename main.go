package main

import (
	"bitbucket.org/davidwallace/go-metal/opc"
	"fmt"
	"github.com/davecheney/profile"
	"github.com/droundy/goopt"
	"os"
	"strings"
	"time"
)

const SPI_MAGIC_WORD = "SPI"
const DEVNULL_MAGIC_WORD = "/dev/null"
const SPI_FN = "/dev/spidev1.0"

// these are pointers to the actual values from the command line parser
var ONCE = goopt.Flag([]string{"-o", "--once"}, []string{}, "quit after one frame", "")
var LAYOUT_FN = goopt.String([]string{"-l", "--layout"}, "...", "layout file (required)")
var DEST = goopt.String([]string{"-d", "--dest"}, "localhost", "destination (either SPI, /dev/null, hostname, or hostname:port)")
var FPS = goopt.Int([]string{"-f", "--fps"}, 40, "max frames per second")
var SECONDS = goopt.Int([]string{"-s", "--seconds"}, 0, "quit after this many seconds")
var PATTERN_NAME = goopt.String([]string{"-p", "--pattern"}, "spatial-stripes", "pattern to show")

// the pattern function we'll be running
var PATTERN_FUNC func(bytesIn chan []byte, bytesOut chan []byte, locations []float64)

// Parse the command line flags.  If invalid, show help and quit.
// Add the default port to ipPort if needed.
// Convert the PATTERN_NAME string to a function and store it in PATTERN_FUNC.
func parseFlags() {
	goopt.Summary = "Available patterns:\n"
	goopt.Summary += "          off \n"
	goopt.Summary += "          raver-plaid \n"
	goopt.Summary += "          spatial-stripes \n"
	goopt.Parse(nil)

	// layout is required
	if *LAYOUT_FN == "..." {
		fmt.Println(goopt.Usage())
		fmt.Println("--------------------------------------------------------------------------------/")
		os.Exit(1)
	}

	// add default port if needed
	if *DEST != SPI_MAGIC_WORD && *DEST != DEVNULL_MAGIC_WORD && !strings.Contains(*DEST, ":") {
		*DEST += ":7890"
	}

	switch *PATTERN_NAME {
	case "off":
		PATTERN_FUNC = opc.PatternOff
	case "raver-plaid":
		PATTERN_FUNC = opc.PatternRaverPlaid
	case "spatial-stripes":
		PATTERN_FUNC = opc.PatternSpatialStripes
	default:
		fmt.Printf("Error: unknown pattern \"%s\"\n", *PATTERN_NAME)
		fmt.Println("--------------------------------------------------------------------------------/")
		os.Exit(1)
	}

}

// Launch the given pixelThread and suck pixels out of it.
// Also launch the SendingToOpcThread and feed the pixels to it.
// Run until timeToRun seconds have passed.
// Maintain the framerate.
// destination should be an ip:port or SPI_MAGIC_WORD or DEVNULL_MAGIC_WORD.
// Set timeToRun <= 0 to run forever.
// Set fps to the number of frames per second you want, or <= 0 for unlimited.
func mainLoop(locations []float64, pixelThread func(chan []byte, chan []byte, []float64), destination string, fps float64, timeToRun float64) {
	if timeToRun > 0 {
		fmt.Printf("[mainLoop] Running for %f seconds with profiling turned on, pixels and network\n", timeToRun)
		defer profile.Start(profile.CPUProfile).Stop()
	} else {
		fmt.Println("[mainLoop] Running forever")
	}

	frame_budget_ms := 1000.0 / fps

	n_pixels := len(locations) / 3
	fillingSlice := make([]byte, n_pixels*3)
	sendingSlice := make([]byte, n_pixels*3)

	bytesToFillChan := make(chan []byte, 0)
	bytesFilledChan := make(chan []byte, 0)
	bytesToSendChan := make(chan []byte, 0)
	bytesSentChan := make(chan []byte, 0)

	// start threads
	switch destination {
	case DEVNULL_MAGIC_WORD:
		go opc.SendToDevNullThread(bytesToSendChan, bytesSentChan)
	case SPI_MAGIC_WORD:
		go opc.SendToLPD8806Thread(bytesToSendChan, bytesSentChan, SPI_FN)
	default:
		go opc.SendToOpcThread(bytesToSendChan, bytesSentChan, destination)
	}
	go pixelThread(bytesToFillChan, bytesFilledChan, locations)

	// main loop
	startTime := float64(time.Now().UnixNano()) / 1.0e9
	lastPrintTime := startTime
	frameStartTime := startTime
	frameEndTime := startTime
	framesSinceLastPrint := 0
	firstIteration := true
	for {
		// if we have any frame budget left from last time around, sleep to control the framerate
		if fps > 0 {
			frameEndTime = float64(time.Now().UnixNano()) / 1.0e9
			timeRemaining := float64(frame_budget_ms)/1000 - (frameEndTime - frameStartTime)
			if timeRemaining > 0 {
				time.Sleep(time.Duration(timeRemaining*1000*1000) * time.Microsecond)
			}
		}

		// fps reporting and bookkeeping
		frameStartTime = float64(time.Now().UnixNano()) / 1.0e9

		// print framerate occasionally
		framesSinceLastPrint += 1
		if frameStartTime > lastPrintTime+1 {
			lastPrintTime = frameStartTime
			fmt.Printf("[mainLoop] %f ms/frame (%d fps)\n", 1000.0/float64(framesSinceLastPrint), framesSinceLastPrint)
			framesSinceLastPrint = 0
		}

		// if profiling, quit after a while
		if timeToRun > 0 && frameStartTime > startTime+timeToRun {
			return
		}

		// start the threads filling and sending slices in parallel.
		// if this is the first time through the loop we have to skip
		//  the sending stage or we'll send out a whole bunch of zeros.
		bytesToFillChan <- fillingSlice
		if !firstIteration {
			bytesToSendChan <- sendingSlice
		}

		// if only sending one frame, let's just get it all over with now
		//  or we'd have to compute two frames worth of pixels because of
		//  the double buffering effect of our parallel threads
		if *ONCE {
			// get filled bytes and send them
			bytesToSendChan <- <-bytesFilledChan
			// wait for sending to complete
			<-bytesSentChan
			fmt.Println("[mainLoop] just running once.  quitting now.")
			return
		}

		// wait until both filling and sending threads are done
		<-bytesFilledChan
		if !firstIteration {
			<-bytesSentChan
		}

		// swap the slices
		sendingSlice, fillingSlice = fillingSlice, sendingSlice

		firstIteration = false
	}
}

func main() {
	fmt.Println("--------------------------------------------------------------------------------\\")
	defer fmt.Println("--------------------------------------------------------------------------------/")

	parseFlags()

	locations := opc.ReadLocations(*LAYOUT_FN)
	mainLoop(locations, PATTERN_FUNC, *DEST, float64(*FPS), float64(*SECONDS))
}
