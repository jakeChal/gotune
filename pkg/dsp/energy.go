package dsp

import "math"

// CalculateRMS calculates the Root Mean Square (energy) of a signal.
// Returns a value between 0.0 (silence) and ~1.0 (full scale).
func CalculateRMS(buffer []float64) float64 {
	if len(buffer) == 0 {
		return 0.0
	}

	var sum float64
	for _, sample := range buffer {
		sum += sample * sample
	}

	return math.Sqrt(sum / float64(len(buffer)))
}
