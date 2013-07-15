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
}
