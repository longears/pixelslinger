pixelslinger
============

Controls LED strips and handles Open Pixel Control messages.

This is a single executable, `pixelslinger`, which can receive or generate pixel values and then send them out in a variety of ways.

Pixel sources
-------------

* `--source localhost:7890` -- Run an OpenPixelControl server and listen for pixels
* `--source fire` -- Use one of the built-in animations.  See the command-line help for a full list.

Pixel destinations
------------------

* `--dest print` -- Print the pixel values to the screen for debugging
* `--dest spi` -- Directly control an LED string attached to the SPI bus
* `--dest hostname:port` -- Send Open Pixel Control messages to the given machine
* `--dest /dev/null` -- Send pixels nowhere.  Useful for benchmarking pixel sources.

Developer documentation
-----------------------

http://godoc.org/github.com/longears/pixelslinger

Usage
-----

```
Usage of ./pixelslinger:
  Available patterns:
          basic-midi
          fire
          moire
          off
          raver-plaid
          sailor-moon
          spatial-stripes
          test
          test-gamma
          test-rgb

Options:
  -l ...              --layout=...              layout file (required)
  -s spatial-stripes  --source=spatial-stripes  pixel source (either a pattern name or localhost[:port])
  -d localhost        --dest=localhost          destination (one of print, spi, /dev/null, or hostname[:port])
  -f 40               --fps=40                  max frames per second
  -n 0                --seconds=0               quit after this many seconds
  -o                  --once                    quit after one frame
                      --help                    show usage message
```
