package main

import (
	"bitbucket.org/davidwallace/go-metal/opc"
	"fmt"
	"github.com/davecheney/profile"
	"os"
	"strconv"
	"strings"
	"time"
)

const SPI_MAGIC_WORD = "SPI"
const DEVNULL_MAGIC_WORD = "/dev/null"
const SPI_FN = "/dev/spidev1.0"

// Display command-line help, then quit.
func helpAndQuit() {
	fmt.Println("--------------------------------------------------------------------------------\\")
	fmt.Println("")
	fmt.Println("Usage:  program-name  <layout.json>  [destination  [fps  [seconds-to-run]]]")
	fmt.Println("")
	fmt.Println("    layout.json       A layout json file")
	fmt.Println("    destination       Where to send the pixels.  Use one of the following values:")
	fmt.Println("                         SPI               send directly out the SPI port")
	fmt.Println("                         /dev/null         send nowhere")
	fmt.Println("                         ip[:port]         send as OPC messages over the network.")
	fmt.Println("                         hostname[:port]     (port defaults to 7890)")
	fmt.Println("    fps               Max frames per second.  Use 0 or -1 for no limit.")
	fmt.Println("    seconds-to-run    Quit after this many seconds.  Defaults to 0, meaning forever.")
	fmt.Println("                        If nonzero, the profiler will be turned on.")
	fmt.Println("")
	fmt.Println("--------------------------------------------------------------------------------/")
	os.Exit(0)
}

// Parse and return the command-line flags.
// Add the default port to ipPort if needed.
func parseFlags() (layoutPath, ipPort string, fps float64, timeToRun float64) {
	layoutPath = "layouts/freespace.json"
	ipPort = "127.0.0.1:7890"
	fps = 40
	timeToRun = 0
	var err error

	if len(os.Args) >= 2 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			helpAndQuit()
		}
		layoutPath = os.Args[1]
	}
	if len(os.Args) >= 3 {
		ipPort = os.Args[2]
		if ipPort != SPI_MAGIC_WORD && ipPort != DEVNULL_MAGIC_WORD && !strings.Contains(ipPort, ":") {
			ipPort += ":7890"
		}
	}
	if len(os.Args) >= 4 {
		fps, err = strconv.ParseFloat(os.Args[3], 64)
		if err != nil {
			helpAndQuit()
		}
	}
	if len(os.Args) >= 5 {
		timeToRun, err = strconv.ParseFloat(os.Args[4], 64)
		if err != nil {
			helpAndQuit()
		}
	}
	if len(os.Args) >= 6 || len(os.Args) <= 1 {
		helpAndQuit()
	}
	return
}

// Launch the given pixelThread and suck pixels out of it.
// Also launch the SendingToOpcThread and feed the pixels to it.
// Run until timeToRun seconds have passed.
// Maintain the framerate.
// destination should be an ip:port or SPI_MAGIC_WORD or DEVNULL_MAGIC_WORD.
// Set timeToRun <= 0 to run forever.
// Set fps to the number of frames per second you want, or <= 0 for unlimited.
func mainLoop(locations []float64, pixelThread func(chan []byte, chan []byte, []float64), layoutPath, destination string, fps float64, timeToRun float64) {
	if timeToRun > 0 {
		fmt.Printf("[opc.mainLoop] Running for %f seconds with profiling turned on, pixels and network\n", timeToRun)
		defer profile.Start(profile.CPUProfile).Stop()
	} else {
		fmt.Println("[opc.mainLoop] Running forever")
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
	firstTime := true
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
			fmt.Printf("[opc.mainLoop] %f ms/frame (%d fps)\n", 1000.0/float64(framesSinceLastPrint), framesSinceLastPrint)
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
		if !firstTime {
			bytesToSendChan <- sendingSlice
		}

		// wait until both are done
		<-bytesFilledChan
		if !firstTime {
			<-bytesSentChan
		}

		// swap the slices
		sendingSlice, fillingSlice = fillingSlice, sendingSlice

		firstTime = false
	}
}

func main() {
	layoutPath, ipPort, fps, timeToRun := parseFlags()

	fmt.Println("--------------------------------------------------------------------------------\\")
	defer fmt.Println("--------------------------------------------------------------------------------/")

	locations := opc.ReadLocations(layoutPath)
	mainLoop(locations, opc.PatternRaverPlaid, layoutPath, ipPort, fps, timeToRun)
}
