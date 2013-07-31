/*
Package beaglebone lets you control the onboard LEDs on a Beaglebone Black.
*/
package beaglebone

import (
	"fmt"
	"os"
)

// Set one of the on-board LEDs on the Beaglebone.
// Fail silently if something doesn't work.
//    ledNum: between 0 and 3 inclusive.  LED 0 is farthest from the Ethernet port.
//    val: 0 or 1.
func SetOnboardLED(ledNum int, val int) {
	ledFn := fmt.Sprintf("/sys/class/leds/beaglebone:green:usr%d/brightness", ledNum)

	// open file
	ledFile, err := os.Create(ledFn)
	if err != nil {
		return
	}

	// close file on exit
	defer func() {
		if err := ledFile.Close(); err != nil {
			return
		}
	}()

	// convert int to string
	// faster to just use an if statement instead of strconv.Itoa()
	var s string
	if val <= 0 {
		s = "0"
	} else {
		s = "1"
	}

	// write to file
	if _, err := ledFile.WriteString(s); err != nil {
		return
	}
}
