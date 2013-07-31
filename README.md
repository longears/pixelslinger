pixelslinger
============

Controls LED strips and handles Open Pixel Control messages.

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
  -d localhost        --dest=localhost          destination (one of screen, spi, /dev/null, or hostname[:port])
  -f 40               --fps=40                  max frames per second
  -n 0                --seconds=0               quit after this many seconds
  -o                  --once                    quit after one frame
                      --help                    show usage message
```
