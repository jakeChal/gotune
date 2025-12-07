package main

import (
	"flag"
	"fmt"
	"os"

	"gitlab.com/jacobidis/gotune/pkg/audio"
	"gitlab.com/jacobidis/gotune/pkg/dsp"
)

const (
	sampleRate = 48000
	bufferSize = 4096

	// Guitar specific
	minFreq = 75.0   // Slightly below E2 (82.4 Hz)
	maxFreq = 1400.0 // Slightly above E6 (1318 Hz)
)

func main() {
	threshold := flag.Float64("t", 0.1, "The MPM algorithm's detection threshold [0 - 1.0]. Low values can lead to false positives")
	flag.Parse()
	if *threshold < 0 || *threshold > 1.0 {
		fmt.Fprintf(os.Stderr, "threshold: %v is out of range!\n", *threshold)
		os.Exit(1)
	}

	ai, err := audio.NewAudioInput(sampleRate)
	if err != nil {
		panic(err)
	}

	if err := ai.Start(); err != nil {
		panic(err)
	}
	defer ai.Stop()

	fmt.Println("Listening...")

	bufferAccum := make([]float64, 0, bufferSize)

	for frame := range ai.Frames {
		frame64 := dsp.Float32ToFloat64(frame)
		bufferAccum = append(bufferAccum, frame64...)

		if len(bufferAccum) >= bufferSize {
			result := dsp.DetectPitch(bufferAccum[:bufferSize], sampleRate, *threshold)

			// display
			if result.HasPitch && result.Frequency >= minFreq && result.Frequency <= maxFreq {
				noteName, _, centsOff := dsp.PitchToNote(result.Frequency)
				fmt.Printf(
					"\rFrequency detected: %9.2f | %-4s %4d cents [confidence: %5.2f]",
					result.Frequency, noteName, int(centsOff), result.Clarity,
				)
			}
			// reset
			bufferAccum = bufferAccum[:0]
		}
	}
}
