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
 go get github.com/davecheney/profile
 go get github.com/droundy/goopt
 ```

1. Compile / run

 ```
 cd $GOPATH/src/github.com/longears/pixelslinger
  
 go run pixelslinger.go     // compile and run in one step
 
 go build pixelslinger.go   // or, just compile
 ./pixelslinger             // and then run after compiling
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

To add your own animation patterns
----------------------------------

1. Start by copying and renaming `opc/pattern-raver-plaid.go`.  Modify it however you want.
1. Add your pattern to the `PATTERN_REGISTRY` map in `opc/opc.go` so you can choose it from the command line.
1. There is a built-in pattern, `midi-switcher`, which uses a MIDI knob to switch between other patterns.  You may want to add your new pattern to its `PATTERN_LIST` in `opc/pattern-midi-switcher.go`.

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
