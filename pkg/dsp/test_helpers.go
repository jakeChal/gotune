package dsp

import "math"

func GenerateSineWave(freq float64, duration float64, sampleRate int) []float64 {
	numSamples := int(duration * float64(sampleRate))
	signal := make([]float64, numSamples)

	for i := 0; i < numSamples; i++ {
		phase := 2 * math.Pi * freq * (float64(i) / float64(sampleRate))
		signal[i] = math.Sin(phase)
	}
	return signal
}

func AlmostEqual(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}
