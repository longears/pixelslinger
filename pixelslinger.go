package main

// TODO: figure out how to handle varying numbers of pixels
// when we're getting pixels via our OPC server source

import (
	"fmt"
	"github.com/droundy/goopt"
	"github.com/longears/pixelslinger/beaglebone"
	"github.com/longears/pixelslinger/config"
	"github.com/longears/pixelslinger/midi"
	"github.com/longears/pixelslinger/opc"
	"github.com/pkg/profile"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

const ONBOARD_LED_HEARTBEAT = 0
const ONBOARD_LED_MIDI = 1

const SPI_MAGIC_WORD = "spi"
const PRINT_MAGIC_WORD = "print"
const DEVNULL_MAGIC_WORD = "/dev/null"
const LOCALHOST = "localhost"
const SPI_FN = "/dev/spidev1.0"

func init() {
	runtime.GOMAXPROCS(2)
}

// these are pointers to the actual values from the command line parser
var LAYOUT_FN = goopt.String([]string{"-l", "--layout"}, "...", "layout file (required)")
var SOURCE = goopt.String([]string{"-s", "--source"}, "spatial-stripes", "pixel source (either a pattern name or "+LOCALHOST+"[:port])")
var DEST = goopt.String([]string{"-d", "--dest"}, "localhost", "destination (one of "+PRINT_MAGIC_WORD+", "+SPI_MAGIC_WORD+", "+DEVNULL_MAGIC_WORD+", or hostname[:port])")
var FPS = goopt.Int([]string{"-f", "--fps"}, 40, "max frames per second")
var SECONDS = goopt.Int([]string{"-n", "--seconds"}, 0, "quit after this many seconds")
var ONCE = goopt.Flag([]string{"-o", "--once"}, []string{}, "quit after one frame", "")

// Parse the command line flags.  If invalid, show help and quit.
// Add default ports if needed.
// Read the layout file.
// Return the number of pixels in the layout, the source and dest thread methods.
func parseFlags() (nPixels int, sourceThread, effectThread, destThread opc.ByteThread) {

	// get sorted pattern names
	patternNames := make([]string, len(opc.PATTERN_REGISTRY))
	ii := 0
	for k, _ := range opc.PATTERN_REGISTRY {
		patternNames[ii] = k
		ii++
	}
	sort.Strings(patternNames)

	goopt.Summary = "Available source patterns:\n"
	for _, patternName := range patternNames {
		goopt.Summary += "          " + patternName + "\n"
	}
	goopt.Parse(nil)

	// layout is required
	if *LAYOUT_FN == "..." {
		fmt.Println(goopt.Usage())
		fmt.Println("--------------------------------------------------------------------------------/")
		os.Exit(1)
	}

	// read locations
	locations := opc.ReadLocations(*LAYOUT_FN)
	nPixels = len(locations) / 3

	// choose source thread method
	if strings.Contains(*SOURCE, LOCALHOST) {
		// source is localhost, so we will start an OPC server.
		// add default port if needed
		if !strings.Contains(*SOURCE, ":") {
			*SOURCE += ":7890"
		}
		sourceThread = opc.MakeOpcServerThread(*SOURCE)
	} else if (*SOURCE)[0] == ':' {
		// source is ":4908"
		*SOURCE = "localhost" + *SOURCE
		sourceThread = opc.MakeOpcServerThread(*SOURCE)
	} else {
		// source is a pattern name
		sourceThreadMaker, ok := opc.PATTERN_REGISTRY[*SOURCE]
		if !ok {
			fmt.Printf("Error: unknown source or pattern \"%s\"\n", *SOURCE)
			fmt.Println("--------------------------------------------------------------------------------/")
			os.Exit(1)
		}
		sourceThread = sourceThreadMaker(locations)
	}

	// choose effect thread method
	effectThread = opc.MakeEffectFader(locations)

	// choose dest thread method
	switch *DEST {
	case DEVNULL_MAGIC_WORD:
		destThread = opc.MakeSendToDevNullThread()
	case PRINT_MAGIC_WORD:
		destThread = opc.MakeSendToScreenThread()
	case SPI_MAGIC_WORD:
		destThread = opc.MakeSendToLPD8806Thread(SPI_FN)
	default:
		// add default port if needed
		if !strings.Contains(*DEST, ":") {
			*DEST += ":7890"
		}
		destThread = opc.MakeSendToOpcThread(*DEST)
	}

	return // returns nPixels, sourceThread, destThread
}

// Launch the sourceThread and destThread methods and coordinate the transfer of bytes from one to the other.
// Run until timeToRun seconds have passed and return.  If timeToRun is 0, run forever.
// Turn on the CPU profiler if timeToRun seconds > 0.
// Limit the framerate to a max of fps unless fps is 0.
func mainLoop(nPixels int, sourceThread, effectThread, destThread opc.ByteThread, fps float64, timeToRun float64) {
	if timeToRun > 0 {
		fmt.Printf("[mainLoop] Running for %f seconds with profiling turned on, pixels and network\n", timeToRun)
		defer profile.Start(profile.CPUProfile).Stop()
	} else {
		fmt.Println("[mainLoop] Running forever")
	}

	// prepare the byte slices and channels that connect the source and dest threads
	fillingSlice := make([]byte, nPixels*3)
	sendingSlice := make([]byte, nPixels*3)

	bytesToFillChan := make(chan []byte, 0)
	toEffectChan := make(chan []byte, 0)
	bytesFilledChan := make(chan []byte, 0)
	bytesToSendChan := make(chan []byte, 0)
	bytesSentChan := make(chan []byte, 0)

	// set up midi
	midiMessageChan := midi.GetMidiMessageStream("/dev/midi1") // this launches the midi thread
	midiState := midi.MidiState{}
	// set initial values for controller knobs
	//  (because the midi hardware only sends us values when the knobs move)
	for knob, defaultVal := range config.DEFAULT_KNOB_VALUES {
		midiState.ControllerValues[knob] = defaultVal
	}

	// launch the threads
	go sourceThread(bytesToFillChan, toEffectChan, &midiState)
	go effectThread(toEffectChan, bytesFilledChan, &midiState)
	go destThread(bytesToSendChan, bytesSentChan, &midiState)

	// main loop
	frame_budget_ms := 1000.0 / fps
	startTime := float64(time.Now().UnixNano()) / 1.0e9
	lastPrintTime := startTime
	frameStartTime := startTime
	frameEndTime := startTime
	framesSinceLastPrint := 0
	firstIteration := true
	flipper := 0
	beaglebone.SetOnboardLED(0, 1)
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
		// print framerate occasionally
		frameStartTime = float64(time.Now().UnixNano()) / 1.0e9
		framesSinceLastPrint += 1
		if frameStartTime > lastPrintTime+1 {
			lastPrintTime = frameStartTime
			fmt.Printf("[mainLoop] %f ms/frame (%d fps)\n", 1000.0/float64(framesSinceLastPrint), framesSinceLastPrint)
			framesSinceLastPrint = 0
			// toggle LED
			beaglebone.SetOnboardLED(ONBOARD_LED_HEARTBEAT, flipper)
			flipper = 1 - flipper
		}

		// if profiling, quit after a while
		if timeToRun > 0 && frameStartTime > startTime+timeToRun {
			return
		}

		// get midi
		midiState.UpdateStateFromChannel(midiMessageChan)
		if len(midiState.RecentMidiMessages) > 0 {
			beaglebone.SetOnboardLED(ONBOARD_LED_MIDI, 1)
		} else {
			beaglebone.SetOnboardLED(ONBOARD_LED_MIDI, 0)
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
		//  the double buffering effect of the two parallel threads
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

	nPixels, sourceThread, effectThread, destThread := parseFlags()
	mainLoop(nPixels, sourceThread, effectThread, destThread, float64(*FPS), float64(*SECONDS))
}
