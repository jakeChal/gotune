package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jakeChal/gotune/pkg/audio"
	"github.com/jakeChal/gotune/pkg/config"
	"github.com/jakeChal/gotune/pkg/dsp"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	sampleRate        = 48000
	bufferSize        = 4096
	silenceThreshold  = 0.001
	minDisplayClarity = 0.3
)

var (
	inTuneStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green
	closeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
	sharpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // Red (too high)
	flatStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Blue (too low)
	noteStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
)

// current state
type model struct {
	frequency float64
	noteName  string
	cents     float64
	// clarity   float64
	hasPitch  bool
	profile   config.InstrumentProfile
	audioChan chan pitchMsg
}

// event carrying new data (to be copied into model)
type pitchMsg struct {
	frequency float64
	noteName  string
	cents     float64
	hasPitch  bool
}

func listenForAudio(audioChan chan pitchMsg) tea.Cmd {
	return func() tea.Msg {
		return <-audioChan // Block until data arrives
	}
}

func initialModel(profile config.InstrumentProfile, audioChan chan pitchMsg) model {
	return model{
		profile:   profile,
		audioChan: audioChan,
	}
}

func (m model) Init() tea.Cmd {
	return listenForAudio(m.audioChan)
}

func (m model) View() string {
	if !m.hasPitch {
		return "Listening ...\n"
	}
	return formatTuningDisplay(m.frequency, m.noteName, m.cents) + "\n"
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case pitchMsg:
		m.frequency = msg.frequency
		m.noteName = msg.noteName
		m.cents = msg.cents
		m.hasPitch = msg.hasPitch
		return m, listenForAudio(m.audioChan)
	}

	return m, nil
}

func processAudio(ai *audio.AudioInput, audioChan chan pitchMsg, profile config.InstrumentProfile) {
	threshold := 0.1
	bufferAccum := make([]float64, 0, bufferSize)

	for frame := range ai.Frames {
		frame64 := dsp.Float32ToFloat64(frame)
		bufferAccum = append(bufferAccum, frame64...)

		if len(bufferAccum) >= bufferSize {
			energy := dsp.CalculateRMS(bufferAccum[:bufferSize])
			if energy < silenceThreshold {
				// fmt.Printf("\rListening...%-60s", "")
				// Keep overflow samples for better continuity
				bufferAccum = bufferAccum[bufferSize:]
				continue
			}

			result := dsp.DetectPitch(bufferAccum[:bufferSize], sampleRate, threshold)

			// display
			if result.HasPitch &&
				result.Clarity >= minDisplayClarity &&
				result.Frequency >= profile.MinFreq &&
				result.Frequency <= profile.MaxFreq {
				noteName, _, centsOff := dsp.PitchToNote(result.Frequency)
				// fmt.Print(formatTuningDisplay(result.Frequency, noteName, centsOff))
				newData := pitchMsg{
					frequency: result.Frequency,
					noteName:  noteName,
					cents:     centsOff,
					hasPitch:  true,
				}
				audioChan <- newData
			}
			// Keep overflow samples for better continuity
			bufferAccum = bufferAccum[bufferSize:]

		}
	}
}

func formatTuningDisplay(freq float64, noteName string, cents float64) string {
	// Visual meter: [<<<|>>>] where | is perfect
	meter := makeMeter(cents)
	status := tuningStatus(cents)
	styledNote := noteStyle.Render(noteName)
	return fmt.Sprintf("%-4s %s %s | %7.2f Hz", styledNote, meter, status, freq)
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
		return inTuneStyle.Render("✓ IN TUNE")
	} else if absCents <= 5 {
		return closeStyle.Render("~ close")
	} else if absCents <= 10 {
		return closeStyle.Render("  adjust")
	} else if cents < 0 {
		return flatStyle.Render("  too low")
	} else {
		return sharpStyle.Render(" too high")
	}
}

func main() {
	// threshold := flag.Float64("t", 0.1, "The MPM algorithm's detection threshold [0 - 1.0]. Low values increase the sensitivity.")
	instrumentName := flag.String("i", "guitar", "Instrument profile (guitar, bouzouki).")
	flag.Parse()

	profile, ok := config.Profiles[*instrumentName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown instrument: %s\n", *instrumentName)
		fmt.Fprintf(os.Stderr, "Supported instruments: %v\n", config.ListInstruments())
		os.Exit(1)
	}

	audioChan := make(chan pitchMsg)
	ai, err := audio.NewAudioInput(sampleRate)
	if err != nil {
		panic(err)
	}

	if err := ai.Start(); err != nil {
		panic(err)
	}
	defer ai.Stop()
	go processAudio(ai, audioChan, profile)

	model := initialModel(profile, audioChan)

	p := tea.NewProgram(model)

	// fmt.Printf("Tuning for: %s (%.0f-%.0f Hz)\n",
	// 	profile.Name, profile.MinFreq, profile.MaxFreq)

	// if *threshold < 0 || *threshold > 1.0 {
	// 	fmt.Fprintf(os.Stderr, "threshold: %v is out of range!\n", *threshold)
	// 	os.Exit(1)
	// }

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
