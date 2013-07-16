package main

// TODO
//  command line parameters (layout file, ip:port)
//  figure out how to have multiple executable files in a directory
//  make lookup tables (Cos, Gamma, ...?)
//  write a pattern that relies on location

import (
	"bitbucket.org/davidwallace/go-tower/colorutils"
	"bitbucket.org/davidwallace/go-tower/opc"
	"github.com/davecheney/profile"
	"math"
    "strings"
    "strconv"
    "os"
    "fmt"
	"time"
)

const FPS float64 = 50
const FRAME_BUDGET_MS float64 = 1000.0 / FPS

func saveToSlice(slice []byte, ii int, r, g, b float64) {
	slice[ii*3+0] = colorutils.FloatToByte(r)
	slice[ii*3+1] = colorutils.FloatToByte(g)
	slice[ii*3+2] = colorutils.FloatToByte(b)
}

func pixelThread(fillThisSlice chan []byte, sliceIsFilled chan int, locations []float64) {
	for {
		// wait for slice to fill
		values := <-fillThisSlice
		n_pixels := len(values) / 3

		t := float64(time.Now().UnixNano()) / 1.0e9

		// fill in values array
		for ii := 0; ii < n_pixels; ii++ {
			//--------------------------------------------------------------------------------

			// make moving stripes for x, y, and z
			x := locations[ii*3+0]
			y := locations[ii*3+1]
			z := locations[ii*3+2]
			r := colorutils.Cos(x, t/4, 1, 0, 0.7) // offset, period, minn, max
			g := colorutils.Cos(y, t/4, 1, 0, 0.7)
			b := colorutils.Cos(z, t/4, 1, 0, 0.7)
			r, g, b = colorutils.RGBContrast(r, g, b, 0.5, 2)

			// make a moving white dot showing the order of the pixels in the layout file
			spark_ii := math.Mod(t*80+float64(n_pixels), float64(n_pixels))
			spark_rad := float64(8)
			spark_val := math.Max(0, (spark_rad-colorutils.ModDist(float64(ii), float64(spark_ii), float64(n_pixels)))/spark_rad)
			spark_val = math.Min(1, spark_val*2)
			r += spark_val
			g += spark_val
			b += spark_val

			// apply gamma curve
			// only do this on live leds, not in the simulator
			//r, g, b = colorutils.RGBGamma(r, g, b, 2.2)

			saveToSlice(values, ii, r, g, b)
			//--------------------------------------------------------------------------------
		}

        // sleep if we still have frame budget left
		t2 := float64(time.Now().UnixNano()) / 1.0e9
        timeUsedSoFar := t2-t
        timeRemaining := float64(FRAME_BUDGET_MS)/1000 - timeUsedSoFar
        if timeRemaining > 0 {
            time.Sleep( time.Duration(timeRemaining*1000*1000) * time.Microsecond)
        }

		//time.Sleep(PIXEL_SLEEP_PER_FRAME * time.Millisecond)
		sliceIsFilled <- 1
	}
}

func helpAndQuit() {
    fmt.Println("")
    fmt.Println("usage: program-name <layout.json> [ip:port [seconds-to-run]]")
    fmt.Println("")
    fmt.Println("    seconds-to-run: use 0 for forever, or negative for benchmarking")
    fmt.Println("")
    os.Exit(1)
}

func getLayoutPathAndIpPort() (layoutPath, ipPort string, seconds float64) {
	layoutPath = "layouts/freespace.json"
	ipPort = "127.0.0.1:7890"
    seconds = 0
    var err error

    if len(os.Args) >= 2 {
        if os.Args[1] == "-h" || os.Args[1] == "--help" {
            helpAndQuit()
        }
        layoutPath = os.Args[1]
    }
    if len(os.Args) >= 3 {
        ipPort = os.Args[2]
    }
    if len(os.Args) >= 4 {
        seconds, err = strconv.ParseFloat(os.Args[3],64)
            if err != nil {
                helpAndQuit()
            }
    }
    if len(os.Args) >= 5 || len(os.Args) <= 1 {
        helpAndQuit()
    }
    if ! strings.Contains(ipPort, ":") {
        ipPort += ":7890"
    }
    return
}

func main() {
    fmt.Println("--------------------------------------------------------------------------------\\")
    defer fmt.Println("--------------------------------------------------------------------------------/")

    layoutPath, ipPort, seconds := getLayoutPathAndIpPort()

    if seconds != 0 {
        if seconds > 0 {
            fmt.Printf("[main] Running for %f seconds with profiling turned on, pixels and network\n", seconds)
        } else if seconds < 0 {
            fmt.Printf("[main] Running for %f seconds with profiling turned on, pixels only\n", -seconds)
        }
        defer profile.Start(profile.CPUProfile).Stop()
    } else {
        fmt.Println("[main] Running forever")
    }

	opc.MainLoop(layoutPath, ipPort, pixelThread, seconds)
}
