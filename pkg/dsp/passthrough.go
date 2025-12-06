package dsp

func PassThrough(frame []float32) float32 {
	// For now: return energy
	var sum float32
	for _, s := range frame {
		sum += s * s
	}
	return sum
}
