package colorutils

import "testing"
import "math"

func TestCosTable(t *testing.T) {
	var correct, approx, diff float64
	for x := -30.0; x < 30; x += 0.1387 {
		correct = math.Cos(x)
		approx = CosTable(x)
		diff = math.Abs(correct - approx)
		if diff > 0.1 {
			t.Errorf("Cos != CosTable: %v - %v = %v", correct, approx, diff)
		}
	}
	var bigOffset float64 = 1373963358.2 * 2 * 3.14159
	for x := bigOffset - 30; x < bigOffset+30; x += 0.1387 {
		correct = math.Cos(x)
		approx = CosTable(x)
		diff = math.Abs(correct - approx)
		if diff > 0.1 {
			t.Errorf("Cos != CosTable: %v - %v = %v", correct, approx, diff)
		}
	}
}

func floatToByteTestHelper(t *testing.T, x float64, result byte) {
	if tmp := FloatToByte(x); tmp != result {
		t.Errorf("FloatToByte(%f) = %d, want %d", x, tmp, result)
	}
}
func TestFloatToByte(t *testing.T) {
	// x
	floatToByteTestHelper(t, -1.0, 0)
	floatToByteTestHelper(t, 0.0, 0)
	floatToByteTestHelper(t, 0.001, 0)
	floatToByteTestHelper(t, 0.999/256, 0)
	floatToByteTestHelper(t, 1.001/256, 1)
	floatToByteTestHelper(t, 0.4999, 127)
	floatToByteTestHelper(t, 0.5001, 128)
	floatToByteTestHelper(t, 255.001/256, 255)
	floatToByteTestHelper(t, 1.0, 255)
	floatToByteTestHelper(t, 2.0, 255)
}
func BenchmarkFloatToByte(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FloatToByte(0.5)
	}
}

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
func BenchmarkRemap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Remap(0.8, 0.0, 1.0, 0.0, 1.0)
	}
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
func BenchmarkClamp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Clamp(0.5, 0, 1)
	}
}

func cosTestHelper(t *testing.T, x, offset, period, minn, maxx, result float64) {
	if tmp := Cos(x, offset, period, minn, maxx); tmp != result {
		t.Errorf("Cos(%f,%f,%f,%f,%f) = %f, want %f", x, offset, period, minn, maxx, tmp, result)
	}
}
func TestCos(t *testing.T) {
	// x, offset, period, minn, maxx
	cosTestHelper(t, 0.0, 0.0, 1.0, 0.0, 1.0, 1.0)
	cosTestHelper(t, 0.5, 0.0, 1.0, 0.0, 1.0, 0.0)
	cosTestHelper(t, 1.0, 0.0, 1.0, 0.0, 1.0, 1.0)

	cosTestHelper(t, 0.0, 0.0, 2.0, 0.0, 1.0, 1.0)
	cosTestHelper(t, 1.0, 0.0, 2.0, 0.0, 1.0, 0.0)
	cosTestHelper(t, 2.0, 0.0, 2.0, 0.0, 1.0, 1.0)

	cosTestHelper(t, 0.0, 0.5, 2.0, 0.0, 1.0, 0.0)

	cosTestHelper(t, 0.5, 0.0, 1.0, 4.0, 5.0, 4.0)
	cosTestHelper(t, 1.0, 0.0, 1.0, 4.0, 5.0, 5.0)
}
func BenchmarkCos(b *testing.B) {
	var x float64 = 0
	for i := 0; i < b.N; i++ {
		x += 0.1028
		if x > 1000 {
			x = 0
		}
		Cos(x, 0.0, 1.0, 0.0, 1.0)
	}
}
func BenchmarkCos2(b *testing.B) {
	var x float64 = 0
	for i := 0; i < b.N; i++ {
		x += 0.1028
		if x > 1000 {
			x = 0
		}
		Cos2(x, 0.0, 1.0, 0.0, 1.0)
	}
}

func contrastTestHelper(t *testing.T, x, center, mult, result float64) {
	if tmp := Contrast(x, center, mult); tmp != result {
		t.Errorf("Contrast(%f,%f,%f) = %f, want %f", x, center, mult, tmp, result)
	}
}
func TestContrast(t *testing.T) {
	// x, center, mult
	contrastTestHelper(t, 0.0, 0.5, 0.5, 0.25)
	contrastTestHelper(t, 0.0, 0.5, 2.0, -0.5)
}
func BenchmarkContrast(b *testing.B) {
	var rr, gg, bb float64
	for i := 0; i < b.N; i++ {
		rr = Contrast(0.3, 0.3, 0.3)
		gg = Contrast(0.3, 0.3, 0.3)
		bb = Contrast(0.3, 0.3, 0.3)
	}
	_ = rr
	_ = gg
	_ = bb
}
func BenchmarkRGBContrast(b *testing.B) {
	var rr, gg, bb float64
	for i := 0; i < b.N; i++ {
		rr, gg, bb = RGBContrast(0.1, 0.2, 0.3, 0.3, 0.3)
	}
	_ = rr
	_ = gg
	_ = bb
}

