package dsp

import (
	"fmt"
	"math"
)

var noteNames = []string{"C", "C#", "D", "D#", "E", "F",
	"F#", "G", "G#", "A", "A#", "B"}

// PitchToNote converts a frequency in Hz to musical note information.
// Returns note name (e.g., "A4"), MIDI number, and cents offset from that note.
func PitchToNote(frequency float64) (noteName string, midiNumber int, cents float64) {
	if frequency <= 0 {
		return "", 0, 0
	}

	//	A4 = 440 Hz = MIDI note 69
	midiNumberOrig := 69 + 12*math.Log2(frequency/440.0)
	midiNumber = int(math.Round(midiNumberOrig))
	cents = 100 * (midiNumberOrig - float64(midiNumber))

	octave := (midiNumber / 12) - 1
	note := noteNames[midiNumber%12]
	noteName = fmt.Sprintf("%s%d", note, octave)
	return noteName, midiNumber, cents
}
