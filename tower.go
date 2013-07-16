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
	"time"
)

// times in ms
const PIXEL_SLEEP_PER_FRAME = 1

func saveToSlice(slice []byte, ii int, r, g, b float64) {
	slice[ii*3+0] = colorutils.FloatToByte(r)
	slice[ii*3+1] = colorutils.FloatToByte(g)
	slice[ii*3+2] = colorutils.FloatToByte(b)
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

			saveToSlice(values, ii, r, g, b)
		}

		// done
		time.Sleep(PIXEL_SLEEP_PER_FRAME * time.Millisecond)
		sliceIsFilled <- 1
	}
}

func main() {
	defer profile.Start(profile.CPUProfile).Stop()

	layoutPath := "layouts/freespace.json"
	ipPort := "127.0.0.1:7890"
	//ipPort := "192.168.11.11:7890"

	opc.MainLoop(layoutPath, ipPort, pixelThread, -1)
}
