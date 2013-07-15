package colorutils

import "testing"

func remapTestHelper(t *testing.T, x, oldmin, oldmax, newmin, newmax, result float64) {
	if tmp := Remap(x, oldmin, oldmax, newmin, newmax); tmp != result {
		t.Errorf("Remap(%f,%f,%f,%f,%f) = %f, want %f", x, oldmin, oldmax, newmin, newmax, tmp, result)
	}
}
func TestRemap(t *testing.T) {
	// x, oldmin, oldmax, newmin, newmax, result
	remapTestHelper(t, -1.0, 0.0, 1.0, 0.0, 1.0, -1.0)
	remapTestHelper(t, 0.0, 0.0, 1.0, 0.0, 1.0, 0.0)
	remapTestHelper(t, 0.8, 0.0, 1.0, 0.0, 1.0, 0.8)
	remapTestHelper(t, 1.0, 0.0, 1.0, 0.0, 1.0, 1.0)
	remapTestHelper(t, 2.0, 0.0, 1.0, 0.0, 1.0, 2.0)

	remapTestHelper(t, -1.0, 0.0, 1.0, 10.0, 20.0, 0.0)
	remapTestHelper(t, 0.0, 0.0, 1.0, 10.0, 20.0, 10.0)
	remapTestHelper(t, 0.8, 0.0, 1.0, 10.0, 20.0, 18.0)
	remapTestHelper(t, 1.0, 0.0, 1.0, 10.0, 20.0, 20.0)
	remapTestHelper(t, 2.0, 0.0, 1.0, 10.0, 20.0, 30.0)

	remapTestHelper(t, 0.0, 10.0, 20.0, 0.0, 1.0, -1.0)
	remapTestHelper(t, 10.0, 10.0, 20.0, 0.0, 1.0, 0.0)
	remapTestHelper(t, 18.0, 10.0, 20.0, 0.0, 1.0, 0.8)
	remapTestHelper(t, 20.0, 10.0, 20.0, 0.0, 1.0, 1.0)
	remapTestHelper(t, 30.0, 10.0, 20.0, 0.0, 1.0, 2.0)

	// degenerate input range
	remapTestHelper(t, 11.0, 11.0, 11.0, 10.0, 20.0, 15.0)
	remapTestHelper(t, 19.0, 11.0, 11.0, 10.0, 20.0, 15.0)

	// degenerate output range
	remapTestHelper(t, 20.0, 10.0, 20.0, 1.0, 1.0, 1.0)
}

func clampTestHelper(t *testing.T, x, minn, maxx, result float64) {
	if tmp := Clamp(x, minn, maxx); tmp != result {
		t.Errorf("Clamp(%f,%f,%f) = %f, want %f", x, minn, maxx, tmp, result)
	}
}
func TestClamp(t *testing.T) {
	// x, minn, maxx
	clampTestHelper(t, -1.0, 0.0, 1.0, 0.0)
	clampTestHelper(t, 0.0, 0.0, 1.0, 0.0)
	clampTestHelper(t, 0.5, 0.0, 1.0, 0.5)
	clampTestHelper(t, 1.0, 0.0, 1.0, 1.0)
	clampTestHelper(t, 2.0, 0.0, 1.0, 1.0)
}
