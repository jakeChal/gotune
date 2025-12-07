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
func NormalizedSquareDifference(buffer []float64) []float64 {
	n := len(buffer)

	// Calculate autocorrelation using FFT (efficient O(N log N))
	// Pad to next power of 2 for FFT efficiency
	fftSize := int(math.Pow(2, math.Ceil(math.Log2(float64(2*n)))))

	// Compute autocorrelation via FFT
	autocorr := autocorrelationFFT(buffer, fftSize)

	// Calculate m(tau) = sum of squared samples for each lag
	// m(tau) = sum(x[j]^2) + sum(x[j+tau]^2) for j in valid range
	cumSumSq := make([]float64, n+1)
	cumSumSq[0] = 0
	for i := 0; i < n; i++ {
		cumSumSq[i+1] = cumSumSq[i] + buffer[i]*buffer[i]
	}

	m := make([]float64, n)
	for tau := 0; tau < n; tau++ {
		// m(tau) = sum(x[0:n-tau]^2) + sum(x[tau:n]^2)
		m[tau] = cumSumSq[n-tau] + (cumSumSq[n] - cumSumSq[tau])
	}

	// Avoid division by zero
	for i := range m {
		if m[i] == 0 {
			m[i] = 1e-10
		}
	}

	// Calculate NSDF: nsdf(tau) = 2 * r(tau) / m(tau)
	nsdf := make([]float64, n)
	for i := 0; i < n; i++ {
		nsdf[i] = 2 * autocorr[i] / m[i]
	}

	return nsdf
}

// Peak represents a detected peak in the NSDF.
type Peak struct {
	Index int     // Lag index (tau)
	Value float64 // NSDF value at this lag
}

// PeakPicking finds positive peaks in the NSDF that exceed the threshold.
// A peak is defined as a point where:
// 1. The value is positive
// 2. The value exceeds the threshold
// 3. The value is greater than its neighbors (local maximum)
func PeakPicking(nsdf []float64, threshold float64) []Peak {
	var peaks []Peak

	// Start from lag=1 to avoid the trivial peak at lag=0
	for i := 1; i < len(nsdf)-1; i++ {
		// Check if it's a positive peak above threshold
		if nsdf[i] > threshold && nsdf[i] > 0 {
			// Check if it's a local maximum
			if nsdf[i] > nsdf[i-1] && nsdf[i] > nsdf[i+1] {
				peaks = append(peaks, Peak{Index: i, Value: nsdf[i]})
			}
		}
	}

	return peaks
}

func autocorrelationFFT(buffer []float64, fftSize int) []float64 {
	n := len(buffer)

	// Create padded buffer
	padded := make([]float64, fftSize)
	copy(padded, buffer)

	// Compute FFT
	fft := fourier.NewFFT(fftSize)
	fftCoeffs := fft.Coefficients(nil, padded)

	// Multiply by conjugate (power spectrum)
	for i := range fftCoeffs {
		fftCoeffs[i] = fftCoeffs[i] * cmplx.Conj(fftCoeffs[i])
	}

	// Inverse FFT to get autocorrelation
	autocorrFull := make([]float64, fftSize)
	fft.Sequence(autocorrFull, fftCoeffs)

	// Normalize by FFT size (gonum doesn't normalize IFFT)
	for i := range autocorrFull {
		autocorrFull[i] /= float64(fftSize)
	}

	// Return only the first n values
	autocorr := make([]float64, n)
	copy(autocorr, autocorrFull[:n])

	return autocorr
}
