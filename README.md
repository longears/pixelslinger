pixelslinger
============

Controls LED strips and handles Open Pixel Control messages.

This is a single executable, `pixelslinger`, which can receive or generate pixel values and then send them out in a variety of ways.


Install & compile
-----------------

1. [Install Go from golang.org](http://golang.org/doc/install)

1. Make a directory where you'll keep your Go code.  I used `~/gocode`.

1. In your `.bash_profile` or wherever, add this:

 `export GOPATH=/path/to/your/gocode`

 That tells Go where your code will live.
 
1. Fetch code and dependencies:

 ```
 go get github.com/longears/pixelslinger
 go get github.com/pkg/profile
 go get github.com/droundy/goopt
 ```
 
 If you receive errors using `go get [repo-url]`, a common solution is `go get -u -v [repo-url]`.

1. Compile / run

 ```
 cd $GOPATH/src/github.com/longears/pixelslinger
  
 go run pixelslinger.go     // compile and run in one step
 
 go build pixelslinger.go   // or, just compile
 ./pixelslinger             // and then run after compiling
 ```


Using with the OpenPixelControl simulator
------------------------------------------

The [main OpenPixelControl repo](https://github.com/zestyping/openpixelcontrol/) comes with an OpenGL
simulator which shows your animation in 3d.

The point of OpenPixelControl is to allow pixel generation and pixel display to be separated into
different programs, possibly running on different machines connected over the network.

Pixelslinger can act as both a pixel source and a pixel display.  In this case we want it to be a
source and the OpenGL simulator will be the display.

1. Download and compile https://github.com/zestyping/openpixelcontrol/

1. Run the simulator which will listen on port 7890 by default

 `openpixelcontrol$ bin/gl_server layouts/freespace.json`

1. In another shell, run pixelslinger to send pixels to the simulator

 ```
 pixelslinger$ ./pixelslinger --layout layouts/freespace.json --source fire --dest localhost:7890
 ```


Pixel sources
-------------

* `--source localhost:7890` -- Run an OpenPixelControl server and listen for pixels from the network
* `--source fire` -- Use one of the built-in animations.  See the command-line help for a full list.


Pixel destinations
------------------

* `--dest print` -- Print the pixel values to the screen for debugging
* `--dest spi` -- Directly control an LED string attached to the SPI bus on a Beaglebone Black
* `--dest hostname:port` -- Send Open Pixel Control messages over the network to the given machine
* `--dest /dev/null` -- Send pixels nowhere.  Useful for benchmarking the framerate of pixel sources.


Adding your own animation patterns
----------------------------------

1. Start by copying and renaming `opc/pattern-raver-plaid.go`.  Modify it however you want.
1. Add your pattern to the `PATTERN_REGISTRY` map in `opc/opc.go` so you can choose it from the command line.
1. There is a built-in pattern, `midi-switcher`, which uses a MIDI knob to switch between other patterns.  You may want to add your new pattern to its `PATTERN_LIST` in `opc/pattern-midi-switcher.go`.


Adding your own layout files
----------------------------

Most of the patterns need to know how the pixels are arranged in space.  OpenPixelControl defines a standard
way of encoding this information into a JSON file:

```
[ 
  {"point": [0.0000, 1.0000, 0.1000]},
  {"point": [0.0393, 0.9992, 0.0000]},
  {"point": [0.0785, 0.9969, 0.0000]}
]
```

Unfortunately this isn't as flexible as real JSON because it has to be arranged exactly as shown above, with
no extra spaces or newlines anywhere.  Also note the lack of a comma on the last line.  If you're generating
your own layouts you should build them as strings instead of trying to use a JSON library, or they might not
match this format pefectly.

Here are [a bunch of Python scripts](https://github.com/longears/openpixelcontrol/tree/metal_tower_2/layouts)
for generating layout files.  For example, there's one called `objToLayout.py` which converts OBJ files to
layout files.


Developer documentation
-----------------------

http://godoc.org/github.com/longears/pixelslinger


Command Line Usage
------------------

```
Available source patterns:
          basic-midi
          fire
          midi-switcher
          moire
          off
          raver-plaid
          sailor-moon
          shield
          spatial-stripes
          test
          test-gamma
          test-rgb
          white

Options:
  -l ...              --layout=...              layout file (required)
  -s spatial-stripes  --source=spatial-stripes  pixel source (either a pattern name or localhost[:port])
  -d localhost        --dest=localhost          destination (one of print, spi, /dev/null, or hostname[:port])
  -f 40               --fps=40                  max frames per second
  -n 0                --seconds=0               quit after this many seconds
  -o                  --once                    quit after one frame
                      --help                    show usage message
```
