package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"github.com/davecheney/profile"
	"bitbucket.org/davidwallace/go-opc/colorutils"
)

// read locations from JSON file into a slice of floats
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
	fmt.Printf("Read %v pixel locations from %s\n", len(locations), fn)
	return locations
}

func main() {
	defer profile.Start(profile.CPUProfile).Stop()

	path := "circle.json"

	LOCATIONS := readLocations(path)
	N_PIXELS := len(LOCATIONS) / 3
	VALUES := make([]byte, N_PIXELS*3)

	// fill in values over and over
	var pct, r, g, b, t float64
	var last_print = float64(time.Now().UnixNano()) / 1.0e9
	var frames = 0
	var start_time = last_print
	t = start_time
	for t < start_time+5 {
		t = float64(time.Now().UnixNano()) / 1.0e9
		frames += 1
		if t > last_print+1 {
			last_print = t
			fmt.Printf("%f ms (%d fps)\n", 1000.0/float64(frames), frames)
			frames = 0
		}
		for ii := 0; ii < N_PIXELS; ii++ {
			pct = float64(ii) / float64(N_PIXELS)

			r = pct
			g = pct
			b = pct

			VALUES[ii*3+0] = colorutils.FloatToByte(r)
			VALUES[ii*3+1] = colorutils.FloatToByte(g)
			VALUES[ii*3+2] = colorutils.FloatToByte(b)
		}

		//for ii, v := range VALUES {
		//    fmt.Printf("VALUES[%d] = %d\n", ii, v)
		//}
	}

}
