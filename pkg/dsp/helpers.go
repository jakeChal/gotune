package dsp

// this helper is for the audio callback in malgo, which needs float64
func Float32ToFloat64(in []float32) []float64 {
	out := make([]float64, len(in))
	for i, v := range in {
		out[i] = float64(v)
	}

	return out
}
