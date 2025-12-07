package dsp

import (
	"math"
	"testing"
)

func TestPitchToNote(t *testing.T) {
	tests := map[string]struct {
		frequency float64
		note      string
		midi      int
		cents     float64
		tolerance float64
	}{
		"A110": {
			frequency: float64(110),
			note:      "A2",
			midi:      45,
			cents:     0,
			tolerance: 0.01,
		},
		"A440": {
			frequency: float64(440),
			note:      "A4",
			midi:      69,
			cents:     0,
			tolerance: 0.01,
		},
		"A440Flat10Cents": {
			frequency: float64(440.0),
			note:      "A4",
			midi:      69,
			cents:     -10,
			tolerance: 0.5,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			detunedFreq := test.frequency * math.Pow(2, test.cents/1200.0)
			note, midi, cents := PitchToNote(detunedFreq)
			if note != test.note {
				t.Errorf("Expected %s, got %s", test.note, note)
			}
			if midi != test.midi {
				t.Errorf("Expected %d, got %d", test.midi, midi)
			}
			if math.Abs(cents-test.cents) > test.tolerance {
				t.Errorf("Expected %s, got %s", test.note, note)
			}
		})
	}
}
