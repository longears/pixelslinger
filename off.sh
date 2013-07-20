#!/bin/sh

# wall.json is the pattern with the most LEDs so we can just always use that one
./main --layout layouts/wall.json --source off --dest spi --once