func clipBlackTestHelper(t *testing.T, x, threshold, result float64) {
	if tmp := ClipBlack(x, threshold); tmp != result {
		t.Errorf("ClipBlack(%f,%f) = %f, want %f", x, threshold, tmp, result)
	}
}
func TestClipBlack(t *testing.T) {
	// x, threshold
	clipBlackTestHelper(t, 0.0, 0.0, 0.0)
	clipBlackTestHelper(t, 0.1, 0.2, 0.0)
	clipBlackTestHelper(t, 0.3, 0.2, 0.3)
}
func BenchmarkClipBlack(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ClipBlack(0.4, 0.3)
	}
}

func modDistTestHelper(t *testing.T, a, b, n, result float64) {
	if tmp := ModDist(a, b, n); tmp != result {
		t.Errorf("ModDist(%f,%f,%f) = %f, want %f", a, b, n, tmp, result)
	}
}
func TestModDist(t *testing.T) {
	// a, b, n
	modDistTestHelper(t, 0.0, 0.0, 10.0, 0.0)
	modDistTestHelper(t, 1.0, 1.0, 10.0, 0.0)
	modDistTestHelper(t, 1.0, 2.0, 10.0, 1.0)
	modDistTestHelper(t, 2.0, 1.0, 10.0, 1.0)
	modDistTestHelper(t, 1.0, 9.0, 10.0, 2.0)
	modDistTestHelper(t, 9.0, 1.0, 10.0, 2.0)

	modDistTestHelper(t, -1.0, 1.0, 10.0, 2.0)

	modDistTestHelper(t, 70.0, 70.0, 10.0, 0.0)
	modDistTestHelper(t, 71.0, 71.0, 10.0, 0.0)
	modDistTestHelper(t, 71.0, 72.0, 10.0, 1.0)
	modDistTestHelper(t, 72.0, 71.0, 10.0, 1.0)
	modDistTestHelper(t, 71.0, 79.0, 10.0, 2.0)
	modDistTestHelper(t, 79.0, 71.0, 10.0, 2.0)

	modDistTestHelper(t, -71.0, -71.0, 10.0, 0.0)
	modDistTestHelper(t, -71.0, -72.0, 10.0, 1.0)
	modDistTestHelper(t, -72.0, -71.0, 10.0, 1.0)
	modDistTestHelper(t, -71.0, -79.0, 10.0, 2.0)
	modDistTestHelper(t, -79.0, -71.0, 10.0, 2.0)
}
func BenchmarkModDist(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ModDist(1.0, 2.0, 10.0)
	}
}

func modDist2TestHelper(t *testing.T, a, b, n, result float64) {
	if tmp := ModDist2(a, b, n); math.Abs(tmp-result) > 0.001 {
		t.Errorf("ModDist2(%f,%f,%f) = %f, want %f", a, b, n, tmp, result)
	}
}
func TestModDist2(t *testing.T) {
	// a, b, n
	modDist2TestHelper(t, 0.0, 0.0, 10.0, 0.0)
	modDist2TestHelper(t, 1.0, 1.0, 10.0, 0.0)
	modDist2TestHelper(t, 1.0, 2.0, 10.0, 1.0)
	modDist2TestHelper(t, 2.0, 1.0, 10.0, 1.0)
	modDist2TestHelper(t, 1.0, 9.0, 10.0, 2.0)
	modDist2TestHelper(t, 9.0, 1.0, 10.0, 2.0)

	modDist2TestHelper(t, -1.0, 1.0, 10.0, 2.0)

	modDist2TestHelper(t, 70.0, 70.0, 10.0, 0.0)
	modDist2TestHelper(t, 71.0, 71.0, 10.0, 0.0)
	modDist2TestHelper(t, 71.0, 72.0, 10.0, 1.0)
	modDist2TestHelper(t, 72.0, 71.0, 10.0, 1.0)
	modDist2TestHelper(t, 71.0, 79.0, 10.0, 2.0)
	modDist2TestHelper(t, 79.0, 71.0, 10.0, 2.0)

	modDist2TestHelper(t, -71.0, -71.0, 10.0, 0.0)
	modDist2TestHelper(t, -71.0, -72.0, 10.0, 1.0)
	modDist2TestHelper(t, -72.0, -71.0, 10.0, 1.0)
	modDist2TestHelper(t, -71.0, -79.0, 10.0, 2.0)
	modDist2TestHelper(t, -79.0, -71.0, 10.0, 2.0)
}
func BenchmarkModDist2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ModDist2(1.0, 2.0, 10.0)
	}
}

func gammaTestHelper(t *testing.T, x, gamma, result float64) {
	if tmp := Gamma(x, gamma); tmp != result {
		t.Errorf("Gamma(%f,%f) = %f, want %f", x, gamma, tmp, result)
	}
}
func TestGamma(t *testing.T) {
	// x, gamma
	gammaTestHelper(t, 0.7, 1.0, 0.7)
	gammaTestHelper(t, 1.0, 0.7, 1.0)
	gammaTestHelper(t, 1.0, 2.2, 1.0)
	gammaTestHelper(t, 2.0, 2.0, 4.0)
	gammaTestHelper(t, 4.0, 0.5, 2.0)
	gammaTestHelper(t, 0.0, 1.0, 0.0)
	gammaTestHelper(t, -1.0, 1.0, 0.0)
}
func BenchmarkGamma(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Gamma(0.7, 0.7)
	}
}
