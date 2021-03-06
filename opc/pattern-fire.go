package opc

// Fire
//   Make a burning fire pattern.
//   This pattern is scaled to fit the layout from top to bottom (z).

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/config"
	"github.com/longears/pixelslinger/midi"
    "math"
	"time"
)

// this is used to cache some per-pixel calculations
type firePixelInfo struct {
    xp float64
    yp float64
    zp float64
    vgrad float64
}

func MakePatternFire(locations []float64) ByteThread {

    const (
        SPEED      = 0.83 // How quick are the flames?  This is applied in addition to the speed knob.
        SIDE_SCALE = 1.7  // Horizontal scale (x and y).  Smaller numbers compress things horizontally.
    )

    // get bounding box
    n_pixels := len(locations) / 3
    var max_coord_x, max_coord_y, max_coord_z float64
    var min_coord_x, min_coord_y, min_coord_z float64
    for ii := 0; ii < n_pixels; ii++ {
        x := locations[ii*3+0]
        y := locations[ii*3+1]
        z := locations[ii*3+2]
        if ii == 0 || x > max_coord_x { max_coord_x = x }
        if ii == 0 || y > max_coord_y { max_coord_y = y }
        if ii == 0 || z > max_coord_z { max_coord_z = z }
        if ii == 0 || x < min_coord_x { min_coord_x = x }
        if ii == 0 || y < min_coord_y { min_coord_y = y }
        if ii == 0 || z < min_coord_z { min_coord_z = z }
    }

    // make array of firePixelInfo structs
    // and fill the cache of per-pixel calculations
    pixelInfoCache := make([]*firePixelInfo, len(locations)/3)
    for ii := range pixelInfoCache {
        thisPixelInfo := &firePixelInfo{}
        pixelInfoCache[ii] = thisPixelInfo

        x := locations[ii*3+0]
        y := locations[ii*3+1]
        z := locations[ii*3+2]

        // scale the height (z) of the layout to fit in the range 0-1
        // and scale x and y accordingly
        z_scale := max_coord_z - min_coord_z
        if z_scale == 0 { // avoid divide by zero
            z_scale = 0.05
        }
        xp := x / z_scale / SIDE_SCALE
        yp := y / z_scale / SIDE_SCALE
        zp := (z-min_coord_z) / z_scale

        // bend space so that things seem to accelerate upwards
        zp = math.Pow(zp + 0.05, 0.7)

        // make basic vertical gradient
        vgrad := colorutils.Cos2(colorutils.Clamp(zp, 0, 1), 0, 2, 0, 1)
        // vgrad := 1 - colorutils.Clamp(zp, 0, 1)

        // save to cache
        thisPixelInfo.xp = xp
        thisPixelInfo.yp = yp
        thisPixelInfo.zp = zp
        thisPixelInfo.vgrad = vgrad
    }

	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
        last_t := 0.0
        t := 0.0
		for bytes := range bytesIn {

            var (
                // hue knob controls hue
                H = 0.05 + float64(midiState.ControllerValues[config.HUE_KNOB]) / 127.0
                S = 0.9
                V = 0.65
                OVERBRIGHT = 1.3
            )

            // fire color
            rFire, gFire, bFire := colorutils.HslToRgb(H, S, V)
            rFire *= OVERBRIGHT
            gFire *= OVERBRIGHT
            bFire *= OVERBRIGHT

			n_pixels := len(bytes) / 3

            // time and speed knob bookkeeping
			this_t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8
			speedKnob := float64(midiState.ControllerValues[config.SPEED_KNOB]) / 127.0
            if speedKnob < 0.5 {
                speedKnob = colorutils.RemapAndClamp(speedKnob, 0, 0.4, 0, 1)
            } else {
                speedKnob = colorutils.RemapAndClamp(speedKnob, 0.6, 1, 1, 4)
            }
			if midiState.KeyVolumes[config.SLOWMO_PAD] > 0 {
                speedKnob *=  0.25
            }
            if last_t != 0 {
                t += (this_t - last_t) * speedKnob * SPEED
            }
            last_t = this_t

			// fill in bytes array
			var r, g, b float64
			for ii := 0; ii < n_pixels; ii++ {
                //--------------------------------------------------------------------------------

                pi := pixelInfoCache[ii]

                // apply various wiggles to coordinate space
                // offset, period, min, max
                zp1 := (  pi.zp + colorutils.Cos2(pi.xp,  t*0.33 + 8.63, 0.15 * 1.7, 0, 1) * 0.2 +
                                  colorutils.Cos2(pi.xp, -t*0.23 + 2.43, 0.34 * 1.7, 0, 1) * 0.3  )
                zp3 := (  pi.zp + colorutils.Cos2(pi.xp, -t*0.42 + 5.62, 0.27 * 1.7, 0, 1) * 0.2 +
                                  colorutils.Cos2(pi.xp,  t*0.20 + 3.07, 0.55 * 1.7, 0, 1) * 0.3  )
                zp4 := (  pi.zp + colorutils.Cos2(pi.xp,  t*0.36 + 4.81, 0.20 * 1.7, 0, 1) * 0.2 +
                                  colorutils.Cos2(pi.xp, -t*0.26 + 7.94, 0.67 * 1.7, 0, 1) * 0.3  )

                // smallest fastest noise
                noise_lit := (  colorutils.Cos2(pi.xp,  -4.37 * t/4, 0.21, 0, 1) +
                                colorutils.Cos2(pi.yp,   4.37 * t/4, 0.21, 0, 1) +
                                colorutils.Cos2(zp1,     4.37 * t,   0.21, 0, 1)  ) / 3

                // small fast noise
                noise_med := (  colorutils.Cos2(pi.xp,  -3 * t/4, 0.3, 0, 1) +
                                colorutils.Cos2(pi.yp,   3 * t/4, 0.3, 0, 1) +
                                colorutils.Cos2(zp3,     3 * t,   0.3, 0, 1)  ) / 3

                // big slow noise
                noise_big := (  colorutils.Cos2(pi.xp,  -0.9 * t/2, 0.8, 0, 1) +
                                colorutils.Cos2(pi.yp,   0.9 * t/2, 0.8, 0, 1) +
                                colorutils.Cos2(zp4,     0.9 * t,   0.8, 0, 1)  ) / 3

                // combine vgradient with noise
                v := (  pi.vgrad  +
                        colorutils.Remap(noise_lit, 0,1, -1,1)*0.17 +
                        colorutils.Remap(noise_med, 0,1, -1,1)*0.20 +
                        colorutils.Remap(noise_big, 0,1, -1,1)*0.80  )

                // apply sine contrast curve
                //v = colorutils.Cos2( colorutils.Clamp(v,0,1), 0, 2, 1, 0 )

                // color map
                r = v * rFire
                g = v * gFire
                b = v * bFire

                r,g,b = colorutils.ContrastRgb(r,g,b, 0.7, 1.1)
                // r,g,b = colorutils.RGBClipBlackByLuminance(r,g,b, 0.2)  // TODO

				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
