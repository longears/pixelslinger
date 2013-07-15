package main

import "fmt"
import "time"
import "bitbucket.org/davidwallace/go-opc/colorutils"

func main() {

	const N_PIXELS = 1000
	var array = make([]byte, N_PIXELS*3)

	var pct, r, g, b, t float64
	var last_print = float64(time.Now().UnixNano()) / 1.0e9
	var frames = 0
	for true {
		t = float64(time.Now().UnixNano()) / 1.0e9
		frames += 1
		if t > last_print+1 {
			last_print = t
			fmt.Printf("%f ms (%d fps)\n", 1000.0/float64(frames), frames)
			frames = 0
		}
		for ii := 0; ii < N_PIXELS; ii++ {
			pct = float64(ii) / N_PIXELS

			r = pct
			g = pct
			b = pct

			array[ii*3+0] = colorutils.FloatToByte(r)
			array[ii*3+1] = colorutils.FloatToByte(g)
			array[ii*3+2] = colorutils.FloatToByte(b)
		}

		//for ii, v := range array {
		//    fmt.Printf("array[%d] = %d\n", ii, v)
		//}
	}

}
