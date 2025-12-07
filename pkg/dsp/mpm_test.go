package dsp

import "testing"

func TestNSDF_Sine(t *testing.T) {
	signal := GenerateSineWave(440, 1, 48000)

	nsdf := NormalizedSquareDifference(signal)

}
