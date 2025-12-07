package dsp

import (
	"math"
	"math/cmplx"

	"gonum.org/v1/gonum/dsp/fourier"
)

// Calculate the Normalized Square Difference Function (NSDF).
// The NSDF is similar to autocorrelation but normalized to account
// for varying signal energy at different lags.
// nsdf(tau) = 2 * r(tau) / [m(0) + m(tau)]
// where:
// - r(tau) is the autocorrelation at lag tau
// - m(tau) is the sum of squared samples in the shifted windows
func NormalizedSquareDifference(buffer []float64) (nsdf []float64) {
	n := len(buffer)

	// Calculate autocorrelation using FFT (efficient O(N log N))
	// Pad to next power of 2 for FFT efficiency
	fftSize := int(math.Pow(2, math.Floor(math.Ceil(math.Log2(float64(2*n))))))

	autocorr := autocorrelationFFT(buffer, fftSize)
	return autocorr
}

func autocorrelationFFT(buffer []float64, fftSize int) []float64 {
	fft := fourier.NewFFT(fftSize)
	fftCoeffs := fft.Coefficients(nil, buffer)

	for i := range fftCoeffs {
		fftCoeffs[i] = fftCoeffs[i] * cmplx.Conj(fftCoeffs[i])
	}
	autocorr := make([]float64, len(buffer))

	return fft.Sequence(autocorr, fftCoeffs)
}
