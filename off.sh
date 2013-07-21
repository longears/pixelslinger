#!/bin/sh

echo 1 > /sys/class/leds/beaglebone\:green\:usr0/brightness
echo 1 > /sys/class/leds/beaglebone\:green\:usr1/brightness
echo 1 > /sys/class/leds/beaglebone\:green\:usr2/brightness
echo 1 > /sys/class/leds/beaglebone\:green\:usr3/brightness

# wall.json is the pattern with the most LEDs so we can just always use that one
./main --layout layouts/wall.json --source off --dest spi --once

echo 0 > /sys/class/leds/beaglebone\:green\:usr0/brightness
echo 0 > /sys/class/leds/beaglebone\:green\:usr1/brightness
echo 0 > /sys/class/leds/beaglebone\:green\:usr2/brightness
echo 0 > /sys/class/leds/beaglebone\:green\:usr3/brightness

