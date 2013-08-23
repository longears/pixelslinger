package opc

import (
	"fmt"
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	"image"
	_ "image/color"
	_ "image/png"
	"math"
	"os"
	"time"
)

func handleErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
}

//================================================================================

type MyColor struct {
	r float64
	g float64
	b float64
}

type MyImage struct {
	xres   int
	yres   int
	pixels [][]*MyColor // 2d array, [x][y]
}

// Init the MyImage pixel array, creating MyColor objects
// from the data in the given image (from the built-in image package).
// HSV is computed here also for each pixel.
func (mi *MyImage) populateFromImage(imgFn string) {
	// read and decode image
	file, err := os.Open(imgFn)
	handleErr(err)
	defer file.Close()
	img, _, err := image.Decode(file)
	handleErr(err)

	// copy and convert pixels
	mi.xres = img.Bounds().Max.X
	mi.yres = img.Bounds().Max.Y
	mi.pixels = make([][]*MyColor, mi.xres)
	for x := 0; x < mi.xres; x++ {
		mi.pixels[x] = make([]*MyColor, mi.yres)
		for y := 0; y < mi.yres; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
			c := &MyColor{float64(r) / 256 / 256, float64(g) / 256 / 256, float64(b) / 256 / 256}
			mi.pixels[x][y] = c
		}
	}
}

func (mi *MyImage) String() string {
	return fmt.Sprintf("<image %v x %v>", mi.xres, mi.yres)
}

// given x and y as floats between 0 and 1,
// return r,g,b as floats between 0 and 1
func (mi *MyImage) getInterpolatedColor(x, y float64, wrapMethod string) (r, g, b float64) {

	switch wrapMethod {
	case "tile":
		// keep x and y between 0 and 1
		_, x = math.Modf(x)
		if x < 0 {
			x += 1
		}
		_, y = math.Modf(y)
		if y < 0 {
			y += 1
		}
	case "extend":
		x = colorutils.Clamp(x, 0, 1)
		y = colorutils.Clamp(y, 0, 1)
	case "mirror":
		x = colorutils.PosMod(x, 2)
		if x > 1 {
			x = 2 - x
		}
		y = colorutils.PosMod(y, 2)
		if y > 1 {
			y = 2 - y
		}
	}

	// float pixel coords
	xp := x * float64(mi.xres-1) * 0.999999
	yp := y * float64(mi.yres-1) * 0.999999

	// integer pixel coords
	x0 := int(xp)
	x1 := x0 + 1
	y0 := int(yp)
	y1 := y0 + 1

	// subpixel fractional coords for interpolation
	_, xPct := math.Modf(xp)
	_, yPct := math.Modf(yp)

	// retrieve colors from image array
	c00 := mi.pixels[x0][y0]
	c10 := mi.pixels[x1][y0]
	c01 := mi.pixels[x0][y1]
	c11 := mi.pixels[x1][y1]

	// interpolate
	r = (c00.r*(1-xPct)+c10.r*xPct)*(1-yPct) + (c01.r*(1-xPct)+c11.r*xPct)*yPct
	g = (c00.g*(1-xPct)+c10.g*xPct)*(1-yPct) + (c01.g*(1-xPct)+c11.g*xPct)*yPct
	b = (c00.b*(1-xPct)+c10.b*xPct)*(1-yPct) + (c01.b*(1-xPct)+c11.b*xPct)*yPct

	return r, g, b
}

//================================================================================

func MakePatternSunset(locations []float64) ByteThread {

	var (
		IMG_PATH = "images/sky3_square.png"
		//IMG_PATH = "images/r.png"
	)

	myImage := &MyImage{}
	myImage.populateFromImage(IMG_PATH)

	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8
			_ = t

			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------

				x := locations[ii*3+0] / 2
				y := locations[ii*3+1] / 2
				z := locations[ii*3+2] / 2
				_ = x
				_ = y
				_ = z

				r, g, b := myImage.getInterpolatedColor(-x, -z, "tile")

				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}

			bytesOut <- bytes
		}
	}
}
