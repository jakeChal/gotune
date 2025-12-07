package main

import (
	"flag"
	"fmt"
	"os"

	"gitlab.com/jacobidis/gotune/pkg/audio"
	"gitlab.com/jacobidis/gotune/pkg/config"
	"gitlab.com/jacobidis/gotune/pkg/dsp"
)

const (
	sampleRate        = 48000
	bufferSize        = 4096
	silenceThreshold  = 0.001
	minDisplayClarity = 0.3
)

func main() {
	threshold := flag.Float64("t", 0.1, "The MPM algorithm's detection threshold [0 - 1.0]. Low values increase the sensitivity.")
	instrumentName := flag.String("i", "guitar", "Instrument profile (guitar, bouzouki).")
	flag.Parse()

	profile, ok := config.GetProfile(*instrumentName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown instrument: %s\n", *instrumentName)
		fmt.Fprintf(os.Stderr, "Supported instruments: %v\n", config.ListInstruments())
		os.Exit(1)
	}

	fmt.Printf("Tuning for: %s (%.0f-%.0f Hz)\n",
		profile.Name, profile.MinFreq, profile.MaxFreq)

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

	bufferAccum := make([]float64, 0, bufferSize)

	for frame := range ai.Frames {
		frame64 := dsp.Float32ToFloat64(frame)
		bufferAccum = append(bufferAccum, frame64...)

		if len(bufferAccum) >= bufferSize {
			energy := dsp.CalculateRMS(bufferAccum[:bufferSize])
			if energy < silenceThreshold {
				fmt.Printf("\rListening...%-60s", "")
				bufferAccum = bufferAccum[:0]
				continue
			}

			result := dsp.DetectPitch(bufferAccum[:bufferSize], sampleRate, *threshold)

			// display
			if result.HasPitch &&
				result.Clarity >= minDisplayClarity &&
				result.Frequency >= profile.MinFreq &&
				result.Frequency <= profile.MaxFreq {
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
