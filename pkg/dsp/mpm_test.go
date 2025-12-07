package dsp

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

type GoldenTestCase struct {
	Name         string    `json:"name"`
	Input        []float64 `json:"input"`
	ExpectedNSDF []float64 `json:"expected_nsdf"`
	SampleRate   int       `json:"sample_rate"`
	Frequency    float64   `json:"frequency"`
}

type GoldenTestData struct {
	Version     string           `json:"version"`
	Description string           `json:"description"`
	TestCases   []GoldenTestCase `json:"test_cases"`
}

func TestNormalizedSquareDifference_WithPython(t *testing.T) {
	// Load golden test data
	data, err := os.ReadFile("../../python/testdata/nsdf_golden.json")
	if err != nil {
		t.Fatalf("Failed to load golden test data: %v", err)
	}

	var golden GoldenTestData
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("Failed to parse golden test data: %v", err)
	}

	const absTolerance = 1e-9
	const relTolerance = 1e-9

	for _, tc := range golden.TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Run the Go implementation
			result := NormalizedSquareDifference(tc.Input)

			// Validate length
			if len(result) != len(tc.ExpectedNSDF) {
				t.Errorf("Length mismatch: got %d, want %d", len(result), len(tc.ExpectedNSDF))
				return
			}

			// Validate values with tolerance
			maxDiff := 0.0
			errorCount := 0
			for i := range result {
				diff := math.Abs(result[i] - tc.ExpectedNSDF[i])
				if diff > maxDiff {
					maxDiff = diff
				}

				// Use relative tolerance for larger values, absolute for near-zero
				threshold := absTolerance
				if math.Abs(tc.ExpectedNSDF[i]) > 1e-6 {
					threshold = relTolerance * math.Abs(tc.ExpectedNSDF[i])
				}

				if diff > threshold {
					if errorCount < 5 {
						t.Errorf("Value mismatch at index %d: got %f, want %f (diff: %e, threshold: %e)",
							i, result[i], tc.ExpectedNSDF[i], diff, threshold)
					}
					errorCount++
				}
			}

			if errorCount > 5 {
				t.Errorf("Too many errors (%d total), stopping after first 5", errorCount)
			}

			t.Logf("Max difference: %e", maxDiff)
		})
	}
}

func TestPeakPicking(t *testing.T) {
	// Create synthetic NSDF with known peaks
	nsdf := make([]float64, 1000)
	nsdf[100] = 0.8
	nsdf[500] = 0.9

	peaks := PeakPicking(nsdf, 0.5)

	if len(peaks) != 2 {
		t.Errorf("Expected 2 peaks, got %d", len(peaks))
	}
}

func TestDetectPitch_A440(t *testing.T) {
	sampleRate := 48000
	freq := 440.0
	duration := 0.5
	thresh := 0.1
	signal := GenerateSineWave(freq, duration, sampleRate)

	result := DetectPitch(signal, sampleRate, thresh)

	if !result.HasPitch {
		t.Fatal("Expected pitch to be detected")
	}

	if !AlmostEqual(result.Frequency, freq, 1.0) {
		t.Errorf("Expected %.3f Hz, got %.3f", freq, result.Frequency)
	}
}
