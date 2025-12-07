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

func formatTuningDisplay(freq float64, noteName string, cents float64) string {
	// Visual meter: [<<<|>>>] where | is perfect
	meter := makeMeter(cents)
	status := tuningStatus(cents)
	return fmt.Sprintf("\r%-4s %s %s | %7.2f Hz", noteName, meter, status, freq)
}

func makeMeter(cents float64) string {
	// cents: -50 to +50 range
	// Display: [<<<<|>>>>]
	if cents < -20 {
		return "[<<<<|    ]"
	}
	if cents < -10 {
		return "[ <<<|    ]"
	}
	if cents < -5 {
		return "[  <<|    ]"
	}
	if cents < -2 {
		return "[   <|    ]"
	}
	if cents <= 2 {
		return "[    |    ]" // In tune!
	}
	if cents <= 5 {
		return "[    |>   ]"
	}
	if cents <= 10 {
		return "[    |>>  ]"
	}
	if cents <= 20 {
		return "[    |>>> ]"
	}
	return "[    |>>>>]"
}

func tuningStatus(cents float64) string {
	absCents := cents
	if absCents < 0 {
		absCents = -absCents
	}

	if absCents <= 2 {
		return "✓ IN TUNE"
	} else if absCents <= 5 {
		return "~ close  "
	} else if absCents <= 10 {
		return "  adjust "
	} else if cents < 0 {
		return "  too low"
	} else {
		return "  too high"
	}
}

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
				// Keep overflow samples for better continuity
				bufferAccum = bufferAccum[bufferSize:]
				continue
			}

			result := dsp.DetectPitch(bufferAccum[:bufferSize], sampleRate, *threshold)

			// display
			if result.HasPitch &&
				result.Clarity >= minDisplayClarity &&
				result.Frequency >= profile.MinFreq &&
				result.Frequency <= profile.MaxFreq {
				noteName, _, centsOff := dsp.PitchToNote(result.Frequency)
				fmt.Print(formatTuningDisplay(result.Frequency, noteName, centsOff))
			}
			// Keep overflow samples for better continuity
			bufferAccum = bufferAccum[bufferSize:]

		}
	}
}
