package main

import "fmt"
//import "bitbucket.org/davidwallace/go-opc/colorutils"

func main() {

    const N_PIXELS = 10

    var array = make([]int, N_PIXELS*3)

    var pct, r, g, b float64
    for ii := 0; ii < N_PIXELS; ii++ {
        pct = float64(ii) / N_PIXELS
        r = pct
        g = pct
        b = pct
        array[ii*3 + 0] = int(r*255)
        array[ii*3 + 1] = int(g*255)
        array[ii*3 + 2] = int(b*255)
    }
    for ii,v := range array {
        fmt.Printf("array[%d] = %d\n", ii, v)
    }

}
