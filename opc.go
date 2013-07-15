package main

import "fmt"
import "bitbucket.org/davidwallace/go-opc/colorutils"

func main() {
    var x float64 = 0.5;
    x = colorutils.Remap(x, 0,1, 3, 3.5)
    fmt.Printf("x = %f\n", x)
}
