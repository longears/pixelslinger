#!/bin/sh

# Turn off all the LEDs by sending one frame of black pixels.
#
# This is meant to be run on a Beaglebone Black.  If you're running it elsewhere,
# remove the echo statements which cause the Beaglebone's onboard LEDs to blink.

echo 1 > /sys/class/leds/beaglebone\:green\:usr0/brightness
echo 1 > /sys/class/leds/beaglebone\:green\:usr1/brightness
echo 1 > /sys/class/leds/beaglebone\:green\:usr2/brightness
echo 1 > /sys/class/leds/beaglebone\:green\:usr3/brightness

# wall.json is the pattern with the most LEDs so let's just use that one all the time.
killall pixelslinger
./pixelslinger --layout layouts/wall.json --source off --dest spi --once

echo 0 > /sys/class/leds/beaglebone\:green\:usr0/brightness
echo 0 > /sys/class/leds/beaglebone\:green\:usr1/brightness
echo 0 > /sys/class/leds/beaglebone\:green\:usr2/brightness
echo 0 > /sys/class/leds/beaglebone\:green\:usr3/brightness

