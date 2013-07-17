package main

// Spatial Stripes
//   Creates spatial sine wave stripes: x in the red channel, y--green, z--blue
//   Also makes a white dot which moves down the strip non-spatially in the order
//   that the LEDs are indexed.

import (
	"bitbucket.org/davidwallace/go-tower/colorutils"
	"bitbucket.org/davidwallace/go-tower/opc"
	"math"
	"time"
)

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
		t := float64(time.Now().UnixNano())/1.0e9 - 1374000000
		// fill in values slice
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
			spark_ii := colorutils.PosMod2(t*80, float64(n_pixels))
			spark_rad := float64(8)
			spark_val := math.Max(0, (spark_rad-colorutils.ModDist2(float64(ii), float64(spark_ii), float64(n_pixels)))/spark_rad)
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
		sliceIsFilled <- 1
	}
}

func main() {
	layoutPath, ipPort, fps, timeToRun := opc.ParseFlags()
	opc.MainLoop(pixelThread, layoutPath, ipPort, fps, timeToRun)
}
