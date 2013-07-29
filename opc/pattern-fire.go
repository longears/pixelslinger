package opc

// Raver plaid
//   A rainbowy pattern with moving diagonal black stripes

import (
	"bitbucket.org/davidwallace/go-metal/colorutils"
	"bitbucket.org/davidwallace/go-metal/midi"
    "math"
	"time"
)

func MakePatternFire(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8
			_ = t

            // slow down time a bit
            t *= 0.83

            // get bounding box
            var max_coord_x, max_coord_y, max_coord_z float64
            var min_coord_x, min_coord_y, min_coord_z float64
            for ii := 0; ii < n_pixels; ii++ {
                x := locations[ii*3 + 0]
                y := locations[ii*3 + 1]
                z := locations[ii*3 + 2]
                if ii == 0 || x > max_coord_x { max_coord_x = x }
                if ii == 0 || y > max_coord_y { max_coord_y = y }
                if ii == 0 || z > max_coord_z { max_coord_z = z }
                if ii == 0 || x < min_coord_x { min_coord_x = x }
                if ii == 0 || y < min_coord_y { min_coord_y = y }
                if ii == 0 || z < min_coord_z { min_coord_z = z }
            }

			// fill in bytes array
			var r, g, b float64
			for ii := 0; ii < n_pixels; ii++ {
                //--------------------------------------------------------------------------------

				x := locations[ii*3+0]
				y := locations[ii*3+1]
				z := locations[ii*3+2]

                // scale the height (z) of the layout to fit in the range 0-1
                // and scale x and y accordingly
                z_scale := max_coord_z - min_coord_z
                side_scale := 1.7 // smaller numbers compress things horizontally
                xp := x / z_scale / side_scale
                yp := y / z_scale / side_scale
                zp := (z-min_coord_z) / z_scale

                // bend space so that things seem to accelerate upwards
                zp = math.Pow(zp + 0.05, 0.7)

                // apply various wiggles to coordinate space
                // offset, period, min, max
                zp1 := (  zp + colorutils.Cos2(xp, t*0.33 + 8.63, 0.15 * 1.7, 0, 1) * 0.2 +
                              colorutils.Cos2(xp, -t*0.23 + 2.43, 0.34 * 1.7, 0, 1) * 0.3  )

                //zp2 := (  zp + colorutils.Cos2(xp, -t*0.38 + 1.23, 0.23 * 1.7, 0, 1) * 0.2 +
                //              colorutils.Cos2(xp, t*0.23 + 2.53, 0.63 * 1.7, 0, 1) * 0.3  )
                zp3 := (  zp + colorutils.Cos2(xp, -t*0.42 + 5.62, 0.27 * 1.7, 0, 1) * 0.2 +
                              colorutils.Cos2(xp, t*0.20 + 3.07, 0.55 * 1.7, 0, 1) * 0.3  )
                zp4 := (  zp + colorutils.Cos2(xp, t*0.36 + 4.81, 0.20 * 1.7, 0, 1) * 0.2 +
                              colorutils.Cos2(xp, -t*0.26 + 7.94, 0.67 * 1.7, 0, 1) * 0.3  )

                // make basic vertical gradient
                vgrad := colorutils.Cos2(colorutils.Clamp(zp, 0, 1), 0, 2, 0, 1)
                // vgrad := 1 - colorutils.Clamp(zp, 0, 1)

                // smallest fastest noise
                noise_lit := (   colorutils.Cos2(xp,  -4.37 * t/4, 0.21, 0, 1) +
                                 colorutils.Cos2(yp,   4.37 * t/4, 0.21, 0, 1) +
                                 colorutils.Cos2(zp1,  4.37 * t,   0.21, 0, 1)  ) / 3

                // small fast noise
                noise_med := (   colorutils.Cos2(xp,  -3 * t/4, 0.3, 0, 1) +
                                 colorutils.Cos2(yp,   3 * t/4, 0.3, 0, 1) +
                                 colorutils.Cos2(zp3,  3 * t,   0.3, 0, 1)  ) / 3

                // big slow noise
                noise_big := (   colorutils.Cos2(xp,  -0.9 * t/2, 0.8, 0, 1) +
                                 colorutils.Cos2(yp,   0.9 * t/2, 0.8, 0, 1) +
                                 colorutils.Cos2(zp4,  0.9 * t,   0.8, 0, 1)  ) / 3

                // combine vgradient with noise
                v := (   vgrad  +
                         colorutils.Remap(noise_lit, 0,1, -1,1)*0.17 +
                         colorutils.Remap(noise_med, 0,1, -1,1)*0.2 +
                         colorutils.Remap(noise_big, 0,1, -1,1)*0.8   )

                // apply sine contrast curve
                //v = colorutils.Cos2( colorutils.Clamp(v,0,1), 0, 2, 1, 0 )

                // color map
                r = v * 1.5
                g = v * 0.65
                b = v * 0.34

                r,g,b = colorutils.RGBContrast(r,g,b, 0.7, 1.2)

                //r,g,b = colorutils.RGBClipBlackByLuminance(r,g,b, 0.2)

				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
